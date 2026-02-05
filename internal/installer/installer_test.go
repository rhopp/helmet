package installer

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/redhat-appstudio/helmet/internal/chartfs"
	"github.com/redhat-appstudio/helmet/internal/config"
	"github.com/redhat-appstudio/helmet/internal/flags"
	"github.com/redhat-appstudio/helmet/internal/k8s"
	"github.com/redhat-appstudio/helmet/internal/resolver"

	o "github.com/onsi/gomega"
	"helm.sh/helm/v3/pkg/chartutil"
)

// TestNewInstaller tests the constructor
func TestNewInstaller(t *testing.T) {
	g := o.NewWithT(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	f := &flags.Flags{}
	kube := k8s.NewKube(f)
	cfs := chartfs.New(os.DirFS("../../test"))
	chart, err := cfs.GetChartFiles("charts/testing")
	g.Expect(err).To(o.Succeed())

	dep := resolver.NewDependencyWithNamespace(chart, "test-namespace")
	tarball := []byte("test-tarball")

	installer := NewInstaller(logger, f, kube, dep, tarball)

	g.Expect(installer).ToNot(o.BeNil())
	g.Expect(installer.flags).To(o.Equal(f))
	g.Expect(installer.kube).To(o.Equal(kube))
	g.Expect(installer.dep).To(o.Equal(dep))
	g.Expect(installer.installerTarball).To(o.Equal(tarball))
}

// TestInstaller_SetValues tests the SetValues method
func TestInstaller_SetValues(t *testing.T) {
	g := o.NewWithT(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	f := &flags.Flags{}
	kube := k8s.NewKube(&flags.Flags{})
	cfs := chartfs.New(os.DirFS("../../test"))
	chart, err := cfs.GetChartFiles("charts/testing")
	g.Expect(err).To(o.Succeed())

	dep := resolver.NewDependencyWithNamespace(chart, "test-namespace")
	installer := NewInstaller(logger, f, kube, dep, nil)

	// Create a simple config for testing - use NewConfigFromFile with test data
	cfsForConfig := chartfs.New(os.DirFS("../../test"))
	cfg, err := config.NewConfigFromFile(cfsForConfig, "config.yaml", "test-namespace")
	g.Expect(err).To(o.Succeed())

	t.Run("set_values_with_simple_template", func(t *testing.T) {
		valuesTmpl := `
namespace: {{ .Installer.Namespace }}
test: value
`
		err := installer.SetValues(context.Background(), cfg, valuesTmpl)
		g.Expect(err).To(o.Succeed())
		g.Expect(installer.valuesBytes).ToNot(o.BeNil())
		g.Expect(string(installer.valuesBytes)).To(o.ContainSubstring("test-namespace"))
	})

	t.Run("set_values_with_empty_template", func(t *testing.T) {
		valuesTmpl := ""
		err := installer.SetValues(context.Background(), cfg, valuesTmpl)
		g.Expect(err).To(o.Succeed())
	})
}

// TestInstaller_PrintRawValues tests the PrintRawValues method
func TestInstaller_PrintRawValues(t *testing.T) {
	g := o.NewWithT(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	f := &flags.Flags{}
	kube := k8s.NewKube(&flags.Flags{})
	cfs := chartfs.New(os.DirFS("../../test"))
	chart, err := cfs.GetChartFiles("charts/testing")
	g.Expect(err).To(o.Succeed())

	dep := resolver.NewDependencyWithNamespace(chart, "test-namespace")
	installer := NewInstaller(logger, f, kube, dep, nil)

	t.Run("print_raw_values_with_values_set", func(t *testing.T) {
		installer.valuesBytes = []byte("test: value")
		// Should not panic
		g.Expect(func() { installer.PrintRawValues() }).ToNot(o.Panic())
	})

	t.Run("print_raw_values_without_values_set", func(t *testing.T) {
		installer.valuesBytes = nil
		// Should not panic even when nil
		g.Expect(func() { installer.PrintRawValues() }).ToNot(o.Panic())
	})
}

// TestInstaller_RenderValues tests the RenderValues method
func TestInstaller_RenderValues(t *testing.T) {
	g := o.NewWithT(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	f := &flags.Flags{}
	kube := k8s.NewKube(&flags.Flags{})
	cfs := chartfs.New(os.DirFS("../../test"))
	chart, err := cfs.GetChartFiles("charts/testing")
	g.Expect(err).To(o.Succeed())

	dep := resolver.NewDependencyWithNamespace(chart, "test-namespace")
	installer := NewInstaller(logger, f, kube, dep, nil)

	t.Run("render_values_without_values_set", func(t *testing.T) {
		installer.valuesBytes = nil
		err := installer.RenderValues()
		g.Expect(err).To(o.HaveOccurred())
		g.Expect(err.Error()).To(o.ContainSubstring("values not set"))
	})

	t.Run("render_values_with_valid_yaml", func(t *testing.T) {
		installer.valuesBytes = []byte(`
test: value
nested:
  key: value
`)
		err := installer.RenderValues()
		g.Expect(err).To(o.Succeed())
		g.Expect(installer.values).ToNot(o.BeNil())
		g.Expect(installer.values["test"]).To(o.Equal("value"))
	})

	t.Run("render_values_with_invalid_yaml", func(t *testing.T) {
		installer.valuesBytes = []byte(`
test: value
  invalid: yaml
`)
		err := installer.RenderValues()
		g.Expect(err).To(o.HaveOccurred())
	})
}

// TestInstaller_PrintValues tests the PrintValues method
func TestInstaller_PrintValues(t *testing.T) {
	g := o.NewWithT(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	f := &flags.Flags{}
	kube := k8s.NewKube(&flags.Flags{})
	cfs := chartfs.New(os.DirFS("../../test"))
	chart, err := cfs.GetChartFiles("charts/testing")
	g.Expect(err).To(o.Succeed())

	dep := resolver.NewDependencyWithNamespace(chart, "test-namespace")
	installer := NewInstaller(logger, f, kube, dep, nil)

	t.Run("print_values_with_values_set", func(t *testing.T) {
		installer.values = chartutil.Values{
			"test": "value",
		}
		// Should not panic
		g.Expect(func() { installer.PrintValues() }).ToNot(o.Panic())
	})

	t.Run("print_values_without_values_set", func(t *testing.T) {
		installer.values = nil
		// Should not panic even when nil
		g.Expect(func() { installer.PrintValues() }).ToNot(o.Panic())
	})
}

// TestInstaller_Install tests the Install method with error cases
func TestInstaller_Install(t *testing.T) {
	g := o.NewWithT(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	f := &flags.Flags{
		DryRun: true, // Use dry-run to avoid actual K8s interactions
	}
	kube := k8s.NewKube(&flags.Flags{})
	cfs := chartfs.New(os.DirFS("../../test"))
	chart, err := cfs.GetChartFiles("charts/testing")
	g.Expect(err).To(o.Succeed())

	dep := resolver.NewDependencyWithNamespace(chart, "test-namespace")
	installer := NewInstaller(logger, f, kube, dep, nil)

	t.Run("install_without_values_set", func(t *testing.T) {
		installer.values = nil
		err := installer.Install(context.Background())
		g.Expect(err).To(o.HaveOccurred())
		g.Expect(err.Error()).To(o.ContainSubstring("values not set"))
	})

	// Note: Testing successful Install would require extensive mocking of
	// deployer.Helm, hooks, and monitor components. This is covered by
	// integration tests. Here we focus on error paths and validation.
}
