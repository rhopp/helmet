package engine

import (
	"os"
	"testing"

	"github.com/redhat-appstudio/helmet/internal/chartfs"
	"github.com/redhat-appstudio/helmet/internal/config"

	o "github.com/onsi/gomega"
)

// TestGetMinorVersion tests the getMinorVersion function
func TestGetMinorVersion(t *testing.T) {
	testCases := []struct {
		name        string
		version     string
		expected    string
		expectError bool
	}{
		{
			name:        "standard_version",
			version:     "4.12.3",
			expected:    "4.12",
			expectError: false,
		},
		{
			name:        "two_part_version",
			version:     "1.2",
			expected:    "1.2",
			expectError: false,
		},
		{
			name:        "version_with_suffix",
			version:     "4.13.0-rc.1",
			expected:    "4.13",
			expectError: false,
		},
		{
			name:        "single_part_version",
			version:     "4",
			expected:    "",
			expectError: true,
		},
		{
			name:        "empty_version",
			version:     "",
			expected:    "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := o.NewWithT(t)

			result, err := getMinorVersion(tc.version)

			if tc.expectError {
				g.Expect(err).To(o.HaveOccurred())
				g.Expect(err.Error()).To(o.ContainSubstring("version does not include a minor part"))
			} else {
				g.Expect(err).To(o.Succeed())
				g.Expect(result).To(o.Equal(tc.expected))
			}
		})
	}
}

// TestVariablesSetInstaller tests the SetInstaller method
func TestVariablesSetInstaller(t *testing.T) {
	g := o.NewWithT(t)

	cfs := chartfs.New(os.DirFS("../../test"))
	configBytes, err := cfs.ReadFile("config.yaml")
	g.Expect(err).To(o.Succeed())

	cfg, err := config.NewConfigFromBytes(configBytes, "test-namespace")
	g.Expect(err).To(o.Succeed())

	vars := NewVariables()
	err = vars.SetInstaller(cfg)
	g.Expect(err).To(o.Succeed())

	// Verify Installer.Namespace was set
	g.Expect(vars.Installer["Namespace"]).To(o.Equal("test-namespace"))

	// Verify Installer.Settings was set
	settings, ok := vars.Installer["Settings"].(map[string]interface{})
	g.Expect(ok).To(o.BeTrue())
	g.Expect(settings).ToNot(o.BeEmpty())

	// Verify Installer.Products was set
	g.Expect(vars.Installer["Products"]).ToNot(o.BeNil())
}

// TestVariablesSetOpenShift is skipped because SetOpenShift requires
// a real OpenShift cluster with specific resources. This function needs
// integration testing rather than unit testing.
// Coverage note: SetOpenShift calls getMinorVersion which is tested separately.

// TestVariablesUnstructured tests the Unstructured method
func TestVariablesUnstructured(t *testing.T) {
	g := o.NewWithT(t)

	vars := NewVariables()
	vars.Installer["test"] = "value"
	vars.OpenShift["version"] = "4.12"

	result, err := vars.Unstructured()
	g.Expect(err).To(o.Succeed())
	g.Expect(result).ToNot(o.BeEmpty())

	// Verify the structure contains Installer and OpenShift
	installer, ok := result["Installer"]
	g.Expect(ok).To(o.BeTrue())
	g.Expect(installer).ToNot(o.BeNil())

	openshift, ok := result["OpenShift"]
	g.Expect(ok).To(o.BeTrue())
	g.Expect(openshift).ToNot(o.BeNil())
}

// TestNewVariables tests the constructor
func TestNewVariables(t *testing.T) {
	g := o.NewWithT(t)

	vars := NewVariables()
	g.Expect(vars).ToNot(o.BeNil())
	g.Expect(vars.Installer).ToNot(o.BeNil())
	g.Expect(vars.OpenShift).ToNot(o.BeNil())
	g.Expect(vars.Installer).To(o.BeEmpty())
	g.Expect(vars.OpenShift).To(o.BeEmpty())
}
