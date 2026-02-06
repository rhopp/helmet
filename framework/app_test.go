package framework

import (
	"archive/tar"
	"bytes"
	"os"
	"testing"
	"testing/fstest"

	"github.com/redhat-appstudio/helmet/api"
	"github.com/redhat-appstudio/helmet/internal/chartfs"
	"github.com/spf13/cobra"

	o "github.com/onsi/gomega"
)

// TestAppCommand tests the Command method
func TestAppCommand(t *testing.T) {
	g := o.NewWithT(t)

	rootCmd := &cobra.Command{
		Use:   "test-app",
		Short: "Test application",
	}

	app := &App{
		rootCmd: rootCmd,
	}

	cmd := app.Command()
	g.Expect(cmd).ToNot(o.BeNil())
	g.Expect(cmd).To(o.Equal(rootCmd))
	g.Expect(cmd.Use).To(o.Equal("test-app"))
	g.Expect(cmd.Short).To(o.Equal("Test application"))
}

// TestAppCommandNil tests Command method with nil rootCmd
func TestAppCommandNil(t *testing.T) {
	g := o.NewWithT(t)

	app := &App{}

	cmd := app.Command()
	g.Expect(cmd).To(o.BeNil())
}

// TestStandardIntegrations tests that StandardIntegrations returns modules
func TestStandardIntegrations(t *testing.T) {
	g := o.NewWithT(t)

	modules := StandardIntegrations()
	g.Expect(modules).ToNot(o.BeNil())
	g.Expect(len(modules)).To(o.BeNumerically(">", 0))

	// Verify that all modules have names
	for _, module := range modules {
		g.Expect(module.Name).ToNot(o.BeEmpty())
	}
}

// TestStandardIntegrationsConsistency tests that multiple calls return the same modules
func TestStandardIntegrationsConsistency(t *testing.T) {
	g := o.NewWithT(t)

	modules1 := StandardIntegrations()
	modules2 := StandardIntegrations()

	g.Expect(len(modules1)).To(o.Equal(len(modules2)))

	// Verify names match (order should be consistent)
	for i := range modules1 {
		g.Expect(modules1[i].Name).To(o.Equal(modules2[i].Name))
	}
}

// TestNewApp tests the NewApp constructor
func TestNewApp(t *testing.T) {
	g := o.NewWithT(t)

	// Create minimal test filesystem
	testFS := fstest.MapFS{
		"config.yaml": &fstest.MapFile{
			Data: []byte("products:\n  - name: test-product\n"),
		},
	}
	cfs := chartfs.New(testFS)

	// Create test AppContext
	appCtx := api.NewAppContext(
		"test-installer",
		api.WithVersion("v1.0.0"),
		api.WithNamespace("test-namespace"),
	)

	// Test with minimal options (no integrations)
	app, err := NewApp(appCtx, cfs,
		WithMCPImage("test-image:latest"),
	)

	g.Expect(err).To(o.BeNil())
	g.Expect(app).ToNot(o.BeNil())
	g.Expect(app.AppCtx).To(o.Equal(appCtx))
	g.Expect(app.ChartFS).To(o.Equal(cfs))
	g.Expect(app.rootCmd).ToNot(o.BeNil())
	g.Expect(app.rootCmd.Use).To(o.Equal("test-installer"))
	g.Expect(app.flags).ToNot(o.BeNil())
	g.Expect(app.kube).ToNot(o.BeNil())
}

// TestNewAppWithShortDescription tests NewApp with custom short description
func TestNewAppWithShortDescription(t *testing.T) {
	g := o.NewWithT(t)

	testFS := fstest.MapFS{
		"config.yaml": &fstest.MapFile{
			Data: []byte("products:\n"),
		},
	}
	cfs := chartfs.New(testFS)

	appCtx := api.NewAppContext(
		"my-app",
		api.WithShortDescription("Custom short description"),
	)

	app, err := NewApp(appCtx, cfs,
		WithMCPImage("image:latest"),
	)

	g.Expect(err).To(o.BeNil())
	g.Expect(app.rootCmd.Short).To(o.Equal("Custom short description"))
}

