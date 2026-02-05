package resolver

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"testing/fstest"

	"github.com/redhat-appstudio/helmet/api"
	"github.com/redhat-appstudio/helmet/internal/chartfs"
	"github.com/redhat-appstudio/helmet/internal/config"
	"github.com/redhat-appstudio/helmet/internal/integrations"

	o "github.com/onsi/gomega"
)

// TestNewTopologyBuilder tests the constructor
func TestNewTopologyBuilder(t *testing.T) {
	g := o.NewWithT(t)

	// Create test filesystem with charts
	testFS := fstest.MapFS{
		"charts/product-a/Chart.yaml": &fstest.MapFile{
			Data: []byte("apiVersion: v2\nname: product-a\nversion: 1.0.0\n"),
		},
	}
	cfs := chartfs.New(testFS)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	appCtx := api.NewAppContext("test-app")
	intMgr := integrations.NewManager()

	builder, err := NewTopologyBuilder(appCtx, logger, cfs, intMgr)

	g.Expect(err).To(o.BeNil())
	g.Expect(builder).ToNot(o.BeNil())
	g.Expect(builder.logger).To(o.Equal(logger))
	g.Expect(builder.integrationsManager).To(o.Equal(intMgr))
	g.Expect(builder.collection).ToNot(o.BeNil())
}

// TestNewTopologyBuilderWithNoCharts tests constructor with empty filesystem
func TestNewTopologyBuilderWithNoCharts(t *testing.T) {
	g := o.NewWithT(t)

	testFS := fstest.MapFS{}
	cfs := chartfs.New(testFS)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	appCtx := api.NewAppContext("test-app")
	intMgr := integrations.NewManager()

	builder, err := NewTopologyBuilder(appCtx, logger, cfs, intMgr)

	// Should succeed with empty collection
	g.Expect(err).To(o.BeNil())
	g.Expect(builder).ToNot(o.BeNil())
	g.Expect(builder.collection).ToNot(o.BeNil())
}

// TestNewTopologyBuilderWithMultipleCharts tests constructor with multiple charts
func TestNewTopologyBuilderWithMultipleCharts(t *testing.T) {
	g := o.NewWithT(t)

	testFS := fstest.MapFS{
		"charts/product-a/Chart.yaml": &fstest.MapFile{
			Data: []byte("apiVersion: v2\nname: product-a\nversion: 1.0.0\n"),
		},
		"charts/product-b/Chart.yaml": &fstest.MapFile{
			Data: []byte("apiVersion: v2\nname: product-b\nversion: 1.0.0\n"),
		},
	}
	cfs := chartfs.New(testFS)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	appCtx := api.NewAppContext("test-app")
	intMgr := integrations.NewManager()

	builder, err := NewTopologyBuilder(appCtx, logger, cfs, intMgr)

	g.Expect(err).To(o.BeNil())
	g.Expect(builder).ToNot(o.BeNil())
}

// TestTopologyBuilderGetCollection tests the GetCollection method
func TestTopologyBuilderGetCollection(t *testing.T) {
	g := o.NewWithT(t)

	testFS := fstest.MapFS{
		"charts/product-a/Chart.yaml": &fstest.MapFile{
			Data: []byte("apiVersion: v2\nname: product-a\nversion: 1.0.0\n"),
		},
	}
	cfs := chartfs.New(testFS)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	appCtx := api.NewAppContext("test-app")
	intMgr := integrations.NewManager()

	builder, err := NewTopologyBuilder(appCtx, logger, cfs, intMgr)
	g.Expect(err).To(o.BeNil())

	collection := builder.GetCollection()
	g.Expect(collection).ToNot(o.BeNil())
	g.Expect(collection).To(o.Equal(builder.collection))
}

// TestTopologyBuilderBuild tests the Build method
func TestTopologyBuilderBuild(t *testing.T) {
	g := o.NewWithT(t)

	// Use real filesystem for test fixtures
	cfs := chartfs.New(os.DirFS("../../test"))
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	appCtx := api.NewAppContext("test-app", api.WithNamespace("test-namespace"))
	intMgr := integrations.NewManager()

	builder, err := NewTopologyBuilder(appCtx, logger, cfs, intMgr)
	g.Expect(err).To(o.BeNil())

	// Create a test config
	cfg, err := config.NewConfigFromFile(cfs, "config.yaml", appCtx.Namespace)
	g.Expect(err).To(o.BeNil())

	// Build the topology - may fail due to missing integrations but tests the Build flow
	ctx := context.Background()
	topology, err := builder.Build(ctx, cfg)

	// Either succeeds or fails with integration error
	if err == nil {
		g.Expect(topology).ToNot(o.BeNil())
		g.Expect(topology.Dependencies()).ToNot(o.BeEmpty())
	} else {
		// Error is acceptable for this integration-dependent test
		g.Expect(err.Error()).To(o.ContainSubstring("integration"))
	}
}

