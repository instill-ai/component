package gen

import (
	"testing"
	"unicode"

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
					Version:     "0.1.0-alpha",
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

func TestFirstTo(t *testing.T) {
	c := qt.New(t)

	testcases := []struct {
		in   string
		mod  func(rune) rune
		want string
	}{
		{in: "Hello world!", mod: unicode.ToLower, want: "hello world!"},
		{in: "hello world!", mod: unicode.ToLower, want: "hello world!"},
		{in: "hello world!", mod: unicode.ToUpper, want: "Hello world!"},
	}

	for _, tc := range testcases {
		c.Run(tc.in, func(c *qt.C) {
			got := firstTo(tc.in, tc.mod)
			c.Check(got, qt.Equals, tc.want)
		})
	}
}

func TestVersionToReleaseStage(t *testing.T) {
	c := qt.New(t)

	testcases := []struct {
		in   string
		want string
	}{
		{in: "0.1.0-alpha", want: "Alpha"},
		{in: "1.0.0-alpha+001", want: "Alpha"},
		{in: "0.1.0-beta", want: "Beta"},
		{in: "1.0.0-beta+exp.sha", want: "Beta"},
		{in: "0.1.0-pre-release", want: "Pre release"},
		{in: "0.1.0", want: "GA"},
	}

	for _, tc := range testcases {
		c.Run(tc.in, func(c *qt.C) {
			got, err := versionToReleaseStage(tc.in)
			c.Check(err, qt.IsNil)
			c.Check(got, qt.Equals, tc.want)
		})
	}
}

func TestComponentType_IndefiniteArticle(t *testing.T) {
	c := qt.New(t)

	testcases := []struct {
		in   ComponentSubtype
		want string
	}{
		{in: cstOperator, want: "an"},
		{in: cstAIConnector, want: "an"},
		{in: cstBlockchainConnector, want: "a"},
		{in: cstDataConnector, want: "a"},
	}

	for _, tc := range testcases {
		c.Run(string(tc.in), func(c *qt.C) {
			c.Check(tc.in.IndefiniteArticle(), qt.Equals, tc.want)
		})
	}
}
