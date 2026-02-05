package chartfs

import (
	"os"
	"testing"

	o "github.com/onsi/gomega"
)

func TestNewChartFS(t *testing.T) {
	g := o.NewWithT(t)

	c := New(os.DirFS("../../test"))
	g.Expect(c).ToNot(o.BeNil())

	t.Run("ReadFile", func(t *testing.T) {
		valuesTmplBytes, err := c.ReadFile("values.yaml.tpl")
		g.Expect(err).To(o.Succeed())
		g.Expect(valuesTmplBytes).ToNot(o.BeEmpty())
	})

	t.Run("GetChartForDep", func(t *testing.T) {
		chart, err := c.GetChartFiles("charts/helmet-product-a")
		g.Expect(err).To(o.Succeed())
		g.Expect(chart).ToNot(o.BeNil())
		g.Expect(chart.Name()).To(o.Equal("helmet-product-a"))
		g.Expect(chart.Templates).ToNot(o.BeEmpty())

		// Asserting the chart templates are present, it should contain at least a
		// few files, plus the presence of the "NOTES.txt" common file.
		names := make([]string, 0, len(chart.Templates))
		for _, tmpl := range chart.Templates {
			names = append(names, tmpl.Name)
		}
		g.Expect(len(names)).To(o.BeNumerically("==", 1))
		g.Expect(names).To(o.ContainElement("templates/NOTES.txt"))
	})

	t.Run("GetAllCharts", func(t *testing.T) {
		charts, err := c.GetAllCharts()
		g.Expect(err).To(o.Succeed())
		g.Expect(charts).ToNot(o.BeNil())
		g.Expect(len(charts)).To(o.BeNumerically(">", 1))
	})

	t.Run("Open", func(t *testing.T) {
		// Test Open method which implements fs.FS interface
		file, err := c.Open("config.yaml")
		g.Expect(err).To(o.Succeed())
		g.Expect(file).ToNot(o.BeNil())
		defer file.Close()

		// Verify file info
		info, err := file.Stat()
		g.Expect(err).To(o.Succeed())
		g.Expect(info.Name()).To(o.Equal("config.yaml"))
		g.Expect(info.IsDir()).To(o.BeFalse())
	})

	t.Run("WithBaseDir", func(t *testing.T) {
		// Test WithBaseDir to create a sub-filesystem
		subFS, err := c.WithBaseDir("charts")
		g.Expect(err).To(o.Succeed())
		g.Expect(subFS).ToNot(o.BeNil())

		// Verify we can access files in the sub-filesystem
		chart, err := subFS.GetChartFiles("helmet-product-a")
		g.Expect(err).To(o.Succeed())
		g.Expect(chart).ToNot(o.BeNil())
		g.Expect(chart.Name()).To(o.Equal("helmet-product-a"))
	})

	t.Run("WithBaseDir_NestedPath", func(t *testing.T) {
		// Test WithBaseDir creates a properly scoped filesystem
		chartsFS, err := c.WithBaseDir("charts")
		g.Expect(err).To(o.Succeed())

		// Now test getting all charts from the scoped filesystem
		charts, err := chartsFS.GetAllCharts()
		g.Expect(err).To(o.Succeed())
		g.Expect(charts).ToNot(o.BeEmpty())
	})
}
