package config

import (
	"testing"

	o "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

// TestFindNode tests the FindNode function for locating keys in YAML nodes
func TestFindNode(t *testing.T) {
	testCases := []struct {
		name        string
		yamlContent string
		key         string
		shouldFind  bool
		expectValue string
	}{
		{
			name: "simple_key_in_mapping",
			yamlContent: `
name: TestProduct
enabled: true
`,
			key:         "name",
			shouldFind:  true,
			expectValue: "TestProduct",
		},
		{
			name: "nested_key_in_mapping",
			yamlContent: `
products:
  developer-hub:
    name: Developer Hub
    enabled: true
`,
			key:         "developer-hub",
			shouldFind:  true,
			expectValue: "",
		},
		{
			name: "key_not_found",
			yamlContent: `
name: TestProduct
enabled: true
`,
			key:        "nonexistent",
			shouldFind: false,
		},
		{
			name: "key_in_nested_structure",
			yamlContent: `
products:
  product1:
    name: Product1
  product2:
    name: Product2
    config:
      setting: value
`,
			key:        "setting",
			shouldFind: true,
		},
		{
			name:        "empty_document",
			yamlContent: ``,
			key:         "name",
			shouldFind:  false,
		},
		{
			name: "key_in_sequence",
			yamlContent: `
items:
  - name: item1
  - name: item2
`,
			key:        "items",
			shouldFind: true,
		},
		{
			name: "document_node_unwrapping",
			yamlContent: `
products:
  test:
    value: data
`,
			key:        "test",
			shouldFind: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := o.NewWithT(t)

			var node yaml.Node
			err := yaml.Unmarshal([]byte(tc.yamlContent), &node)
			g.Expect(err).To(o.Succeed())

			result, err := FindNode(&node, tc.key)

			if tc.shouldFind {
				g.Expect(err).To(o.Succeed())
				g.Expect(result).ToNot(o.BeNil(), "Expected to find key %q", tc.key)
			} else {
				// Key not found - result should be nil
				g.Expect(result).To(o.BeNil(), "Expected NOT to find key %q", tc.key)
			}
		})
	}
}

// TestFindNode_DocumentNodeUnwrapping tests that DocumentNode is properly unwrapped
func TestFindNode_DocumentNodeUnwrapping(t *testing.T) {
	g := o.NewWithT(t)

	yamlContent := `
name: TestProduct
enabled: true
`

	var node yaml.Node
	err := yaml.Unmarshal([]byte(yamlContent), &node)
	g.Expect(err).To(o.Succeed())

	// The root node should be a DocumentNode
	g.Expect(node.Kind).To(o.Equal(yaml.DocumentNode))

	// FindNode should unwrap it and find the key
	result, err := FindNode(&node, "name")
	g.Expect(err).To(o.Succeed())
	g.Expect(result).ToNot(o.BeNil())
}

// TestFindNode_RecursiveSearch tests recursive searching through nested structures
func TestFindNode_RecursiveSearch(t *testing.T) {
	g := o.NewWithT(t)

	yamlContent := `
level1:
  level2:
    level3:
      target: found
`

	var node yaml.Node
	err := yaml.Unmarshal([]byte(yamlContent), &node)
	g.Expect(err).To(o.Succeed())

	// Should find deeply nested key
	result, err := FindNode(&node, "target")
	g.Expect(err).To(o.Succeed())
	g.Expect(result).ToNot(o.BeNil())
}

