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

	// SrcConnDefJSONSchema represents the SourceConnectorDefinition JSON Schema
	SrcConnDefJSONSchema *jsonschema.Schema

	// DstConnDefJSONSchema represents the DestinationConnectorDefinition JSON Schema
	DstConnDefJSONSchema *jsonschema.Schema

	// SrcConnJSONSchema represents the SourceConnector JSON Schema
	SrcConnJSONSchema *jsonschema.Schema

	// DstConnJSONSchema represents the DestinationConnector JSON Schema
	DstConnJSONSchema *jsonschema.Schema
}

// InitJSONSchema initialise JSON Schema instances with the given files
func InitJSONSchema(logger *zap.Logger) ConfigLoader {

	compiler := jsonschema.NewCompiler()
	var err error

	connectorDefBytes, _ := f.ReadFile("config/connector_definition.json")
	if err := compiler.AddResource("https://github.com/instill-ai/connector/blob/main/pkg/configLoader/config/connector_definition.json", bytes.NewReader(connectorDefBytes)); err != nil {
		logger.Fatal(fmt.Sprintf("%#v\n", err.Error()))
	}

	connectorBytes, _ := f.ReadFile("config/connector.json")
	if err := compiler.AddResource("https://github.com/instill-ai/connector/blob/main/pkg/configLoader/config/connector.json", bytes.NewReader(connectorBytes)); err != nil {
		logger.Fatal(fmt.Sprintf("%#v\n", err.Error()))
	}

	srcConnectorDefBytes, _ := f.ReadFile("config/source_connector_definition.json")
	if err := compiler.AddResource("https://github.com/instill-ai/connector/blob/main/pkg/configLoader/config/source_connector_definition.json", bytes.NewReader(srcConnectorDefBytes)); err != nil {
		logger.Fatal(fmt.Sprintf("%#v\n", err.Error()))
	}
	srcConnDefJSONSchema, err := compiler.Compile("https://github.com/instill-ai/connector/blob/main/pkg/configLoader/config/source_connector_definition.json")
	if err != nil {
		logger.Fatal(fmt.Sprintf("%#v\n", err.Error()))
	}

	destConnectorDefBytes, _ := f.ReadFile("config/destination_connector_definition.json")
	if err := compiler.AddResource("https://github.com/instill-ai/connector/blob/main/pkg/configLoader/config/destination_connector_definition.json", bytes.NewReader(destConnectorDefBytes)); err != nil {
		logger.Fatal(fmt.Sprintf("%#v\n", err.Error()))
	}
	dstConnDefJSONSchema, err := compiler.Compile("https://github.com/instill-ai/connector/blob/main/pkg/configLoader/config/destination_connector_definition.json")
	if err != nil {
		logger.Fatal(fmt.Sprintf("%#v\n", err.Error()))
	}

	srcConnectorBytes, _ := f.ReadFile("config/source_connector.json")
	if err := compiler.AddResource("https://github.com/instill-ai/connector/blob/main/pkg/configLoader/config/source_connector.json", bytes.NewReader(srcConnectorBytes)); err != nil {
		logger.Fatal(fmt.Sprintf("%#v\n", err.Error()))
	}

	srcConnJSONSchema, err := compiler.Compile("https://github.com/instill-ai/connector/blob/main/pkg/configLoader/config/source_connector.json")
	if err != nil {
		logger.Fatal(fmt.Sprintf("%#v\n", err.Error()))
	}

	destConnectorBytes, _ := f.ReadFile("config/destination_connector.json")
	if err := compiler.AddResource("https://github.com/instill-ai/connector/blob/main/pkg/configLoader/config/destination_connector.json", bytes.NewReader(destConnectorBytes)); err != nil {
		logger.Fatal(fmt.Sprintf("%#v\n", err.Error()))
	}
	dstConnJSONSchema, err := compiler.Compile("https://github.com/instill-ai/connector/blob/main/pkg/configLoader/config/destination_connector.json")
	if err != nil {
		logger.Fatal(fmt.Sprintf("%#v\n", err.Error()))
	}

	return ConfigLoader{
		Logger:               logger,
		SrcConnDefJSONSchema: srcConnDefJSONSchema,
		DstConnDefJSONSchema: dstConnDefJSONSchema,
		SrcConnJSONSchema:    srcConnJSONSchema,
		DstConnJSONSchema:    dstConnJSONSchema,
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
