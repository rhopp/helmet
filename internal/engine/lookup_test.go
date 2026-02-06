package engine

import (
	"testing"

	"github.com/redhat-appstudio/helmet/internal/k8s"

	o "github.com/onsi/gomega"
)

// TestLookupFuncsLookupMethod tests the Lookup method
func TestLookupFuncsLookupMethod(t *testing.T) {
	g := o.NewWithT(t)

	// Create a minimal k8s.Kube instance
	// Note: We can't fully test lookup() without a real cluster or complex mocking
	// but we can test the method structure
	kube := &k8s.Kube{}
	lookupFuncs := NewLookupFuncs(kube)

	// Test that Lookup() returns a function
	fn := lookupFuncs.Lookup()
	g.Expect(fn).ToNot(o.BeNil())

	// Verify the function has the correct signature
	// by checking it can be assigned to LookupFn type
	var lookupFn LookupFn
	lookupFn = fn
	g.Expect(lookupFn).ToNot(o.BeNil())
}

// TestNewLookupFuncs tests the constructor
func TestNewLookupFuncs(t *testing.T) {
	g := o.NewWithT(t)

	kube := &k8s.Kube{}
	lookupFuncs := NewLookupFuncs(kube)

	g.Expect(lookupFuncs).ToNot(o.BeNil())
	g.Expect(lookupFuncs.kube).To(o.Equal(kube))
}

// TestLookupFuncsStructure tests the LookupFuncs struct
func TestLookupFuncsStructure(t *testing.T) {
	g := o.NewWithT(t)

	kube := &k8s.Kube{}
	lookupFuncs := &LookupFuncs{kube: kube}

	g.Expect(lookupFuncs.kube).To(o.Equal(kube))

	// Test Lookup method returns the internal lookup function
	fn := lookupFuncs.Lookup()
	g.Expect(fn).ToNot(o.BeNil())
}
