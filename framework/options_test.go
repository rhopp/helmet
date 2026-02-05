package framework

import (
	"testing"

	"github.com/redhat-appstudio/helmet/api"
	"github.com/redhat-appstudio/helmet/internal/mcptools"

	o "github.com/onsi/gomega"
)

// TestWithIntegrations tests the WithIntegrations functional option
func TestWithIntegrations(t *testing.T) {
	g := o.NewWithT(t)

	module1 := api.IntegrationModule{Name: "github"}
	module2 := api.IntegrationModule{Name: "gitlab"}

	app := &App{}
	opt := WithIntegrations(module1, module2)
	opt(app)

	g.Expect(app.integrations).To(o.HaveLen(2))
	g.Expect(app.integrations[0].Name).To(o.Equal("github"))
	g.Expect(app.integrations[1].Name).To(o.Equal("gitlab"))
}

// TestWithIntegrationsMultipleCalls tests that multiple calls append integrations
func TestWithIntegrationsMultipleCalls(t *testing.T) {
	g := o.NewWithT(t)

	module1 := api.IntegrationModule{Name: "github"}
	module2 := api.IntegrationModule{Name: "gitlab"}
	module3 := api.IntegrationModule{Name: "quay"}

	app := &App{}

	// First call
	opt1 := WithIntegrations(module1, module2)
	opt1(app)

	// Second call should append, not replace
	opt2 := WithIntegrations(module3)
	opt2(app)

	g.Expect(app.integrations).To(o.HaveLen(3))
	g.Expect(app.integrations[0].Name).To(o.Equal("github"))
	g.Expect(app.integrations[1].Name).To(o.Equal("gitlab"))
	g.Expect(app.integrations[2].Name).To(o.Equal("quay"))
}

// TestWithIntegrationsEmpty tests calling WithIntegrations with no modules
func TestWithIntegrationsEmpty(t *testing.T) {
	g := o.NewWithT(t)

	app := &App{}
	opt := WithIntegrations()
	opt(app)

	g.Expect(app.integrations).To(o.BeEmpty())
}

// TestWithMCPImage tests the WithMCPImage functional option
func TestWithMCPImage(t *testing.T) {
	g := o.NewWithT(t)

	app := &App{}
	opt := WithMCPImage("quay.io/example/mcp-server:latest")
	opt(app)

	g.Expect(app.mcpImage).To(o.Equal("quay.io/example/mcp-server:latest"))
}

// TestWithMCPImageEmpty tests setting empty MCP image
func TestWithMCPImageEmpty(t *testing.T) {
	g := o.NewWithT(t)

	app := &App{}
	opt := WithMCPImage("")
	opt(app)

	g.Expect(app.mcpImage).To(o.Equal(""))
}

// TestWithMCPImageOverride tests that later calls override earlier ones
func TestWithMCPImageOverride(t *testing.T) {
	g := o.NewWithT(t)

	app := &App{}

	opt1 := WithMCPImage("first-image:v1")
	opt1(app)

	opt2 := WithMCPImage("second-image:v2")
	opt2(app)

	g.Expect(app.mcpImage).To(o.Equal("second-image:v2"))
}

// TestWithMCPToolsBuilder tests the WithMCPToolsBuilder functional option
func TestWithMCPToolsBuilder(t *testing.T) {
	g := o.NewWithT(t)

	// MCPToolsBuilder is a function type, so create a simple test function
	builderCalled := false
	builder := func(ctx mcptools.MCPToolsContext) ([]mcptools.Interface, error) {
		builderCalled = true
		return nil, nil
	}

	app := &App{}
	opt := WithMCPToolsBuilder(builder)
	opt(app)

	g.Expect(app.mcpToolsBuilder).ToNot(o.BeNil())

	// Verify the builder function works
	_, err := app.mcpToolsBuilder(mcptools.MCPToolsContext{})
	g.Expect(err).To(o.Succeed())
	g.Expect(builderCalled).To(o.BeTrue())
}

// TestWithMCPToolsBuilderNil tests setting nil builder
func TestWithMCPToolsBuilderNil(t *testing.T) {
	g := o.NewWithT(t)

	app := &App{}
	opt := WithMCPToolsBuilder(nil)
	opt(app)

	g.Expect(app.mcpToolsBuilder).To(o.BeNil())
}

// TestWithInstallerTarball tests the WithInstallerTarball functional option
func TestWithInstallerTarball(t *testing.T) {
	g := o.NewWithT(t)

	tarballData := []byte("fake-tarball-content")

	app := &App{}
	opt := WithInstallerTarball(tarballData)
	opt(app)

	g.Expect(app.installerTarball).To(o.Equal(tarballData))
}

// TestWithInstallerTarballEmpty tests setting empty tarball
func TestWithInstallerTarballEmpty(t *testing.T) {
	g := o.NewWithT(t)

	app := &App{}
	opt := WithInstallerTarball([]byte{})
	opt(app)

	g.Expect(app.installerTarball).To(o.BeEmpty())
}

// TestWithInstallerTarballNil tests setting nil tarball
func TestWithInstallerTarballNil(t *testing.T) {
	g := o.NewWithT(t)

	app := &App{}
	opt := WithInstallerTarball(nil)
	opt(app)

	g.Expect(app.installerTarball).To(o.BeNil())
}

// TestMultipleOptions tests applying multiple functional options together
func TestMultipleOptions(t *testing.T) {
	g := o.NewWithT(t)

	module := api.IntegrationModule{Name: "github"}
	builder := func(ctx mcptools.MCPToolsContext) ([]mcptools.Interface, error) {
		return nil, nil
	}
	tarballData := []byte("tarball")
	mcpImage := "quay.io/mcp:latest"

	app := &App{}

	// Apply all options
	WithIntegrations(module)(app)
	WithMCPImage(mcpImage)(app)
	WithMCPToolsBuilder(builder)(app)
	WithInstallerTarball(tarballData)(app)

	// Verify all were applied
	g.Expect(app.integrations).To(o.HaveLen(1))
	g.Expect(app.integrations[0].Name).To(o.Equal("github"))
	g.Expect(app.mcpImage).To(o.Equal(mcpImage))
	g.Expect(app.mcpToolsBuilder).ToNot(o.BeNil())
	g.Expect(app.installerTarball).To(o.Equal(tarballData))
}
