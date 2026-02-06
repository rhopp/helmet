package mcptools

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	"github.com/redhat-appstudio/helmet/api"
	"github.com/redhat-appstudio/helmet/internal/chartfs"
	"github.com/redhat-appstudio/helmet/internal/flags"
	"github.com/redhat-appstudio/helmet/internal/integrations"
	"github.com/redhat-appstudio/helmet/internal/k8s"

	o "github.com/onsi/gomega"
)

// TestNewMCPToolsContext tests the MCPToolsContext constructor
func TestNewMCPToolsContext(t *testing.T) {
	g := o.NewWithT(t)

	// Create test dependencies
	appCtx := &api.AppContext{
		Name:    "test-app",
		Version: "1.0.0",
	}

	f := flags.NewFlags()
	cfs := &chartfs.ChartFS{}
	kube := &k8s.Kube{}
	integrationMgr := integrations.NewManager()
	image := "test-image:v1.0"

	// Create context
	ctx := NewMCPToolsContext(appCtx, f, cfs, kube, integrationMgr, image)

	// Verify all fields are set correctly
	g.Expect(ctx.AppCtx).To(o.Equal(appCtx))
	g.Expect(ctx.Flags).To(o.Equal(f))
	g.Expect(ctx.ChartFS).To(o.Equal(cfs))
	g.Expect(ctx.Kube).To(o.Equal(kube))
	g.Expect(ctx.IntegrationManager).To(o.Equal(integrationMgr))
	g.Expect(ctx.Image).To(o.Equal(image))

	// Verify logger is configured
	g.Expect(ctx.Logger).ToNot(o.BeNil())
}

// TestMCPToolsContextLoggerUsesDiscard tests that the logger writes to io.Discard
func TestMCPToolsContextLoggerUsesDiscard(t *testing.T) {
	g := o.NewWithT(t)

	// Create test dependencies
	appCtx := &api.AppContext{Name: "test-app"}
	f := flags.NewFlags()

	// Create context
	ctx := NewMCPToolsContext(appCtx, f, nil, nil, nil, "")

	// Test that logger doesn't write to stdout/stderr
	// We can't directly test io.Discard, but we can verify the logger exists
	g.Expect(ctx.Logger).ToNot(o.BeNil())

	// Try to write a log message - it should not panic
	g.Expect(func() {
		ctx.Logger.Info("test message")
		ctx.Logger.Debug("debug message")
		ctx.Logger.Error("error message")
	}).ToNot(o.Panic())
}

// TestMCPToolsContextWithNilValues tests context creation with nil values
func TestMCPToolsContextWithNilValues(t *testing.T) {
	g := o.NewWithT(t)

	appCtx := &api.AppContext{Name: "test-app"}
	f := flags.NewFlags()

	// Create context with some nil values
	ctx := NewMCPToolsContext(appCtx, f, nil, nil, nil, "")

	g.Expect(ctx.AppCtx).ToNot(o.BeNil())
	g.Expect(ctx.Flags).ToNot(o.BeNil())
	g.Expect(ctx.Logger).ToNot(o.BeNil())
	g.Expect(ctx.ChartFS).To(o.BeNil())
	g.Expect(ctx.Kube).To(o.BeNil())
	g.Expect(ctx.IntegrationManager).To(o.BeNil())
	g.Expect(ctx.Image).To(o.BeEmpty())
}

// TestMCPToolsContextFields tests direct field access
func TestMCPToolsContextFields(t *testing.T) {
	g := o.NewWithT(t)

	ctx := MCPToolsContext{
		AppCtx: &api.AppContext{
			Name:    "direct-app",
			Version: "2.0.0",
		},
		Logger:             slog.New(slog.NewTextHandler(io.Discard, nil)),
		Flags:              flags.NewFlags(),
		ChartFS:            &chartfs.ChartFS{},
		Kube:               &k8s.Kube{},
		IntegrationManager: integrations.NewManager(),
		Image:              "direct-image:v2.0",
	}

	g.Expect(ctx.AppCtx.Name).To(o.Equal("direct-app"))
	g.Expect(ctx.AppCtx.Version).To(o.Equal("2.0.0"))
	g.Expect(ctx.Logger).ToNot(o.BeNil())
	g.Expect(ctx.Flags).ToNot(o.BeNil())
	g.Expect(ctx.ChartFS).ToNot(o.BeNil())
	g.Expect(ctx.Kube).ToNot(o.BeNil())
	g.Expect(ctx.IntegrationManager).ToNot(o.BeNil())
	g.Expect(ctx.Image).To(o.Equal("direct-image:v2.0"))
}

// TestMCPToolsBuilderType tests the MCPToolsBuilder function type
func TestMCPToolsBuilderType(t *testing.T) {
	g := o.NewWithT(t)

	// Create a sample builder function
	builder := func(ctx MCPToolsContext) ([]Interface, error) {
		return []Interface{}, nil
	}

	g.Expect(builder).ToNot(o.BeNil())

	// Test calling the builder
	ctx := MCPToolsContext{
		AppCtx: &api.AppContext{Name: "test"},
		Flags:  flags.NewFlags(),
	}

	tools, err := builder(ctx)
	g.Expect(err).To(o.Succeed())
	g.Expect(tools).ToNot(o.BeNil())
	g.Expect(tools).To(o.BeEmpty())
}

// TestLoggerOutputToDiscard verifies logger output goes to discard
func TestLoggerOutputToDiscard(t *testing.T) {
	g := o.NewWithT(t)

	// Capture stdout/stderr to verify nothing is written
	oldStdout := slog.Default()
	defer slog.SetDefault(oldStdout)

	var buf bytes.Buffer
	testLogger := slog.New(slog.NewTextHandler(&buf, nil))
	slog.SetDefault(testLogger)

	// Create context
	appCtx := &api.AppContext{Name: "test"}
	f := flags.NewFlags()
	ctx := NewMCPToolsContext(appCtx, f, nil, nil, nil, "")

	// Write logs using the context logger
	ctx.Logger.Info("this should not appear in buf")

	// Verify the test logger buffer is empty (context logger uses io.Discard)
	g.Expect(buf.String()).To(o.BeEmpty())
}
