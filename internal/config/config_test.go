package config

import (
	"os"
	"testing"

	"github.com/redhat-appstudio/helmet/internal/chartfs"

	o "github.com/onsi/gomega"
)

func TestNewConfigFromFile(t *testing.T) {
	g := o.NewWithT(t)

	cfs := chartfs.New(os.DirFS("../../test"))

	cfg, err := NewConfigFromFile(cfs, "config.yaml", "test-namespace")
	g.Expect(err).To(o.Succeed())
	g.Expect(cfg).NotTo(o.BeNil())
	g.Expect(cfg.Installer).NotTo(o.BeNil())

	t.Run("Validate", func(t *testing.T) {
		err := cfg.Validate()
		g.Expect(err).To(o.Succeed())
	})

	t.Run("GetEnabledProducts", func(t *testing.T) {
		products := cfg.GetEnabledProducts()
		g.Expect(products).NotTo(o.BeEmpty())
		g.Expect(len(products)).To(o.BeNumerically(">", 1))
	})

	t.Run("GetProduct", func(t *testing.T) {
		_, err := cfg.GetProduct("product1")
		g.Expect(err).NotTo(o.Succeed())

		product, err := cfg.GetProduct("Product A")
		g.Expect(err).To(o.Succeed())
		g.Expect(product).NotTo(o.BeNil())
		g.Expect(product.GetNamespace()).NotTo(o.BeEmpty())
	})

	t.Run("MarshalYAML and UnmarshalYAML", func(t *testing.T) {
		payload, err := cfg.MarshalYAML()
		g.Expect(err).To(o.Succeed())
		g.Expect(string(payload)).To(o.ContainSubstring("tssc:"))

		err = cfg.UnmarshalYAML(payload)
		g.Expect(err).To(o.Succeed())
	})

	t.Run("DecodeNode", func(t *testing.T) {
		err := cfg.DecodeNode()
		g.Expect(err).To(o.Succeed())
		g.Expect(cfg.Installer).NotTo(o.BeNil())
	})

	t.Run("String", func(t *testing.T) {
		original, err := cfs.ReadFile("config.yaml")
		g.Expect(err).To(o.Succeed())

		configString := cfg.String()
		g.Expect(err).To(o.Succeed())
		g.Expect(configString).To(o.ContainSubstring("tssc:"))

		// Asserting the original configuration looks like the marshaled one.
		g.Expect(string(original)).To(o.Equal(configString))
	})

	t.Run("SetSettings", func(t *testing.T) {
		data := map[string]interface{}{
			"crc": true,
			"ci": map[string]interface{}{
				"debug": true,
			},
		}
		err := cfg.Set("tssc.settings", data)
		g.Expect(err).To(o.Succeed())
		configString := cfg.String()
		g.Expect(configString).To(o.ContainSubstring("crc: true"))
		g.Expect(configString).To(o.ContainSubstring("debug: true"))
	})

	t.Run("SetProducts", func(t *testing.T) {
		// Product A is product 0
		err := cfg.Set("tssc.products.0.namespace", "productAtest")
		g.Expect(err).To(o.Succeed())

		// Product 2
		err = cfg.Set("tssc.products.2.enabled", false)
		g.Expect(err).To(o.Succeed())

		// Product D is product 3
		dData := map[string]any{
			"catalogURL":   "https://someIP.io",
			"authProvider": "gitlab",
		}
		err = cfg.Set("tssc.products.3.properties", dData)
		g.Expect(err).To(o.Succeed())

		configString := cfg.String()
		g.Expect(configString).To(o.ContainSubstring("namespace: productAtest"))
		g.Expect(configString).To(o.ContainSubstring("enabled: false"))
		g.Expect(configString).To(o.ContainSubstring("catalogURL: https://someIP.io"))
		g.Expect(configString).To(o.ContainSubstring("authProvider: gitlab"))
	})

	t.Run("FlattenMap", func(t *testing.T) {
		data := map[string]interface{}{
			"key1": "value1",
			"key2": map[string]interface{}{
				"key3": "value2",
			},
		}
		expectedKeyValues := map[string]interface{}{
			"prefix.key1":      "value1",
			"prefix.key2.key3": "value2",
		}
		keyPaths, err := FlattenMap(data, "prefix")
		g.Expect(err).To(o.Succeed())
		g.Expect(keyPaths).To(o.HaveLen(len(expectedKeyValues)))
		for key, value := range keyPaths {
			g.Expect(value).To(o.Equal(expectedKeyValues[key]))
		}
	})

	t.Run("SetProduct", func(t *testing.T) {
		// Get an existing product
		product, err := cfg.GetProduct("Product D")
		g.Expect(err).To(o.Succeed())
		g.Expect(product).NotTo(o.BeNil())

		// Modify it
		product.Enabled = false
		newNamespace := "new-productD-namespace"
		product.Namespace = &newNamespace
		product.Properties["catalogURL"] = "http://new.url/catalog.yaml"

		// Call SetProduct
		err = cfg.SetProduct("Product D", *product)
		g.Expect(err).To(o.Succeed())

		// Assert changes
		configString := cfg.String()
		g.Expect(configString).To(o.ContainSubstring("enabled: false"))
		g.Expect(configString).To(o.ContainSubstring("namespace: new-productD-namespace"))
		g.Expect(configString).To(o.ContainSubstring(
			"catalogURL: http://new.url/catalog.yaml"))

		// Test non-existent product
		err = cfg.SetProduct("NonExistentProduct", Product{})
		g.Expect(err).NotTo(o.Succeed())
		g.Expect(err.Error()).To(o.ContainSubstring(
			"product \"NonExistentProduct\" not found"))
	})

	t.Run("Namespace", func(t *testing.T) {
		namespace := cfg.Namespace()
		g.Expect(namespace).To(o.Equal("test-namespace"))
	})

	t.Run("ApplyDefaults", func(t *testing.T) {
		// ApplyDefaults should propagate the installer namespace to products that don't have one
		cfg.ApplyDefaults()

		// Verify products have namespaces
		products := cfg.GetEnabledProducts()
		for _, product := range products {
			g.Expect(product.Namespace).ToNot(o.BeNil())
			g.Expect(*product.Namespace).ToNot(o.BeEmpty())
		}
	})
}

