package hooks

import (
	"bytes"
	"os"
	"testing"

	"github.com/redhat-appstudio/helmet/api"
	"github.com/redhat-appstudio/helmet/internal/chartfs"
	"github.com/redhat-appstudio/helmet/internal/resolver"

	o "github.com/onsi/gomega"
)

func TestNewHooks(t *testing.T) {
	g := o.NewWithT(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cfs := chartfs.New(os.DirFS("../../test"))

	chart, err := cfs.GetChartFiles("charts/testing")
	g.Expect(err).To(o.Succeed())

	appCtx := api.NewAppContext("tssc")
	h := NewHooks(
		resolver.NewDependencyWithNamespace(chart, appCtx.Namespace),
		&stdout,
		&stderr,
	)

	vals := map[string]interface{}{
		"key": map[string]interface{}{
			"nested": "value",
		},
	}

	t.Run("PreDeploy", func(t *testing.T) {
		err := h.PreDeploy(vals)
		g.Expect(err).To(o.Succeed())

		t.Logf("stdout: %s", stdout.String())
		t.Logf("stderr: %s", stderr.String())
		// Asserting the environment variable is printed out by the hook script,
		// the variable is passed by the informed values.
		g.Expect(stdout.String()).
			To(o.ContainSubstring("# INSTALLER__KEY__NESTED='value'"))

		stdout.Reset()
		stderr.Reset()
	})

	t.Run("PostDeploy", func(t *testing.T) {
		err := h.PostDeploy(vals)
		g.Expect(err).To(o.Succeed())

		t.Logf("stdout: %s", stdout.String())
		t.Logf("stderr: %s", stderr.String())
		g.Expect(stdout.String()).To(o.ContainSubstring("script runs after"))

		stdout.Reset()
		stderr.Reset()
	})
}

// TestHooksWithNoScript tests hooks when chart has no hook scripts
func TestHooksWithNoScript(t *testing.T) {
	g := o.NewWithT(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cfs := chartfs.New(os.DirFS("../../test"))

	// Use a chart without hook scripts
	chart, err := cfs.GetChartFiles("charts/helmet-product-a")
	g.Expect(err).To(o.Succeed())

	appCtx := api.NewAppContext("tssc")
	h := NewHooks(
		resolver.NewDependencyWithNamespace(chart, appCtx.Namespace),
		&stdout,
		&stderr,
	)

	vals := map[string]interface{}{"key": "value"}

	// Should not error when hook script doesn't exist
	err = h.PreDeploy(vals)
	g.Expect(err).To(o.BeNil())

	err = h.PostDeploy(vals)
	g.Expect(err).To(o.BeNil())
}

// TestHooksExec tests the exec function directly
func TestHooksExec(t *testing.T) {
	g := o.NewWithT(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cfs := chartfs.New(os.DirFS("../../test"))

	chart, err := cfs.GetChartFiles("charts/testing")
	g.Expect(err).To(o.Succeed())

	appCtx := api.NewAppContext("tssc")
	h := NewHooks(
		resolver.NewDependencyWithNamespace(chart, appCtx.Namespace),
		&stdout,
		&stderr,
	)

	t.Run("successful script execution", func(t *testing.T) {
		// Create a simple successful script
		tmpFile, err := os.CreateTemp("/tmp", "test-hook-*.sh")
		g.Expect(err).To(o.BeNil())
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString("#!/bin/bash\necho 'success'\n")
		g.Expect(err).To(o.BeNil())
		tmpFile.Close()
		os.Chmod(tmpFile.Name(), 0o755)

		vals := map[string]interface{}{"test": "value"}
		err = h.exec(tmpFile.Name(), vals)
		g.Expect(err).To(o.BeNil())
		g.Expect(stdout.String()).To(o.ContainSubstring("success"))

		stdout.Reset()
	})

	t.Run("failing script execution", func(t *testing.T) {
		// Create a script that exits with error
		tmpFile, err := os.CreateTemp("/tmp", "test-hook-*.sh")
		g.Expect(err).To(o.BeNil())
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString("#!/bin/bash\nexit 1\n")
		g.Expect(err).To(o.BeNil())
		tmpFile.Close()
		os.Chmod(tmpFile.Name(), 0o755)

		vals := map[string]interface{}{"test": "value"}
		err = h.exec(tmpFile.Name(), vals)
		g.Expect(err).ToNot(o.BeNil())
	})
}
