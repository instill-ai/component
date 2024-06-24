package gen

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	_ "embed"

	"github.com/go-playground/validator/v10"
	"github.com/russross/blackfriday/v2"

	componentbase "github.com/instill-ai/component/base"
)

const (
	definitionsFile = "definition.json"
	setupFile       = "setup.json"
	tasksFile       = "tasks.json"
)

//go:embed resources/templates/readme.mdx.tmpl
var readmeTmpl string

// READMEGenerator is used to generate the README file of a component.
type READMEGenerator struct {
	validate *validator.Validate

	configDir         string
	outputFile        string
	extraContentPaths map[string]string
}

// NewREADMEGenerator returns an initialized generator.
func NewREADMEGenerator(configDir, outputFile string, extraContentPaths map[string]string) *READMEGenerator {
	return &READMEGenerator{
		validate: validator.New(validator.WithRequiredStructEnabled()),

		configDir:         configDir,
		outputFile:        outputFile,
		extraContentPaths: extraContentPaths,
	}
}

func (g *READMEGenerator) parseDefinition(configDir string) (d definition, err error) {
	definitionJSON, err := os.ReadFile(filepath.Join(configDir, definitionsFile))
	if err != nil {
		return d, err
	}

	renderedDefinitionJSON, err := componentbase.RenderJSON(definitionJSON, nil)
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

	return def, nil
}

func (g *READMEGenerator) parseSetup(configDir string) (s *objectSchema, err error) {
	setupJSON, err := os.ReadFile(filepath.Join(configDir, setupFile))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	renderedSetupJSON, err := componentbase.RenderJSON(setupJSON, nil)
	if err != nil {
		return nil, err
	}

	setup := &objectSchema{}
	if err := json.Unmarshal(renderedSetupJSON, &setup); err != nil {
		return nil, err
	}

	if err := g.validate.Var(setup, "len=1,dive"); err != nil {
		return nil, fmt.Errorf("invalid definitions file:\n%w", asValidationError(err))
	}

	return setup, nil
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

	renderedTasksJSON, err := componentbase.RenderJSON(tasksJSON, additionalJSONs)
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

	setup, err := g.parseSetup(g.configDir)
	if err != nil {
		return err
	}

	tasks, err := g.parseTasks(g.configDir)
	if err != nil {
		return err
	}

	readme, err := template.New("readme").Funcs(template.FuncMap{
		"firstToLower":     firstToLower,
		"asAnchor":         blackfriday.SanitizedAnchorName,
		"loadExtraContent": g.loadExtraContent,
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

	p, err := readmeParams{}.parseDefinition(def, setup, tasks)
	if err != nil {
		return fmt.Errorf("converting to template params: %w", err)
	}

	return readme.Execute(out, p)
}

func (g *READMEGenerator) loadExtraContent(section string) (string, error) {
	if g.extraContentPaths[section] == "" {
		return "", nil
	}

	extra, err := os.ReadFile(g.extraContentPaths[section])
	if err != nil {
		return "", fmt.Errorf("reading extra contents for sectino %s: %w", section, err)
	}

	return string(extra), nil
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

type setupConfig struct {
	Prerequisites string
	Properties    []resourceProperty
}

type readmeParams struct {
	ID            string
	Title         string
	Description   string
	IsDraft       bool
	ComponentType ComponentType
	ReleaseStage  releaseStage
	SourceURL     string
	SetupConfig   setupConfig

	Tasks []readmeTask
}

// parseDefinition converts a component definition and its tasks to the README
// template params.
func (p readmeParams) parseDefinition(d definition, s *objectSchema, tasks map[string]task) (readmeParams, error) {
	p.ComponentType = toComponentType[d.Type]

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

	p.SetupConfig = setupConfig{Prerequisites: d.Prerequisites}

	if s != nil {
		p.SetupConfig.Properties = parseResourceProperties(s)
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
			rt.Title = componentbase.TaskIDToTitle(at)
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
		prop.Description = strings.ReplaceAll(prop.Description, "\n", " ")
		prop.Description = strings.ReplaceAll(prop.Description, "{", "\\{")
		prop.Description = strings.ReplaceAll(prop.Description, "}", "\\{")

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
		if cmp := cmp.Compare(*i.Order, *j.Order); cmp != 0 {
			return cmp
		}
		return cmp.Compare(i.ID, j.ID)
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
