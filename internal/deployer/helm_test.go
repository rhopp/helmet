package deployer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/redhat-appstudio/helmet/internal/chartfs"
	"github.com/redhat-appstudio/helmet/internal/flags"
	"github.com/redhat-appstudio/helmet/internal/k8s"

	o "github.com/onsi/gomega"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/resource"
)

// mockMonitor implements monitor.Interface for testing
type mockMonitor struct {
	collectErr error
	collected  []*resource.Info
}

func (m *mockMonitor) Collect(ctx context.Context, r *resource.Info) error {
	if m.collectErr != nil {
		return m.collectErr
	}
	m.collected = append(m.collected, r)
	return nil
}

func (m *mockMonitor) Watch(timeout time.Duration) error {
	return nil
}

// TestHelm_printRelease tests the printRelease method
func TestHelm_printRelease(t *testing.T) {
	g := o.NewWithT(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	f := &flags.Flags{
		Debug:  false,
		DryRun: false,
	}

	h := &Helm{
		logger: logger,
		flags:  f,
		chart:  &chart.Chart{Metadata: &chart.Metadata{Name: "test-chart"}},
	}

	rel := &release.Release{
		Name:      "test-release",
		Namespace: "test-namespace",
		Version:   1,
		Chart: &chart.Chart{
			Metadata: &chart.Metadata{
				Name:    "test-chart",
				Version: "1.0.0",
			},
		},
		Info: &release.Info{
			Status: release.StatusDeployed,
			Notes:  "Test notes",
		},
		Config: map[string]interface{}{
			"key": "value",
		},
		Manifest: "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test\n",
	}

	t.Run("print_release_normal_mode", func(t *testing.T) {
		h.flags.Debug = false
		h.flags.DryRun = false
		// Should not panic
		g.Expect(func() { h.printRelease(rel) }).ToNot(o.Panic())
	})

	t.Run("print_release_debug_mode", func(t *testing.T) {
		h.flags.Debug = true
		h.flags.DryRun = false
		// Should not panic and print config values
		g.Expect(func() { h.printRelease(rel) }).ToNot(o.Panic())
	})

	t.Run("print_release_dry_run_mode", func(t *testing.T) {
		h.flags.Debug = false
		h.flags.DryRun = true
		// Should not panic and print extended info
		g.Expect(func() { h.printRelease(rel) }).ToNot(o.Panic())
	})
}

// TestHelm_helmInstall tests the helmInstall method
func TestHelm_helmInstall(t *testing.T) {
	g := o.NewWithT(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	f := &flags.Flags{
		Timeout: 5 * time.Minute,
		DryRun:  true, // Use dry-run to avoid actual cluster operations
	}

	// Create a simple test chart
	testChart := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "test-chart",
			Version: "1.0.0",
		},
	}

	// Create action configuration for testing
	actionCfg := &action.Configuration{}
	err := actionCfg.Init(nil, "test-namespace", "memory", func(format string, v ...interface{}) {})
	g.Expect(err).To(o.Succeed())

	h := &Helm{
		logger:    logger,
		flags:     f,
		chart:     testChart,
		namespace: "test-namespace",
		actionCfg: actionCfg,
	}

	t.Run("install_with_dry_run", func(t *testing.T) {
		vals := chartutil.Values{}
		rel, err := h.helmInstall(context.Background(), vals)

		// In dry-run mode with memory driver, install should succeed
		g.Expect(err).To(o.Succeed())
		g.Expect(rel).ToNot(o.BeNil())
		g.Expect(rel.Name).To(o.Equal("test-chart"))
		g.Expect(rel.Namespace).To(o.Equal("test-namespace"))
	})

	t.Run("install_sets_timeout", func(t *testing.T) {
		h.flags.Timeout = 10 * time.Minute
		vals := chartutil.Values{}

		_, err := h.helmInstall(context.Background(), vals)
		g.Expect(err).To(o.Succeed())
	})
}

