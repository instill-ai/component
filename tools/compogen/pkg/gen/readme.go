package gen

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
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
		"enumValues":       enumValues,
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
		return "", fmt.Errorf("reading extra contents for section %s: %w", section, err)
	}

	return string(extra), nil
}

type readmeTask struct {
	ID            string
	Title         string
	Description   string
	Input         []resourceProperty
	InputObjects  []map[string]objectSchema
	OneOf         map[string][]objectSchema
	Output        []resourceProperty
	OutputObjects []map[string]objectSchema
}

type resourceProperty struct {
	property
	ID       string
	Required bool
}

type setupConfig struct {
	Prerequisites string
	Properties    []resourceProperty
	OneOf         map[string][]objectSchema
}

type readmeParams struct {
	ID            string
	Title         string
	Description   string
	Vendor        string
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
	p.Vendor = d.Vendor
	p.Description = d.Description
	p.IsDraft = !d.Public
	p.ReleaseStage = d.ReleaseStage
	p.SourceURL = d.SourceURL

	p.SetupConfig = setupConfig{Prerequisites: d.Prerequisites}

	if s != nil {
		p.SetupConfig.Properties = parseResourceProperties(s)
		p.SetupConfig.parseOneOfProperties(s.Properties)
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

		rt.parseObjectProperties(t.Input.Properties, true)
		rt.parseObjectProperties(t.Output.Properties, false)
		rt.parseOneOfProperties(t.Input.Properties)

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
		if op.Deprecated {
			continue
		}

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
		prop.Description = strings.ReplaceAll(prop.Description, "}", "\\}")

		propMap[k] = prop
	}

	for _, k := range o.Required {
		if prop, ok := propMap[k]; ok {
			prop.Required = true
			propMap[k] = prop
		}
	}

	props := make([]resourceProperty, len(propMap))
	idx := 0
	for k := range propMap {
		props[idx] = propMap[k]
		idx++
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

func (rt *readmeTask) parseObjectProperties(properties map[string]property, isInput bool) {
	if properties == nil {
		return
	}

	sortedProperties := sortPropertiesByOrder(properties)

	for _, op := range sortedProperties {
		if op.Deprecated {
			continue
		}

		if op.Type != "object" && (op.Type != "array" || op.Items.Type != "object") {
			continue
		}

		if op.Type == "object" {

			if isInput {
				rt.InputObjects = append(rt.InputObjects, map[string]objectSchema{
					op.Title: {
						Properties: op.Properties,
					},
				})
				rt.parseObjectProperties(op.Properties, isInput)
			} else {
				rt.OutputObjects = append(rt.OutputObjects, map[string]objectSchema{
					op.Title: {
						Properties: op.Properties,
					},
				})
				rt.parseObjectProperties(op.Properties, isInput)
			}
		} else { // else if op.Type == "array" && op.Items.Type == "object"

			if isInput {
				rt.InputObjects = append(rt.InputObjects, map[string]objectSchema{
					op.Title: {
						Properties: op.Items.Properties,
					},
				})

				rt.parseObjectProperties(op.Items.Properties, isInput)
			} else {
				rt.OutputObjects = append(rt.OutputObjects, map[string]objectSchema{
					op.Title: {
						Properties: op.Items.Properties,
					},
				})
				rt.parseObjectProperties(op.Items.Properties, isInput)
			}
		}
	}

	return
}

func sortPropertiesByOrder(properties map[string]property) []property {
	// Extract the keys
	keys := make([]string, 0, len(properties))
	for k := range properties {
		keys = append(keys, k)
	}

	// Sort the keys based on the Order field in the property
	sort.Slice(keys, func(i, j int) bool {
		// Default to 0 if Order is nil
		orderI := 0
		if properties[keys[i]].Order != nil {
			orderI = *properties[keys[i]].Order
		}

		orderJ := 0
		if properties[keys[j]].Order != nil {
			orderJ = *properties[keys[j]].Order
		}

		return orderI < orderJ
	})

	sortedProperties := make([]property, 0, len(properties))
	for _, key := range keys {
		sortedProperties = append(sortedProperties, properties[key])
	}

	return sortedProperties
}

func (rt *readmeTask) parseOneOfProperties(properties map[string]property) {
	if properties == nil {
		return
	}

	for key, op := range properties {
		if op.Deprecated {
			continue
		}

		if op.Type != "object" {
			continue
		}

		if op.OneOf != nil {
			if rt.OneOf[key] == nil {
				rt.OneOf = map[string][]objectSchema{
					key: op.OneOf,
				}
			} else {
				rt.OneOf[key] = append(rt.OneOf[key], op.OneOf...)
			}
		}
		rt.parseOneOfProperties(op.Properties)
	}

	return
}

func (sc *setupConfig) parseOneOfProperties(properties map[string]property) {
	if properties == nil {
		return
	}

	for key, op := range properties {
		if op.Deprecated {
			continue
		}

		// Now, we only have 1 layer. So, we do not have to recursively parse.
		if op.OneOf != nil {
			sc.OneOf = map[string][]objectSchema{
				key: op.OneOf,
			}
		}
	}

	return
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

func enumValues(enum []string) string {
	return strings.Join(enum, "<br/>- ")
}
