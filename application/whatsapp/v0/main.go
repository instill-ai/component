package whatsapp

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	_ "embed"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
)

const (
	taskSendTemplate = "TASK_SEND_TEMPLATE"

	basePath = "https://graph.facebook.com"
	version  = "v20.0"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	//go:embed config/setup.json
	setupJSON []byte

	once sync.Once
	comp *component
)

type component struct {
	base.Component
}

type execution struct {
	base.ComponentExecution
	client  WhatsappInterface
	execute func(*structpb.Struct) (*structpb.Struct, error)
}

// Init returns an implementation of IComponent that implements the greeting
// task.
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
	e := &execution{
		ComponentExecution: base.ComponentExecution{Component: c, SystemVariables: sysVars, Task: task},
		client:             newClient(setup, c.GetLogger()),
	}

	switch task {
	case taskSendTemplate:
		e.execute = e.SendTemplate
	default:
		return nil, fmt.Errorf("unsupported task")
	}

	return &base.ExecutionWrapper{Execution: e}, nil
}

func (e *execution) Execute(_ context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := make([]*structpb.Struct, len(inputs))

	// An execution  might take several inputs. One result will be returned for
	// each one of them, containing the execution output for that set of
	// parameters.
	for i, input := range inputs {
		output, err := e.execute(input)
		if err != nil {
			return nil, err
		}

		outputs[i] = output
	}

	return outputs, nil
}

