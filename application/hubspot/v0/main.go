//go:generate compogen readme ./config ./README.mdx

package hubspot

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"sync"

	hubspot "github.com/belong-inc/go-hubspot"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	taskGetContact          = "TASK_GET_CONTACT"
	taskCreateContact       = "TASK_CREATE_CONTACT"
	taskGetDeal             = "TASK_GET_DEAL"
	taskCreateDeal          = "TASK_CREATE_DEAL"
	taskGetThread           = "TASK_GET_THREAD"
	taskRetrieveAssociation = "TASK_RETRIEVE_ASSOCIATION"
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
	client *CustomClient
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

func getToken(setup *structpb.Struct) string {
	return setup.GetFields()["token"].GetStringValue()
}

// custom client to support thread task
func hubspotNewCustomClient(setup *structpb.Struct) *CustomClient {
	client, err := NewCustomClient(hubspot.SetPrivateAppToken(getToken(setup)))

	if err != nil {
		panic(err)
	}

	return client
}

func (c *component) CreateExecution(sysVars map[string]any, setup *structpb.Struct, task string) (*base.ExecutionWrapper, error) {

	e := &execution{
		ComponentExecution: base.ComponentExecution{Component: c, SystemVariables: sysVars, Task: task},
		client:             hubspotNewCustomClient(setup),
	}

	return &base.ExecutionWrapper{Execution: e}, nil
}

func (e *execution) Execute(_ context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {

	outputs := []*structpb.Struct{}

	for _, input := range inputs {

		switch e.Task {
		case taskGetContact:

			uniqueKey := input.Fields["contact-id-or-email"].GetStringValue()

			// If user enter email instead of contact ID
			if strings.Contains(uniqueKey, "@") {
				uniqueKey += "?idProperty=email"
			}

			res, err := e.client.CRM.Contact.Get(uniqueKey, &ContactInfoHSFormat{}, &hubspot.RequestQueryOption{CustomProperties: []string{"phone"}})

			if err != nil {
				return nil, err
			}

			contactInfo := res.Properties.(*ContactInfoHSFormat)

			// convert to another struct in order to utilize base.ConvertToStructpb
			contactInfoOutput := ContactInfoTaskFormat(*contactInfo)

			output, err := base.ConvertToStructpb(contactInfoOutput)
			if err != nil {
				return nil, err
			}

			outputs = append(outputs, output)

		case taskCreateContact:
			contactInfoInput := ContactInfoTaskFormat{}

			err := base.ConvertFromStructpb(input, &contactInfoInput)

			if err != nil {
				return nil, err
			}

			contactInfoReq := ContactInfoHSFormat(contactInfoInput)

			res, err := e.client.CRM.Contact.Create(&contactInfoReq)

			if err != nil {
				return nil, err
			}

			output := new(structpb.Struct)
			output.Fields = map[string]*structpb.Value{
				"contact-id": structpb.NewStringValue(res.Properties.(*ContactInfoHSFormat).ContactId),
			}

			outputs = append(outputs, output)

		case taskGetDeal:

			uniqueKey := input.Fields["deal-id"].GetStringValue()

			res, err := e.client.CRM.Deal.Get(uniqueKey, &DealInfoHSFormat{}, nil)

			if err != nil {
				return nil, err
			}

			dealInfo := res.Properties.(*DealInfoHSFormat)

			// convert to another struct in order to utilize base.ConvertToStructpb
			dealInfoOutput := DealInfoTaskFormat(*dealInfo)

			output, err := base.ConvertToStructpb(dealInfoOutput)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, output)

		case taskCreateDeal:

			dealInfoInput := DealInfoTaskFormat{}
			err := base.ConvertFromStructpb(input, &dealInfoInput)

			if err != nil {
				return nil, err
			}
			dealInfoReq := DealInfoHSFormat(dealInfoInput)

			res, err := e.client.CRM.Deal.Create(&dealInfoReq)

			if err != nil {
				return nil, err
			}

			// get deal Id
			dealId := res.Properties.(*DealInfoHSFormat).DealId

			output := new(structpb.Struct)
			output.Fields = map[string]*structpb.Value{
				"deal-id": structpb.NewStringValue(dealId),
			}

			// this section of the code is used to associate contact with deal if there is any
			uniqueKey := input.Fields["contact-id-or-email"].GetStringValue()

			if uniqueKey != "" {
				if strings.Contains(uniqueKey, "@") {
					uniqueKey += "?idProperty=email"

					//get contact id first using email
					res, err := e.client.CRM.Contact.Get(uniqueKey, &ContactIDHSFormat{}, nil)

					if err != nil {
						return nil, err
					}

					uniqueKey = res.Properties.(*ContactIDHSFormat).ContactId

				}

				e.client.CRM.Contact.AssociateAnotherObj(uniqueKey, &hubspot.AssociationConfig{
					ToObject:   hubspot.ObjectTypeDeal,
					ToObjectID: dealId,
					Type:       hubspot.AssociationTypeContactToDeal,
				})
			}

			outputs = append(outputs, output)

		case taskGetThread:
			uniqueKey := input.Fields["thread-id"].GetStringValue()

			res, err := e.client.Thread.Get(uniqueKey)

			if err != nil {
				return nil, err
			}

			outputTaskFormat := e.client.Thread.ConvertToTaskFormat(res)

			output, err := base.ConvertToStructpb(outputTaskFormat)
			if err != nil {
				return nil, err
			}

			outputs = append(outputs, output)

		case taskRetrieveAssociation:

			retrieveInput := RetrieveAssociationInput{}
			err := base.ConvertFromStructpb(input, &retrieveInput)

			var output *structpb.Struct

			switch retrieveInput.ObjectType {
			case "Threads":
				// To handle Threads
				res, err := e.client.RetrieveAssociation.GetThreadId(retrieveInput.ContactId)

				if err != nil {
					return nil, err
				}

				output, err = base.ConvertToStructpb(res)

				if err != nil {
					return nil, err
				}

			default:
				// To handle CRM objects
				res, err := e.client.RetrieveAssociation.GetCrmId(retrieveInput.ContactId, retrieveInput.ObjectType)

				if err != nil {
					return nil, err
				}

				var crmIdOutput RetrieveCrmIdResultTaskFormat
				if len(res.Results) != 0 {
					// convert to another struct in order to utilize base.ConvertToStructpb
					// Only uses Results[0] because the input is only one contact ID
					crmIdOutput = RetrieveCrmIdResultTaskFormat(res.Results[0])
				} else {
					// if there is no object ID associated with contact ID, assign an empty array (to prevent nil error)
					crmIdOutput = RetrieveCrmIdResultTaskFormat{
						IdArray: []RetrieveCrmId{},
					}
				}
				output, err = base.ConvertToStructpb(crmIdOutput)

				if err != nil {
					return nil, err
				}

			}

			outputs = append(outputs, output)

			if err != nil {
				return nil, err
			}

		default:
			return nil, fmt.Errorf("unsupported task")
		}
	}

	return outputs, nil
}
