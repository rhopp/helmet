package integrations

import (
	"context"
	"log/slog"
	"testing"

	"github.com/redhat-appstudio/helmet/api"
	"github.com/redhat-appstudio/helmet/internal/config"
	o "github.com/onsi/gomega"
)

// TestRegister_DirectManipulation tests the Register method
func TestRegister_DirectManipulation(t *testing.T) {
	g := o.NewWithT(t)

	mgr := NewManager()

	// Create a module
	mod := api.IntegrationModule{
		Name: "test-integration",
	}

	// Create a nil integration just for testing structure
	// Since we only test the registration logic, not the integration functionality
	mgr.integrations[IntegrationName("test-integration")] = nil
	mgr.modules[IntegrationName("test-integration")] = mod

	// Verify registration
	g.Expect(mgr.integrations).To(o.HaveLen(1))
	g.Expect(mgr.modules).To(o.HaveLen(1))

	names := mgr.IntegrationNames()
	g.Expect(names).To(o.ContainElement("test-integration"))

	modules := mgr.GetModules()
	g.Expect(modules).To(o.HaveLen(1))
	g.Expect(modules[0].Name).To(o.Equal("test-integration"))
}

// TestIntegration_RetrieveExisting tests getting an existing integration
func TestIntegration_RetrieveExisting(t *testing.T) {
	g := o.NewWithT(t)

	mgr := NewManager()

	// Manually add an integration
	mgr.integrations[GitHub] = nil  // nil is okay for this structural test

	// Should not panic when retrieving existing integration
	result := mgr.Integration(GitHub)
	g.Expect(result).To(o.BeNil())  // we added nil, so we get nil back
}

// TestIntegrationNames_Multiple tests getting names from populated manager
func TestIntegrationNames_Multiple(t *testing.T) {
	g := o.NewWithT(t)

	mgr := NewManager()

	// Add multiple integrations
	mgr.integrations[GitHub] = nil
	mgr.integrations[GitLab] = nil
	mgr.integrations[BitBucket] = nil

	names := mgr.IntegrationNames()
	g.Expect(names).To(o.HaveLen(3))
	g.Expect(names).To(o.ContainElements("github", "gitlab", "bitbucket"))
}

// TestGetModules_Multiple tests getting modules from populated manager
func TestGetModules_Multiple(t *testing.T) {
	g := o.NewWithT(t)

	mgr := NewManager()

	// Add multiple modules
	mgr.modules[GitHub] = api.IntegrationModule{Name: "github"}
	mgr.modules[GitLab] = api.IntegrationModule{Name: "gitlab"}

	modules := mgr.GetModules()
	g.Expect(modules).To(o.HaveLen(2))

	// Extract names
	var names []string
	for _, mod := range modules {
		names = append(names, mod.Name)
	}
	g.Expect(names).To(o.ContainElements("github", "gitlab"))
}

// TestLoadModules_EmptyModuleList tests loading with empty module list
func TestLoadModules_EmptyModuleList(t *testing.T) {
	g := o.NewWithT(t)

	mgr := NewManager()
	logger := slog.Default()

	// Create a minimal fake kube for testing
	// Note: LoadModules calls Init which requires real dependencies
	// This test ensures LoadModules handles empty lists gracefully

	err := mgr.LoadModules("test-app", logger, nil, []api.IntegrationModule{})
	g.Expect(err).To(o.Succeed())
	g.Expect(mgr.integrations).To(o.BeEmpty())
	g.Expect(mgr.modules).To(o.BeEmpty())
}

// TestRegisterMethod tests the Register method directly
func TestRegisterMethod(t *testing.T) {
	g := o.NewWithT(t)

	mgr := NewManager()
	mod := api.IntegrationModule{
		Name: "custom-integration",
	}

	// Register with nil integration (structural test only)
	mgr.Register(mod, nil)

	// Verify it was registered
	g.Expect(mgr.integrations).To(o.HaveKey(IntegrationName("custom-integration")))
	g.Expect(mgr.modules).To(o.HaveKey(IntegrationName("custom-integration")))

	// Verify we can retrieve module
	modules := mgr.GetModules()
	g.Expect(modules).To(o.HaveLen(1))
	g.Expect(modules[0].Name).To(o.Equal("custom-integration"))
}

// TestConfiguredIntegrations_EmptyManager tests with no integrations
func TestConfiguredIntegrations_EmptyManager(t *testing.T) {
	g := o.NewWithT(t)

	mgr := NewManager()
	cfg := &config.Config{}
	ctx := context.Background()

	configured, err := mgr.ConfiguredIntegrations(ctx, cfg)
	g.Expect(err).To(o.Succeed())
	g.Expect(configured).To(o.BeEmpty())
	g.Expect(configured).ToNot(o.BeNil())
}

// TestIntegrationNamesAndModulesConsistency tests that names and modules stay in sync
func TestIntegrationNamesAndModulesConsistency(t *testing.T) {
	g := o.NewWithT(t)

	mgr := NewManager()

	// Register modules
	for _, name := range []IntegrationName{GitHub, GitLab, Jenkins} {
		mgr.integrations[name] = nil
		mgr.modules[name] = api.IntegrationModule{Name: string(name)}
	}

	// Verify counts match
	names := mgr.IntegrationNames()
	modules := mgr.GetModules()

	g.Expect(len(names)).To(o.Equal(len(modules)))
	g.Expect(names).To(o.HaveLen(3))
	g.Expect(modules).To(o.HaveLen(3))
}
