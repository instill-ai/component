package base

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/instill-ai/component/pkg/jsonref"
	"github.com/lestrrat-go/jsref/provider"
	"go.uber.org/zap"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

const conditionJSON = `
{
	"type": "string",
	"instillUIOrder": 1,
	"instillShortDescription": "config whether the component will be executed or skipped",
	"instillAcceptFormats": ["string"],
    "instillUpstreamTypes": ["value", "template"]
}
`

// IComponent is the interface that wraps the basic component methods.
// All component need to implement this interface.
type IComponent interface {
	GetID() string
	GetUID() uuid.UUID
	GetLogger() *zap.Logger
	GetTaskInputSchemas() map[string]string
	GetTaskOutputSchemas() map[string]string

	UsageHandlerCreator() func(IExecution) UsageHandler
}

func convertDataSpecToCompSpec(dataSpec *structpb.Struct) (*structpb.Struct, error) {
	// var err error
	compSpec := proto.Clone(dataSpec).(*structpb.Struct)
	if _, ok := compSpec.Fields["const"]; ok {
		return compSpec, nil
	}

	isFreeform := checkFreeForm(compSpec)

	if _, ok := compSpec.Fields["type"]; !ok && !isFreeform {
		return nil, fmt.Errorf("type missing: %+v", compSpec)
	} else if _, ok := compSpec.Fields["instillUpstreamTypes"]; !ok && compSpec.Fields["type"].GetStringValue() == "object" {

		if _, ok := compSpec.Fields["instillUIOrder"]; !ok {
			compSpec.Fields["instillUIOrder"] = structpb.NewNumberValue(0)
		}
		if _, ok := compSpec.Fields["required"]; !ok {
			return nil, fmt.Errorf("required missing: %+v", compSpec)
		}
		if _, ok := compSpec.Fields["instillEditOnNodeFields"]; !ok {
			compSpec.Fields["instillEditOnNodeFields"] = compSpec.Fields["required"]
		}

		if _, ok := compSpec.Fields["properties"]; ok {
			for k, v := range compSpec.Fields["properties"].GetStructValue().AsMap() {
				s, err := structpb.NewStruct(v.(map[string]interface{}))
				if err != nil {
					return nil, err
				}
				converted, err := convertDataSpecToCompSpec(s)
				if err != nil {
					return nil, err
				}
				compSpec.Fields["properties"].GetStructValue().Fields[k] = structpb.NewStructValue(converted)

			}
		}
		if _, ok := compSpec.Fields["patternProperties"]; ok {
			for k, v := range compSpec.Fields["patternProperties"].GetStructValue().AsMap() {
				s, err := structpb.NewStruct(v.(map[string]interface{}))
				if err != nil {
					return nil, err
				}
				converted, err := convertDataSpecToCompSpec(s)
				if err != nil {
					return nil, err
				}
				compSpec.Fields["patternProperties"].GetStructValue().Fields[k] = structpb.NewStructValue(converted)

			}
		}
		for _, target := range []string{"allOf", "anyOf", "oneOf"} {
			if _, ok := compSpec.Fields[target]; ok {
				for idx, item := range compSpec.Fields[target].GetListValue().AsSlice() {
					s, err := structpb.NewStruct(item.(map[string]interface{}))
					if err != nil {
						return nil, err
					}
					converted, err := convertDataSpecToCompSpec(s)
					if err != nil {
						return nil, err
					}
					compSpec.Fields[target].GetListValue().AsSlice()[idx] = structpb.NewStructValue(converted)
				}
			}
		}

	} else {
		if _, ok := compSpec.Fields["instillUIOrder"]; !ok {
			compSpec.Fields["instillUIOrder"] = structpb.NewNumberValue(0)
		}
		original := proto.Clone(compSpec).(*structpb.Struct)
		delete(original.Fields, "title")
		delete(original.Fields, "description")
		delete(original.Fields, "instillShortDescription")
		delete(original.Fields, "instillAcceptFormats")
		delete(original.Fields, "instillUIOrder")
		delete(original.Fields, "instillUpstreamTypes")

		newCompSpec := &structpb.Struct{Fields: make(map[string]*structpb.Value)}

		newCompSpec.Fields["title"] = structpb.NewStringValue(compSpec.Fields["title"].GetStringValue())
		newCompSpec.Fields["description"] = structpb.NewStringValue(compSpec.Fields["description"].GetStringValue())
		if _, ok := compSpec.Fields["instillShortDescription"]; ok {
			newCompSpec.Fields["instillShortDescription"] = compSpec.Fields["instillShortDescription"]
		} else {
			newCompSpec.Fields["instillShortDescription"] = newCompSpec.Fields["description"]
		}
		newCompSpec.Fields["instillUIOrder"] = structpb.NewNumberValue(compSpec.Fields["instillUIOrder"].GetNumberValue())
		if compSpec.Fields["instillAcceptFormats"] != nil {
			newCompSpec.Fields["instillAcceptFormats"] = structpb.NewListValue(compSpec.Fields["instillAcceptFormats"].GetListValue())
		}
		newCompSpec.Fields["instillUpstreamTypes"] = structpb.NewListValue(compSpec.Fields["instillUpstreamTypes"].GetListValue())
		newCompSpec.Fields["anyOf"] = structpb.NewListValue(&structpb.ListValue{Values: []*structpb.Value{}})

		for _, v := range compSpec.Fields["instillUpstreamTypes"].GetListValue().GetValues() {
			if v.GetStringValue() == "value" {
				original.Fields["instillUpstreamType"] = v
				newCompSpec.Fields["anyOf"].GetListValue().Values = append(newCompSpec.Fields["anyOf"].GetListValue().Values, structpb.NewStructValue(original))
			}
			if v.GetStringValue() == "reference" {
				item, err := structpb.NewValue(
					map[string]interface{}{
						"type":                "string",
						"pattern":             "^\\{.*\\}$",
						"instillUpstreamType": "reference",
					},
				)
				if err != nil {
					return nil, err
				}
				newCompSpec.Fields["anyOf"].GetListValue().Values = append(newCompSpec.Fields["anyOf"].GetListValue().Values, item)
			}
			if v.GetStringValue() == "template" {
				item, err := structpb.NewValue(
					map[string]interface{}{
						"type":                "string",
						"instillUpstreamType": "template",
					},
				)
				if err != nil {
					return nil, err
				}
				newCompSpec.Fields["anyOf"].GetListValue().Values = append(newCompSpec.Fields["anyOf"].GetListValue().Values, item)
			}

		}

		compSpec = newCompSpec

	}
	return compSpec, nil
}