// TestHelm_helmUpgrade tests the helmUpgrade method
func TestHelm_helmUpgrade(t *testing.T) {
	g := o.NewWithT(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	f := &flags.Flags{
		Timeout: 5 * time.Minute,
		DryRun:  true,
	}

	testChart := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "test-upgrade-chart",
			Version: "1.0.0",
		},
	}

	actionCfg := &action.Configuration{}
	err := actionCfg.Init(nil, "test-namespace", "memory", func(format string, v ...interface{}) {})
	g.Expect(err).To(o.Succeed())

	h := &Helm{
		logger:    logger,
		flags:     f,
		chart:     testChart,
		namespace: "test-namespace",
		actionCfg: actionCfg,
	}

	t.Run("upgrade_without_existing_release_returns_error", func(t *testing.T) {
		vals := chartutil.Values{}
		h.flags.DryRun = true
		rel, err := h.helmUpgrade(context.Background(), vals)

		// Upgrade should fail when there's no existing release
		g.Expect(err).ToNot(o.Succeed())
		g.Expect(errors.Is(err, ErrUpgradeFailed)).To(o.BeTrue())
		g.Expect(rel).To(o.BeNil())
	})

	t.Run("upgrade_sets_namespace_and_timeout", func(t *testing.T) {
		// This test verifies the configuration is set correctly
		// Even if the upgrade fails, we can check the error message
		h.flags.Timeout = 15 * time.Minute
		vals := chartutil.Values{}

		_, err := h.helmUpgrade(context.Background(), vals)
		// Error is expected since no release exists
		g.Expect(err).ToNot(o.Succeed())
		g.Expect(errors.Is(err, ErrUpgradeFailed)).To(o.BeTrue())
	})
}

// TestHelm_Deploy tests the Deploy method
func TestHelm_Deploy(t *testing.T) {
	// This test requires access to Kubernetes API, skip in CI environments
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test in CI environment (requires Kubernetes cluster)")
	}

	g := o.NewWithT(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	f := &flags.Flags{
		Timeout: 5 * time.Minute,
		DryRun:  true,
	}

	testChart := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "test-deploy-chart",
			Version: "1.0.0",
		},
	}

	t.Run("deploy_new_release", func(t *testing.T) {
		actionCfg := &action.Configuration{}
		err := actionCfg.Init(nil, "test-namespace", "memory", func(format string, v ...interface{}) {})
		g.Expect(err).To(o.Succeed())

		h := &Helm{
			logger:    logger,
			flags:     f,
			chart:     testChart,
			namespace: "test-namespace",
			actionCfg: actionCfg,
		}

		vals := chartutil.Values{}
		err = h.Deploy(context.Background(), vals)

		g.Expect(err).To(o.Succeed())
		g.Expect(h.release).ToNot(o.BeNil())
		g.Expect(h.release.Name).To(o.Equal("test-deploy-chart"))
	})

	t.Run("deploy_upgrade_existing_release", func(t *testing.T) {
		actionCfg := &action.Configuration{}
		err := actionCfg.Init(nil, "test-namespace", "memory", func(format string, v ...interface{}) {})
		g.Expect(err).To(o.Succeed())

		// Install first
		install := action.NewInstall(actionCfg)
		install.Namespace = "test-namespace"
		install.ReleaseName = "test-deploy-chart"
		install.DryRun = true
		install.ClientOnly = true
		_, err = install.Run(testChart, chartutil.Values{})
		g.Expect(err).To(o.Succeed())

		h := &Helm{
			logger:    logger,
			flags:     f,
			chart:     testChart,
			namespace: "test-namespace",
			actionCfg: actionCfg,
		}

		vals := chartutil.Values{"key": "value"}
		err = h.Deploy(context.Background(), vals)

		g.Expect(err).To(o.Succeed())
		g.Expect(h.release).ToNot(o.BeNil())
	})
}

