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
	validStruct := func() *definition {
		return &definition{
			ID:             "foo",
			Title:          "Foo",
			Description:    "Foo bar",
			Public:         false,
			Version:        "0.1.0-alpha",
			AvailableTasks: []string{"TASK_1", "TASK_2"},
			SourceURL:      "https://github.com/instill-ai",
		}
	}

	c.Run("ok", func(c *qt.C) {
		err := validate.Struct(validStruct())
		c.Check(err, qt.IsNil)
	})

	testcases := []struct {
		name     string
		modifier func(*definition)
		wantErr  string
	}{
		{
			name: "nok - no ID",
			modifier: func(d *definition) {
				d.ID = ""
			},
			wantErr: "ID field is required",
		},
		{
			name: "nok - no title",
			modifier: func(d *definition) {
				d.Title = ""
			},
			wantErr: "Title field is required",
		},
		{
			name: "nok - no description",
			modifier: func(d *definition) {
				d.Description = ""
			},
			wantErr: "Description field is required",
		},
		{
			name: "nok - no version",
			modifier: func(d *definition) {
				d.Version = ""
			},
			wantErr: "Version field is required",
		},
		{
			name: "nok - invalid version",
			modifier: func(d *definition) {
				d.Version = "v0.1.0-alpha"
			},
			wantErr: "Version field must be valid SemVer 2.0.0",
		},
		{
			name: "nok - zero tasks",
			modifier: func(d *definition) {
				d.AvailableTasks = []string{}
			},
			wantErr: "AvailableTasks field doesn't reach the minimum value / number of elements",
		},
		{
			name: "nok - invalid source URL",
			modifier: func(d *definition) {
				d.SourceURL = "github.com/instill-ai"
			},
			wantErr: "SourceURL field must be a valid URL",
		},
		{
			// This validates the omitnil tag and the nested validation.
			// Resource specification validation details are covered in a
			// separate test.
			name: "nok - resource specification must be valid if present",
			modifier: func(d *definition) {
				d.Spec = spec{
					ResourceSpecification: &objectSchema{},
				}
			},
			wantErr: "Properties field doesn't reach the minimum value / number of elements",
		},
		{
			name: "nok - multiple errors",
			modifier: func(d *definition) {
				d.Title = ""
				d.Description = ""
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
