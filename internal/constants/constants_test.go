package constants

import (
	"testing"

	o "github.com/onsi/gomega"
)

// TestConfigFilename tests that ConfigFilename constant has expected value
func TestConfigFilename(t *testing.T) {
	g := o.NewWithT(t)
	g.Expect(ConfigFilename).To(o.Equal("config.yaml"))
}

// TestValuesFilename tests that ValuesFilename constant has expected value
func TestValuesFilename(t *testing.T) {
	g := o.NewWithT(t)
	g.Expect(ValuesFilename).To(o.Equal("values.yaml.tpl"))
}

// TestInstructionsFilename tests that InstructionsFilename constant has expected value
func TestInstructionsFilename(t *testing.T) {
	g := o.NewWithT(t)
	g.Expect(InstructionsFilename).To(o.Equal("instructions.md"))
}

// TestConstantsNotEmpty tests that all constants are non-empty
func TestConstantsNotEmpty(t *testing.T) {
	g := o.NewWithT(t)
	g.Expect(ConfigFilename).ToNot(o.BeEmpty())
	g.Expect(ValuesFilename).ToNot(o.BeEmpty())
	g.Expect(InstructionsFilename).ToNot(o.BeEmpty())
}

// TestConstantsUnique tests that all constants have unique values
func TestConstantsUnique(t *testing.T) {
	g := o.NewWithT(t)

	constants := map[string]bool{
		ConfigFilename:       false,
		ValuesFilename:       false,
		InstructionsFilename: false,
	}

	// Check no duplicates
	g.Expect(len(constants)).To(o.Equal(3))
}
