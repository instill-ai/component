package gen

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
)

const (
	definitionsFile = "definitions.json"
)

//go:embed resources/templates/readme.mdx.tmpl
var readmeTmpl string

// READMEGenerator is used to generate the README file of a component.
type READMEGenerator struct {
	validate   *validator.Validate
	configDir  string
	outputFile string
}

// NewREADMEGenerator returns an initialized generator.
func NewREADMEGenerator(configDir, outputFile string) *READMEGenerator {
	return &READMEGenerator{
		validate:   validator.New(validator.WithRequiredStructEnabled()),
		configDir:  configDir,
		outputFile: outputFile,
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

	return defs.Definitions[0], nil
}

// Generate creates a MDX file with the component documentation from the
// component schema.
func (g *READMEGenerator) Generate() error {
	def, err := g.parseDefinition(g.configDir)
	if err != nil {
		return err
	}

	readme, err := template.New("readme").Parse(readmeTmpl)
	if err != nil {
		return err
	}

	out, err := os.Create(g.outputFile)
	if err != nil {
		return err
	}

	defer out.Close()

	return readme.Execute(out, def.toREADMEParams())
}

// This struct is used to validate the definitions schema.
type definitions struct {
	// Definitions is an array for legacy reasons: Airbyte used to have several
	// definitions. These were merged into one but the structure remained. It
	// should be refactored to remove the array nesting in the future.
	Definitions []definition `validate:"len=1,dive"`
}

type definition struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
	Public      bool   `json:"public"`
}

func (d definition) toREADMEParams() readmeParams {
	return readmeParams{
		Title:         d.Title,
		Description:   "performs an action",
		IsDraft:       !d.Public,
		ComponentType: "operator",
		ReleaseStage:  "Alpha",
	}
}

type readmeParams struct {
	Title         string
	Description   string
	IsDraft       bool
	ComponentType string
	ReleaseStage  string
}
