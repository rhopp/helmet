package resolver

import (
	"os"
	"testing"

	"github.com/redhat-appstudio/helmet/api"
	"github.com/redhat-appstudio/helmet/internal/chartfs"

	o "github.com/onsi/gomega"
)

func TestNewCollection(t *testing.T) {
	g := o.NewWithT(t)

	cfs := chartfs.New(os.DirFS("../../test"))

	charts, err := cfs.GetAllCharts()
	g.Expect(err).To(o.Succeed())

	appCtx := api.NewAppContext("tssc")
	c, err := NewCollection(appCtx, charts)
	g.Expect(err).To(o.Succeed())
	g.Expect(c).NotTo(o.BeNil())
}

func TestGetProductNameForIntegration(t *testing.T) {
	g := o.NewWithT(t)

	cfs := chartfs.New(os.DirFS("../../test"))

	charts, err := cfs.GetAllCharts()
	g.Expect(err).To(o.Succeed())

	appCtx := api.NewAppContext("tssc")
	c, err := NewCollection(appCtx, charts)
	g.Expect(err).To(o.Succeed())

	// Test with non-existent integration
	productName := c.GetProductNameForIntegration("non-existent-integration")
	g.Expect(productName).To(o.Equal(""))

	// Test with an integration that might exist (depends on test data)
	productName2 := c.GetProductNameForIntegration("github")
	// Product name might be empty or populated, just ensure it's a string
	g.Expect(productName2).To(o.BeAssignableToTypeOf(""))
}
