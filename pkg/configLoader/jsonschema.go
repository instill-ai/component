package configLoader

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// TODO: refactor configLoader

//go:embed config/*
var f embed.FS

type ConfigLoader struct {
	// logger
	Logger *zap.Logger

	// ConnDefJSONSchema represents the ConnectorDefinition JSON Schema
	ConnDefJSONSchema *jsonschema.Schema

	// ConnJSONSchema represents the Connector JSON Schema
	ConnJSONSchema *jsonschema.Schema
}

// InitJSONSchema initialise JSON Schema instances with the given files
func InitJSONSchema(logger *zap.Logger) ConfigLoader {

	compiler := jsonschema.NewCompiler()
	var err error

	connectorDefBytes, _ := f.ReadFile("config/connector_definition.json")
	if err := compiler.AddResource("https://github.com/instill-ai/component/blob/main/pkg/configLoader/config/connector_definition.json", bytes.NewReader(connectorDefBytes)); err != nil {
		logger.Fatal(fmt.Sprintf("%#v\n", err.Error()))
	}

	connectorBytes, _ := f.ReadFile("config/connector.json")
	if err := compiler.AddResource("https://github.com/instill-ai/component/blob/main/pkg/configLoader/config/connector.json", bytes.NewReader(connectorBytes)); err != nil {
		logger.Fatal(fmt.Sprintf("%#v\n", err.Error()))
	}

	connDefJSONSchema, err := compiler.Compile("https://github.com/instill-ai/component/blob/main/pkg/configLoader/config/connector_definition.json")
	if err != nil {
		logger.Fatal(fmt.Sprintf("%#v\n", err.Error()))
	}

	connJSONSchema, err := compiler.Compile("https://github.com/instill-ai/component/blob/main/pkg/configLoader/config/connector.json")
	if err != nil {
		logger.Fatal(fmt.Sprintf("%#v\n", err.Error()))
	}

	return ConfigLoader{
		Logger:            logger,
		ConnDefJSONSchema: connDefJSONSchema,
		ConnJSONSchema:    connJSONSchema,
	}

}

// ValidateJSONSchema validates the Protobuf message data
func ValidateJSONSchema(schema *jsonschema.Schema, msg proto.Message, emitUnpopulated bool) error {

	b, err := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: emitUnpopulated}.Marshal(msg)
	if err != nil {
		return err
	}

	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	if err := schema.Validate(v); err != nil {
		switch e := err.(type) {
		case *jsonschema.ValidationError:
			b, err := json.MarshalIndent(e.DetailedOutput(), "", "  ")
			if err != nil {
				return err
			}
			return fmt.Errorf(string(b))
		case jsonschema.InvalidJSONTypeError:
			return e
		default:
			return e
		}
	}

	return nil
}

// ValidateJSONSchemaString validates the string data given a string schema
func ValidateJSONSchemaString(schema string, data *structpb.Struct) error {

	sch, err := jsonschema.CompileString("schema.json", schema)
	if err != nil {
		return err
	}

	b, err := data.MarshalJSON()
	if err != nil {
		return err
	}

	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	if err = sch.Validate(v); err != nil {
		switch e := err.(type) {
		case *jsonschema.ValidationError:
			b, err := json.MarshalIndent(e.DetailedOutput(), "", "  ")
			if err != nil {
				return err
			}
			return fmt.Errorf(string(b))
		case jsonschema.InvalidJSONTypeError:
			return e
		default:
			return e
		}
	}

	return nil
}
