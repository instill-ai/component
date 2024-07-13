package hubspot

import (
	"context"
	"testing"

	"github.com/belong-inc/go-hubspot"
	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	bearerToken = "123"
)

// Mock Contact struct and its functions
type MockContact struct{}

func (s *MockContact) Get(contactID string, contact interface{}, option *hubspot.RequestQueryOption) (*hubspot.ResponseResource, error) {

	var fakeContact ContactInfoHSFormat
	if contactID == "32027696539" || contactID == "bh@hubspot.com" {

		fakeContact = ContactInfoHSFormat{
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

	fakeContactInfo := contact.(*ContactInfoHSFormat)

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

// Mock Deal struct and its functions
type MockDeal struct{}

func (s *MockDeal) Get(dealID string, deal interface{}, option *hubspot.RequestQueryOption) (*hubspot.ResponseResource, error) {

	var fakeDeal DealInfoHSFormat
	if dealID == "20620806729" {
		fakeDeal = DealInfoHSFormat{
			DealName:   "Fake deal",
			Pipeline:   "default",
			DealStage:  "qualifiedtobuy",
			CreateDate: "2024-07-09T02:22:06.140Z",
			DealId:     "20620806729",
		}
	}

	ret := &hubspot.ResponseResource{
		Properties: &fakeDeal,
	}

	return ret, nil
}

func (s *MockDeal) Create(deal interface{}) (*hubspot.ResponseResource, error) {
	arbitraryDealId := "12345678900"

	fakeDealInfo := deal.(*DealInfoHSFormat)

	fakeDealInfo.DealId = arbitraryDealId

	ret := &hubspot.ResponseResource{
		Properties: fakeDealInfo,
	}

	return ret, nil
}

func (s *MockDeal) Update(dealID string, deal interface{}) (*hubspot.ResponseResource, error) {
	return nil, nil
}

func (s *MockDeal) AssociateAnotherObj(dealID string, conf *hubspot.AssociationConfig) (*hubspot.ResponseResource, error) {
	return nil, nil
}

// Mock Thread struct and its functions
type MockThread struct{}

func (s *MockThread) Get(threadId string) (*ThreadResponseHSFormat, error) {

	var fakeThread ThreadResponseHSFormat
	if threadId == "7509711154" {
		fakeThread = ThreadResponseHSFormat{
			Results: []ThreadResultHSFormat{
				{
					CreatedAt: "2024-07-02T10:42:15Z",
					Senders: []ThreadUserHSFormat{
						{
							Name: "Brian Halligan (Sample Contact)",
							DeliveryIdentifier: ThreadDeliveryIdentifier{
								Type:  "HS_EMAIL_ADDRESS",
								Value: "bh@hubspot.com",
							},
						},
					},
					Recipients: []ThreadUserHSFormat{
						{
							DeliveryIdentifier: ThreadDeliveryIdentifier{
								Type:  "HS_EMAIL_ADDRESS",
								Value: "fake_email@gmail.com",
							},
						},
					},
					Text:    "Just random content inside",
					Subject: "A fake message",
				},
			},
		}
	}

	return &fakeThread, nil
}

// Mock Retrieve Association struct and its functions

type MockRetrieveAssociation struct{}

func (s *MockRetrieveAssociation) GetThreadId(contactId string) (*RetrieveThreadIdResponse, error) {

	var fakeThreadId RetrieveThreadIdResponse
	if contactId == "32027696539" {
		fakeThreadId = RetrieveThreadIdResponse{
			Results: []RetrieveThreadIdResult{
				{
					Id: "7509711154",
				},
			},
		}
	}
	return &fakeThreadId, nil
}

func (s *MockRetrieveAssociation) GetCrmId(contactId string, objectType string) (*RetrieveCrmIdResponseHSFormat, error) {

	var fakeCrmId RetrieveCrmIdResponseHSFormat
	if contactId == "32027696539" {
		fakeCrmId = RetrieveCrmIdResponseHSFormat{
			Results: []RetrieveCrmIdResultHSFormat{
				{
					IdArray: []RetrieveCrmId{
						{
							Id: "12345678900",
						},
					},
				},
			},
		}
	}
	return &fakeCrmId, nil

}

func createMockClient() *CustomClient {

	mockCRM := &hubspot.CRM{
		Contact: &MockContact{},
		Deal:    &MockDeal{},
	}

	mockClient := &CustomClient{
		Client: &hubspot.Client{
			CRM: mockCRM,
		},
		Thread:              &MockThread{},
		RetrieveAssociation: &MockRetrieveAssociation{},
	}

	return mockClient
}

// Testing functions

func TestComponent_ExecuteGetContactTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	tc := struct {
		name     string
		input    string
		wantResp ContactInfoTaskFormat
	}{
		name:  "ok - get contact",
		input: "32027696539",
		wantResp: ContactInfoTaskFormat{
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
		input    ContactInfoTaskFormat
		wantResp string
	}{
		name: "ok - create contact",
		input: ContactInfoTaskFormat{
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
		exec := &base.ExecutionWrapper{Execution: e}

		pbInput, err := base.ConvertToStructpb(tc.input)

		c.Assert(err, qt.IsNil)

		res, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbInput})
		c.Assert(err, qt.IsNil)

		resString := res[0].Fields["contact-id"].GetStringValue()

		c.Check(resString, qt.Equals, tc.wantResp)

	})
}

func TestComponent_ExecuteGetDealTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	tc := struct {
		name     string
		input    string
		wantResp DealInfoTaskFormat
	}{
		name:  "ok - get deal",
		input: "20620806729",
		wantResp: DealInfoTaskFormat{
			DealName:   "Fake deal",
			Pipeline:   "default",
			DealStage:  "qualifiedtobuy",
			CreateDate: "2024-07-09T02:22:06.140Z",
		},
	}

	c.Run(tc.name, func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"token": bearerToken,
		})
		c.Assert(err, qt.IsNil)

		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: taskGetDeal},
			client:             createMockClient(),
		}
		exec := &base.ExecutionWrapper{Execution: e}

		pbInput, err := structpb.NewStruct(map[string]any{
			"deal-id": tc.input,
		})

		c.Assert(err, qt.IsNil)

		res, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbInput})
		c.Assert(err, qt.IsNil)

		resJSON, err := protojson.Marshal(res[0])
		c.Assert(err, qt.IsNil)

		c.Check(resJSON, qt.JSONEquals, tc.wantResp)

	})
}

