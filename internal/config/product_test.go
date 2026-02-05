package config

import (
	"testing"

	o "github.com/onsi/gomega"
)

// TestProductKeyName tests the KeyName sanitization function
func TestProductKeyName(t *testing.T) {
	testCases := []struct {
		name     string
		product  Product
		expected string
	}{
		{
			name:     "simple_name",
			product:  Product{Name: "SimpleProduct"},
			expected: "SimpleProduct",
		},
		{
			name:     "name_with_spaces",
			product:  Product{Name: "Developer Hub"},
			expected: "Developer_Hub",
		},
		{
			name:     "name_with_hyphens",
			product:  Product{Name: "openshift-pipelines"},
			expected: "openshift_pipelines",
		},
		{
			name:     "name_with_special_chars",
			product:  Product{Name: "Product@#$%Name!"},
			expected: "Product_Name",
		},
		{
			name:     "name_starting_with_digit",
			product:  Product{Name: "123Product"},
			expected: "_123Product",
		},
		{
			name:     "name_with_multiple_spaces",
			product:  Product{Name: "Red   Hat   Developer   Hub"},
			expected: "Red_Hat_Developer_Hub",
		},
		{
			name:     "name_with_underscores",
			product:  Product{Name: "my_product_name"},
			expected: "my_product_name",
		},
		{
			name:     "name_with_leading_trailing_special",
			product:  Product{Name: "!!!Product!!!"},
			expected: "Product",
		},
		{
			name:     "name_with_mixed_special_chars",
			product:  Product{Name: "Product-Name.v2"},
			expected: "Product_Name_v2",
		},
		{
			name:     "empty_name",
			product:  Product{Name: ""},
			expected: "",
		},
		{
			name:     "only_special_chars",
			product:  Product{Name: "!!!@@@###"},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := o.NewWithT(t)
			result := tc.product.KeyName()
			g.Expect(result).To(o.Equal(tc.expected))
		})
	}
}

// TestProductGetNamespace tests the GetNamespace function
func TestProductGetNamespace(t *testing.T) {
	t.Run("namespace_set", func(t *testing.T) {
		g := o.NewWithT(t)

		namespace := "my-namespace"
		product := Product{
			Name:      "TestProduct",
			Namespace: &namespace,
		}

		result := product.GetNamespace()
		g.Expect(result).To(o.Equal("my-namespace"))
	})

	t.Run("namespace_nil", func(t *testing.T) {
		g := o.NewWithT(t)

		product := Product{
			Name:      "TestProduct",
			Namespace: nil,
		}

		result := product.GetNamespace()
		g.Expect(result).To(o.BeEmpty())
	})

	t.Run("namespace_empty_string", func(t *testing.T) {
		g := o.NewWithT(t)

		namespace := ""
		product := Product{
			Name:      "TestProduct",
			Namespace: &namespace,
		}

		result := product.GetNamespace()
		g.Expect(result).To(o.BeEmpty())
	})
}

// TestProductValidate tests the Validate function
func TestProductValidate(t *testing.T) {
	t.Run("enabled_with_namespace", func(t *testing.T) {
		g := o.NewWithT(t)

		namespace := "test-namespace"
		product := Product{
			Name:      "TestProduct",
			Enabled:   true,
			Namespace: &namespace,
		}

		err := product.Validate()
		g.Expect(err).To(o.Succeed())
	})

	t.Run("enabled_without_namespace", func(t *testing.T) {
		g := o.NewWithT(t)

		product := Product{
			Name:      "TestProduct",
			Enabled:   true,
			Namespace: nil,
		}

		err := product.Validate()
		g.Expect(err).To(o.HaveOccurred())
		g.Expect(err).To(o.MatchError(o.ContainSubstring("missing namespace")))
		g.Expect(err).To(o.MatchError(o.ContainSubstring("TestProduct")))
	})

	t.Run("disabled_without_namespace", func(t *testing.T) {
		g := o.NewWithT(t)

		product := Product{
			Name:      "TestProduct",
			Enabled:   false,
			Namespace: nil,
		}

		err := product.Validate()
		g.Expect(err).To(o.Succeed())
	})

	t.Run("enabled_with_empty_namespace", func(t *testing.T) {
		g := o.NewWithT(t)

		namespace := ""
		product := Product{
			Name:      "TestProduct",
			Enabled:   true,
			Namespace: &namespace,
		}

		err := product.Validate()
		g.Expect(err).To(o.HaveOccurred())
		g.Expect(err).To(o.MatchError(o.ContainSubstring("missing namespace")))
	})
}

// TestProductConstants tests the product name constants
func TestProductConstants(t *testing.T) {
	g := o.NewWithT(t)

	g.Expect(DeveloperHub).To(o.Equal("Developer Hub"))
	g.Expect(OpenShiftPipelines).To(o.Equal("OpenShift Pipelines"))
}

// TestProductStruct tests the Product struct fields
func TestProductStruct(t *testing.T) {
	g := o.NewWithT(t)

	namespace := "test-ns"
	properties := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	product := Product{
		Name:       "MyProduct",
		Enabled:    true,
		Namespace:  &namespace,
		Properties: properties,
	}

	g.Expect(product.Name).To(o.Equal("MyProduct"))
	g.Expect(product.Enabled).To(o.BeTrue())
	g.Expect(product.Namespace).ToNot(o.BeNil())
	g.Expect(*product.Namespace).To(o.Equal("test-ns"))
	g.Expect(product.Properties).To(o.HaveLen(2))
	g.Expect(product.Properties["key1"]).To(o.Equal("value1"))
	g.Expect(product.Properties["key2"]).To(o.Equal(42))
}
