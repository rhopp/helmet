package resolver

import (
	"testing"

	o "github.com/onsi/gomega"
)

// TestIntegrationsErrorConstants tests error constant definitions
func TestIntegrationsErrorConstants(t *testing.T) {
	g := o.NewWithT(t)

	g.Expect(ErrUnknownIntegration).ToNot(o.BeNil())
	g.Expect(ErrUnknownIntegration.Error()).To(o.Equal("unknown integration"))

	g.Expect(ErrPrerequisiteIntegration).ToNot(o.BeNil())
	g.Expect(ErrPrerequisiteIntegration.Error()).To(o.Equal("dependency prerequisite integration(s) missing"))
}

// TestIntegrationsStruct tests the Integrations struct
func TestIntegrationsStruct(t *testing.T) {
	g := o.NewWithT(t)

	integs := &Integrations{
		configured: map[string]bool{
			"github": true,
			"gitlab": false,
		},
	}

	g.Expect(integs.configured).To(o.HaveLen(2))
	g.Expect(integs.configured["github"]).To(o.BeTrue())
	g.Expect(integs.configured["gitlab"]).To(o.BeFalse())
}
