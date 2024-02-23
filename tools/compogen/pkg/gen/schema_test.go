package gen

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/go-playground/validator/v10"
)

func TestObjectSchema_Validate(t *testing.T) {
	c := qt.New(t)

	validate := validator.New(validator.WithRequiredStructEnabled())

	zero, one, two := 0, 1, 2
	// Returns a valid struct
	validStruct := func() *objectSchema {
		return &objectSchema{
			Properties: map[string]property{
				"stringval": property{
					Description: "a string",
					Title:       "String Value",
					Type:        "string",
					Order:       &zero,
				},
				"intval": property{
					Description: "an integer number",
					Title:       "Integer value",
					Type:        "integer",
					Order:       &one,
				},
			},
			Required: []string{"a"},
		}
	}

	c.Run("ok", func(c *qt.C) {
		err := validate.Struct(validStruct())
		c.Check(err, qt.IsNil)
	})

	testcases := []struct {
		name     string
		modifier func(*objectSchema)
		wantErr  string
	}{
		{
			name: "nok - no properties",
			modifier: func(rs *objectSchema) {
				rs.Properties = map[string]property{}
			},
			wantErr: "Properties field doesn't reach the minimum value / number of elements",
		},
		{
			name: "nok - no title",
			modifier: func(rs *objectSchema) {
				rs.Properties["wrong"] = property{
					Description: "foo",
					Type:        "zoot",
					Order:       &two,
				}
			},
			wantErr: "Title field is required",
		},
		{
			name: "nok - no description",
			modifier: func(rs *objectSchema) {
				rs.Properties["wrong"] = property{
					Title: "bar",
					Type:  "zot",
					Order: &two,
				}
			},
			wantErr: "Description field is required",
		},
		{
			name: "nok - no order",
			modifier: func(rs *objectSchema) {
				rs.Properties["wrong"] = property{
					Description: "foo",
					Title:       "bar",
					Type:        "zot",
				}
			},
			wantErr: "Order field is required",
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
