package api

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/redhat-appstudio/helmet/internal/config"
	"github.com/redhat-appstudio/helmet/internal/integration"
	"github.com/redhat-appstudio/helmet/internal/k8s"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"

	o "github.com/onsi/gomega"
)

// TestIntegrationModuleStructure tests the IntegrationModule struct fields
func TestIntegrationModuleStructure(t *testing.T) {
	g := o.NewWithT(t)

	// Create a simple integration module
	module := IntegrationModule{
		Name: "test-integration",
		Init: func(logger *slog.Logger, kube *k8s.Kube) integration.Interface {
			return &mockIntegration{}
		},
		Command: func(ctx *AppContext, logger *slog.Logger, kube *k8s.Kube, intg *integration.Integration) SubCommand {
			return newMockSubCommand()
		},
	}

	g.Expect(module.Name).To(o.Equal("test-integration"))
	g.Expect(module.Init).ToNot(o.BeNil())
	g.Expect(module.Command).ToNot(o.BeNil())
}

// TestIntegrationModuleInit tests the Init function
func TestIntegrationModuleInit(t *testing.T) {
	g := o.NewWithT(t)

	module := IntegrationModule{
		Name: "test-integration",
		Init: func(logger *slog.Logger, kube *k8s.Kube) integration.Interface {
			return &mockIntegration{name: "initialized"}
		},
		Command: nil,
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	kube := &k8s.Kube{}

	intg := module.Init(logger, kube)
	g.Expect(intg).ToNot(o.BeNil())

	mockIntg, ok := intg.(*mockIntegration)
	g.Expect(ok).To(o.BeTrue())
	g.Expect(mockIntg.name).To(o.Equal("initialized"))
}

// TestIntegrationModuleCommand tests the Command function
func TestIntegrationModuleCommand(t *testing.T) {
	g := o.NewWithT(t)

	module := IntegrationModule{
		Name: "test-integration",
		Init: nil,
		Command: func(ctx *AppContext, logger *slog.Logger, kube *k8s.Kube, intg *integration.Integration) SubCommand {
			return newMockSubCommand()
		},
	}

	appCtx := NewAppContext("test-app")
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	kube := &k8s.Kube{}
	intg := &integration.Integration{}

	subCmd := module.Command(appCtx, logger, kube, intg)
	g.Expect(subCmd).ToNot(o.BeNil())
	g.Expect(subCmd.Cmd()).ToNot(o.BeNil())
	g.Expect(subCmd.Cmd().Use).To(o.Equal("mock"))
}

// TestIntegrationModuleMultiple tests creating multiple integration modules
func TestIntegrationModuleMultiple(t *testing.T) {
	g := o.NewWithT(t)

	modules := []IntegrationModule{
		{
			Name: "github",
			Init: func(logger *slog.Logger, kube *k8s.Kube) integration.Interface {
				return &mockIntegration{name: "github"}
			},
			Command: func(ctx *AppContext, logger *slog.Logger, kube *k8s.Kube, intg *integration.Integration) SubCommand {
				return newMockSubCommand()
			},
		},
		{
			Name: "acs",
			Init: func(logger *slog.Logger, kube *k8s.Kube) integration.Interface {
				return &mockIntegration{name: "acs"}
			},
			Command: func(ctx *AppContext, logger *slog.Logger, kube *k8s.Kube, intg *integration.Integration) SubCommand {
				return newMockSubCommand()
			},
		},
	}

	g.Expect(len(modules)).To(o.Equal(2))
	g.Expect(modules[0].Name).To(o.Equal("github"))
	g.Expect(modules[1].Name).To(o.Equal("acs"))

	// Test that each can be initialized
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	kube := &k8s.Kube{}

	for _, module := range modules {
		intg := module.Init(logger, kube)
		g.Expect(intg).ToNot(o.BeNil())

		mockIntg, ok := intg.(*mockIntegration)
		g.Expect(ok).To(o.BeTrue())
		g.Expect(mockIntg.name).To(o.Equal(module.Name))
	}
}

// mockIntegration is a test implementation of integration.Interface
type mockIntegration struct {
	name string
}

func (m *mockIntegration) PersistentFlags(cmd *cobra.Command) {}

func (m *mockIntegration) LoggerWith(logger *slog.Logger) *slog.Logger {
	return logger.With("integration", m.name)
}

func (m *mockIntegration) Validate() error {
	return nil
}

func (m *mockIntegration) Type() corev1.SecretType {
	return corev1.SecretTypeOpaque
}

func (m *mockIntegration) SetArgument(key, value string) error {
	return nil
}

func (m *mockIntegration) Data(ctx context.Context, cfg *config.Config) (map[string][]byte, error) {
	return map[string][]byte{
		"test": []byte("data"),
	}, nil
}
