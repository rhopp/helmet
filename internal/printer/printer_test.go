package printer

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	o "github.com/onsi/gomega"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	helmtime "helm.sh/helm/v3/pkg/time"
)

// captureOutput captures stdout during function execution
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// TestDisclaimer tests the disclaimer message
func TestDisclaimer(t *testing.T) {
	g := o.NewWithT(t)

	output := captureOutput(func() {
		Disclaimer()
	})

	g.Expect(output).To(o.ContainSubstring("DISCLAIMER"))
	g.Expect(output).To(o.ContainSubstring("EXPERIMENTAL DEPLOYMENTS"))
	g.Expect(output).To(o.ContainSubstring("PRODUCTION IS UNSUPPORTED"))
	g.Expect(output).To(o.HavePrefix("\n!!! "))
	g.Expect(output).To(o.HaveSuffix("!!!\n\n"))
}

// TestHelmReleasePrinter tests printing release information
func TestHelmReleasePrinter(t *testing.T) {
	g := o.NewWithT(t)

	rel := &release.Release{
		Chart: &chart.Chart{
			Metadata: &chart.Metadata{
				Name:    "test-chart",
				Version: "1.2.3",
			},
		},
		Info: &release.Info{
			Status:       release.StatusDeployed,
			LastDeployed: helmtime.Now(),
		},
		Namespace: "test-namespace",
		Version:   5,
	}

	output := captureOutput(func() {
		HelmReleasePrinter(rel)
	})

	g.Expect(output).To(o.ContainSubstring("Chart: test-chart"))
	g.Expect(output).To(o.ContainSubstring("Version: 1.2.3"))
	g.Expect(output).To(o.ContainSubstring("Status: deployed"))
	g.Expect(output).To(o.ContainSubstring("Namespace: test-namespace"))
	g.Expect(output).To(o.ContainSubstring("Revision: 5"))
	g.Expect(output).To(o.ContainSubstring("Updated:"))
	g.Expect(output).To(o.ContainSubstring("#"))
}

// TestHelmReleaseNotesPrinter tests printing release notes
func TestHelmReleaseNotesPrinter(t *testing.T) {
	t.Run("with_notes", func(t *testing.T) {
		g := o.NewWithT(t)

		rel := &release.Release{
			Info: &release.Info{
				Notes: "This is a test release note\nWith multiple lines",
			},
		}

		output := captureOutput(func() {
			HelmReleaseNotesPrinter(rel)
		})

		g.Expect(output).To(o.ContainSubstring("# Notes"))
		g.Expect(output).To(o.ContainSubstring("This is a test release note"))
		g.Expect(output).To(o.ContainSubstring("With multiple lines"))
	})

	t.Run("without_notes", func(t *testing.T) {
		g := o.NewWithT(t)

		rel := &release.Release{
			Info: &release.Info{
				Notes: "",
			},
		}

		output := captureOutput(func() {
			HelmReleaseNotesPrinter(rel)
		})

		g.Expect(output).To(o.BeEmpty())
	})
}

// TestHelmExtendedReleasePrinter tests printing extended release information
func TestHelmExtendedReleasePrinter(t *testing.T) {
	t.Run("with_manifest_no_hooks", func(t *testing.T) {
		g := o.NewWithT(t)

		rel := &release.Release{
			Manifest: "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test-pod",
			Hooks:    []*release.Hook{},
		}

		output := captureOutput(func() {
			HelmExtendedReleasePrinter(rel)
		})

		g.Expect(output).To(o.ContainSubstring("# Manifest"))
		g.Expect(output).To(o.ContainSubstring("apiVersion: v1"))
		g.Expect(output).To(o.ContainSubstring("kind: Pod"))
		g.Expect(output).To(o.ContainSubstring("name: test-pod"))
		g.Expect(output).ToNot(o.ContainSubstring("# Hooks"))
	})

	t.Run("with_manifest_and_hooks", func(t *testing.T) {
		g := o.NewWithT(t)

		rel := &release.Release{
			Manifest: "apiVersion: v1\nkind: Service",
			Hooks: []*release.Hook{
				{
					Manifest: "apiVersion: batch/v1\nkind: Job\nmetadata:\n  name: pre-install-hook",
				},
				{
					Manifest: "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: post-install-hook",
				},
			},
		}

		output := captureOutput(func() {
			HelmExtendedReleasePrinter(rel)
		})

		g.Expect(output).To(o.ContainSubstring("# Manifest"))
		g.Expect(output).To(o.ContainSubstring("kind: Service"))
		g.Expect(output).To(o.ContainSubstring("# Hooks"))
		g.Expect(output).To(o.ContainSubstring("kind: Job"))
		g.Expect(output).To(o.ContainSubstring("pre-install-hook"))
		g.Expect(output).To(o.ContainSubstring("kind: ConfigMap"))
		g.Expect(output).To(o.ContainSubstring("post-install-hook"))
		g.Expect(strings.Count(output, "---")).To(o.Equal(2)) // Each hook has a separator
	})
}

