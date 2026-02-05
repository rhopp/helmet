package subcmd

import (
	"testing"

	"github.com/redhat-appstudio/helmet/internal/integrations"

	o "github.com/onsi/gomega"
)

// TestStandardModules tests the StandardModules function
func TestStandardModules(t *testing.T) {
	g := o.NewWithT(t)

	modules := StandardModules()

	// Should return exactly 11 modules
	g.Expect(modules).To(o.HaveLen(11))

	// Verify module names are set correctly
	moduleNames := make([]string, len(modules))
	for i, m := range modules {
		moduleNames[i] = m.Name
	}

	expectedNames := []string{
		string(integrations.ACS),
		string(integrations.Artifactory),
		string(integrations.Azure),
		string(integrations.BitBucket),
		string(integrations.GitHub),
		string(integrations.GitLab),
		string(integrations.Jenkins),
		string(integrations.Nexus),
		string(integrations.Quay),
		string(integrations.TrustedArtifactSigner),
		string(integrations.Trustification),
	}

	g.Expect(moduleNames).To(o.Equal(expectedNames))
}

// TestACSModule tests the ACS module configuration
func TestACSModule(t *testing.T) {
	g := o.NewWithT(t)

	g.Expect(ACSModule.Name).To(o.Equal(string(integrations.ACS)))
	g.Expect(ACSModule.Init).ToNot(o.BeNil())
	g.Expect(ACSModule.Command).ToNot(o.BeNil())
}

// TestArtifactoryModule tests the Artifactory module configuration
func TestArtifactoryModule(t *testing.T) {
	g := o.NewWithT(t)

	g.Expect(ArtifactoryModule.Name).To(o.Equal(string(integrations.Artifactory)))
	g.Expect(ArtifactoryModule.Init).ToNot(o.BeNil())
	g.Expect(ArtifactoryModule.Command).ToNot(o.BeNil())
}

// TestAzureModule tests the Azure module configuration
func TestAzureModule(t *testing.T) {
	g := o.NewWithT(t)

	g.Expect(AzureModule.Name).To(o.Equal(string(integrations.Azure)))
	g.Expect(AzureModule.Init).ToNot(o.BeNil())
	g.Expect(AzureModule.Command).ToNot(o.BeNil())
}

// TestBitBucketModule tests the BitBucket module configuration
func TestBitBucketModule(t *testing.T) {
	g := o.NewWithT(t)

	g.Expect(BitBucketModule.Name).To(o.Equal(string(integrations.BitBucket)))
	g.Expect(BitBucketModule.Init).ToNot(o.BeNil())
	g.Expect(BitBucketModule.Command).ToNot(o.BeNil())
}

// TestGitHubModule tests the GitHub module configuration
func TestGitHubModule(t *testing.T) {
	g := o.NewWithT(t)

	g.Expect(GitHubModule.Name).To(o.Equal(string(integrations.GitHub)))
	g.Expect(GitHubModule.Init).ToNot(o.BeNil())
	g.Expect(GitHubModule.Command).ToNot(o.BeNil())
}

// TestGitLabModule tests the GitLab module configuration
func TestGitLabModule(t *testing.T) {
	g := o.NewWithT(t)

	g.Expect(GitLabModule.Name).To(o.Equal(string(integrations.GitLab)))
	g.Expect(GitLabModule.Init).ToNot(o.BeNil())
	g.Expect(GitLabModule.Command).ToNot(o.BeNil())
}

// TestJenkinsModule tests the Jenkins module configuration
func TestJenkinsModule(t *testing.T) {
	g := o.NewWithT(t)

	g.Expect(JenkinsModule.Name).To(o.Equal(string(integrations.Jenkins)))
	g.Expect(JenkinsModule.Init).ToNot(o.BeNil())
	g.Expect(JenkinsModule.Command).ToNot(o.BeNil())
}

// TestNexusModule tests the Nexus module configuration
func TestNexusModule(t *testing.T) {
	g := o.NewWithT(t)

	g.Expect(NexusModule.Name).To(o.Equal(string(integrations.Nexus)))
	g.Expect(NexusModule.Init).ToNot(o.BeNil())
	g.Expect(NexusModule.Command).ToNot(o.BeNil())
}

// TestQuayModule tests the Quay module configuration
func TestQuayModule(t *testing.T) {
	g := o.NewWithT(t)

	g.Expect(QuayModule.Name).To(o.Equal(string(integrations.Quay)))
	g.Expect(QuayModule.Init).ToNot(o.BeNil())
	g.Expect(QuayModule.Command).ToNot(o.BeNil())
}

// TestTrustedArtifactSignerModule tests the TrustedArtifactSigner module configuration
func TestTrustedArtifactSignerModule(t *testing.T) {
	g := o.NewWithT(t)

	g.Expect(TrustedArtifactSignerModule.Name).To(o.Equal(string(integrations.TrustedArtifactSigner)))
	g.Expect(TrustedArtifactSignerModule.Init).ToNot(o.BeNil())
	g.Expect(TrustedArtifactSignerModule.Command).ToNot(o.BeNil())
}

// TestTrustificationModule tests the Trustification module configuration
func TestTrustificationModule(t *testing.T) {
	g := o.NewWithT(t)

	g.Expect(TrustificationModule.Name).To(o.Equal(string(integrations.Trustification)))
	g.Expect(TrustificationModule.Init).ToNot(o.BeNil())
	g.Expect(TrustificationModule.Command).ToNot(o.BeNil())
}

// TestAllModulesHaveUniqueNames ensures no duplicate module names
func TestAllModulesHaveUniqueNames(t *testing.T) {
	g := o.NewWithT(t)

	modules := StandardModules()
	nameMap := make(map[string]bool)

	for _, m := range modules {
		_, exists := nameMap[m.Name]
		g.Expect(exists).To(o.BeFalse(), "duplicate module name: %s", m.Name)
		nameMap[m.Name] = true
	}
}

// TestAllModulesHaveValidCallbacks ensures all modules have non-nil Init and Command functions
func TestAllModulesHaveValidCallbacks(t *testing.T) {
	g := o.NewWithT(t)

	modules := StandardModules()

	for _, m := range modules {
		g.Expect(m.Init).ToNot(o.BeNil(), "module %s has nil Init function", m.Name)
		g.Expect(m.Command).ToNot(o.BeNil(), "module %s has nil Command function", m.Name)
	}
}
