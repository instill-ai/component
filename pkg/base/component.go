package base

import (
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

// All component need to implement this interface
type IComponent interface {

	// Functions that need to be implemented in component implementation
	// Create a execution by definition uid and component configuration
	CreateExecution(defUid uuid.UUID, config *structpb.Struct, logger *zap.Logger) (IExecution, error)
}

type IExecution interface {
	// Functions that shared for all connectors
	// Validate the input and output format
	ValidateInput(data []*structpb.Struct, task string) error
	ValidateOutput(data []*structpb.Struct, task string) error

	// Functions that need to be implemented in connector implementation
	// Execute
	Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error)
}

type BaseExecution struct {
	// Logger for connection
	Logger                *zap.Logger
	DefUid                uuid.UUID
	OpenAPISpecifications *structpb.Struct
	Config                *structpb.Struct
}

func (conn *BaseExecution) ValidateInput(data []*structpb.Struct, task string) error {
	schema, err := conn.getInputSchema(task)
	if err != nil {
		return err
	}
	return conn.validate(data, string(schema))
}

func (conn *BaseExecution) ValidateOutput(data []*structpb.Struct, task string) error {
	schema, err := conn.getOutputSchema(task)
	if err != nil {
		return err
	}
	return conn.validate(data, string(schema))

}

func (conn *BaseExecution) validate(data []*structpb.Struct, jsonSchema string) error {
	sch, err := jsonschema.CompileString("schema.json", jsonSchema)
	if err != nil {
		return err
	}
	for idx := range data {
		var v interface{}
		jsonData, err := protojson.Marshal(data[idx])
		if err != nil {
			return err
		}

		if err := json.Unmarshal(jsonData, &v); err != nil {
			return err
		}

		if err = sch.Validate(v); err != nil {
			return err
		}
	}
	return nil
}

func (conn *BaseExecution) getInputSchema(task string) ([]byte, error) {

	if _, ok := conn.OpenAPISpecifications.GetFields()[task]; !ok {
		return nil, fmt.Errorf("task %s not exist", task)
	}
	walk := conn.OpenAPISpecifications.GetFields()[task]
	for _, key := range []string{"paths", "/execute", "post", "requestBody", "content", "application/json", "schema", "properties", "inputs", "items"} {
		walk = walk.GetStructValue().Fields[key]
	}
	walkBytes, err := protojson.Marshal(walk)
	return walkBytes, err
}

func (conn *BaseExecution) getOutputSchema(task string) ([]byte, error) {
	if _, ok := conn.OpenAPISpecifications.GetFields()[task]; !ok {
		return nil, fmt.Errorf("task %s not exist", task)
	}
	walk := conn.OpenAPISpecifications.GetFields()[task]
	for _, key := range []string{"paths", "/execute", "post", "responses", "200", "content", "application/json", "schema", "properties", "outputs", "items"} {
		walk = walk.GetStructValue().Fields[key]
	}
	walkBytes, err := protojson.Marshal(walk)
	return walkBytes, err
}

func ConvertFromStructpb(from *structpb.Struct, to interface{}) error {
	inputJson, err := protojson.Marshal(from)
	if err != nil {
		return err
	}

	err = json.Unmarshal(inputJson, to)
	if err != nil {
		return err
	}
	return nil
}

func ConvertToStructpb(from interface{}) (*structpb.Struct, error) {
	to := &structpb.Struct{}
	outputJson, err := json.Marshal(from)
	if err != nil {
		return nil, err
	}

	err = protojson.Unmarshal(outputJson, to)
	if err != nil {
		return nil, err
	}
	return to, nil
}
