package base

import (
	"testing"

	"github.com/frankban/quicktest"
)

func TestConvertToKebab(t *testing.T) {
	c := quicktest.New(t)

	transformer := InstillDynamicFormatTransformer{}

	tests := []struct {
		input    string
		expected string
	}{
		{"hello_world", "hello-world"},
		{"HelloWorld", "hello-world"},
		{"helloWorld", "hello-world"},
		{"HELLO_WORLD", "hello-world"},
		{"", ""},
	}

	for _, tt := range tests {
		got := transformer.ConvertToKebab(tt.input)
		c.Assert(got, quicktest.Equals, tt.expected)
	}
}