// TestNewAppDefaultShortDescription tests NewApp generates default short description
func TestNewAppDefaultShortDescription(t *testing.T) {
	g := o.NewWithT(t)

	testFS := fstest.MapFS{
		"config.yaml": &fstest.MapFile{
			Data: []byte("products:\n"),
		},
	}
	cfs := chartfs.New(testFS)

	appCtx := api.NewAppContext("my-app")

	app, err := NewApp(appCtx, cfs,
		WithMCPImage("image:latest"),
	)

	g.Expect(err).To(o.BeNil())
	g.Expect(app.rootCmd.Short).To(o.Equal("my-app installer"))
}

// TestNewAppWithLongDescription tests NewApp with long description
func TestNewAppWithLongDescription(t *testing.T) {
	g := o.NewWithT(t)

	testFS := fstest.MapFS{
		"config.yaml": &fstest.MapFile{
			Data: []byte("products:\n"),
		},
	}
	cfs := chartfs.New(testFS)

	longDesc := "This is a very long description\nwith multiple lines"
	appCtx := api.NewAppContext(
		"my-app",
		api.WithLongDescription(longDesc),
	)

	app, err := NewApp(appCtx, cfs,
		WithMCPImage("image:latest"),
	)

	g.Expect(err).To(o.BeNil())
	g.Expect(app.rootCmd.Long).To(o.Equal(longDesc))
}

// TestNewAppMissingMCPImage tests that NewApp fails without MCP image
func TestNewAppMissingMCPImage(t *testing.T) {
	g := o.NewWithT(t)

	testFS := fstest.MapFS{
		"config.yaml": &fstest.MapFile{
			Data: []byte("products:\n"),
		},
	}
	cfs := chartfs.New(testFS)
	appCtx := api.NewAppContext("test-app")

	app, err := NewApp(appCtx, cfs)

	g.Expect(err).ToNot(o.BeNil())
	g.Expect(err.Error()).To(o.ContainSubstring("MCP server image not configured"))
	g.Expect(app).To(o.BeNil())
}

// TestNewAppWithIntegrations tests NewApp with custom integrations
func TestNewAppWithIntegrations(t *testing.T) {
	g := o.NewWithT(t)

	testFS := fstest.MapFS{
		"config.yaml": &fstest.MapFile{
			Data: []byte("products:\n"),
		},
	}
	cfs := chartfs.New(testFS)
	appCtx := api.NewAppContext("test-app")

	// Use standard integrations
	integrations := StandardIntegrations()

	app, err := NewApp(appCtx, cfs,
		WithMCPImage("image:latest"),
		WithIntegrations(integrations...),
	)

	g.Expect(err).To(o.BeNil())
	g.Expect(app).ToNot(o.BeNil())
	g.Expect(app.integrations).To(o.HaveLen(len(integrations)))
}

// TestNewAppFromTarball tests the NewAppFromTarball constructor
func TestNewAppFromTarball(t *testing.T) {
	g := o.NewWithT(t)

	// Create a test tarball using existing helper
	testTarball := createTarballFromMap(map[string]string{
		"config.yaml": "products:\n",
	})

	appCtx := api.NewAppContext(
		"tarball-app",
		api.WithVersion("v2.0.0"),
	)

	// Get current directory
	cwd, err := os.Getwd()
	g.Expect(err).To(o.BeNil())

	app, err := NewAppFromTarball(appCtx, testTarball, cwd,
		WithMCPImage("tarball-image:latest"),
	)

	g.Expect(err).To(o.BeNil())
	g.Expect(app).ToNot(o.BeNil())
	g.Expect(app.AppCtx).To(o.Equal(appCtx))
	g.Expect(app.ChartFS).ToNot(o.BeNil())
	g.Expect(app.installerTarball).To(o.Equal(testTarball))
}