func TestComponent_ExecuteCreateDealTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	tc := struct {
		name           string
		inputDeal      DealInfoTaskFormat
		inputContactId string //used to associate contact with deal
		wantResp       string
	}{
		name: "ok - create deal",
		inputDeal: DealInfoTaskFormat{
			DealName:  "Test Creating Deal",
			Pipeline:  "default",
			DealStage: "contractsent",
			Amount:    900,
		},
		inputContactId: "32027696539",
		wantResp:       "12345678900",
	}

	c.Run(tc.name, func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"token": bearerToken,
		})
		c.Assert(err, qt.IsNil)

		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: taskCreateDeal},
			client:             createMockClient(),
		}
		exec := &base.ExecutionWrapper{Execution: e}

		pbInput, err := base.ConvertToStructpb(tc.inputDeal)
		pbInput.Fields["contact-id-or-email"] = structpb.NewStringValue(tc.inputContactId)

		c.Assert(err, qt.IsNil)

		res, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbInput})
		c.Assert(err, qt.IsNil)

		resString := res[0].Fields["deal-id"].GetStringValue()

		c.Check(resString, qt.Equals, tc.wantResp)

	})
}

func TestComponent_ExecuteGetThreadTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	tc := struct {
		name     string
		input    string
		wantResp ThreadResponseTaskFormat
	}{
		name:  "ok - get thread",
		input: "7509711154",
		wantResp: ThreadResponseTaskFormat{
			Results: []ThreadResultTaskFormat{
				{
					CreatedAt: "2024-07-02T10:42:15Z",
					Senders: []ThreadUserTaskFormat{
						{
							Name:  "Brian Halligan (Sample Contact)",
							Type:  "HS_EMAIL_ADDRESS",
							Value: "bh@hubspot.com",
						},
					},
					Recipients: []ThreadUserTaskFormat{
						{
							Type:  "HS_EMAIL_ADDRESS",
							Value: "fake_email@gmail.com",
						},
					},
					Text:    "Just random content inside",
					Subject: "A fake message",
				},
			},
		},
	}

	c.Run(tc.name, func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"token": bearerToken,
		})
		c.Assert(err, qt.IsNil)

		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: taskGetThread},
			client:             createMockClient(),
		}
		exec := &base.ExecutionWrapper{Execution: e}

		pbInput, err := structpb.NewStruct(map[string]any{
			"thread-id": tc.input,
		})

		c.Assert(err, qt.IsNil)

		res, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbInput})
		c.Assert(err, qt.IsNil)

		resJSON, err := protojson.Marshal(res[0])
		c.Assert(err, qt.IsNil)

		c.Check(resJSON, qt.JSONEquals, tc.wantResp)

	})
}

func TestComponent_ExecuteRetrieveAssociationTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    RetrieveAssociationInput
		wantResp interface{}
	}{
		{
			name: "ok - retrieve association: thread ID",
			input: RetrieveAssociationInput{
				ContactId:  "32027696539",
				ObjectType: "Threads",
			},
			wantResp: RetrieveThreadIdResponse{
				Results: []RetrieveThreadIdResult{
					{
						Id: "7509711154",
					},
				},
			},
		},
		{
			name: "ok - retrieve association: deal ID",
			input: RetrieveAssociationInput{
				ContactId:  "32027696539",
				ObjectType: "Deals",
			},
			wantResp: RetrieveCrmIdResultTaskFormat{
				IdArray: []RetrieveCrmId{
					{
						Id: "12345678900",
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"token": bearerToken,
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: taskRetrieveAssociation},
				client:             createMockClient(),
			}
			exec := &base.ExecutionWrapper{Execution: e}

			pbInput, err := base.ConvertToStructpb(tc.input)

			c.Assert(err, qt.IsNil)

			res, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbInput})
			c.Assert(err, qt.IsNil)

			resJSON, err := protojson.Marshal(res[0])
			c.Assert(err, qt.IsNil)

			c.Check(resJSON, qt.JSONEquals, tc.wantResp)

		})
	}

}
