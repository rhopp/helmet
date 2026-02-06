package resolver

import (
	"log/slog"
	"os"
	"testing"

	"github.com/redhat-appstudio/helmet/internal/chartfs"

	o "github.com/onsi/gomega"
)

func TestNewDependency(t *testing.T) {
	g := o.NewWithT(t)

	cfs := chartfs.New(os.DirFS("../../test"))

	productA, err := cfs.GetChartFiles("charts/helmet-product-a")
	g.Expect(err).To(o.Succeed())

	d := NewDependency(productA)

	t.Run("Chart", func(t *testing.T) {
		g.Expect(d.Chart()).NotTo(o.BeNil())
	})

	t.Run("Name", func(t *testing.T) {
		g.Expect(d.Name()).To(o.Equal("helmet-product-a"))
	})

	t.Run("Namespace", func(t *testing.T) {
		g.Expect(d.Namespace()).To(o.Equal(""))
	})

	t.Run("DependsOn", func(t *testing.T) {
		dependsOn := d.DependsOn()
		g.Expect(len(dependsOn)).To(o.BeNumerically(">", 1))
		g.Expect(dependsOn[0]).To(o.Equal("helmet-foundation"))
	})

	t.Run("ProductName", func(t *testing.T) {
		g.Expect(d.ProductName()).To(o.Equal("Product A"))
	})

	t.Run("UseProductNamespace", func(t *testing.T) {
		g.Expect(d.UseProductNamespace()).To(o.BeEmpty())
	})

	t.Run("IntegrationsProvided", func(t *testing.T) {
		// IntegrationsProvided should return a slice (empty or populated)
		provided := d.IntegrationsProvided()
		g.Expect(provided).ToNot(o.BeNil())
	})

	t.Run("IntegrationsRequired", func(t *testing.T) {
		// IntegrationsRequired should return a string (empty or populated)
		required := d.IntegrationsRequired()
		g.Expect(required).To(o.BeAssignableToTypeOf(""))
	})

	t.Run("LoggerWith", func(t *testing.T) {
		g := o.NewWithT(t)
		// Create a basic logger
		baseLogger := slog.New(slog.NewTextHandler(os.Stderr, nil))
		decoratedLogger := d.LoggerWith(baseLogger)
		g.Expect(decoratedLogger).ToNot(o.BeNil())
		g.Expect(decoratedLogger).ToNot(o.Equal(baseLogger))
	})
}
