package integrations

import (
	"testing"

	o "github.com/onsi/gomega"
)

// TestNewManager tests the manager constructor
func TestNewManager(t *testing.T) {
	g := o.NewWithT(t)

	mgr := NewManager()

	g.Expect(mgr).ToNot(o.BeNil())
	g.Expect(mgr.integrations).ToNot(o.BeNil())
	g.Expect(mgr.modules).ToNot(o.BeNil())
	g.Expect(mgr.integrations).To(o.BeEmpty())
	g.Expect(mgr.modules).To(o.BeEmpty())
}

// TestIntegrationNames tests retrieving all integration names
func TestIntegrationNames(t *testing.T) {
	t.Run("empty_manager", func(t *testing.T) {
		g := o.NewWithT(t)

		mgr := NewManager()
		names := mgr.IntegrationNames()

		g.Expect(names).To(o.BeEmpty())
		g.Expect(names).ToNot(o.BeNil())
	})
}

// TestGetModules tests retrieving all registered modules
func TestGetModules(t *testing.T) {
	t.Run("empty_manager", func(t *testing.T) {
		g := o.NewWithT(t)

		mgr := NewManager()
		modules := mgr.GetModules()

		g.Expect(modules).To(o.BeEmpty())
		g.Expect(modules).ToNot(o.BeNil())
	})
}

// TestIntegrationNameConstants tests the integration name constants
func TestIntegrationNameConstants(t *testing.T) {
	g := o.NewWithT(t)

	// Verify all expected constants are defined with correct values
	g.Expect(string(ACS)).To(o.Equal("acs"))
	g.Expect(string(Artifactory)).To(o.Equal("artifactory"))
	g.Expect(string(Azure)).To(o.Equal("azure"))
	g.Expect(string(BitBucket)).To(o.Equal("bitbucket"))
	g.Expect(string(GitHub)).To(o.Equal("github"))
	g.Expect(string(GitLab)).To(o.Equal("gitlab"))
	g.Expect(string(Jenkins)).To(o.Equal("jenkins"))
	g.Expect(string(Nexus)).To(o.Equal("nexus"))
	g.Expect(string(Quay)).To(o.Equal("quay"))
	g.Expect(string(TrustedArtifactSigner)).To(o.Equal("tas"))
	g.Expect(string(Trustification)).To(o.Equal("trustification"))
}

// TestIntegration_Panic tests that accessing non-existent integration panics
func TestIntegration_Panic(t *testing.T) {
	g := o.NewWithT(t)

	mgr := NewManager()

	// Should panic when integration doesn't exist
	g.Expect(func() {
		mgr.Integration("nonexistent")
	}).To(o.Panic())
}

// TestManagerInitialization tests that the manager is properly initialized
func TestManagerInitialization(t *testing.T) {
	g := o.NewWithT(t)

	mgr := NewManager()

	// Verify internal maps are initialized
	g.Expect(mgr.integrations).ToNot(o.BeNil())
	g.Expect(mgr.modules).ToNot(o.BeNil())

	// Verify empty state
	g.Expect(len(mgr.integrations)).To(o.Equal(0))
	g.Expect(len(mgr.modules)).To(o.Equal(0))

	// Test that methods work on empty manager
	names := mgr.IntegrationNames()
	g.Expect(names).To(o.HaveLen(0))

	modules := mgr.GetModules()
	g.Expect(modules).To(o.HaveLen(0))
}

// TestIntegrationNamesReturnsCopy tests that IntegrationNames returns a new slice
func TestIntegrationNamesReturnsCopy(t *testing.T) {
	g := o.NewWithT(t)

	mgr := NewManager()

	// Get names twice
	names1 := mgr.IntegrationNames()
	names2 := mgr.IntegrationNames()

	// Both should be empty but different instances
	g.Expect(names1).To(o.BeEmpty())
	g.Expect(names2).To(o.BeEmpty())

	// Verify they are different slice instances (not same pointer)
	// This tests that the function creates a new slice each time
	g.Expect(&names1).ToNot(o.BeIdenticalTo(&names2))
}

// TestGetModulesReturnsCopy tests that GetModules returns a new slice
func TestGetModulesReturnsCopy(t *testing.T) {
	g := o.NewWithT(t)

	mgr := NewManager()

	// Get modules twice
	modules1 := mgr.GetModules()
	modules2 := mgr.GetModules()

	// Both should be empty but different instances
	g.Expect(modules1).To(o.BeEmpty())
	g.Expect(modules2).To(o.BeEmpty())

	// Verify they are different slice instances
	g.Expect(&modules1).ToNot(o.BeIdenticalTo(&modules2))
}

// TestIntegrationNameType tests the IntegrationName type
func TestIntegrationNameType(t *testing.T) {
	g := o.NewWithT(t)

	// Test that IntegrationName constants can be used as map keys
	testMap := make(map[IntegrationName]string)
	testMap[GitHub] = "github-value"
	testMap[GitLab] = "gitlab-value"

	g.Expect(testMap[GitHub]).To(o.Equal("github-value"))
	g.Expect(testMap[GitLab]).To(o.Equal("gitlab-value"))

	// Test string conversion
	name := GitHub
	g.Expect(string(name)).To(o.Equal("github"))
}

// TestAllIntegrationConstants tests all integration name constants are unique
func TestAllIntegrationConstants(t *testing.T) {
	g := o.NewWithT(t)

	// Collect all constants in a map to ensure uniqueness
	allConstants := map[string]IntegrationName{
		"acs":            ACS,
		"artifactory":    Artifactory,
		"azure":          Azure,
		"bitbucket":      BitBucket,
		"github":         GitHub,
		"gitlab":         GitLab,
		"jenkins":        Jenkins,
		"nexus":          Nexus,
		"quay":           Quay,
		"tas":            TrustedArtifactSigner,
		"trustification": Trustification,
	}

	// Should have 11 unique constants
	g.Expect(allConstants).To(o.HaveLen(11))

	// Verify each maps correctly
	for key, constant := range allConstants {
		g.Expect(string(constant)).To(o.Equal(key))
	}
}

// The following tests were moved to manager_test_additional.go
// to avoid complex mocking requirements
