package gen

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
	"github.com/launchdarkly/go-semver"
	"github.com/russross/blackfriday/v2"

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

// Generate creates a MDX file with the component documentation from the
// component schema.
func (g *READMEGenerator) Generate() error {
	def, err := g.parseDefinition(g.configDir)
	if err != nil {
		return err
	}

	readme, err := template.New("readme").Funcs(template.FuncMap{
		"asAnchor": blackfriday.SanitizedAnchorName,
	}).Parse(readmeTmpl)
	if err != nil {
		return err
	}

	out, err := os.Create(g.outputFile)
	if err != nil {
		return err
	}

	defer out.Close()

	p, err := def.toREADMEParams(g.componentType)
	if err != nil {
		return err
	}

	return readme.Execute(out, p)
}

// This struct is used to validate the definitions schema.
type definitions struct {
	// Definitions is an array for legacy reasons: Airbyte used to have several
	// definitions. These were merged into one but the structure remained. It
	// should be refactored to remove the array nesting in the future.
	Definitions []definition `validate:"len=1,dive"`
}

type definition struct {
	Title          string   `json:"title" validate:"required"`
	Description    string   `json:"description" validate:"required"`
	Version        string   `json:"version" validate:"required,semver"`
	AvailableTasks []string `json:"available_tasks" validate:"gt=0"`

	Public bool   `json:"public"`
	Type   string `json:"type"`
}

func (d definition) toREADMEParams(ct ComponentType) (readmeParams, error) {
	p := readmeParams{}

	prerelease, err := versionToReleaseStage(d.Version)
	if err != nil {
		return p, err
	}

	p.Title = d.Title
	p.Description = firstTo(d.Description, unicode.ToLower)
	p.ComponentType = ct
	p.IsDraft = !d.Public
	p.ReleaseStage = prerelease

	p.Tasks = make([]Task, len(d.AvailableTasks))
	for i, at := range d.AvailableTasks {
		p.Tasks[i] = Task{
			ID:    at,
			Title: component.TaskIDToTitle(at),
		}
	}

	switch ct {
	case ComponentTypeConnector:
		p.ComponentSubtype = toComponentSubtype[d.Type]
	case ComponentTypeOperator:
		p.ComponentSubtype = cstOperator
	}

	return p, nil
}

// Task contains the information of a component task.
type Task struct {
	ID    string
	Title string
}

type readmeParams struct {
	Title            string
	Description      string
	IsDraft          bool
	ComponentType    ComponentType
	ComponentSubtype ComponentSubtype
	ReleaseStage     string

	Tasks []Task
}

func firstTo(s string, modifier func(rune) rune) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size <= 1 {
		return s
	}

	mod := modifier(r)
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
		// "pre-release" -> "Pre release"
		rs := strings.ReplaceAll(firstTo(prerelease, unicode.ToUpper), "-", " ")
		return rs, nil
	}

	return generallyAvailable, nil
}
