package gen

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
	"github.com/launchdarkly/go-semver"
	"github.com/russross/blackfriday/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	component "github.com/instill-ai/component/pkg/base"
)

const (
	definitionsFile = "definitions.json"
)

//go:embed resources/templates/readme.mdx.tmpl
var readmeTmpl string

// READMEGenerator is used to generate the README file of a component.
type READMEGenerator struct {
	validate *validator.Validate

	componentType ComponentType
	configDir     string
	outputFile    string
}

// NewREADMEGenerator returns an initialized generator.
func NewREADMEGenerator(configDir, outputFile string, componentType ComponentType) *READMEGenerator {
	return &READMEGenerator{
		validate: validator.New(validator.WithRequiredStructEnabled()),

		componentType: componentType,
		configDir:     configDir,
		outputFile:    outputFile,
	}
}

func (g *READMEGenerator) parseDefinition(configDir string) (d definition, err error) {
	definitionsJSON, err := os.ReadFile(filepath.Join(configDir, definitionsFile))
	if err != nil {
		return d, err
	}

	defs := definitions{}
	if err := json.Unmarshal(definitionsJSON, &defs.Definitions); err != nil {
		return d, err
	}

	if err := g.validate.Struct(defs); err != nil {
		return d, fmt.Errorf("invalid definitions file:\n%w", asValidationError(err))
	}

	_, ok := toComponentSubtype[defs.Definitions[0].Type]
	if g.componentType == ComponentTypeConnector && !ok {
		return d, fmt.Errorf("invalid definitions file:\nType field is invalid")
	}

	return defs.Definitions[0], nil
}

// This is used to build the cURL examples for Instill Core and Cloud.
type host struct {
	Name string
	URL  string
}

// Generate creates a MDX file with the component documentation from the
// component schema.
func (g *READMEGenerator) Generate() error {
	def, err := g.parseDefinition(g.configDir)
	if err != nil {
		return err
	}

	readme, err := template.New("readme").Funcs(template.FuncMap{
		"firstToLower": firstToLower,
		"asAnchor":     blackfriday.SanitizedAnchorName,
		"hosts": func() []host {
			return []host{
				{Name: "Instill-Cloud", URL: "https://api.instill.tech"},
				{Name: "Instill-Core", URL: "http://localhost:8080"},
			}
		},
	}).Parse(readmeTmpl)
	if err != nil {
		return err
	}

	out, err := os.Create(g.outputFile)
	if err != nil {
		return err
	}

	defer out.Close()

	p, err := readmeParams{ComponentType: g.componentType}.parseDefinition(def)
	if err != nil {
		return err
	}

	return readme.Execute(out, p)
}

// Task contains the information of a component task.
type Task struct {
	ID    string
	Title string
}

type resourceProperty struct {
	property
	ID       string
	Required bool
}

type resourceConfig struct {
	Prerequisites string
	Properties    []resourceProperty
}

type readmeParams struct {
	ID               string
	Title            string
	Description      string
	IsDraft          bool
	ComponentType    ComponentType
	ComponentSubtype ComponentSubtype
	ReleaseStage     string
	SourceURL        string
	ResourceConfig   resourceConfig

	Tasks []Task
}

func (p readmeParams) parseDefinition(d definition) (readmeParams, error) {
	switch p.ComponentType {
	case ComponentTypeConnector:
		p.ComponentSubtype = toComponentSubtype[d.Type]
	case ComponentTypeOperator:
		p.ComponentSubtype = cstOperator
	default:
		return p, fmt.Errorf("invalid component type")
	}

	prerelease, err := versionToReleaseStage(d.Version)
	if err != nil {
		return p, err
	}

	p.ID = d.ID
	p.Title = d.Title
	p.Description = d.Description
	p.IsDraft = !d.Public
	p.ReleaseStage = prerelease
	p.SourceURL = d.SourceURL
	p.ResourceConfig = parseResourceConfig(d)

	p.Tasks = make([]Task, len(d.AvailableTasks))
	for i, at := range d.AvailableTasks {
		p.Tasks[i] = Task{
			ID:    at,
			Title: component.TaskIDToTitle(at),
		}
	}

	return p, nil
}

func parseResourceConfig(d definition) resourceConfig {
	rc := resourceConfig{Prerequisites: d.Prerequisites}
	if d.Spec.ResourceSpecification == nil {
		return rc
	}

	specProps := d.Spec.ResourceSpecification.Properties
	rc.Properties = make([]resourceProperty, len(specProps))

	// We need a map first to set the Required property, then we'll
	// transform it to a slice.
	propMap := make(map[string]resourceProperty)
	for k, sp := range specProps {
		propMap[k] = resourceProperty{
			ID:       k,
			property: sp,
		}
	}

	for _, k := range d.Spec.ResourceSpecification.Required {
		if prop, ok := propMap[k]; ok {
			prop.Required = true
			propMap[k] = prop
		}
	}

	for _, rp := range propMap {
		// We can safely access the order pointer because it has been
		// previously validated by the caller.
		rc.Properties[*rp.Order] = rp
	}

	return rc
}

func firstToLower(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size <= 1 {
		return s
	}

	mod := unicode.ToLower(r)
	if r == mod {
		return s
	}

	return string(mod) + s[size:]
}

const generallyAvailable = "GA"

func versionToReleaseStage(s string) (string, error) {
	v, err := semver.Parse(s)
	if err != nil {
		return "", err
	}

	if prerelease := v.GetPrerelease(); prerelease != "" {
		// If prerelease has several bits, use spaces. E.g.:
		// "pre-release" -> "Pre Release"
		rs := cases.Title(language.English).String(strings.ReplaceAll(prerelease, "-", " "))
		return rs, nil
	}

	return generallyAvailable, nil
}
