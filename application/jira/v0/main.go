//go:generate compogen readme ./config ./README.mdx
package jira

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/x/errmsg"
)

const (
	apiBaseURL      = "https://api.atlassian.com"
	taskListBoards  = "TASK_LIST_BOARDS"
	taskListIssues  = "TASK_LIST_ISSUES"
	taskListSprints = "TASK_LIST_SPRINTS"
	taskGetIssue    = "TASK_GET_ISSUE"
	taskGetSprint   = "TASK_GET_SPRINT"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/setup.json
	setupJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte

	once sync.Once
	comp *component
)

type component struct {
	base.Component
}

type execution struct {
	base.ComponentExecution
	execute func(context.Context, *structpb.Struct) (*structpb.Struct, error)
	client  Client
}

// Init returns an implementation of IConnector that interacts with Slack.
func Init(bc base.Component) *component {
	once.Do(func() {
		comp = &component{Component: bc}
		err := comp.LoadDefinition(definitionJSON, setupJSON, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})

	return comp
}

func (c *component) CreateExecution(sysVars map[string]any, setup *structpb.Struct, task string) (base.IExecution, error) {
	ctx := context.Background()
	jiraClient, err := newClient(ctx, setup)
	if err != nil {
		return nil, err
	}
	e := &execution{
		ComponentExecution: base.ComponentExecution{Component: c, SystemVariables: sysVars, Setup: setup, Task: task},
		client:             *jiraClient,
	}
	// docs: https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/#about
	switch task {
	case taskListBoards:
		e.execute = e.client.listBoardsTask
	case taskListIssues:
		e.execute = e.client.listIssuesTask
	case taskListSprints:
		e.execute = e.client.listSprintsTask
	case taskGetIssue:
		e.execute = e.client.getIssueTask
	case taskGetSprint:
		e.execute = e.client.getSprintTask
	default:
		return nil, errmsg.AddMessage(
			fmt.Errorf("not supported task: %s", task),
			fmt.Sprintf("%s task is not supported.", task),
		)
	}

	return e, nil
}

func (e *execution) getInputSchemaJSON(task string) (map[string]interface{}, error) {
	taskSpec, ok := e.Component.GetTaskInputSchemas()[task]
	if !ok {
		return nil, errmsg.AddMessage(
			fmt.Errorf("task %s not found", task),
			fmt.Sprintf("Task %s not found", task),
		)
	}
	var taskSpecMap map[string]interface{}
	err := json.Unmarshal([]byte(taskSpec), &taskSpecMap)
	if err != nil {
		return nil, errmsg.AddMessage(
			err,
			"Failed to unmarshal input",
		)
	}
	inputMap := taskSpecMap["properties"].(map[string]interface{})
	return inputMap, nil
}
func (e *execution) fillInDefaultValues(input *structpb.Struct) (*structpb.Struct, error) {
	inputMap, err := e.getInputSchemaJSON(e.Task)
	if err != nil {
		return nil, err
	}
	return e.fillInDefaultValuesWithReference(input, inputMap)
}
func hasNextLevel(valueMap map[string]interface{}) bool {
	if valType, ok := valueMap["type"]; ok {
		if valType != "object" {
			return false
		}
	}
	if _, ok := valueMap["properties"]; ok {
		return true
	}
	for _, target := range []string{"allOf", "anyOf", "oneOf"} {
		if _, ok := valueMap[target]; ok {
			items := valueMap[target].([]interface{})
			for _, v := range items {
				if _, ok := v.(map[string]interface{})["properties"].(map[string]interface{}); ok {
					return true
				}
			}
		}
	}
	return false
}
func optionMatch(valueMap *structpb.Struct, reference map[string]interface{}, checkFields []string) bool {
	for _, checkField := range checkFields {
		if _, ok := valueMap.GetFields()[checkField]; !ok {
			return false
		}
		if val, ok := reference[checkField].(map[string]interface{})["const"]; ok {
			if valueMap.GetFields()[checkField].GetStringValue() != val {
				return false
			}
		}
	}
	return true
}
func (e *execution) fillInDefaultValuesWithReference(input *structpb.Struct, reference map[string]interface{}) (*structpb.Struct, error) {
	for key, value := range reference {
		valueMap, ok := value.(map[string]interface{})
		if !ok {
			continue
		}
		if _, ok := valueMap["default"]; !ok {
			if !hasNextLevel(valueMap) {
				continue
			}
			if _, ok := input.GetFields()[key]; !ok {
				input.GetFields()[key] = &structpb.Value{
					Kind: &structpb.Value_StructValue{
						StructValue: &structpb.Struct{
							Fields: make(map[string]*structpb.Value),
						},
					},
				}
			}
			var properties map[string]interface{}
			if _, ok := valueMap["properties"]; !ok {
				var requiredFieldsRaw []interface{}
				if requiredFieldsRaw, ok = valueMap["required"].([]interface{}); !ok {
					continue
				}
				requiredFields := make([]string, len(requiredFieldsRaw))
				for idx, v := range requiredFieldsRaw {
					requiredFields[idx] = fmt.Sprintf("%v", v)
				}
				for _, target := range []string{"allOf", "anyOf", "oneOf"} {
					var items []interface{}
					if items, ok = valueMap[target].([]interface{}); !ok {
						continue
					}
					for _, v := range items {
						if properties, ok = v.(map[string]interface{})["properties"].(map[string]interface{}); !ok {
							continue
						}
						inputSubField := input.GetFields()[key].GetStructValue()
						if target == "oneOf" && !optionMatch(inputSubField, properties, requiredFields) {
							continue
						}
						subField, err := e.fillInDefaultValuesWithReference(inputSubField, properties)
						if err != nil {
							return nil, err
						}
						input.GetFields()[key] = &structpb.Value{
							Kind: &structpb.Value_StructValue{
								StructValue: subField,
							},
						}
					}
				}
			} else {
				if properties, ok = valueMap["properties"].(map[string]interface{}); !ok {
					continue
				}
				subField, err := e.fillInDefaultValuesWithReference(input.GetFields()[key].GetStructValue(), properties)
				if err != nil {
					return nil, err
				}
				input.GetFields()[key] = &structpb.Value{
					Kind: &structpb.Value_StructValue{
						StructValue: subField,
					},
				}
			}
			continue
		}
		if _, ok := input.GetFields()[key]; ok {
			continue
		}
		defaultValue := valueMap["default"]
		typeValue := valueMap["type"]
		switch typeValue {
		case "string":
			input.GetFields()[key] = &structpb.Value{
				Kind: &structpb.Value_StringValue{
					StringValue: fmt.Sprintf("%v", defaultValue),
				},
			}
		case "integer", "number":
			input.GetFields()[key] = &structpb.Value{
				Kind: &structpb.Value_NumberValue{
					NumberValue: defaultValue.(float64),
				},
			}
		case "boolean":
			input.GetFields()[key] = &structpb.Value{
				Kind: &structpb.Value_BoolValue{
					BoolValue: defaultValue.(bool),
				},
			}
		case "array":
			input.GetFields()[key] = &structpb.Value{
				Kind: &structpb.Value_ListValue{
					ListValue: &structpb.ListValue{
						Values: []*structpb.Value{},
					},
				},
			}
			itemType := valueMap["items"].(map[string]interface{})["type"]
			switch itemType {
			case "string":
				for _, v := range defaultValue.([]interface{}) {
					input.GetFields()[key].GetListValue().Values = append(input.GetFields()[key].GetListValue().Values, &structpb.Value{
						Kind: &structpb.Value_StringValue{
							StringValue: fmt.Sprintf("%v", v),
						},
					})
				}
			case "integer", "number":
				for _, v := range defaultValue.([]interface{}) {
					input.GetFields()[key].GetListValue().Values = append(input.GetFields()[key].GetListValue().Values, &structpb.Value{
						Kind: &structpb.Value_NumberValue{
							NumberValue: v.(float64),
						},
					})
				}
			case "boolean":
				for _, v := range defaultValue.([]interface{}) {
					input.GetFields()[key].GetListValue().Values = append(input.GetFields()[key].GetListValue().Values, &structpb.Value{
						Kind: &structpb.Value_BoolValue{
							BoolValue: v.(bool),
						},
					})
				}
			}
		}
	}
	return input, nil
}

func (e *execution) Execute(ctx context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := make([]*structpb.Struct, len(inputs))

	for i, input := range inputs {
		input, err := e.fillInDefaultValues(input)
		if err != nil {
			return nil, err
		}
		output, err := e.execute(ctx, input)
		if err != nil {
			return nil, err
		}

		outputs[i] = output
	}

	return outputs, nil
}
