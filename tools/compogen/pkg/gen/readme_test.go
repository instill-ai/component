package gen

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestFirstToLower(t *testing.T) {
	c := qt.New(t)

	testcases := []struct {
		in   string
		mod  func(rune) rune
		want string
	}{
		{in: "Hello world!", want: "hello world!"},
		{in: "hello world!", want: "hello world!"},
	}

	for _, tc := range testcases {
		c.Run(tc.in, func(c *qt.C) {
			got := firstToLower(tc.in)
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
		{in: "0.1.0-pre-release", want: "Pre Release"},
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