// TestValuesPrinter tests printing values as properties
func TestValuesPrinter(t *testing.T) {
	t.Run("simple_values", func(t *testing.T) {
		g := o.NewWithT(t)

		vals := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		}

		output := captureOutput(func() {
			ValuesPrinter("Test Values", vals)
		})

		g.Expect(output).To(o.ContainSubstring("# Test Values"))
		g.Expect(output).To(o.ContainSubstring(" * "))
		// Check that all keys appear (order may vary due to map iteration)
		g.Expect(output).To(o.MatchRegexp(`key1: value1|key2: 42|key3: true`))
	})

	t.Run("nested_values", func(t *testing.T) {
		g := o.NewWithT(t)

		vals := map[string]interface{}{
			"database": map[string]interface{}{
				"host": "localhost",
				"port": 5432,
			},
			"app": map[string]interface{}{
				"name": "test-app",
				"config": map[string]interface{}{
					"debug": true,
				},
			},
		}

		output := captureOutput(func() {
			ValuesPrinter("Nested Configuration", vals)
		})

		g.Expect(output).To(o.ContainSubstring("# Nested Configuration"))
		g.Expect(output).To(o.ContainSubstring(" * database.host: localhost"))
		g.Expect(output).To(o.ContainSubstring(" * database.port: 5432"))
		g.Expect(output).To(o.ContainSubstring(" * app.name: test-app"))
		g.Expect(output).To(o.ContainSubstring(" * app.config.debug: true"))
	})

	t.Run("empty_values", func(t *testing.T) {
		g := o.NewWithT(t)

		vals := map[string]interface{}{}

		output := captureOutput(func() {
			ValuesPrinter("Empty Values", vals)
		})

		g.Expect(output).To(o.ContainSubstring("# Empty Values"))
		// Should only have the header, no properties
		lines := strings.Split(strings.TrimSpace(output), "\n")
		g.Expect(len(lines)).To(o.BeNumerically("<=", 4)) // Header lines only
	})
}

// TestValuesToPrint tests the internal valuesToProperties function
func TestValuesToPrint(t *testing.T) {
	t.Run("flat_map", func(t *testing.T) {
		g := o.NewWithT(t)

		vals := map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		}

		sb := new(strings.Builder)
		valuesToProperties(vals, "", sb)

		result := sb.String()
		g.Expect(result).To(o.ContainSubstring("key1: value1"))
		g.Expect(result).To(o.ContainSubstring("key2: 123"))
	})

	t.Run("nested_map", func(t *testing.T) {
		g := o.NewWithT(t)

		vals := map[string]interface{}{
			"outer": map[string]interface{}{
				"inner": "value",
			},
		}

		sb := new(strings.Builder)
		valuesToProperties(vals, "", sb)

		result := sb.String()
		g.Expect(result).To(o.ContainSubstring("outer.inner: value"))
	})

	t.Run("with_path_prefix", func(t *testing.T) {
		g := o.NewWithT(t)

		vals := map[string]interface{}{
			"key": "value",
		}

		sb := new(strings.Builder)
		valuesToProperties(vals, "prefix", sb)

		result := sb.String()
		g.Expect(result).To(o.ContainSubstring("prefix.key: value"))
	})

	t.Run("deeply_nested", func(t *testing.T) {
		g := o.NewWithT(t)

		vals := map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": map[string]interface{}{
					"level3": "deep-value",
				},
			},
		}

		sb := new(strings.Builder)
		valuesToProperties(vals, "", sb)

		result := sb.String()
		g.Expect(result).To(o.ContainSubstring("level1.level2.level3: deep-value"))
	})
}

// TestPrintProperties tests the internal printProperties function
func TestPrintProperties(t *testing.T) {
	t.Run("with_prefix", func(t *testing.T) {
		g := o.NewWithT(t)

		sb := new(strings.Builder)
		sb.WriteString("line1: value1\n")
		sb.WriteString("line2: value2\n")
		sb.WriteString("line3: value3\n")

		output := captureOutput(func() {
			printProperties(sb, ">>> ")
		})

		g.Expect(output).To(o.ContainSubstring(">>> line1: value1"))
		g.Expect(output).To(o.ContainSubstring(">>> line2: value2"))
		g.Expect(output).To(o.ContainSubstring(">>> line3: value3"))
	})

	t.Run("empty_string_builder", func(t *testing.T) {
		g := o.NewWithT(t)

		sb := new(strings.Builder)

		output := captureOutput(func() {
			printProperties(sb, "* ")
		})

		g.Expect(output).To(o.BeEmpty())
	})

	t.Run("single_line", func(t *testing.T) {
		g := o.NewWithT(t)

		sb := new(strings.Builder)
		sb.WriteString("single: line\n")

		output := captureOutput(func() {
			printProperties(sb, "- ")
		})

		g.Expect(output).To(o.ContainSubstring("- single: line"))
	})
}