const taskPrefix = "TASK_"

// TaskIDToTitle builds a Task title from its ID. This is used when the `title`
// key in the task definition isn't present.
func TaskIDToTitle(id string) string {
	title := strings.ReplaceAll(id, taskPrefix, "")
	title = strings.ReplaceAll(title, "_", " ")
	return cases.Title(language.English).String(title)
}

func generateComponentTaskCards(tasks map[string]*structpb.Struct) []*pipelinePB.ComponentTask {
	taskCards := make([]*pipelinePB.ComponentTask, 0, len(tasks))
	for k := range tasks {
		title := tasks[k].Fields["title"].GetStringValue()
		if title == "" {
			title = TaskIDToTitle(k)
		}

		description := tasks[k].Fields["instillShortDescription"].GetStringValue()

		taskCards = append(taskCards, &pipelinePB.ComponentTask{
			Name:        k,
			Title:       title,
			Description: description,
		})
	}

	return taskCards
}

func generateComponentSpec(title string, tasks []*pipelinePB.ComponentTask, taskStructs map[string]*structpb.Struct) (*structpb.Struct, error) {
	var err error
	componentSpec := &structpb.Struct{Fields: map[string]*structpb.Value{}}
	componentSpec.Fields["$schema"] = structpb.NewStringValue("http://json-schema.org/draft-07/schema#")
	componentSpec.Fields["title"] = structpb.NewStringValue(fmt.Sprintf("%s Component", title))
	componentSpec.Fields["type"] = structpb.NewStringValue("object")

	if err != nil {
		return nil, err
	}

	oneOfList := &structpb.ListValue{
		Values: []*structpb.Value{},
	}
	for _, task := range tasks {
		taskName := task.Name

		oneOf := &structpb.Struct{Fields: map[string]*structpb.Value{}}
		oneOf.Fields["type"] = structpb.NewStringValue("object")
		oneOf.Fields["properties"] = structpb.NewStructValue(&structpb.Struct{Fields: make(map[string]*structpb.Value)})

		oneOf.Fields["properties"].GetStructValue().Fields["task"], err = structpb.NewValue(map[string]interface{}{
			"const": task.Name,
			"title": task.Title,
		})
		if err != nil {
			return nil, err
		}

		if taskStructs[taskName].Fields["description"].GetStringValue() != "" {
			oneOf.Fields["properties"].GetStructValue().Fields["task"].GetStructValue().Fields["description"] = structpb.NewStringValue(taskStructs[taskName].Fields["description"].GetStringValue())
		}

		if task.Description != "" {
			oneOf.Fields["properties"].GetStructValue().Fields["task"].GetStructValue().Fields["instillShortDescription"] = structpb.NewStringValue(task.Description)
		}
		taskJSONStruct := proto.Clone(taskStructs[taskName]).(*structpb.Struct).Fields["input"].GetStructValue()

		compInputStruct, err := convertDataSpecToCompSpec(taskJSONStruct)
		if err != nil {
			return nil, fmt.Errorf("task %s: %s error: %+v", title, task, err)
		}

		condition := &structpb.Struct{}
		err = protojson.Unmarshal([]byte(conditionJSON), condition)
		if err != nil {
			if err != nil {
				panic(err)
			}
		}
		oneOf.Fields["properties"].GetStructValue().Fields["condition"] = structpb.NewStructValue(condition)
		oneOf.Fields["properties"].GetStructValue().Fields["input"] = structpb.NewStructValue(compInputStruct)
		if taskStructs[taskName].Fields["metadata"] != nil {
			metadataStruct := proto.Clone(taskStructs[taskName]).(*structpb.Struct).Fields["metadata"].GetStructValue()
			oneOf.Fields["properties"].GetStructValue().Fields["metadata"] = structpb.NewStructValue(metadataStruct)
		}

		// oneOf
		oneOfList.Values = append(oneOfList.Values, structpb.NewStructValue(oneOf))
	}

	componentSpec.Fields["oneOf"] = structpb.NewListValue(oneOfList)

	if err != nil {
		return nil, err
	}

	return componentSpec, nil

}

