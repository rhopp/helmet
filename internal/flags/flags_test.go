package flags

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/pflag"

	o "github.com/onsi/gomega"
)

func TestFlags_PersistentFlags(t *testing.T) {
	g := o.NewWithT(t)

	logLevel := slog.LevelWarn
	flags := &Flags{
		Debug:          false,
		DryRun:         false,
		KubeConfigPath: "/path/to/kubeconfig",
		LogLevel:       &logLevel,
		Timeout:        15 * time.Minute,
		Version:        false,
	}

	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.PersistentFlags(flagSet)

	// Verify all flags are registered
	g.Expect(flagSet.Lookup("debug")).ToNot(o.BeNil())
	g.Expect(flagSet.Lookup("dry-run")).ToNot(o.BeNil())
	g.Expect(flagSet.Lookup("version")).ToNot(o.BeNil())
	g.Expect(flagSet.Lookup("kube-config")).ToNot(o.BeNil())
	g.Expect(flagSet.Lookup("log-level")).ToNot(o.BeNil())
	g.Expect(flagSet.Lookup("timeout")).ToNot(o.BeNil())

	// Verify default values
	debugFlag, err := flagSet.GetBool("debug")
	g.Expect(err).To(o.Succeed())
	g.Expect(debugFlag).To(o.BeFalse())

	dryRunFlag, err := flagSet.GetBool("dry-run")
	g.Expect(err).To(o.Succeed())
	g.Expect(dryRunFlag).To(o.BeFalse())

	kubeConfigFlag, err := flagSet.GetString("kube-config")
	g.Expect(err).To(o.Succeed())
	g.Expect(kubeConfigFlag).To(o.Equal("/path/to/kubeconfig"))
}

func TestFlags_PersistentFlags_SetValues(t *testing.T) {
	g := o.NewWithT(t)

	logLevel := slog.LevelWarn
	flags := &Flags{
		Debug:          false,
		DryRun:         false,
		KubeConfigPath: "/default/path",
		LogLevel:       &logLevel,
		Timeout:        15 * time.Minute,
		Version:        false,
	}

	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.PersistentFlags(flagSet)

	// Set values via flags
	err := flagSet.Set("debug", "true")
	g.Expect(err).To(o.Succeed())
	g.Expect(flags.Debug).To(o.BeTrue())

	err = flagSet.Set("dry-run", "true")
	g.Expect(err).To(o.Succeed())
	g.Expect(flags.DryRun).To(o.BeTrue())

	err = flagSet.Set("kube-config", "/custom/path")
	g.Expect(err).To(o.Succeed())
	g.Expect(flags.KubeConfigPath).To(o.Equal("/custom/path"))

	err = flagSet.Set("log-level", "debug")
	g.Expect(err).To(o.Succeed())
	g.Expect(*flags.LogLevel).To(o.Equal(slog.LevelDebug))

	err = flagSet.Set("timeout", "30m")
	g.Expect(err).To(o.Succeed())
	g.Expect(flags.Timeout).To(o.Equal(30 * time.Minute))
}

func TestFlags_GetLogger(t *testing.T) {
	t.Run("with warn level", func(t *testing.T) {
		g := o.NewWithT(t)
		logLevel := slog.LevelWarn
		flags := &Flags{LogLevel: &logLevel}

		var buf bytes.Buffer
		logger := flags.GetLogger(&buf)

		g.Expect(logger).ToNot(o.BeNil())

		// Log at different levels
		logger.Debug("debug message")   // should not appear
		logger.Info("info message")     // should not appear
		logger.Warn("warning message")  // should appear
		logger.Error("error message")   // should appear

		output := buf.String()
		g.Expect(output).ToNot(o.ContainSubstring("debug message"))
		g.Expect(output).ToNot(o.ContainSubstring("info message"))
		g.Expect(output).To(o.ContainSubstring("warning message"))
		g.Expect(output).To(o.ContainSubstring("error message"))
	})

	t.Run("with debug level", func(t *testing.T) {
		g := o.NewWithT(t)
		logLevel := slog.LevelDebug
		flags := &Flags{LogLevel: &logLevel}

		var buf bytes.Buffer
		logger := flags.GetLogger(&buf)

		logger.Debug("debug message")
		logger.Info("info message")

		output := buf.String()
		g.Expect(output).To(o.ContainSubstring("debug message"))
		g.Expect(output).To(o.ContainSubstring("info message"))
	})
}

