package api

import (
	"testing"

	o "github.com/onsi/gomega"
)

// TestNewAppContext tests the constructor with default values
func TestNewAppContext(t *testing.T) {
	g := o.NewWithT(t)

	appCtx := NewAppContext("test-app")
	g.Expect(appCtx).ToNot(o.BeNil())
	g.Expect(appCtx.Name).To(o.Equal("test-app"))
	g.Expect(appCtx.Namespace).To(o.Equal("test-app")) // defaults to name
	g.Expect(appCtx.Version).To(o.Equal("v0.0.0-SNAPSHOT"))
	g.Expect(appCtx.CommitID).To(o.Equal("unknown"))
	g.Expect(appCtx.Short).To(o.Equal(""))
	g.Expect(appCtx.Long).To(o.Equal(""))
}

// TestWithNamespace tests the WithNamespace functional option
func TestWithNamespace(t *testing.T) {
	g := o.NewWithT(t)

	appCtx := NewAppContext("test-app", WithNamespace("custom-namespace"))
	g.Expect(appCtx.Namespace).To(o.Equal("custom-namespace"))
	g.Expect(appCtx.Name).To(o.Equal("test-app")) // name unchanged
}

// TestWithVersion tests the WithVersion functional option
func TestWithVersion(t *testing.T) {
	g := o.NewWithT(t)

	appCtx := NewAppContext("test-app", WithVersion("v1.2.3"))
	g.Expect(appCtx.Version).To(o.Equal("v1.2.3"))
}

// TestWithCommitID tests the WithCommitID functional option
func TestWithCommitID(t *testing.T) {
	g := o.NewWithT(t)

	appCtx := NewAppContext("test-app", WithCommitID("abc123def"))
	g.Expect(appCtx.CommitID).To(o.Equal("abc123def"))
}

// TestWithShortDescription tests the WithShortDescription functional option
func TestWithShortDescription(t *testing.T) {
	g := o.NewWithT(t)

	appCtx := NewAppContext("test-app", WithShortDescription("A short description"))
	g.Expect(appCtx.Short).To(o.Equal("A short description"))
}

// TestWithLongDescription tests the WithLongDescription functional option
func TestWithLongDescription(t *testing.T) {
	g := o.NewWithT(t)

	longDesc := "This is a much longer description that provides more detail about the application."
	appCtx := NewAppContext("test-app", WithLongDescription(longDesc))
	g.Expect(appCtx.Long).To(o.Equal(longDesc))
}

// TestMultipleOptions tests using multiple functional options together
func TestMultipleOptions(t *testing.T) {
	g := o.NewWithT(t)

	appCtx := NewAppContext("test-app",
		WithNamespace("prod-namespace"),
		WithVersion("v2.0.0"),
		WithCommitID("sha256abc"),
		WithShortDescription("Production app"),
		WithLongDescription("This is the production version of the application"),
	)

	g.Expect(appCtx.Name).To(o.Equal("test-app"))
	g.Expect(appCtx.Namespace).To(o.Equal("prod-namespace"))
	g.Expect(appCtx.Version).To(o.Equal("v2.0.0"))
	g.Expect(appCtx.CommitID).To(o.Equal("sha256abc"))
	g.Expect(appCtx.Short).To(o.Equal("Production app"))
	g.Expect(appCtx.Long).To(o.Equal("This is the production version of the application"))
}

// TestOptionOrderIndependence tests that option order doesn't matter
func TestOptionOrderIndependence(t *testing.T) {
	g := o.NewWithT(t)

	// Apply options in one order
	appCtx1 := NewAppContext("test-app",
		WithVersion("v1.0.0"),
		WithNamespace("ns1"),
	)

	// Apply same options in different order
	appCtx2 := NewAppContext("test-app",
		WithNamespace("ns1"),
		WithVersion("v1.0.0"),
	)

	g.Expect(appCtx1.Version).To(o.Equal(appCtx2.Version))
	g.Expect(appCtx1.Namespace).To(o.Equal(appCtx2.Namespace))
}

// TestOptionOverride tests that later options override earlier ones
func TestOptionOverride(t *testing.T) {
	g := o.NewWithT(t)

	// Apply same option twice - last one should win
	appCtx := NewAppContext("test-app",
		WithVersion("v1.0.0"),
		WithVersion("v2.0.0"), // This should override the first
	)

	g.Expect(appCtx.Version).To(o.Equal("v2.0.0"))
}