// TestNewAppFromTarballInvalidTarball tests NewAppFromTarball with invalid tarball
func TestNewAppFromTarballInvalidTarball(t *testing.T) {
	g := o.NewWithT(t)

	appCtx := api.NewAppContext("invalid-tarball-app")
	invalidTarball := []byte("not a tarball")

	cwd, err := os.Getwd()
	g.Expect(err).To(o.BeNil())

	app, err := NewAppFromTarball(appCtx, invalidTarball, cwd,
		WithMCPImage("image:latest"),
	)

	g.Expect(err).ToNot(o.BeNil())
	g.Expect(app).To(o.BeNil())
}

// TestAppRun tests the Run method
func TestAppRun(t *testing.T) {
	g := o.NewWithT(t)

	testFS := fstest.MapFS{
		"config.yaml": &fstest.MapFile{
			Data: []byte("products:\n"),
		},
	}
	cfs := chartfs.New(testFS)
	appCtx := api.NewAppContext("run-test-app")

	app, err := NewApp(appCtx, cfs,
		WithMCPImage("image:latest"),
	)
	g.Expect(err).To(o.BeNil())

	// Set args to trigger help (no error expected)
	app.rootCmd.SetArgs([]string{"--help"})

	err = app.Run()
	// Help should not return error
	g.Expect(err).To(o.BeNil())
}

// TestAppRunWithInvalidCommand tests Run with invalid command
func TestAppRunWithInvalidCommand(t *testing.T) {
	g := o.NewWithT(t)

	testFS := fstest.MapFS{
		"config.yaml": &fstest.MapFile{
			Data: []byte("products:\n"),
		},
	}
	cfs := chartfs.New(testFS)
	appCtx := api.NewAppContext("invalid-cmd-app")

	app, err := NewApp(appCtx, cfs,
		WithMCPImage("image:latest"),
	)
	g.Expect(err).To(o.BeNil())

	// Set invalid command
	app.rootCmd.SetArgs([]string{"nonexistent-command"})

	err = app.Run()
	g.Expect(err).ToNot(o.BeNil())
}

// TestAppSubcommands tests that subcommands are properly registered
func TestAppSubcommands(t *testing.T) {
	g := o.NewWithT(t)

	testFS := fstest.MapFS{
		"config.yaml": &fstest.MapFile{
			Data: []byte("products:\n"),
		},
	}
	cfs := chartfs.New(testFS)
	appCtx := api.NewAppContext("subcmd-app")

	app, err := NewApp(appCtx, cfs,
		WithMCPImage("image:latest"),
	)
	g.Expect(err).To(o.BeNil())

	// Check that standard subcommands are registered
	subcommands := app.rootCmd.Commands()
	g.Expect(len(subcommands)).To(o.BeNumerically(">", 0))

	// Find specific subcommands by name
	cmdNames := make(map[string]bool)
	for _, cmd := range subcommands {
		cmdNames[cmd.Name()] = true
	}

	// Expected subcommands based on setupRootCmd
	expectedCmds := []string{"integration", "config", "deploy", "installer", "template", "topology"}
	for _, expected := range expectedCmds {
		g.Expect(cmdNames[expected]).To(o.BeTrue(), "Expected subcommand %s to be registered", expected)
	}
}

// Helper function to create a test tarball
func createTestTarball(t *testing.T, buf *bytes.Buffer) []byte {
	t.Helper()
	return createMinimalTarball(t)
}

// createMinimalTarball creates a minimal valid tar archive for testing
func createMinimalTarball(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	// Add a simple config.yaml file
	configContent := []byte("products:\n")
	hdr := &tar.Header{
		Name: "config.yaml",
		Mode: 0600,
		Size: int64(len(configContent)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("failed to write tar header: %v", err)
	}
	if _, err := tw.Write(configContent); err != nil {
		t.Fatalf("failed to write tar content: %v", err)
	}

	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}

	return buf.Bytes()
}