func (e *execution) SendTemplate(in *structpb.Struct) (*structpb.Struct, error) {

	inputStruct := TemplateInput{}
	err := base.ConvertFromStructpb(in, &inputStruct)

	if err != nil {
		return nil, err
	}

	req := MessageObjectReq{
		MessagingProduct: "whatsapp",
		To:               inputStruct.To,
		Type:             "template",
		Template: &TemplateObject{
			Name: inputStruct.TemplateName,
			Language: LanguageObject{
				Code: inputStruct.LanguageCode,
			},
		},
	}

	// to assign HeaderParameters
	// Header can have 6 types: none, text, image, video, document & location.

	// create a header component

	var component *ComponentObject
	switch inputStruct.HeaderType {
	case "text":
		// the text header type can have 0 parameter, so there is no need to have an error message if there is no parameter
		if len(inputStruct.HeaderParameters) != 0 {
			component = &ComponentObject{
				Type:       "header",
				Parameters: make([]interface{}, len(inputStruct.HeaderParameters)),
			}

			for index, value := range inputStruct.HeaderParameters {
				component.Parameters[index] = TextParameter{
					Type: "text",
					Text: value,
				}
			}
		}

	case "image":

		// the image header type only has 1 parameter, which is the id/link of the image
		if len(inputStruct.HeaderParameters) == 1 {

			component = &ComponentObject{
				Type:       "header",
				Parameters: make([]interface{}, 1),
			}

			firstParam := inputStruct.HeaderParameters[0]
			if strings.Contains(firstParam, "http") {
				component.Parameters[0] = ImageParameter{
					Type: "image",
					Image: MediaObject{
						Link: firstParam,
					},
				}
			} else {
				component.Parameters[0] = ImageParameter{
					Type: "image",
					Image: MediaObject{
						Id: firstParam,
					},
				}
			}
		} else {
			return nil, fmt.Errorf("the image header type requires one parameter (in the header parameters), which is the id/link of the image. format: [id/link]")
		}

	case "video":
		// the video header type only has 1 parameter, which is the id/link of the video

		if len(inputStruct.HeaderParameters) == 1 {
			component = &ComponentObject{
				Type:       "header",
				Parameters: make([]interface{}, 1),
			}

			firstParam := inputStruct.HeaderParameters[0]
			if strings.Contains(firstParam, "http") {
				component.Parameters[0] = VideoParameter{
					Type: "video",
					Video: MediaObject{
						Link: firstParam,
					},
				}
			} else {
				component.Parameters[0] = VideoParameter{
					Type: "video",
					Video: MediaObject{
						Id: firstParam,
					},
				}
			}
		} else {
			return nil, fmt.Errorf("the video header type requires one parameter (in the header parameters), which is the id/link of the video. format: [id/link]")
		}

	case "document":

		if len(inputStruct.HeaderParameters) == 1 || len(inputStruct.HeaderParameters) == 2 {

			component = &ComponentObject{
				Type: "header",
			}

			firstParam := inputStruct.HeaderParameters[0]

			var secondParam string
			if len(inputStruct.HeaderParameters) > 1 {
				secondParam = inputStruct.HeaderParameters[1]
				component.Parameters = make([]interface{}, 2)
			} else {
				component.Parameters = make([]interface{}, 1)
			}

			if strings.Contains(firstParam, "http") {
				component.Parameters[0] = DocumentParameter{
					Type: "document",
					Document: MediaObject{
						Link:     firstParam,
						Filename: secondParam,
					},
				}
			} else {
				component.Parameters[0] = DocumentParameter{
					Type: "document",
					Document: MediaObject{
						Id:       firstParam,
						Filename: secondParam,
					},
				}
			}
		} else {
			return nil, fmt.Errorf("the document header type requires one or two parameter (in the header parameters). The first parameter is the id/link of the document (required), and the second parameter is the filename which can be used to specify the extension of the file as well (optional). format: [id/link, filename]")
		}

	case "location":

		if len(inputStruct.HeaderParameters) == 4 {

			component = &ComponentObject{
				Type:       "header",
				Parameters: make([]interface{}, 1),
			}

			component.Parameters[0] =
				LocationParameter{
					Type: "location",
					Location: LocationObject{
						Latitude:  inputStruct.HeaderParameters[0],
						Longitude: inputStruct.HeaderParameters[1],
						Name:      inputStruct.HeaderParameters[2],
						Address:   inputStruct.HeaderParameters[3],
					},
				}
		} else {
			return nil, fmt.Errorf("the location header type requires 4 parameters which are: latitude, longitude, name, address. format: [latitude, longitude, name, address]")
		}

	}

	req.Template.Components = append(req.Template.Components, component)

	// create a body component if there is any body parameters

	if len(inputStruct.BodyParameters) != 0 {
		component := &ComponentObject{
			Type:       "body",
			Parameters: make([]interface{}, len(inputStruct.BodyParameters)),
		}

		for index, value := range inputStruct.BodyParameters {
			component.Parameters[index] = TextParameter{
				Type: "text",
				Text: value,
			}
		}

		req.Template.Components = append(req.Template.Components, component)
	}

	// create button component if there is any
	// one parameter -> one button component

	for index, value := range inputStruct.ButtonParameters {
		splitParam := strings.Split(value, ";")

		if len(splitParam) != 2 {
			return nil, fmt.Errorf("format is wrong, it must be 'parameter_type;value_of_the_parameter'. Example: payload;randomvalue")
		}

		var param ButtonParameter
		if splitParam[0] == "payload" {
			param = ButtonParameter{
				Type:    "payload",
				Payload: splitParam[1],
			}
		} else if splitParam[0] == "url" {
			param = ButtonParameter{
				Type: "url",
				Text: splitParam[1],
			}

		} else {
			return nil, fmt.Errorf("wrong parameter_type. parameter_type is either 'payload' or 'url'")
		}

		component := &ComponentObject{
			Type:          "button",
			ButtonSubType: splitParam[0],
			ButtonIndex:   strconv.Itoa(index),
		}

		component.Parameters = append(component.Parameters, param)

		req.Template.Components = append(req.Template.Components, component)
	}

	resp, err := e.client.SendMessageAPI(&req, inputStruct.PhoneNumberId)

	if err != nil {
		return nil, err
	}

	// only take the first index because we are sending a template to an individual, so there will only be one contact and one message.
	outputStruct := SendTemplateOutput{
		WaId:          resp.Contacts[0].WaId,
		Id:            resp.Messages[0].Id,
		MessageStatus: resp.Messages[0].MessageStatus,
	}

	output, err := base.ConvertToStructpb(outputStruct)

	if err != nil {
		return nil, err
	}

	return output, nil
}