func formatDataSpec(dataSpec *structpb.Struct) (*structpb.Struct, error) {
	// var err error
	compSpec := proto.Clone(dataSpec).(*structpb.Struct)
	if _, ok := compSpec.Fields["const"]; ok {
		return compSpec, nil
	}

	isFreeform := checkFreeForm(compSpec)

	if _, ok := compSpec.Fields["type"]; !ok && !isFreeform {
		return nil, fmt.Errorf("type missing: %+v", compSpec)
	} else if compSpec.Fields["type"].GetStringValue() == "array" {

		if _, ok := compSpec.Fields["instillUIOrder"]; !ok {
			compSpec.Fields["instillUIOrder"] = structpb.NewNumberValue(0)
		}

		converted, err := formatDataSpec(compSpec.Fields["items"].GetStructValue())
		if err != nil {
			return nil, err
		}
		compSpec.Fields["items"] = structpb.NewStructValue(converted)
	} else if compSpec.Fields["type"].GetStringValue() == "object" {

		if _, ok := compSpec.Fields["instillUIOrder"]; !ok {
			compSpec.Fields["instillUIOrder"] = structpb.NewNumberValue(0)
		}
		if _, ok := compSpec.Fields["required"]; !ok {
			return nil, fmt.Errorf("required missing: %+v", compSpec)
		}
		if _, ok := compSpec.Fields["instillEditOnNodeFields"]; !ok {
			compSpec.Fields["instillEditOnNodeFields"] = compSpec.Fields["required"]
		}

		if _, ok := compSpec.Fields["properties"]; ok {
			for k, v := range compSpec.Fields["properties"].GetStructValue().AsMap() {
				s, err := structpb.NewStruct(v.(map[string]interface{}))
				if err != nil {
					return nil, err
				}
				converted, err := formatDataSpec(s)
				if err != nil {
					return nil, err
				}
				compSpec.Fields["properties"].GetStructValue().Fields[k] = structpb.NewStructValue(converted)

			}
		}
		if _, ok := compSpec.Fields["patternProperties"]; ok {
			for k, v := range compSpec.Fields["patternProperties"].GetStructValue().AsMap() {
				s, err := structpb.NewStruct(v.(map[string]interface{}))
				if err != nil {
					return nil, err
				}
				converted, err := formatDataSpec(s)
				if err != nil {
					return nil, err
				}
				compSpec.Fields["patternProperties"].GetStructValue().Fields[k] = structpb.NewStructValue(converted)

			}
		}
		for _, target := range []string{"allOf", "anyOf", "oneOf"} {
			if _, ok := compSpec.Fields[target]; ok {
				for idx, item := range compSpec.Fields[target].GetListValue().AsSlice() {
					s, err := structpb.NewStruct(item.(map[string]interface{}))
					if err != nil {
						return nil, err
					}
					converted, err := formatDataSpec(s)
					if err != nil {
						return nil, err
					}
					compSpec.Fields[target].GetListValue().AsSlice()[idx] = structpb.NewStructValue(converted)
				}
			}
		}

	} else {
		if _, ok := compSpec.Fields["instillUIOrder"]; !ok {
			compSpec.Fields["instillUIOrder"] = structpb.NewNumberValue(0)
		}

		newCompSpec := &structpb.Struct{Fields: make(map[string]*structpb.Value)}

		newCompSpec.Fields["type"] = structpb.NewStringValue(compSpec.Fields["type"].GetStringValue())
		newCompSpec.Fields["title"] = structpb.NewStringValue(compSpec.Fields["title"].GetStringValue())
		newCompSpec.Fields["description"] = structpb.NewStringValue(compSpec.Fields["description"].GetStringValue())
		if _, ok := newCompSpec.Fields["instillShortDescription"]; ok {
			newCompSpec.Fields["instillShortDescription"] = compSpec.Fields["instillShortDescription"]
		} else {
			newCompSpec.Fields["instillShortDescription"] = newCompSpec.Fields["description"]
		}
		newCompSpec.Fields["instillUIOrder"] = structpb.NewNumberValue(compSpec.Fields["instillUIOrder"].GetNumberValue())
		if compSpec.Fields["instillFormat"] != nil {
			newCompSpec.Fields["instillFormat"] = structpb.NewStringValue(compSpec.Fields["instillFormat"].GetStringValue())
		}

		compSpec = newCompSpec

	}
	return compSpec, nil
}

