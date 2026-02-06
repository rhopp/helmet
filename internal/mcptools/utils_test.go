package mcptools

import (
	"errors"
	"strings"
	"testing"

	o "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

// TestMissingClusterConfigErrorFromErr tests error message generation
func TestMissingClusterConfigErrorFromErr(t *testing.T) {
	g := o.NewWithT(t)

	appName := "test-app"
	originalErr := errors.New("config not found in cluster")

	errMsg := missingClusterConfigErrorFromErr(appName, originalErr)

	// Verify the error message contains expected elements
	g.Expect(errMsg).To(o.ContainSubstring(appName))
	g.Expect(errMsg).To(o.ContainSubstring("not configured yet"))
	g.Expect(errMsg).To(o.ContainSubstring("config not found in cluster"))
	g.Expect(errMsg).To(o.ContainSubstring(appName + configInitSuffix))
	g.Expect(errMsg).To(o.ContainSubstring(appName + statusSuffix))
}

// TestMissingClusterConfigErrorFromErr_WithDifferentErrors tests with various error types
func TestMissingClusterConfigErrorFromErr_WithDifferentErrors(t *testing.T) {
	testCases := []struct {
		name        string
		appName     string
		err         error
		shouldFind  []string
	}{
		{
			name:    "simple_error",
			appName: "my-app",
			err:     errors.New("simple error"),
			shouldFind: []string{
				"my-app",
				"simple error",
				"my-app_config_init",
				"my-app_status",
			},
		},
		{
			name:    "complex_error_message",
			appName: "complex-app",
			err:     errors.New("failed to connect: network timeout after 30s"),
			shouldFind: []string{
				"complex-app",
				"failed to connect",
				"network timeout",
			},
		},
		{
			name:    "empty_app_name",
			appName: "",
			err:     errors.New("test error"),
			shouldFind: []string{
				"test error",
				"_config_init",
				"_status",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := o.NewWithT(t)

			result := missingClusterConfigErrorFromErr(tc.appName, tc.err)

			for _, expected := range tc.shouldFind {
				g.Expect(result).To(o.ContainSubstring(expected))
			}
		})
	}
}

// TestGenerateIntegrationSubCmdUsage tests usage string generation
func TestGenerateIntegrationSubCmdUsage(t *testing.T) {
	g := o.NewWithT(t)

	appName := "test-app"
	cmd := &cobra.Command{
		Use:  "github",
		Long: "Configure GitHub integration",
	}

	// Add a required flag
	cmd.PersistentFlags().String("token", "", "GitHub token")
	cmd.PersistentFlags().String("org", "", "GitHub organization")

	// Mark token as required using annotations
	if cmd.PersistentFlags().Lookup("token") != nil {
		cmd.PersistentFlags().Lookup("token").Annotations = map[string][]string{
			cobra.BashCompOneRequiredFlag: {"true"},
		}
	}

	usage := generateIntegrationSubCmdUsage(appName, cmd)

	g.Expect(usage).To(o.ContainSubstring("test-app integration github"))
	g.Expect(usage).To(o.ContainSubstring("Configure GitHub integration"))
	g.Expect(usage).To(o.ContainSubstring("--token=\"OVERWRITE_ME\""))
	g.Expect(usage).To(o.ContainSubstring("Example:"))
}

// TestGenerateIntegrationSubCmdUsage_NoRequiredFlags tests with no required flags
func TestGenerateIntegrationSubCmdUsage_NoRequiredFlags(t *testing.T) {
	g := o.NewWithT(t)

	appName := "simple-app"
	cmd := &cobra.Command{
		Use:  "test",
		Long: "Test command with no required flags",
	}

	// Add optional flags (no required annotation)
	cmd.PersistentFlags().String("optional", "", "Optional flag")

	usage := generateIntegrationSubCmdUsage(appName, cmd)

	g.Expect(usage).To(o.ContainSubstring("simple-app integration test"))
	g.Expect(usage).To(o.ContainSubstring("Test command with no required flags"))
	g.Expect(usage).ToNot(o.ContainSubstring("--optional"))
}

// TestGenerateIntegrationSubCmdUsage_MultipleRequiredFlags tests with multiple required flags
func TestGenerateIntegrationSubCmdUsage_MultipleRequiredFlags(t *testing.T) {
	g := o.NewWithT(t)

	appName := "multi-app"
	cmd := &cobra.Command{
		Use:  "complex",
		Long: "Complex integration with multiple requirements",
	}

	// Add multiple required flags
	flags := []string{"api-key", "endpoint", "region"}
	for _, flagName := range flags {
		cmd.PersistentFlags().String(flagName, "", flagName+" description")
		if flag := cmd.PersistentFlags().Lookup(flagName); flag != nil {
			flag.Annotations = map[string][]string{
				cobra.BashCompOneRequiredFlag: {"true"},
			}
		}
	}

	usage := generateIntegrationSubCmdUsage(appName, cmd)

	g.Expect(usage).To(o.ContainSubstring("multi-app integration complex"))
	g.Expect(usage).To(o.ContainSubstring("--api-key=\"OVERWRITE_ME\""))
	g.Expect(usage).To(o.ContainSubstring("--endpoint=\"OVERWRITE_ME\""))
	g.Expect(usage).To(o.ContainSubstring("--region=\"OVERWRITE_ME\""))
}

// TestGenerateIntegrationSubCmdUsage_Structure tests the overall structure
func TestGenerateIntegrationSubCmdUsage_Structure(t *testing.T) {
	g := o.NewWithT(t)

	cmd := &cobra.Command{
		Use:  "structure-test",
		Long: "Testing structure formatting",
	}

	usage := generateIntegrationSubCmdUsage("app", cmd)

	// Verify structure contains expected sections
	g.Expect(usage).To(o.ContainSubstring("## `structure-test` Subcommand Usage"))
	g.Expect(usage).To(o.ContainSubstring("Testing structure formatting"))
	g.Expect(usage).To(o.ContainSubstring("Example:"))
	g.Expect(usage).To(o.ContainSubstring("\t"))

	// Should have proper markdown formatting
	lines := strings.Split(usage, "\n")
	g.Expect(len(lines)).To(o.BeNumerically(">", 2))
}

// TestConstantSuffixes tests the suffix constants
func TestConstantSuffixes(t *testing.T) {
	g := o.NewWithT(t)

	// These constants are defined in configtools.go and statustool.go
	// We're testing them indirectly through the utility functions
	appName := "test"

	// Test through error message
	err := errors.New("test")
	msg := missingClusterConfigErrorFromErr(appName, err)

	g.Expect(msg).To(o.ContainSubstring("test_config_init"))
	g.Expect(msg).To(o.ContainSubstring("test_status"))
}

// TestGenerateIntegrationSubCmdUsage_EmptyCommand tests with minimal command
func TestGenerateIntegrationSubCmdUsage_EmptyCommand(t *testing.T) {
	g := o.NewWithT(t)

	cmd := &cobra.Command{
		Use: "minimal",
	}

	usage := generateIntegrationSubCmdUsage("app", cmd)

	g.Expect(usage).To(o.ContainSubstring("app integration minimal"))
	g.Expect(usage).To(o.ContainSubstring("Example:"))
}
