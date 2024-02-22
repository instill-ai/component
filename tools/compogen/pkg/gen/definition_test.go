package gen

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/go-playground/validator/v10"
)

func TestDefinition_Validate(t *testing.T) {
	c := qt.New(t)

	validate := validator.New(validator.WithRequiredStructEnabled())

	// Returns a valid struct
	validStruct := func() *definitions {
		return &definitions{
			Definitions: []definition{
				{
					ID:             "foo",
					Title:          "Foo",
					Description:    "Foo bar",
					Public:         false,
					Version:        "0.1.0-alpha",
					AvailableTasks: []string{"TASK_1", "TASK_2"},
					SourceURL:      "https://github.com/instill-ai",
				},
			}}
	}

	c.Run("ok", func(c *qt.C) {
		err := validate.Struct(validStruct())
		c.Check(err, qt.IsNil)
	})

	testcases := []struct {
		name     string
		modifier func(*definitions)
		wantErr  string
	}{
		{
			name: "nok - no definitions",
			modifier: func(defs *definitions) {
				defs.Definitions = []definition{}
			},
			wantErr: "Definitions field has an invalid length",
		},
		{
			name: "nok - > 1 definitions",
			modifier: func(defs *definitions) {
				v := validStruct()
				defs.Definitions = append(defs.Definitions, v.Definitions[0])
			},
			wantErr: "Definitions field has an invalid length",
		},
		{
			name: "nok - no ID",
			modifier: func(defs *definitions) {
				defs.Definitions[0].ID = ""
			},
			wantErr: "ID field is required",
		},
		{
			name: "nok - no title",
			modifier: func(defs *definitions) {
				defs.Definitions[0].Title = ""
			},
			wantErr: "Title field is required",
		},
		{
			name: "nok - no description",
			modifier: func(defs *definitions) {
				defs.Definitions[0].Description = ""
			},
			wantErr: "Description field is required",
		},
		{
			name: "nok - no version",
			modifier: func(defs *definitions) {
				defs.Definitions[0].Version = ""
			},
			wantErr: "Version field is required",
		},
		{
			name: "nok - invalid version",
			modifier: func(defs *definitions) {
				defs.Definitions[0].Version = "v0.1.0-alpha"
			},
			wantErr: "Version field must be valid SemVer 2.0.0",
		},
		{
			name: "nok - zero tasks",
			modifier: func(defs *definitions) {
				defs.Definitions[0].AvailableTasks = []string{}
			},
			wantErr: "AvailableTasks field doesn't reach the minimum value / number of elements",
		},
		{
			name: "nok - invalid source URL",
			modifier: func(defs *definitions) {
				defs.Definitions[0].SourceURL = "github.com/instill-ai"
			},
			wantErr: "SourceURL field must be a valid URL",
		},
		{
			// This validates the omitnil tag and the nested validation.
			// Resource specification validation details are covered in a
			// separate test.
			name: "nok - resource specification must be valid if present",
			modifier: func(defs *definitions) {
				defs.Definitions[0].Spec = spec{
					ResourceSpecification: &objectSchema{},
				}
			},
			wantErr: "Properties field doesn't reach the minimum value / number of elements",
		},
		{
			name: "nok - multiple errors",
			modifier: func(defs *definitions) {
				defs.Definitions[0].Title = ""
				defs.Definitions[0].Description = ""
			},
			wantErr: "Title field is required\nDescription field is required",
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			got := validStruct()
			tc.modifier(got)

			err := validate.Struct(got)
			c.Check(err, qt.IsNotNil)
			c.Check(asValidationError(err), qt.ErrorMatches, tc.wantErr)
		})
	}
}