// TestUpdateMappingValue tests updating values in mapping nodes
func TestUpdateMappingValue(t *testing.T) {
	t.Run("update_existing_key", func(t *testing.T) {
		g := o.NewWithT(t)

		yamlContent := `
name: OldName
enabled: false
`

		var node yaml.Node
		err := yaml.Unmarshal([]byte(yamlContent), &node)
		g.Expect(err).To(o.Succeed())

		newValue := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "NewName",
		}

		err = UpdateMappingValue(&node, "name", newValue)
		g.Expect(err).To(o.Succeed())

		// Verify the value was updated
		result, err := FindNode(&node, "name")
		g.Expect(err).To(o.Succeed())
		g.Expect(result).ToNot(o.BeNil())
		g.Expect(result.Value).To(o.Equal("NewName"))
	})

	t.Run("key_not_found_no_error", func(t *testing.T) {
		g := o.NewWithT(t)

		yamlContent := `
name: TestProduct
`

		var node yaml.Node
		err := yaml.Unmarshal([]byte(yamlContent), &node)
		g.Expect(err).To(o.Succeed())

		newValue := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "true",
		}

		// UpdateMappingValue returns nil when key doesn't exist (doesn't create it)
		err = UpdateMappingValue(&node, "enabled", newValue)
		g.Expect(err).To(o.Succeed())

		// Verify the key was NOT created
		result, _ := FindNode(&node, "enabled")
		g.Expect(result).To(o.BeNil())
	})

	t.Run("create_root_mapping_on_empty_document", func(t *testing.T) {
		g := o.NewWithT(t)

		// Create an empty DocumentNode
		node := yaml.Node{
			Kind: yaml.DocumentNode,
		}

		newValue := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "TestValue",
		}

		err := UpdateMappingValue(&node, "testkey", newValue)
		g.Expect(err).To(o.Succeed())

		// Verify root mapping was created inside the document
		g.Expect(node.Content).To(o.HaveLen(1))
		g.Expect(node.Content[0].Kind).To(o.Equal(yaml.MappingNode))
	})

	t.Run("preserve_anchor_on_update", func(t *testing.T) {
		g := o.NewWithT(t)

		yamlContent := `
name: &anchor TestProduct
reference: *anchor
`

		var node yaml.Node
		err := yaml.Unmarshal([]byte(yamlContent), &node)
		g.Expect(err).To(o.Succeed())

		newValue := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "NewProduct",
		}

		err = UpdateMappingValue(&node, "name", newValue)
		g.Expect(err).To(o.Succeed())
	})

	t.Run("unsupported_node_kind", func(t *testing.T) {
		g := o.NewWithT(t)

		// Create a ScalarNode (unsupported for UpdateMappingValue)
		node := yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "test",
		}

		newValue := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "updated",
		}

		err := UpdateMappingValue(&node, "name", newValue)
		g.Expect(err).To(o.HaveOccurred())
		g.Expect(err.Error()).To(o.ContainSubstring("cannot set value on node kind"))
	})
}

// TestUpdateNestedValue tests updating nested values in YAML structures
func TestUpdateNestedValue(t *testing.T) {
	t.Run("single_key_path", func(t *testing.T) {
		g := o.NewWithT(t)

		yamlContent := `
name: OldName
enabled: false
`

		var node yaml.Node
		err := yaml.Unmarshal([]byte(yamlContent), &node)
		g.Expect(err).To(o.Succeed())

		newValue := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "NewName",
		}

		err = UpdateNestedValue(&node, []string{"name"}, newValue)
		g.Expect(err).To(o.Succeed())

		result, err := FindNode(&node, "name")
		g.Expect(err).To(o.Succeed())
		g.Expect(result).ToNot(o.BeNil())
		g.Expect(result.Value).To(o.Equal("NewName"))
	})

	t.Run("multi_key_path", func(t *testing.T) {
		g := o.NewWithT(t)

		yamlContent := `
products:
  developer-hub:
    name: Developer Hub
    enabled: false
`

		var node yaml.Node
		err := yaml.Unmarshal([]byte(yamlContent), &node)
		g.Expect(err).To(o.Succeed())

		newValue := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "true",
		}

		err = UpdateNestedValue(&node, []string{"products", "developer-hub", "enabled"}, newValue)
		g.Expect(err).To(o.Succeed())
	})

	t.Run("key_not_found_in_path", func(t *testing.T) {
		g := o.NewWithT(t)

		yamlContent := `
products:
  existing-product:
    name: Existing
`

		var node yaml.Node
		err := yaml.Unmarshal([]byte(yamlContent), &node)
		g.Expect(err).To(o.Succeed())

		newValue := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "NewProduct",
		}

		// UpdateNestedValue will return an error if a key in the path doesn't exist
		err = UpdateNestedValue(&node, []string{"products", "new-product", "name"}, newValue)
		g.Expect(err).To(o.HaveOccurred())
		g.Expect(err.Error()).To(o.ContainSubstring("key not found"))
	})

	t.Run("empty_path", func(t *testing.T) {
		g := o.NewWithT(t)

		yamlContent := `
name: TestProduct
`

		var node yaml.Node
		err := yaml.Unmarshal([]byte(yamlContent), &node)
		g.Expect(err).To(o.Succeed())

		newValue := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "NewValue",
		}

		err = UpdateNestedValue(&node, []string{}, newValue)
		g.Expect(err).To(o.HaveOccurred())
		g.Expect(err.Error()).To(o.ContainSubstring("path is missing"))
	})

	t.Run("invalid_node_kind", func(t *testing.T) {
		g := o.NewWithT(t)

		// Create a ScalarNode
		node := yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "test",
		}

		newValue := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "updated",
		}

		err := UpdateNestedValue(&node, []string{"key1", "key2"}, newValue)
		g.Expect(err).To(o.HaveOccurred())
		g.Expect(err.Error()).To(o.ContainSubstring("cannot navigate through node kind"))
	})

	t.Run("sequence_node_in_path", func(t *testing.T) {
		g := o.NewWithT(t)

		yamlContent := `
items:
  - name: item1
    value: val1
  - name: item2
    value: val2
`

		var node yaml.Node
		err := yaml.Unmarshal([]byte(yamlContent), &node)
		g.Expect(err).To(o.Succeed())

		newValue := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "updated_val2",
		}

		// UpdateNestedValue supports sequence nodes with integer indices
		err = UpdateNestedValue(&node, []string{"items", "1", "value"}, newValue)
		g.Expect(err).To(o.Succeed())
	})

	t.Run("deep_nested_path", func(t *testing.T) {
		g := o.NewWithT(t)

		yamlContent := `
level1:
  level2:
    level3:
      level4:
        value: original
`

		var node yaml.Node
		err := yaml.Unmarshal([]byte(yamlContent), &node)
		g.Expect(err).To(o.Succeed())

		newValue := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "updated",
		}

		err = UpdateNestedValue(&node, []string{"level1", "level2", "level3", "level4", "value"}, newValue)
		g.Expect(err).To(o.Succeed())
	})
}

