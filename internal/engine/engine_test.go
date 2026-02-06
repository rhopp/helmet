package engine

import (
	"os"
	"testing"

	"github.com/redhat-appstudio/helmet/internal/chartfs"
	"github.com/redhat-appstudio/helmet/internal/config"

	o "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

// testYamlTmpl is a template to render a YAML payload based on the information
// available for the installer template engine.
const testYamlTmpl = `
---
root:
  namespace: {{ .Installer.Namespace }} 
  settings:
    key: value
  products:
{{- range $k, $v := .Installer.Products }}
  {{- $k | nindent 4 }}:
  {{- $v | toYaml | nindent 6 }}
{{- end }}
  catalogURL: {{ .Installer.Products.Product_D.Properties.catalogURL }}
`

func TestEngine_Render(t *testing.T) {
	g := o.NewWithT(t)

	cfs := chartfs.New(os.DirFS("../../test"))

	cfg, err := config.NewConfigFromFile(cfs, "config.yaml", "test-namespace")
	g.Expect(err).To(o.Succeed())

	variables := NewVariables()
	err = variables.SetInstaller(cfg)
	g.Expect(err).To(o.Succeed())

	t.Logf("Template: %s", testYamlTmpl)

	e := NewEngine(nil, testYamlTmpl)
	payload, err := e.Render(variables)
	g.Expect(err).To(o.Succeed())
	g.Expect(payload).NotTo(o.BeEmpty())

	t.Logf("Output: %s", payload)

	// Unmarshal the rendered payload to check the actual structure of the YAML
	// file created with the template engine.
	var outputMap map[string]interface{}
	err = yaml.Unmarshal(payload, &outputMap)
	g.Expect(err).To(o.Succeed())
	g.Expect(outputMap).NotTo(o.BeEmpty())

	g.Expect(outputMap).To(o.HaveKey("root"))
	g.Expect(outputMap["root"]).NotTo(o.BeNil())

	root := outputMap["root"].(map[string]interface{})
	g.Expect(root).To(o.HaveKey("namespace"))
	g.Expect(root["namespace"]).To(o.Equal(cfg.Namespace()))

	g.Expect(root).To(o.HaveKey("settings"))
	g.Expect(root["settings"]).NotTo(o.BeNil())

	g.Expect(root).To(o.HaveKey("products"))
	g.Expect(root["products"]).NotTo(o.BeNil())

	g.Expect(root).To(o.HaveKey("catalogURL"))
	g.Expect(root["catalogURL"]).NotTo(o.BeNil())

	product, err := cfg.GetProduct("Product D")
	g.Expect(err).To(o.Succeed())
	g.Expect(root["catalogURL"]).To(o.Equal(product.Properties["catalogURL"]))
}

// TestEngine_RenderWithInvalidTemplate tests Render with invalid template
func TestEngine_RenderWithInvalidTemplate(t *testing.T) {
	g := o.NewWithT(t)

	// Template with invalid syntax
	invalidTemplate := `{{ .MissingField | invalid }}`

	variables := NewVariables()
	e := NewEngine(nil, invalidTemplate)

	_, err := e.Render(variables)
	g.Expect(err).ToNot(o.BeNil())
}

// TestEngine_RenderWithExecutionError tests Render with template execution error
func TestEngine_RenderWithExecutionError(t *testing.T) {
	g := o.NewWithT(t)

	// Template that references non-existent field
	errorTemplate := `{{ .NonExistent.Field }}`

	variables := NewVariables()
	e := NewEngine(nil, errorTemplate)

	_, err := e.Render(variables)
	g.Expect(err).ToNot(o.BeNil())
}

// TestEngine_RenderEmptyTemplate tests Render with empty template
func TestEngine_RenderEmptyTemplate(t *testing.T) {
	g := o.NewWithT(t)

	emptyTemplate := ``

	variables := NewVariables()
	e := NewEngine(nil, emptyTemplate)

	payload, err := e.Render(variables)
	g.Expect(err).To(o.BeNil())
	g.Expect(payload).To(o.BeEmpty())
}

// TestNewEngine tests the NewEngine constructor
func TestNewEngine(t *testing.T) {
	g := o.NewWithT(t)

	templatePayload := "test template"
	e := NewEngine(nil, templatePayload)

	g.Expect(e).ToNot(o.BeNil())
	g.Expect(e.templatePayload).To(o.Equal(templatePayload))
	g.Expect(e.funcMap).ToNot(o.BeNil())

	// Check that custom functions are registered
	g.Expect(e.funcMap).To(o.HaveKey("toYaml"))
	g.Expect(e.funcMap).To(o.HaveKey("fromYaml"))
	g.Expect(e.funcMap).To(o.HaveKey("fromYamlArray"))
	g.Expect(e.funcMap).To(o.HaveKey("toJson"))
	g.Expect(e.funcMap).To(o.HaveKey("fromJson"))
	g.Expect(e.funcMap).To(o.HaveKey("fromJsonArray"))
	g.Expect(e.funcMap).To(o.HaveKey("required"))
	g.Expect(e.funcMap).To(o.HaveKey("lookup"))
}

// TestNewEngineWithNilKube tests NewEngine with nil Kube
func TestNewEngineWithNilKube(t *testing.T) {
	g := o.NewWithT(t)

	e := NewEngine(nil, "template")
	g.Expect(e).ToNot(o.BeNil())
	g.Expect(e.funcMap).ToNot(o.BeNil())
	g.Expect(e.funcMap).To(o.HaveKey("lookup"))
}
