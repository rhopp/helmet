package resolver

import (
	"os"
	"testing"

	"github.com/redhat-appstudio/helmet/internal/chartfs"

	o "github.com/onsi/gomega"
)

func TestNewTopology(t *testing.T) {
	g := o.NewWithT(t)

	cfs := chartfs.New(os.DirFS("../../test"))
	g.Expect(cfs).ToNot(o.BeNil())

	ns := "default"
	foundationChart, err := cfs.GetChartFiles("charts/helmet-foundation")
	g.Expect(err).To(o.Succeed())
	foundationDep := NewDependencyWithNamespace(foundationChart, ns)

	operatorsChart, err := cfs.GetChartFiles("charts/helmet-operators")
	g.Expect(err).To(o.Succeed())
	operatorsDep := NewDependencyWithNamespace(operatorsChart, ns)

	infrastructureChart, err := cfs.GetChartFiles("charts/helmet-infrastructure")
	g.Expect(err).To(o.Succeed())
	infrastructureDep := NewDependencyWithNamespace(infrastructureChart, ns)

	networkingChart, err := cfs.GetChartFiles("charts/helmet-networking")
	g.Expect(err).To(o.Succeed())
	networkingDep := NewDependencyWithNamespace(networkingChart, ns)

	topology := NewTopology()

	t.Run("Append", func(t *testing.T) {
		topology.Append(*networkingDep)
	})

	t.Run("PrependBefore", func(t *testing.T) {
		topology.PrependBefore(
			networkingDep.Name(),
			*foundationDep,
			*infrastructureDep,
		)
	})

	t.Run("AppendAfter", func(t *testing.T) {
		topology.AppendAfter(foundationChart.Name(), *operatorsDep)
	})

	t.Run("GetDependencies", func(t *testing.T) {
		deps := topology.Dependencies()
		g.Expect(deps).ToNot(o.BeNil())
		names := make([]string, 0, len(deps))
		for _, d := range deps {
			names = append(names, d.Name())
		}
		g.Expect(names).To(o.Equal([]string{
			"helmet-foundation",
			"helmet-operators",
			"helmet-infrastructure",
			"helmet-networking",
		}))
	})

	t.Run("GetDependency", func(t *testing.T) {
		// Test finding an existing dependency
		dep, err := topology.GetDependency("helmet-foundation")
		g.Expect(err).To(o.Succeed())
		g.Expect(dep).ToNot(o.BeNil())
		g.Expect(dep.Name()).To(o.Equal("helmet-foundation"))

		// Test finding a non-existent dependency
		dep2, err2 := topology.GetDependency("non-existent")
		g.Expect(err2).To(o.HaveOccurred())
		g.Expect(dep2).To(o.BeNil())
	})

	t.Run("Walk", func(t *testing.T) {
		// Test walking through all dependencies
		count := 0
		err := topology.Walk(func(name string, d Dependency) error {
			count++
			g.Expect(name).ToNot(o.BeEmpty())
			g.Expect(d.Name()).To(o.Equal(name))
			return nil
		})
		g.Expect(err).To(o.Succeed())
		g.Expect(count).To(o.Equal(4))
	})
}
