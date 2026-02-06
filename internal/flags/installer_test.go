package flags

import (
	"testing"

	"github.com/redhat-appstudio/helmet/internal/constants"
	"github.com/spf13/pflag"

	o "github.com/onsi/gomega"
)

// TestValuesTemplateFlagConstant tests the constant value
func TestValuesTemplateFlagConstant(t *testing.T) {
	g := o.NewWithT(t)
	g.Expect(ValuesTemplateFlag).To(o.Equal("values-template"))
}

// TestSetValuesTmplFlag tests setting the values-template flag
func TestSetValuesTmplFlag(t *testing.T) {
	g := o.NewWithT(t)

	var flagValue string
	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)

	SetValuesTmplFlag(flagSet, &flagValue)

	// Check flag was registered
	flag := flagSet.Lookup(ValuesTemplateFlag)
	g.Expect(flag).ToNot(o.BeNil())
	g.Expect(flag.Name).To(o.Equal("values-template"))
	g.Expect(flag.Usage).To(o.Equal("Path to the values template file"))

	// Check default value
	g.Expect(flag.DefValue).To(o.Equal(constants.ValuesFilename))
	g.Expect(flagValue).To(o.Equal(constants.ValuesFilename))
}

// TestSetValuesTmplFlagWithCustomValue tests setting custom value
func TestSetValuesTmplFlagWithCustomValue(t *testing.T) {
	g := o.NewWithT(t)

	var flagValue string
	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)

	SetValuesTmplFlag(flagSet, &flagValue)

	// Set custom value
	err := flagSet.Set(ValuesTemplateFlag, "custom-values.yaml")
	g.Expect(err).To(o.BeNil())
	g.Expect(flagValue).To(o.Equal("custom-values.yaml"))
}

// TestSetValuesTmplFlagMultipleFlagSets tests using function with multiple flag sets
func TestSetValuesTmplFlagMultipleFlagSets(t *testing.T) {
	g := o.NewWithT(t)

	var flagValue1 string
	var flagValue2 string

	flagSet1 := pflag.NewFlagSet("test1", pflag.ContinueOnError)
	flagSet2 := pflag.NewFlagSet("test2", pflag.ContinueOnError)

	SetValuesTmplFlag(flagSet1, &flagValue1)
	SetValuesTmplFlag(flagSet2, &flagValue2)

	// Both should have default value
	g.Expect(flagValue1).To(o.Equal(constants.ValuesFilename))
	g.Expect(flagValue2).To(o.Equal(constants.ValuesFilename))

	// Setting one should not affect the other
	err := flagSet1.Set(ValuesTemplateFlag, "custom1.yaml")
	g.Expect(err).To(o.BeNil())
	g.Expect(flagValue1).To(o.Equal("custom1.yaml"))
	g.Expect(flagValue2).To(o.Equal(constants.ValuesFilename))
}