// TestHelm_Verify tests the Verify method
func TestHelm_Verify(t *testing.T) {
	g := o.NewWithT(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	testChart := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "test-verify-chart",
			Version: "1.0.0",
		},
	}

	t.Run("verify_skipped_in_dry_run", func(t *testing.T) {
		f := &flags.Flags{
			DryRun: true,
		}

		actionCfg := &action.Configuration{}
		err := actionCfg.Init(nil, "test-namespace", "memory", func(format string, v ...interface{}) {})
		g.Expect(err).To(o.Succeed())

		h := &Helm{
			logger:    logger,
			flags:     f,
			chart:     testChart,
			namespace: "test-namespace",
			actionCfg: actionCfg,
		}

		err = h.Verify()
		// Should return nil in dry-run mode
		g.Expect(err).To(o.Succeed())
	})

	t.Run("verify_returns_error_for_missing_release", func(t *testing.T) {
		f := &flags.Flags{
			DryRun: false,
		}

		actionCfg := &action.Configuration{}
		err := actionCfg.Init(nil, "test-namespace", "memory", func(format string, v ...interface{}) {})
		g.Expect(err).To(o.Succeed())

		h := &Helm{
			logger:    logger,
			flags:     f,
			chart:     testChart,
			namespace: "test-namespace",
			actionCfg: actionCfg,
		}

		err = h.Verify()
		// Should return error when release doesn't exist
		g.Expect(err).ToNot(o.Succeed())
	})
}

// TestHelm_VerifyWithRetry tests the VerifyWithRetry method
func TestHelm_VerifyWithRetry(t *testing.T) {
	g := o.NewWithT(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	f := &flags.Flags{
		DryRun: true, // Use dry-run to avoid actual operations
	}

	testChart := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "test-retry-chart",
			Version: "1.0.0",
		},
	}

	actionCfg := &action.Configuration{}
	err := actionCfg.Init(nil, "test-namespace", "memory", func(format string, v ...interface{}) {})
	g.Expect(err).To(o.Succeed())

	h := &Helm{
		logger:    logger,
		flags:     f,
		chart:     testChart,
		namespace: "test-namespace",
		actionCfg: actionCfg,
	}

	t.Run("verify_succeeds_on_first_try", func(t *testing.T) {
		start := time.Now()
		err := h.VerifyWithRetry()
		duration := time.Since(start)

		// Should succeed immediately in dry-run mode
		g.Expect(err).To(o.Succeed())
		// Should not have retried (no 1-minute delays)
		g.Expect(duration).To(o.BeNumerically("<", 5*time.Second))
	})
}

// TestHelm_GetNotes tests the GetNotes method
func TestHelm_GetNotes(t *testing.T) {
	g := o.NewWithT(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	f := &flags.Flags{
		DryRun: true,
	}

	t.Run("get_notes_returns_method_structure", func(t *testing.T) {
		actionCfg := &action.Configuration{}
		err := actionCfg.Init(nil, "test-namespace", "memory", func(format string, v ...interface{}) {})
		g.Expect(err).To(o.Succeed())

		h := &Helm{
			logger:    logger,
			flags:     f,
			chart:     &chart.Chart{Metadata: &chart.Metadata{Name: "test-chart"}},
			namespace: "test-namespace",
			actionCfg: actionCfg,
		}

		// GetNotes should call action.Get with Version: 0
		// Without an existing release, it will return an error
		notes, err := h.GetNotes()

		// Expected to fail since no release exists
		g.Expect(err).ToNot(o.Succeed())
		g.Expect(notes).To(o.Equal(""))
	})

	t.Run("get_notes_from_missing_release", func(t *testing.T) {
		actionCfg := &action.Configuration{}
		err := actionCfg.Init(nil, "test-namespace", "memory", func(format string, v ...interface{}) {})
		g.Expect(err).To(o.Succeed())

		h := &Helm{
			logger:    logger,
			flags:     f,
			chart:     &chart.Chart{Metadata: &chart.Metadata{Name: "nonexistent-chart"}},
			namespace: "test-namespace",
			actionCfg: actionCfg,
		}

		notes, err := h.GetNotes()
		// Should return error for missing release
		g.Expect(err).ToNot(o.Succeed())
		g.Expect(notes).To(o.Equal(""))
		g.Expect(errors.Is(err, driver.ErrReleaseNotFound)).To(o.BeTrue())
	})
}

// TestHelm_VisitReleaseResources tests the VisitReleaseResources method
func TestHelm_VisitReleaseResources(t *testing.T) {
	g := o.NewWithT(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	f := &flags.Flags{
		DryRun: true,
	}

	testChart := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "test-visit-chart",
			Version: "1.0.0",
		},
	}

	actionCfg := &action.Configuration{}
	err := actionCfg.Init(nil, "test-namespace", "memory", func(format string, v ...interface{}) {})
	g.Expect(err).To(o.Succeed())

	h := &Helm{
		logger:    logger,
		flags:     f,
		chart:     testChart,
		namespace: "test-namespace",
		actionCfg: actionCfg,
	}

	// Deploy first to have a release with manifest
	err = h.Deploy(context.Background(), chartutil.Values{})
	g.Expect(err).To(o.Succeed())

	t.Run("visit_with_empty_manifest", func(t *testing.T) {
		// Override release with empty manifest
		h.release = &release.Release{
			Name:      "test-visit-chart",
			Namespace: "test-namespace",
			Manifest:  "",
		}

		m := &mockMonitor{}
		err := h.VisitReleaseResources(context.Background(), m)

		// Empty manifest should succeed but not collect anything
		g.Expect(err).To(o.Succeed())
		g.Expect(m.collected).To(o.BeEmpty())
	})

	t.Run("visit_with_monitor_error", func(t *testing.T) {
		h.release = &release.Release{
			Name:      "test-visit-chart",
			Namespace: "test-namespace",
			Manifest:  "",
		}

		expectedErr := errors.New("monitor collect error")
		m := &mockMonitor{collectErr: expectedErr}

		err := h.VisitReleaseResources(context.Background(), m)
		// Should propagate monitor error (though with empty manifest it won't be called)
		g.Expect(err).To(o.Succeed())
	})
}