func generateDataSpecs(tasks map[string]*structpb.Struct) (map[string]*pipelinePB.DataSpecification, error) {

	specs := map[string]*pipelinePB.DataSpecification{}
	for k := range tasks {
		spec := &pipelinePB.DataSpecification{}
		var err error
		taskJSONStruct := proto.Clone(tasks[k]).(*structpb.Struct)
		spec.Input, err = formatDataSpec(taskJSONStruct.Fields["input"].GetStructValue())
		if err != nil {
			return nil, err
		}
		spec.Output, err = formatDataSpec(taskJSONStruct.Fields["output"].GetStructValue())
		if err != nil {
			return nil, err
		}
		specs[k] = spec
	}

	return specs, nil
}

func loadTasks(availableTasks []string, tasksJSONBytes []byte) ([]*pipelinePB.ComponentTask, map[string]*structpb.Struct, error) {

	taskStructs := map[string]*structpb.Struct{}
	var err error

	tasksJSONMap := map[string]map[string]interface{}{}
	err = json.Unmarshal(tasksJSONBytes, &tasksJSONMap)
	if err != nil {
		return nil, nil, err
	}

	for _, t := range availableTasks {
		if v, ok := tasksJSONMap[t]; ok {
			taskStructs[t], err = structpb.NewStruct(v)
			if err != nil {
				return nil, nil, err
			}

		}
	}
	tasks := generateComponentTaskCards(taskStructs)
	return tasks, taskStructs, nil
}

// ConvertFromStructpb converts from structpb.Struct to a struct
func ConvertFromStructpb(from *structpb.Struct, to interface{}) error {
	inputJSON, err := protojson.Marshal(from)
	if err != nil {
		return err
	}

	err = json.Unmarshal(inputJSON, to)
	if err != nil {
		return err
	}
	return nil
}

// ConvertToStructpb converts from a struct to structpb.Struct
func ConvertToStructpb(from interface{}) (*structpb.Struct, error) {
	to := &structpb.Struct{}
	outputJSON, err := json.Marshal(from)
	if err != nil {
		return nil, err
	}

	err = protojson.Unmarshal(outputJSON, to)
	if err != nil {
		return nil, err
	}
	return to, nil
}

func RenderJSON(tasksJSONBytes []byte, additionalJSONBytes map[string][]byte) ([]byte, error) {
	var err error
	mp := provider.NewMap()
	for k, v := range additionalJSONBytes {
		var i interface{}
		err = json.Unmarshal(v, &i)
		if err != nil {
			return nil, err
		}
		err = mp.Set(k, i)
		if err != nil {
			return nil, err
		}
	}
	res := jsonref.New()
	err = res.AddProvider(mp)
	if err != nil {
		return nil, err
	}
	err = res.AddProvider(provider.NewHTTP())
	if err != nil {
		return nil, err
	}

	var tasksJSON interface{}
	err = json.Unmarshal(tasksJSONBytes, &tasksJSON)
	if err != nil {
		return nil, err
	}

	result, err := res.Resolve(tasksJSON, "", jsonref.WithRecursiveResolution(true))
	if err != nil {
		return nil, err
	}
	renderedTasksJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return renderedTasksJSON, nil

}

// For formats such as `*`, `semi-structured/*`, and `semi-structured/json` we
// treat them as freeform data. Thus, there is no need to set the `type` in the
// JSON schema.
func checkFreeForm(compSpec *structpb.Struct) bool {
	acceptFormats := compSpec.Fields["instillAcceptFormats"].GetListValue().AsSlice()

	formats := make([]any, 0, len(acceptFormats)+1) // This avoids reallocations when appending values to the slice.
	formats = append(formats, acceptFormats...)

	if instillFormat := compSpec.Fields["instillFormat"].GetStringValue(); instillFormat != "" {
		formats = append(formats, instillFormat)
	}

	for _, v := range formats {
		if v.(string) == "*" || v.(string) == "semi-structured/*" || v.(string) == "semi-structured/json" {
			return true
		}
	}

	return false
}