// TestUpdateNestedValue_EdgeCases tests edge cases for UpdateNestedValue
func TestUpdateNestedValue_EdgeCases(t *testing.T) {
	t.Run("document_node_unwrapping", func(t *testing.T) {
		g := o.NewWithT(t)

		yamlContent := `
name: TestProduct
`

		var node yaml.Node
		err := yaml.Unmarshal([]byte(yamlContent), &node)
		g.Expect(err).To(o.Succeed())

		// Ensure we have a DocumentNode
		g.Expect(node.Kind).To(o.Equal(yaml.DocumentNode))

		newValue := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "NewProduct",
		}

		// Should handle DocumentNode unwrapping
		err = UpdateNestedValue(&node, []string{"name"}, newValue)
		g.Expect(err).To(o.Succeed())
	})
}

// TestFindNode_SequenceNodes tests FindNode behavior with sequence nodes
func TestFindNode_SequenceNodes(t *testing.T) {
	g := o.NewWithT(t)

	yamlContent := `
products:
  - name: Product1
    enabled: true
  - name: Product2
    enabled: false
`

	var node yaml.Node
	err := yaml.Unmarshal([]byte(yamlContent), &node)
	g.Expect(err).To(o.Succeed())

	// Should find the products key (which contains a sequence)
	result, err := FindNode(&node, "products")
	g.Expect(err).To(o.Succeed())
	g.Expect(result).ToNot(o.BeNil())
	g.Expect(result.Kind).To(o.Equal(yaml.SequenceNode))
}

// TestUpdateMappingValue_ComplexStructures tests with complex YAML structures
func TestUpdateMappingValue_ComplexStructures(t *testing.T) {
	g := o.NewWithT(t)

	yamlContent := `
products:
  developer-hub:
    name: Developer Hub
    enabled: true
    properties:
      key1: value1
      key2: value2
`

	var node yaml.Node
	err := yaml.Unmarshal([]byte(yamlContent), &node)
	g.Expect(err).To(o.Succeed())

	// Find the products node
	productsNode, err := FindNode(&node, "products")
	g.Expect(err).To(o.Succeed())
	g.Expect(productsNode).ToNot(o.BeNil())

	newValue := &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "name"},
			{Kind: yaml.ScalarNode, Value: "New Product"},
		},
	}

	err = UpdateMappingValue(productsNode, "new-product", newValue)
	g.Expect(err).To(o.Succeed())

	// UpdateMappingValue doesn't create new keys, so this key won't exist
	result, _ := FindNode(&node, "new-product")
	g.Expect(result).To(o.BeNil())
}
