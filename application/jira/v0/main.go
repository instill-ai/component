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
	apiBaseURL     = "https://api.atlassian.com"
	taskListBoards = "TASK_LIST_BOARDS"
	taskGetIssue   = "TASK_GET_ISSUE"
	taskGetSprint  = "TASK_GET_SPRINT"
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

func (c *component) CreateExecution(sysVars map[string]any, setup *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
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

	return &base.ExecutionWrapper{Execution: e}, nil
}

func (e *execution) fillInDefaultValues(input *structpb.Struct) (*structpb.Struct, error) {
	task := e.Task
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
	for key, value := range inputMap {
		valueMap, ok := value.(map[string]interface{})
		if !ok {
			continue
		}
		if _, ok := valueMap["default"]; !ok {
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
