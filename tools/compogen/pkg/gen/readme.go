package gen

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"path/filepath"

	"html/template"
	"os"
)

const (
	definitionsFile = "definitions.json"
)

//go:embed resources/templates/readme.mdx.tmpl
var readmeTmpl string

func parseDefinition(configDir string) (d definition, err error) {
	definitionsJSON, err := os.ReadFile(filepath.Join(configDir, definitionsFile))
	if err != nil {
		return d, err
	}

	var defs []definition
	if err := json.Unmarshal(definitionsJSON, &defs); err != nil {
		return d, err
	}

	if len(defs) < 1 {
		return d, fmt.Errorf("%s must have at least 1 valid definition", definitionsFile)
	}

	return defs[0], nil
}

// GenerateREADME generates the component documentation from the component
// schema.
func GenerateREADME(configDir, outputFile string) error {
	def, err := parseDefinition(configDir)
	if err != nil {
		return err
	}

	readme, err := template.New("readme").Parse(readmeTmpl)
	if err != nil {
		return err
	}

	out, err := os.Create(outputFile)
	if err != nil {
		return err
	}

	defer out.Close()

	return readme.Execute(out, def.toREADMEParams())
}

type definition struct {
	Title       string `json:"title"`
	Description string `json:"description"`
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
