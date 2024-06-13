package base

import (
	"encoding/base64"
	"regexp"
	"strings"
	"unicode"

	"github.com/gabriel-vasile/mimetype"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"google.golang.org/protobuf/types/known/structpb"
)

type InstillAcceptFormatsCompiler struct{}

func (InstillAcceptFormatsCompiler) Compile(ctx jsonschema.CompilerContext, m map[string]interface{}) (jsonschema.ExtSchema, error) {
	if instillAcceptFormats, ok := m["instillAcceptFormats"]; ok {

		formats := []string{}
		for _, instillAcceptFormat := range instillAcceptFormats.([]interface{}) {
			formats = append(formats, instillAcceptFormat.(string))
		}
		return InstillAcceptFormatsSchema(formats), nil
	}

	return nil, nil
}

type InstillAcceptFormatsSchema []string

func (s InstillAcceptFormatsSchema) Validate(ctx jsonschema.ValidationContext, v interface{}) error {

	// TODO: We should design a better approach to validate the Base64 data.

	switch v := v.(type) {

	case string:
		mimeType := ""
		for _, instillAcceptFormat := range s {

			switch {
			case instillAcceptFormat == "string",
				instillAcceptFormat == "*",
				instillAcceptFormat == "*/*",
				strings.HasPrefix(instillAcceptFormat, "semi-structured"),
				strings.HasPrefix(instillAcceptFormat, "structured"):
				return nil

			// For other types, we assume they are Base64 strings and need to validate the Base64 encoding.
			default:

				b, err := base64.StdEncoding.DecodeString(TrimBase64Mime(v))
				if err != nil {
					return ctx.Error("instillAcceptFormats", "can not decode file")
				}

				mimeType = strings.Split(mimetype.Detect(b).String(), ";")[0]
				if strings.Split(mimeType, "/")[0] == strings.Split(instillAcceptFormat, "/")[0] && strings.Split(instillAcceptFormat, "/")[1] == "*" {
					return nil
				} else if mimeType == instillAcceptFormat {
					return nil
				}
			}
		}
		return ctx.Error("instillAcceptFormats", "expected one of %v, but got %s", s, mimeType)

	default:
		return nil
	}
}

var InstillAcceptFormatsMeta = jsonschema.MustCompileString("instillAcceptFormats.json", `{
	"properties" : {
		"instillAcceptFormats": {
			"type": "array",
			"items": {
				"type": "string"
			}
		}
	}
}`)

type InstillFormatCompiler struct{}

func (InstillFormatCompiler) Compile(ctx jsonschema.CompilerContext, m map[string]interface{}) (jsonschema.ExtSchema, error) {
	if _, ok := m["instillFormat"]; ok {

		return InstillFormatSchema(m["instillFormat"].(string)), nil
	}

	return nil, nil
}

type InstillFormatSchema string

func (s InstillFormatSchema) Validate(ctx jsonschema.ValidationContext, v interface{}) error {

	// TODO: We should design a better approach to validate the Base64 data.
	switch v := v.(type) {

	case string:
		switch {
		case s == "string",
			s == "*",
			s == "*/*",
			strings.HasPrefix(string(s), "semi-structured"),
			strings.HasPrefix(string(s), "structured"):
			return nil

		// For other types, we assume they are Base64 strings and need to validate the Base64 encoding.
		default:
			mimeType := ""
			if !strings.HasPrefix(v, "data:") {
				b, err := base64.StdEncoding.DecodeString(TrimBase64Mime(v))
				if err != nil {
					return ctx.Error("instillFormat", "can not decode file")
				}
				mimeType = strings.Split(mimetype.Detect(b).String(), ";")[0]
			} else {
				mimeType = strings.Split(strings.Split(v, ";")[0], ":")[1]
			}

			if strings.Split(mimeType, "/")[0] == strings.Split(string(s), "/")[0] && strings.Split(string(s), "/")[1] == "*" {
				return nil
			} else if mimeType == string(s) {
				return nil
			} else {
				return ctx.Error("instillFormat", "expected %v, but got %s", s, mimeType)
			}

		}

	default:
		return nil
	}
}

var InstillFormatMeta = jsonschema.MustCompileString("instillFormat.json", `{
	"properties" : {
		"instillFormat": {
			"type": "string"
		}
	}
}`)

func CompileInstillAcceptFormats(sch *structpb.Struct) error {
	var err error
	for k, v := range sch.Fields {
		if v.GetStructValue() != nil {
			err = CompileInstillAcceptFormats(v.GetStructValue())
			if err != nil {
				return err
			}
		}
		if k == "instillAcceptFormats" {
			itemInstillAcceptFormats := []interface{}{}
			for _, item := range v.GetListValue().AsSlice() {
				if strings.HasPrefix(item.(string), "array:") {
					itemInstillAcceptFormats = append(itemInstillAcceptFormats, strings.Split(item.(string), ":")[1])
				}
			}
			if len(itemInstillAcceptFormats) > 0 {
				sch.Fields["items"].GetStructValue().Fields["instillAcceptFormats"], err = structpb.NewValue(itemInstillAcceptFormats)
				if err != nil {
					return err
				}
			}
		}

	}
	return nil
}

func CompileInstillFormat(sch *structpb.Struct) error {
	var err error
	for k, v := range sch.Fields {
		if v.GetStructValue() != nil {
			err = CompileInstillFormat(v.GetStructValue())
			if err != nil {
				return err
			}
		}
		if k == "instillFormat" {
			if strings.HasPrefix(v.GetStringValue(), "array:") {
				itemInstillFormat := strings.Split(v.GetStringValue(), ":")[1]
				sch.Fields["items"] = structpb.NewStructValue(&structpb.Struct{Fields: make(map[string]*structpb.Value)})
				sch.Fields["items"].GetStructValue().Fields["instillFormat"], err = structpb.NewValue(itemInstillFormat)
				if err != nil {
					return err
				}
			}
		}

	}
	return nil
}

func TrimBase64Mime(b64 string) string {
	splitB64 := strings.Split(b64, ",")
	return splitB64[len(splitB64)-1]
}

type InstillDynamicFormatTransformer struct{}

func (InstillDynamicFormatTransformer) ConvertToKebab(str string) string {
	if strings.Index(str, "_") >= 0 {
		re := regexp.MustCompile(`_`)
		return strings.ToLower(re.ReplaceAllString(str, "-"))
	} else if containsCapital(str) {
		return camelToKebab(str)
	}
	return str
}

func containsCapital(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func camelToKebab(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i != 0 {
				result.WriteRune('-')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
