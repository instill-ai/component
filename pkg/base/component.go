package base

import (
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// All component need to implement this interface
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
	generateComponentSpec(title string, availableTasks []string) (*structpb.Struct, error)
	// Generate OpenAPI Specifications
	generateOpenAPISpecs(title string, availableTasks []string) (*structpb.Struct, error)

	// Load tasks
	loadTasks(tasksJson []byte) error
	// Get task input schemas
	GetTaskInputSchemas() map[string]string
	// Get task output schemas
	GetTaskOutputSchemas() map[string]string
}

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
	if _, ok := compSpec.Fields["type"]; !ok || compSpec.Fields["type"].GetStringValue() == "object" {
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
		// var err error
		if _, ok := compSpec.Fields["instillFormat"]; !ok {
			return compSpec, nil
		}
		original := proto.Clone(compSpec).(*structpb.Struct)
		delete(original.Fields, "title")
		delete(original.Fields, "description")
		delete(original.Fields, "instillFormat")
		delete(original.Fields, "instillUpstreamTypes")

		newCompSpec := &structpb.Struct{Fields: make(map[string]*structpb.Value)}

		newCompSpec.Fields["title"] = structpb.NewStringValue(compSpec.Fields["title"].GetStringValue())
		newCompSpec.Fields["description"] = structpb.NewStringValue(compSpec.Fields["description"].GetStringValue())
		newCompSpec.Fields["instillFormat"] = structpb.NewStringValue(compSpec.Fields["instillFormat"].GetStringValue())
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

		}

		compSpec = newCompSpec

	}
	return compSpec, nil
}

func (comp *Component) generateComponentSpec(title string, availableTasks []string) (*structpb.Struct, error) {
	var err error
	componentSpec := &structpb.Struct{Fields: map[string]*structpb.Value{}}
	componentSpec.Fields["$schema"] = structpb.NewStringValue("http://json-schema.org/draft-07/schema#")
	componentSpec.Fields["title"] = structpb.NewStringValue(fmt.Sprintf("%s Component", title))
	componentSpec.Fields["type"] = structpb.NewStringValue("object")

	if err != nil {
		return nil, err
	}

	if len(availableTasks) > 1 {
		oneOfList := &structpb.ListValue{
			Values: []*structpb.Value{},
		}
		for _, availableTask := range availableTasks {

			oneOf := &structpb.Struct{Fields: map[string]*structpb.Value{}}
			oneOf.Fields["type"] = structpb.NewStringValue("object")
			oneOf.Fields["properties"] = structpb.NewStructValue(&structpb.Struct{Fields: make(map[string]*structpb.Value)})

			oneOf.Fields["properties"].GetStructValue().Fields["task"], err = structpb.NewValue(map[string]interface{}{
				"const": availableTask,
			})
			if err != nil {
				return nil, err
			}

			taskJsonStruct := proto.Clone(comp.tasks[availableTask]).(*structpb.Struct).Fields["input"].GetStructValue()

			compInputStruct, err := convertDataSpecToCompSpec(taskJsonStruct)
			if err != nil {
				return nil, err
			}
			oneOf.Fields["properties"].GetStructValue().Fields["input"] = structpb.NewStructValue(compInputStruct)
			if comp.tasks[availableTask].Fields["metadata"] != nil {
				metadataStruct := proto.Clone(comp.tasks[availableTask]).(*structpb.Struct).Fields["metadata"].GetStructValue()
				oneOf.Fields["properties"].GetStructValue().Fields["metadata"] = structpb.NewStructValue(metadataStruct)
			}

			// oneOf
			oneOfList.Values = append(oneOfList.Values, structpb.NewStructValue(oneOf))
		}

		componentSpec.Fields["oneOf"] = structpb.NewListValue(oneOfList)

		if err != nil {
			return nil, err
		}
	} else {

		taskJsonStruct := proto.Clone(comp.tasks[availableTasks[0]]).(*structpb.Struct).Fields["input"].GetStructValue()
		compInputStruct, err := convertDataSpecToCompSpec(taskJsonStruct)
		if err != nil {
			return nil, err
		}
		c := &structpb.Struct{Fields: map[string]*structpb.Value{}}
		c.Fields["input"] = structpb.NewStructValue(compInputStruct)
		componentSpec.Fields["properties"] = structpb.NewStructValue(c)
		if comp.tasks[availableTasks[0]].Fields["metadata"] != nil {
			metadataStruct := proto.Clone(comp.tasks[availableTasks[0]]).(*structpb.Struct).Fields["metadata"].GetStructValue()
			componentSpec.Fields["properties"].GetStructValue().Fields["metadata"] = structpb.NewStructValue(metadataStruct)
		}
	}

	return componentSpec, nil

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

		taskJsonStruct := proto.Clone(comp.tasks[availableTask]).(*structpb.Struct)
		walk.Fields["items"] = taskJsonStruct.Fields["input"]

		walk = openAPITemplate
		for _, key := range []string{"paths", "/execute", "post", "responses", "200", "content", "application/json", "schema", "properties", "outputs"} {
			walk = walk.Fields[key].GetStructValue()
		}

		walk.Fields["items"] = taskJsonStruct.Fields["output"]

		openAPITemplates.Fields[availableTask] = structpb.NewStructValue(openAPITemplate)
	}

	return openAPITemplates, nil
}

func (comp *Component) loadTasks(tasksJsonBytes []byte) error {

	var err error

	tasksJsonMap := map[string]map[string]interface{}{}
	err = json.Unmarshal(tasksJsonBytes, &tasksJsonMap)
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

	for k, v := range tasksJsonMap {
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

func (comp *Component) GetTaskInputSchemas() map[string]string {
	return comp.taskInputSchemas
}
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
			comp.Logger.Error("get connector definition error")
		}
		definitions = append(definitions, val)
	}

	return definitions
}

func (comp *Component) getDefinitionByUID(defUID uuid.UUID) (interface{}, error) {
	val, ok := comp.definitionMapByUID[defUID]
	if !ok {
		return nil, fmt.Errorf("get connector definition error")
	}
	return val, nil
}

func (comp *Component) getDefinitionByID(defID string) (interface{}, error) {

	val, ok := comp.definitionMapByID[defID]
	if !ok {
		return nil, fmt.Errorf("get connector definition error")
	}
	return val, nil
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