// TestNewConfigFromBytes tests creating config from byte array
func TestNewConfigFromBytes(t *testing.T) {
	g := o.NewWithT(t)

	cfs := chartfs.New(os.DirFS("../../test"))
	configBytes, err := cfs.ReadFile("config.yaml")
	g.Expect(err).To(o.Succeed())

	cfg, err := NewConfigFromBytes(configBytes, "test-namespace")
	g.Expect(err).To(o.Succeed())
	g.Expect(cfg).ToNot(o.BeNil())
	g.Expect(cfg.Namespace()).To(o.Equal("test-namespace"))
	g.Expect(cfg.Installer).ToNot(o.BeNil())

	// Test with invalid YAML
	invalidYAML := []byte("invalid: yaml: content: :")
	_, err = NewConfigFromBytes(invalidYAML, "test-namespace")
	g.Expect(err).To(o.HaveOccurred())
}

// TestNewConfigDefault tests creating default config
func TestNewConfigDefault(t *testing.T) {
	g := o.NewWithT(t)

	cfs := chartfs.New(os.DirFS("../../test"))

	cfg, err := NewConfigDefault(cfs, "default-namespace")
	g.Expect(err).To(o.Succeed())
	g.Expect(cfg).ToNot(o.BeNil())
	g.Expect(cfg.Namespace()).To(o.Equal("default-namespace"))

	// Validate that it created a valid config
	err = cfg.Validate()
	g.Expect(err).To(o.Succeed())
}

// TestConvertStringMapToAny tests the conversion utility
func TestConvertStringMapToAny(t *testing.T) {
	g := o.NewWithT(t)

	stringMap := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	anyMap := ConvertStringMapToAny(stringMap)
	g.Expect(anyMap).To(o.HaveLen(2))
	g.Expect(anyMap["key1"]).To(o.Equal("value1"))
	g.Expect(anyMap["key2"]).To(o.Equal("value2"))

	// Test with empty map
	emptyMap := ConvertStringMapToAny(map[string]string{})
	g.Expect(emptyMap).To(o.HaveLen(0))
}

// TestValidateErrors tests error conditions in Validate
func TestValidateErrors(t *testing.T) {
	g := o.NewWithT(t)

	cfs := chartfs.New(os.DirFS("../../test"))

	t.Run("missing_settings", func(t *testing.T) {
		cfg, err := NewConfigFromFile(cfs, "config.yaml", "test-namespace")
		g.Expect(err).To(o.Succeed())

		// Clear settings to trigger error
		cfg.Installer.Settings = nil
		err = cfg.Validate()
		g.Expect(err).To(o.HaveOccurred())
		g.Expect(err.Error()).To(o.ContainSubstring("missing settings"))
	})
}