func TestFlags_LoggerWith(t *testing.T) {
	g := o.NewWithT(t)

	logLevel := slog.LevelInfo
	flags := &Flags{
		Debug:    true,
		DryRun:   true,
		Timeout:  10 * time.Minute,
		LogLevel: &logLevel,
	}

	var buf bytes.Buffer
	baseLogger := flags.GetLogger(&buf)
	contextLogger := flags.LoggerWith(baseLogger)

	g.Expect(contextLogger).ToNot(o.BeNil())

	// Log a message with the contextual logger
	contextLogger.Info("test message")

	output := buf.String()
	g.Expect(output).To(o.ContainSubstring("test message"))
	g.Expect(output).To(o.ContainSubstring("debug=true"))
	g.Expect(output).To(o.ContainSubstring("dry-run=true"))
	g.Expect(output).To(o.ContainSubstring("timeout=10m0s"))
}

func TestFlags_ShowVersion(t *testing.T) {
	g := o.NewWithT(t)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	flags := &Flags{}
	flags.ShowVersion("test-app", "v1.2.3", "abc123def456")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	g.Expect(err).To(o.Succeed())

	output := buf.String()
	g.Expect(output).To(o.ContainSubstring("test-app"))
	g.Expect(output).To(o.ContainSubstring("Version: v1.2.3"))
	g.Expect(output).To(o.ContainSubstring("Commit: abc123def456"))
}

func TestNewFlags(t *testing.T) {
	g := o.NewWithT(t)

	// Unset KUBECONFIG to test default behavior
	oldKubeConfig := os.Getenv("KUBECONFIG")
	os.Unsetenv("KUBECONFIG")
	defer func() {
		if oldKubeConfig != "" {
			os.Setenv("KUBECONFIG", oldKubeConfig)
		}
	}()

	flags := NewFlags()

	g.Expect(flags).ToNot(o.BeNil())
	g.Expect(flags.Debug).To(o.BeFalse())
	g.Expect(flags.DryRun).To(o.BeFalse())
	g.Expect(flags.Version).To(o.BeFalse())
	g.Expect(flags.LogLevel).ToNot(o.BeNil())
	g.Expect(*flags.LogLevel).To(o.Equal(slog.LevelWarn))
	g.Expect(flags.Timeout).To(o.Equal(15 * time.Minute))
	// KubeConfigPath should be set to default path when KUBECONFIG is not set
	g.Expect(flags.KubeConfigPath).To(o.ContainSubstring(".kube/config"))
}

func TestNewFlags_KubeConfigFromEnv(t *testing.T) {
	g := o.NewWithT(t)

	// Set KUBECONFIG environment variable
	customPath := "/custom/kubeconfig"
	os.Setenv("KUBECONFIG", customPath)
	defer os.Unsetenv("KUBECONFIG")

	flags := NewFlags()

	g.Expect(flags.KubeConfigPath).To(o.Equal(customPath))
}

func TestNewFlags_KubeConfigDefault(t *testing.T) {
	g := o.NewWithT(t)

	// Ensure KUBECONFIG is not set
	os.Unsetenv("KUBECONFIG")

	flags := NewFlags()

	// Should default to ~/.kube/config
	g.Expect(flags.KubeConfigPath).To(o.ContainSubstring(".kube/config"))
	g.Expect(strings.HasSuffix(flags.KubeConfigPath, "/.kube/config")).To(o.BeTrue())
}
