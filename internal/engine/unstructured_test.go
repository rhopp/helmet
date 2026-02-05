package engine

import (
	"testing"

	o "github.com/onsi/gomega"
)

func TestUnstructuredType(t *testing.T) {
	t.Run("valid map", func(t *testing.T) {
		g := o.NewWithT(t)
		input := map[string]interface{}{
			"key1": "value1",
			"key2": 123,
			"nested": map[string]interface{}{
				"key3": true,
			},
		}

		result, err := UnstructuredType(input)
		g.Expect(err).To(o.Succeed())
		g.Expect(result).ToNot(o.BeNil())
		g.Expect(result["key1"]).To(o.Equal("value1"))
		g.Expect(result["key2"]).To(o.Equal(float64(123))) // JSON numbers are float64
		g.Expect(result["nested"]).ToNot(o.BeNil())
	})

	t.Run("empty map", func(t *testing.T) {
		g := o.NewWithT(t)
		input := map[string]interface{}{}

		result, err := UnstructuredType(input)
		g.Expect(err).To(o.Succeed())
		g.Expect(result).ToNot(o.BeNil())
		g.Expect(result).To(o.BeEmpty())
	})

	t.Run("nil input", func(t *testing.T) {
		g := o.NewWithT(t)
		result, err := UnstructuredType(nil)
		g.Expect(err).To(o.Succeed())
		g.Expect(result).To(o.BeNil())
	})

	t.Run("unmarshalable value", func(t *testing.T) {
		g := o.NewWithT(t)
		// Channels cannot be marshaled to JSON
		input := make(chan int)

		_, err := UnstructuredType(input)
		g.Expect(err).ToNot(o.Succeed())
		g.Expect(err.Error()).To(o.ContainSubstring("json"))
	})

	t.Run("complex struct", func(t *testing.T) {
		g := o.NewWithT(t)
		type TestStruct struct {
			Name  string
			Value int
			Enabled bool
		}
		input := TestStruct{
			Name:  "test",
			Value: 42,
			Enabled: true,
		}

		result, err := UnstructuredType(input)
		g.Expect(err).To(o.Succeed())
		g.Expect(result).ToNot(o.BeNil())
		g.Expect(result["Name"]).To(o.Equal("test"))
		g.Expect(result["Value"]).To(o.Equal(float64(42)))
		g.Expect(result["Enabled"]).To(o.BeTrue())
	})
}

func TestUnstructured(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		g := o.NewWithT(t)
		payload := []byte(`{"key1":"value1","key2":123}`)

		result, err := Unstructured(payload)
		g.Expect(err).To(o.Succeed())
		g.Expect(result).ToNot(o.BeNil())
		g.Expect(result["key1"]).To(o.Equal("value1"))
		g.Expect(result["key2"]).To(o.Equal(float64(123)))
	})

	t.Run("empty JSON object", func(t *testing.T) {
		g := o.NewWithT(t)
		payload := []byte(`{}`)

		result, err := Unstructured(payload)
		g.Expect(err).To(o.Succeed())
		g.Expect(result).ToNot(o.BeNil())
		g.Expect(result).To(o.BeEmpty())
	})

	t.Run("empty payload", func(t *testing.T) {
		g := o.NewWithT(t)
		payload := []byte(``)

		_, err := Unstructured(payload)
		g.Expect(err).ToNot(o.Succeed())
	})

	t.Run("invalid JSON", func(t *testing.T) {
		g := o.NewWithT(t)
		payload := []byte(`{invalid json}`)

		_, err := Unstructured(payload)
		g.Expect(err).ToNot(o.Succeed())
		g.Expect(err.Error()).To(o.ContainSubstring("invalid"))
	})

	t.Run("nested JSON", func(t *testing.T) {
		g := o.NewWithT(t)
		payload := []byte(`{
			"level1": {
				"level2": {
					"key": "value"
				}
			}
		}`)

		result, err := Unstructured(payload)
		g.Expect(err).To(o.Succeed())
		g.Expect(result).ToNot(o.BeNil())

		level1, ok := result["level1"].(map[string]interface{})
		g.Expect(ok).To(o.BeTrue())
		g.Expect(level1).ToNot(o.BeNil())

		level2, ok := level1["level2"].(map[string]interface{})
		g.Expect(ok).To(o.BeTrue())
		g.Expect(level2["key"]).To(o.Equal("value"))
	})

	t.Run("JSON with array", func(t *testing.T) {
		g := o.NewWithT(t)
		payload := []byte(`{"items":["item1","item2","item3"]}`)

		result, err := Unstructured(payload)
		g.Expect(err).To(o.Succeed())
		g.Expect(result).ToNot(o.BeNil())

		items, ok := result["items"].([]interface{})
		g.Expect(ok).To(o.BeTrue())
		g.Expect(items).To(o.HaveLen(3))
		g.Expect(items[0]).To(o.Equal("item1"))
	})
}
