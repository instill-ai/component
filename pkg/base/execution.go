package base

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

type ExecutionWrapper struct {
	Execution IExecution
}

type IExecution interface {
	GetTask() string
	GetLogger() *zap.Logger
	GetUsageHandler() UsageHandler
	GetTaskInputSchema() string
	GetTaskOutputSchema() string

	Execute([]*structpb.Struct) ([]*structpb.Struct, error)
}

func FormatErrors(inputPath string, e jsonschema.Detailed, errors *[]string) {

	path := inputPath + e.InstanceLocation

	pathItems := strings.Split(path, "/")
	formatedPath := pathItems[0]
	for _, pathItem := range pathItems[1:] {
		if _, err := strconv.Atoi(pathItem); err == nil {
			formatedPath += fmt.Sprintf("[%s]", pathItem)
		} else {
			formatedPath += fmt.Sprintf(".%s", pathItem)
		}

	}
	*errors = append(*errors, fmt.Sprintf("%s: %s", formatedPath, e.Error))

}

// Validate the input and output format
func Validate(data []*structpb.Struct, jsonSchema string, target string) error {

	schStruct := &structpb.Struct{}
	err := protojson.Unmarshal([]byte(jsonSchema), schStruct)
	if err != nil {
		return err
	}

	err = CompileInstillAcceptFormats(schStruct)
	if err != nil {
		return err
	}

	schStr, err := protojson.Marshal(schStruct)
	if err != nil {
		return err
	}

	c := jsonschema.NewCompiler()
	c.RegisterExtension("instillAcceptFormats", InstillAcceptFormatsMeta, InstillAcceptFormatsCompiler{})
	c.RegisterExtension("instillFormat", InstillFormatMeta, InstillFormatCompiler{})
	if err := c.AddResource("schema.json", strings.NewReader(string(schStr))); err != nil {
		return err
	}
	sch, err := c.Compile("schema.json")
	if err != nil {
		return err
	}
	errors := []string{}

	for idx := range data {
		var v interface{}
		jsonData, err := protojson.Marshal(data[idx])
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s[%d]: data error", target, idx))
			continue
		}

		if err := json.Unmarshal(jsonData, &v); err != nil {
			errors = append(errors, fmt.Sprintf("%s[%d]: data error", target, idx))
			continue
		}

		if err = sch.Validate(v); err != nil {
			e := err.(*jsonschema.ValidationError)

			for _, valErr := range e.DetailedOutput().Errors {
				inputPath := fmt.Sprintf("%s/%d", target, idx)
				FormatErrors(inputPath, valErr, &errors)
				for _, subValErr := range valErr.Errors {
					FormatErrors(inputPath, subValErr, &errors)
				}
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}

	return nil
}

// ExecuteWithValidation executes the execution with validation
func (e *ExecutionWrapper) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {

	if err := Validate(inputs, e.Execution.GetTaskInputSchema(), "inputs"); err != nil {
		return nil, err
	}

	if e.Execution.GetUsageHandler() != nil {
		if err := e.Execution.GetUsageHandler().Check(); err != nil {
			return nil, err
		}
	}

	outputs, err := e.Execution.Execute(inputs)
	if err != nil {
		return nil, err
	}

	if err := Validate(outputs, e.Execution.GetTaskOutputSchema(), "outputs"); err != nil {
		return nil, err
	}

	if e.Execution.GetUsageHandler() != nil {
		if err := e.Execution.GetUsageHandler().Collect(); err != nil {
			return nil, err
		}
	}

	return outputs, err
}
