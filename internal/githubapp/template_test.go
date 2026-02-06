package githubapp

import (
	"strings"
	"testing"

	o "github.com/onsi/gomega"
)

// TestGitHubNewAppForTmpl tests the HTML template for creating a new app
func TestGitHubNewAppForTmpl(t *testing.T) {
	g := o.NewWithT(t)

	g.Expect(gitHubNewAppForTmpl).ToNot(o.BeEmpty())
	g.Expect(gitHubNewAppForTmpl).To(o.ContainSubstring("<html>"))
	g.Expect(gitHubNewAppForTmpl).To(o.ContainSubstring("<form method=\"post\""))
	g.Expect(gitHubNewAppForTmpl).To(o.ContainSubstring("action=\"%s/settings/apps/new\""))
	g.Expect(gitHubNewAppForTmpl).To(o.ContainSubstring("<input type=\"submit\""))
	g.Expect(gitHubNewAppForTmpl).To(o.ContainSubstring("Create your GitHub App"))
	g.Expect(gitHubNewAppForTmpl).To(o.ContainSubstring("<input type=\"hidden\" name=\"manifest\""))
	g.Expect(gitHubNewAppForTmpl).To(o.ContainSubstring("value='%s'"))
}

// TestGitHubAppSuccessfullyCreatedTmpl tests the success message template
func TestGitHubAppSuccessfullyCreatedTmpl(t *testing.T) {
	g := o.NewWithT(t)

	g.Expect(gitHubAppSuccessfullyCreatedTmpl).ToNot(o.BeEmpty())
	g.Expect(gitHubAppSuccessfullyCreatedTmpl).To(o.ContainSubstring("<html>"))
	g.Expect(gitHubAppSuccessfullyCreatedTmpl).To(o.ContainSubstring("successfully created"))
	g.Expect(gitHubAppSuccessfullyCreatedTmpl).To(o.ContainSubstring("Install the GitHub App"))
	g.Expect(gitHubAppSuccessfullyCreatedTmpl).To(o.ContainSubstring("<form method=\"get\""))
	g.Expect(gitHubAppSuccessfullyCreatedTmpl).To(o.ContainSubstring("action=\"%s\""))
}

// TestTemplateVariableCount tests that templates have the right number of format variables
func TestTemplateVariableCount(t *testing.T) {
	g := o.NewWithT(t)

	// gitHubNewAppForTmpl should have 2 %s placeholders (URL and manifest)
	count := strings.Count(gitHubNewAppForTmpl, "%s")
	g.Expect(count).To(o.Equal(2))

	// gitHubAppSuccessfullyCreatedTmpl should have 1 %s placeholder (URL)
	count = strings.Count(gitHubAppSuccessfullyCreatedTmpl, "%s")
	g.Expect(count).To(o.Equal(1))
}
