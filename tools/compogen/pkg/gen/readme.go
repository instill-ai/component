package gen

import (
	"cmp"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"text/template"
	"unicode"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
	"github.com/russross/blackfriday/v2"

	component "github.com/instill-ai/component/pkg/base"
)

const (
	definitionsFile = "definition.json"
	tasksFile       = "tasks.json"
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
	definitionJSON, err := os.ReadFile(filepath.Join(configDir, definitionsFile))
	if err != nil {
		return d, err
	}

	renderedDefinitionJSON, err := component.RenderJSON(definitionJSON, nil)
	if err != nil {
		return d, err
	}

	def := definition{}
	if err := json.Unmarshal(renderedDefinitionJSON, &def); err != nil {
		return d, err
	}

	if err := g.validate.Var(def, "len=1,dive"); err != nil {
		return d, fmt.Errorf("invalid definitions file:\n%w", asValidationError(err))
	}

	_, ok := toComponentSubtype[def.Type]
	if g.componentType == ComponentTypeConnector && !ok {
		return d, fmt.Errorf("invalid definitions file:\nType field is invalid")
	}

	return def, nil
}

func (g *READMEGenerator) parseTasks(configDir string) (map[string]task, error) {
	tasksJSON, err := os.ReadFile(filepath.Join(configDir, tasksFile))
	if err != nil {
		return nil, err
	}
	files, err := os.ReadDir(configDir)
	if err != nil {
		return nil, err
	}
	additionalJSONs := map[string][]byte{}
	for _, file := range files {
		additionalJSON, err := os.ReadFile(filepath.Join(configDir, file.Name()))
		if err != nil {
			return nil, err
		}
		additionalJSONs[file.Name()] = additionalJSON

	}

	renderedTasksJSON, err := component.RenderJSON(tasksJSON, additionalJSONs)
	if err != nil {
		return nil, err
	}
	tasks := map[string]task{}
	if err := json.Unmarshal(renderedTasksJSON, &tasks); err != nil {
		return nil, err
	}

	if err := g.validate.Var(tasks, "gt=0,dive"); err != nil {
		return nil, fmt.Errorf("invalid tasks file:\n%w", asValidationError(err))
	}

	return tasks, nil
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

	tasks, err := g.parseTasks(g.configDir)
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

	p, err := readmeParams{ComponentType: g.componentType}.parseDefinition(def, tasks)
	if err != nil {
		return err
	}

	return readme.Execute(out, p)
}

type readmeTask struct {
	ID          string
	Title       string
	Description string
	Input       []resourceProperty
	Output      []resourceProperty
}

type resourceProperty struct {
	property
	ID       string
	Required bool
}

type connectionConfig struct {
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
	ReleaseStage     releaseStage
	SourceURL        string
	ConnectionConfig connectionConfig

	Tasks []readmeTask
}

func (p readmeParams) parseDefinition(d definition, tasks map[string]task) (readmeParams, error) {
	switch p.ComponentType {
	case ComponentTypeConnector:
		p.ComponentSubtype = toComponentSubtype[d.Type]
	case ComponentTypeOperator:
		p.ComponentSubtype = cstOperator
	default:
		return p, fmt.Errorf("invalid component type")
	}

	var err error
	if p.Tasks, err = parseREADMETasks(d.AvailableTasks, tasks); err != nil {
		return p, err
	}

	p.ID = d.ID
	p.Title = d.Title
	p.Description = d.Description
	p.IsDraft = !d.Public
	p.ReleaseStage = d.ReleaseStage
	p.SourceURL = d.SourceURL

	p.ConnectionConfig = connectionConfig{Prerequisites: d.Prerequisites}
	if d.Spec.ConnectionSpecification != nil {
		p.ConnectionConfig.Properties = parseResourceProperties(d.Spec.ConnectionSpecification)
	}

	return p, nil
}

func parseREADMETasks(availableTasks []string, tasks map[string]task) ([]readmeTask, error) {
	readmeTasks := make([]readmeTask, len(availableTasks))
	for i, at := range availableTasks {
		t, ok := tasks[at]
		if !ok {
			return nil, fmt.Errorf("invalid tasks file:\nmissing %s", at)
		}

		rt := readmeTask{
			ID:          at,
			Description: t.Description,
			Input:       parseResourceProperties(t.Input),
			Output:      parseResourceProperties(t.Output),
		}

		if rt.Title = t.Title; rt.Title == "" {
			rt.Title = component.TaskIDToTitle(at)
		}

		readmeTasks[i] = rt
	}

	return readmeTasks, nil
}

func parseResourceProperties(o *objectSchema) []resourceProperty {
	if o == nil {
		return []resourceProperty{}
	}

	// We need a map first to set the Required property, then we'll
	// transform it to a slice.
	propMap := make(map[string]resourceProperty)
	for k, op := range o.Properties {
		prop := resourceProperty{
			ID:       k,
			property: op,
		}
		// If type is map, extend the type with the element type.
		switch prop.Type {
		case "array":
			if prop.Items.Type != "" {
				prop.Type += fmt.Sprintf("[%s]", prop.Items.Type)
			}
		case "":
			prop.Type = "any"
		}

		propMap[k] = prop
	}

	for _, k := range o.Required {
		if prop, ok := propMap[k]; ok {
			prop.Required = true
			propMap[k] = prop
		}
	}

	props := make([]resourceProperty, len(o.Properties))
	idx := 0
	for k := range propMap {
		props[idx] = propMap[k]
		idx += 1
	}

	// Note: The order might not be consecutive numbers.
	slices.SortFunc(props, func(i, j resourceProperty) int {
		return cmp.Compare(*i.Order, *j.Order)
	})

	return props
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
