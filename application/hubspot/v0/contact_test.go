package hubspot

import (
	"context"
	"testing"

	hubspot "github.com/belong-inc/go-hubspot"
	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	bearerToken = "123"
)

// mockClient is a custom client that will be used for testing

func createMockClient() *CustomClient {

	mockCRM := &hubspot.CRM{
		Contact: &MockContact{},
		Deal:    &MockDeal{},
		Company: &MockCompany{},
	}

	mockClient := &CustomClient{
		Client: &hubspot.Client{
			CRM: mockCRM,
		},
		Thread:              &MockThread{},
		RetrieveAssociation: &MockRetrieveAssociation{},
		Ticket:              &MockTicket{},
	}

	return mockClient
}

// Mock Contact struct and its functions
type MockContact struct{}

func (s *MockContact) Get(contactID string, contact interface{}, option *hubspot.RequestQueryOption) (*hubspot.ResponseResource, error) {

	var fakeContact TaskGetContactResp
	if contactID == "32027696539" || contactID == "bh@hubspot.com" {

		fakeContact = TaskGetContactResp{
			FirstName:      "Brian",
			LastName:       "Halligan (Sample Contact)",
			Email:          "bh@hubspot.com",
			Company:        "HubSpot",
			JobTitle:       "CEO",
			LifecycleStage: "lead",
			ContactId:      "32027696539",
		}
	}
	ret := &hubspot.ResponseResource{
		Properties: &fakeContact,
	}

	return ret, nil
}

func (s *MockContact) Create(contact interface{}) (*hubspot.ResponseResource, error) {

	// in the actual create function, if the user created a contact, it will return all the information back to the user, so I will be mimicking that

	arbitraryContactId := "12345678"

	fakeContactInfo := contact.(*TaskCreateContactReq)

	fakeContactInfo.ContactId = arbitraryContactId

	ret := &hubspot.ResponseResource{
		Properties: fakeContactInfo,
	}

	return ret, nil
}

func (s *MockContact) Update(contactID string, contact interface{}) (*hubspot.ResponseResource, error) {
	return nil, nil
}
func (s *MockContact) Delete(contactID string) error {
	return nil
}
func (s *MockContact) AssociateAnotherObj(contactID string, conf *hubspot.AssociationConfig) (*hubspot.ResponseResource, error) {
	return nil, nil
}

func TestComponent_ExecuteGetContactTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	tc := struct {
		name     string
		input    string
		wantResp TaskGetContactOutput
	}{
		name:  "ok - get contact",
		input: "32027696539",
		wantResp: TaskGetContactOutput{
			FirstName:      "Brian",
			LastName:       "Halligan (Sample Contact)",
			Email:          "bh@hubspot.com",
			Company:        "HubSpot",
			JobTitle:       "CEO",
			LifecycleStage: "lead",
			ContactId:      "32027696539",
		},
	}

	c.Run(tc.name, func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"token": bearerToken,
		})
		c.Assert(err, qt.IsNil)

		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: taskGetContact},
			client:             createMockClient(),
		}

		e.execute = e.GetContact
		exec := &base.ExecutionWrapper{Execution: e}

		pbInput, err := structpb.NewStruct(map[string]any{
			"contact-id-or-email": tc.input,
		})

		c.Assert(err, qt.IsNil)

		res, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbInput})

		c.Assert(err, qt.IsNil)

		resJSON, err := protojson.Marshal(res[0])
		c.Assert(err, qt.IsNil)

		c.Check(resJSON, qt.JSONEquals, tc.wantResp)

	})
}

func TestComponent_ExecuteCreateContactTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	tc := struct {
		name     string
		input    TaskCreateContactInput
		wantResp string
	}{
		name: "ok - create contact",
		input: TaskCreateContactInput{
			FirstName: "Test",
			LastName:  "Name",
			Email:     "test_name@gmail.com",
		},
		wantResp: "12345678",
	}

	c.Run(tc.name, func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"token": bearerToken,
		})
		c.Assert(err, qt.IsNil)

		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: taskCreateContact},
			client:             createMockClient(),
		}
		e.execute = e.CreateContact
		exec := &base.ExecutionWrapper{Execution: e}

		pbInput, err := base.ConvertToStructpb(tc.input)

		c.Assert(err, qt.IsNil)

		res, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbInput})
		c.Assert(err, qt.IsNil)

		resString := res[0].Fields["contact-id"].GetStringValue()

		c.Check(resString, qt.Equals, tc.wantResp)

	})
}
