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

	// Functions that need to be implemented in each component implementation

	// Create a execution by definition uid and component configuration
	CreateExecution(defUID uuid.UUID, task string, config *structpb.Struct, logger *zap.Logger) (IExecution, error)

	// Internal functions

	// Get the list of definitions under this connector
	listDefinitions() []interface{}
	// Add definition
	addDefinition(def interface{}) error
	// Get the definition by definition uid
	getDefinitionByUID(defUID uuid.UUID) (interface{}, error)
	// Get the definition by definition id
	getDefinitionByID(defID string) (interface{}, error)

	// Generate Component Specification
	generateComponentSpec(title string, availableTasks []*pipelinePB.ComponentTask) (*structpb.Struct, error)
	// Generate OpenAPI Specifications
	generateOpenAPISpecs(title string, availableTasks []string) (*structpb.Struct, error)

	// Load tasks
	loadTasks(tasksJSON []byte) error
	// Get task input schemas
	GetTaskInputSchemas() map[string]string
	// Get task output schemas
	GetTaskOutputSchemas() map[string]string
}

// Component is the basic component struct
type Component struct {
	Name string

	// Store all the component definitions in the component
	definitionMapByUID map[uuid.UUID]interface{}
	definitionMapByID  map[string]interface{}

	// Used for ordered
	definitionUIDs []uuid.UUID

	tasks             map[string]*structpb.Struct
	taskInputSchemas  map[string]string
	taskOutputSchemas map[string]string

	// Logger
	Logger *zap.Logger
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

func (comp *Component) generateComponentTasks(availableTasks []string) []*pipelinePB.ComponentTask {
	tasks := make([]*pipelinePB.ComponentTask, 0, len(availableTasks))
	for _, t := range availableTasks {
		title := comp.tasks[t].Fields["title"].GetStringValue()
		if title == "" {
			title = TaskIDToTitle(t)
		}

		description := comp.tasks[t].Fields["instillShortDescription"].GetStringValue()

		tasks = append(tasks, &pipelinePB.ComponentTask{
			Name:        t,
			Title:       title,
			Description: description,
		})
	}

	return tasks
}

func (comp *Component) generateComponentSpec(title string, availableTasks []*pipelinePB.ComponentTask) (*structpb.Struct, error) {
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
	for _, availableTask := range availableTasks {
		taskName := availableTask.Name

		oneOf := &structpb.Struct{Fields: map[string]*structpb.Value{}}
		oneOf.Fields["type"] = structpb.NewStringValue("object")
		oneOf.Fields["properties"] = structpb.NewStructValue(&structpb.Struct{Fields: make(map[string]*structpb.Value)})

		oneOf.Fields["properties"].GetStructValue().Fields["task"], err = structpb.NewValue(map[string]interface{}{
			"const": availableTask.Name,
			"title": availableTask.Title,
		})
		if err != nil {
			return nil, err
		}

		if comp.tasks[taskName].Fields["description"].GetStringValue() != "" {
			oneOf.Fields["properties"].GetStructValue().Fields["task"].GetStructValue().Fields["description"] = structpb.NewStringValue(comp.tasks[taskName].Fields["description"].GetStringValue())
		}

		if availableTask.Description != "" {
			oneOf.Fields["properties"].GetStructValue().Fields["task"].GetStructValue().Fields["instillShortDescription"] = structpb.NewStringValue(availableTask.Description)
		}
		taskJSONStruct := proto.Clone(comp.tasks[taskName]).(*structpb.Struct).Fields["input"].GetStructValue()

		compInputStruct, err := convertDataSpecToCompSpec(taskJSONStruct)
		if err != nil {
			return nil, fmt.Errorf("task %s: %s error: %+v", title, availableTask, err)
		}

		condition := &structpb.Struct{}
		err = protojson.Unmarshal([]byte(conditionJSON), condition)
		if err != nil {
			panic(err)
		}
		oneOf.Fields["properties"].GetStructValue().Fields["condition"] = structpb.NewStructValue(condition)
		oneOf.Fields["properties"].GetStructValue().Fields["input"] = structpb.NewStructValue(compInputStruct)
		if comp.tasks[taskName].Fields["metadata"] != nil {
			metadataStruct := proto.Clone(comp.tasks[taskName]).(*structpb.Struct).Fields["metadata"].GetStructValue()
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

func convertDataSpecToOpenAPISpec(dataSpec *structpb.Struct) (*structpb.Struct, error) {
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

		converted, err := convertDataSpecToOpenAPISpec(compSpec.Fields["items"].GetStructValue())
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
				converted, err := convertDataSpecToOpenAPISpec(s)
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
				converted, err := convertDataSpecToOpenAPISpec(s)
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
					converted, err := convertDataSpecToOpenAPISpec(s)
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

func (comp *Component) generateOpenAPISpecs(title string, availableTasks []string) (*structpb.Struct, error) {

	openAPITemplates := &structpb.Struct{Fields: map[string]*structpb.Value{}}
	for _, availableTask := range availableTasks {
		openAPITemplate := &structpb.Struct{}

		err := protojson.Unmarshal([]byte(OpenAPITemplate), openAPITemplate)
		if err != nil {
			return nil, err
		}
		openAPITemplate.Fields["info"].GetStructValue().Fields["title"] = structpb.NewStringValue(fmt.Sprintf("%s Component - %s", title, availableTask))

		walk := openAPITemplate
		for _, key := range []string{"paths", "/execute", "post", "requestBody", "content", "application/json", "schema", "properties", "inputs"} {
			walk = walk.Fields[key].GetStructValue()
		}

		taskJSONStruct := proto.Clone(comp.tasks[availableTask]).(*structpb.Struct)

		inputStruct, err := convertDataSpecToOpenAPISpec(taskJSONStruct.Fields["input"].GetStructValue())
		if err != nil {
			return nil, fmt.Errorf("task %s: %s error: %+v", title, availableTask, err)
		}
		walk.Fields["items"] = structpb.NewStructValue(inputStruct)

		walk = openAPITemplate
		for _, key := range []string{"paths", "/execute", "post", "responses", "200", "content", "application/json", "schema", "properties", "outputs"} {
			walk = walk.Fields[key].GetStructValue()
		}

		outputStruct, err := convertDataSpecToOpenAPISpec(taskJSONStruct.Fields["output"].GetStructValue())
		if err != nil {
			return nil, fmt.Errorf("task %s: %s error: %+v", title, availableTask, err)
		}
		walk.Fields["items"] = structpb.NewStructValue(outputStruct)

		openAPITemplates.Fields[availableTask] = structpb.NewStructValue(openAPITemplate)
	}

	return openAPITemplates, nil
}

func (comp *Component) loadTasks(tasksJSONBytes []byte) error {

	var err error

	tasksJSONMap := map[string]map[string]interface{}{}
	err = json.Unmarshal(tasksJSONBytes, &tasksJSONMap)
	if err != nil {
		return err
	}

	if comp.tasks == nil {
		comp.tasks = map[string]*structpb.Struct{}
	}
	if comp.taskInputSchemas == nil {
		comp.taskInputSchemas = map[string]string{}
	}
	if comp.taskOutputSchemas == nil {
		comp.taskOutputSchemas = map[string]string{}
	}

	for k, v := range tasksJSONMap {
		if k != "$defs" {
			comp.tasks[k], err = structpb.NewStruct(v)
			if err != nil {
				return err
			}
			var s []byte
			s, err = protojson.Marshal(comp.tasks[k].Fields["input"].GetStructValue())
			if err != nil {
				return err
			}
			comp.taskInputSchemas[k] = string(s)

			s, err = protojson.Marshal(comp.tasks[k].Fields["output"].GetStructValue())
			if err != nil {
				return err
			}
			comp.taskOutputSchemas[k] = string(s)
		}
	}
	return nil
}

// GetTaskInputSchemas returns the task input schemas
func (comp *Component) GetTaskInputSchemas() map[string]string {
	return comp.taskInputSchemas
}

// GetTaskOutputSchemas returns the task output schemas
func (comp *Component) GetTaskOutputSchemas() map[string]string {
	return comp.taskOutputSchemas
}

func (comp *Component) addDefinition(def interface{}) error {

	type definition interface {
		GetId() string
		GetUid() string
	}

	if comp.definitionMapByUID == nil {
		comp.definitionMapByUID = map[uuid.UUID]interface{}{}
	}
	if comp.definitionMapByID == nil {
		comp.definitionMapByID = map[string]interface{}{}
	}
	uid := uuid.FromStringOrNil(def.(definition).GetUid())
	id := def.(definition).GetId()
	comp.definitionUIDs = append(comp.definitionUIDs, uid)
	comp.definitionMapByUID[uid] = def
	comp.definitionMapByID[id] = def
	return nil
}

func (comp *Component) listDefinitions() []interface{} {
	definitions := []interface{}{}
	for _, uid := range comp.definitionUIDs {
		val, ok := comp.definitionMapByUID[uid]
		if !ok {
			// logger
			comp.Logger.Error("get connector/operator definition error")
		}
		definitions = append(definitions, val)
	}

	return definitions
}

func (comp *Component) getDefinitionByUID(defUID uuid.UUID) (interface{}, error) {
	val, ok := comp.definitionMapByUID[defUID]
	if !ok {
		return nil, fmt.Errorf("component definition UID doesn't exist")
	}
	return val, nil
}

func (comp *Component) getDefinitionByID(defID string) (interface{}, error) {

	val, ok := comp.definitionMapByID[defID]
	if !ok {
		return nil, fmt.Errorf("component definition ID doesn't exist")
	}
	return val, nil
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

func renderTaskJSON(tasksJSONBytes []byte, additionalJSONBytes map[string][]byte) ([]byte, error) {
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
