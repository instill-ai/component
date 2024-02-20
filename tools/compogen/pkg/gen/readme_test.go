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
					Title:       "foo",
					Description: "bar",
					Public:      false,
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
			name: "nok - multiple errors",
			modifier: func(defs *definitions) {
				defs.Definitions[0] = definition{}
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