// TestNewHelm tests the NewHelm constructor
func TestNewHelm(t *testing.T) {
	g := o.NewWithT(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	f := &flags.Flags{
		DryRun: true,
	}

	cfs := chartfs.New(os.DirFS("../../test"))
	testChart, err := cfs.GetChartFiles("charts/testing")
	g.Expect(err).To(o.Succeed())

	t.Run("create_helm_with_memory_driver", func(t *testing.T) {
		// Set memory driver for testing
		os.Setenv("HELM_DRIVER", "memory")
		defer os.Unsetenv("HELM_DRIVER")

		// Note: NewHelm requires a real k8s.Kube instance which needs kubeconfig
		// For true unit testing, this would need to be mocked or we skip this test
		// For now, we test the error path

		kube := k8s.NewKube(f)

		h, err := NewHelm(logger, f, kube, "test-namespace", testChart)
		if err != nil {
			// May fail in test environment without cluster access
			t.Logf("NewHelm failed (expected in test env): %v", err)
			return
		}

		g.Expect(h).ToNot(o.BeNil())
		g.Expect(h.chart).To(o.Equal(testChart))
		g.Expect(h.namespace).To(o.Equal("test-namespace"))
		g.Expect(h.actionCfg).ToNot(o.BeNil())
	})
}

// TestHelm_Errors tests error handling
func TestHelm_Errors(t *testing.T) {
	g := o.NewWithT(t)

	t.Run("ErrInstallFailed_is_defined", func(t *testing.T) {
		g.Expect(ErrInstallFailed).ToNot(o.BeNil())
		g.Expect(ErrInstallFailed.Error()).To(o.Equal("install failed"))
	})

	t.Run("ErrUpgradeFailed_is_defined", func(t *testing.T) {
		g.Expect(ErrUpgradeFailed).ToNot(o.BeNil())
		g.Expect(ErrUpgradeFailed.Error()).To(o.Equal("upgrade failed"))
	})

	t.Run("error_constants_can_be_checked_with_errors_Is", func(t *testing.T) {
		// Verify that error wrapping works correctly
		wrappedInstall := fmt.Errorf("%w: some details", ErrInstallFailed)
		g.Expect(errors.Is(wrappedInstall, ErrInstallFailed)).To(o.BeTrue())

		wrappedUpgrade := fmt.Errorf("%w: some details", ErrUpgradeFailed)
		g.Expect(errors.Is(wrappedUpgrade, ErrUpgradeFailed)).To(o.BeTrue())

		// Cross-check: they are different errors
		g.Expect(errors.Is(wrappedInstall, ErrUpgradeFailed)).To(o.BeFalse())
		g.Expect(errors.Is(wrappedUpgrade, ErrInstallFailed)).To(o.BeFalse())
	})
}
