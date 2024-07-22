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

// This file is for testing CRM objects: Contacts, Deals, Companies, Tickets
// notcrm_test.go is for Threads and Retrieve Association

const (
	bearerToken = "123"
)

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

// Mock Deal struct and its functions
type MockDeal struct{}

func (s *MockDeal) Get(dealID string, deal interface{}, option *hubspot.RequestQueryOption) (*hubspot.ResponseResource, error) {

	var fakeDeal TaskGetDealResp
	if dealID == "20620806729" {
		fakeDeal = TaskGetDealResp{
			DealName:   "Fake deal",
			Pipeline:   "default",
			DealStage:  "qualifiedtobuy",
			CreateDate: "2024-07-09T02:22:06.140Z",
		}
	}

	ret := &hubspot.ResponseResource{
		Properties: &fakeDeal,
	}

	return ret, nil
}

func (s *MockDeal) Create(deal interface{}) (*hubspot.ResponseResource, error) {
	arbitraryDealId := "12345678900"

	fakeDealInfo := deal.(*TaskCreateDealReq)

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

// Mock Company struct and its functions
type MockCompany struct{}

func (s *MockCompany) Get(companyID string, company interface{}, option *hubspot.RequestQueryOption) (*hubspot.ResponseResource, error) {

	var fakeCompany TaskGetCompanyResp
	if companyID == "20620806729" {
		fakeCompany = TaskGetCompanyResp{
			CompanyName:   "HubSpot",
			CompanyDomain: "hubspot.com",
			Description:   "HubSpot offers a comprehensive cloud-based marketing and sales platform with integrated applications for attracting, converting, and delighting customers through inbound marketing strategies.",
			PhoneNumber:   "+1 888-482-7768",
			Industry:      "COMPUTER_SOFTWARE",
			AnnualRevenue: "10000000000",
		}
	}

	ret := &hubspot.ResponseResource{
		Properties: &fakeCompany,
	}

	return ret, nil
}

func (s *MockCompany) Create(company interface{}) (*hubspot.ResponseResource, error) {
	arbitraryCompanyId := "99999999999"

	fakeCompanyInfo := company.(*TaskCreateCompanyReq)

	fakeCompanyInfo.CompanyId = arbitraryCompanyId

	ret := &hubspot.ResponseResource{
		Properties: fakeCompanyInfo,
	}

	return ret, nil
}
func (s *MockCompany) Update(companyID string, company interface{}) (*hubspot.ResponseResource, error) {
	return nil, nil
}
func (s *MockCompany) Delete(companyID string) error {
	return nil
}
func (s *MockCompany) AssociateAnotherObj(companyID string, conf *hubspot.AssociationConfig) (*hubspot.ResponseResource, error) {
	return nil, nil
}

// Mock Ticket struct and its functions
type MockTicket struct{}

func (s *MockTicket) Get(ticketId string) (*hubspot.ResponseResource, error) {
	var fakeTicket TaskGetTicketResp
	if ticketId == "2865646368" {
		fakeTicket = TaskGetTicketResp{
			TicketName:   "HubSpot - New Query (Sample Query)",
			TicketStatus: "1",
			Pipeline:     "0",
			Category:     "PRODUCT_ISSUE;BILLING_ISSUE",
		}
	}

	ret := &hubspot.ResponseResource{
		Properties: &fakeTicket,
	}

	return ret, nil
}
func (s *MockTicket) Create(ticket *TaskCreateTicketReq) (*hubspot.ResponseResource, error) {
	arbitraryTicketId := "99987654321"

	fakeTicketInfo := ticket

	fakeTicketInfo.TicketId = arbitraryTicketId

	ret := &hubspot.ResponseResource{
		Properties: fakeTicketInfo,
	}

	return ret, nil
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
		wantResp TaskGetDealOutput
	}{
		name:  "ok - get deal",
		input: "20620806729",
		wantResp: TaskGetDealOutput{
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
		inputDeal      TaskCreateDealInput
		inputContactId string //used to associate contact with deal
		wantResp       string
	}{
		name: "ok - create deal",
		inputDeal: TaskCreateDealInput{
			DealName:  "Test Creating Deal",
			Pipeline:  "default",
			DealStage: "contractsent",
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

func TestComponent_ExecuteGetCompanyTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	tc := struct {
		name     string
		input    string
		wantResp TaskGetCompanyOutput
	}{
		name:  "ok - get company",
		input: "20620806729",
		wantResp: TaskGetCompanyOutput{
			CompanyName:   "HubSpot",
			CompanyDomain: "hubspot.com",
			Description:   "HubSpot offers a comprehensive cloud-based marketing and sales platform with integrated applications for attracting, converting, and delighting customers through inbound marketing strategies.",
			PhoneNumber:   "+1 888-482-7768",
			Industry:      "COMPUTER_SOFTWARE",
			AnnualRevenue: 10000000000,
		},
	}

	c.Run(tc.name, func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"token": bearerToken,
		})
		c.Assert(err, qt.IsNil)

		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: taskGetCompany},
			client:             createMockClient(),
		}
		exec := &base.ExecutionWrapper{Execution: e}

		pbInput, err := structpb.NewStruct(map[string]any{
			"company-id": tc.input,
		})

		c.Assert(err, qt.IsNil)

		res, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbInput})

		c.Assert(err, qt.IsNil)

		resJSON, err := protojson.Marshal(res[0])
		c.Assert(err, qt.IsNil)

		c.Check(resJSON, qt.JSONEquals, tc.wantResp)

	})
}

func TestComponent_ExecuteCreateCompanyTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	tc := struct {
		name           string
		inputCompany   TaskCreateCompanyInput
		inputContactId string //used to associate contact with company
		wantResp       string
	}{
		name: "ok - create company",
		inputCompany: TaskCreateCompanyInput{
			CompanyName:   "Fake Company",
			CompanyDomain: "fakecompany.com",
			AnnualRevenue: 5000000,
		},
		inputContactId: "32027696539",
		wantResp:       "99999999999",
	}

	c.Run(tc.name, func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"token": bearerToken,
		})
		c.Assert(err, qt.IsNil)

		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: taskCreateCompany},
			client:             createMockClient(),
		}
		exec := &base.ExecutionWrapper{Execution: e}

		pbInput, err := base.ConvertToStructpb(tc.inputCompany)
		pbInput.Fields["contact-id-or-email"] = structpb.NewStringValue(tc.inputContactId)

		c.Assert(err, qt.IsNil)

		res, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbInput})
		c.Assert(err, qt.IsNil)

		resString := res[0].Fields["company-id"].GetStringValue()

		c.Check(resString, qt.Equals, tc.wantResp)

	})
}

func TestComponent_ExecuteGetTicketTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	tc := struct {
		name     string
		input    string
		wantResp TaskGetTicketOutput
	}{
		name:  "ok - get ticket",
		input: "2865646368",
		wantResp: TaskGetTicketOutput{
			TicketName:   "HubSpot - New Query (Sample Query)",
			TicketStatus: "1",
			Pipeline:     "0",
			Category:     []string{"PRODUCT_ISSUE", "BILLING_ISSUE"},
		},
	}

	c.Run(tc.name, func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"token": bearerToken,
		})
		c.Assert(err, qt.IsNil)

		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: taskGetTicket},
			client:             createMockClient(),
		}
		exec := &base.ExecutionWrapper{Execution: e}

		pbInput, err := structpb.NewStruct(map[string]any{
			"ticket-id": tc.input,
		})

		c.Assert(err, qt.IsNil)

		res, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbInput})

		c.Assert(err, qt.IsNil)

		resJSON, err := protojson.Marshal(res[0])
		c.Assert(err, qt.IsNil)

		c.Check(resJSON, qt.JSONEquals, tc.wantResp)

	})
}

func TestComponent_ExecuteCreateTicketTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	tc := struct {
		name           string
		inputTicket    TaskCreateTicketInput
		inputContactId string //used to associate contact with ticket
		wantResp       string
	}{
		name: "ok - create ticket",
		inputTicket: TaskCreateTicketInput{
			TicketName:   "Fake Ticket",
			TicketStatus: "2",
			Pipeline:     "0",
			Category:     []string{"FEATURE_REQUEST", "GENERAL_INQUIRY"},
		},
		inputContactId: "32027696539",
		wantResp:       "99987654321",
	}

	c.Run(tc.name, func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"token": bearerToken,
		})
		c.Assert(err, qt.IsNil)

		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: taskCreateTicket},
			client:             createMockClient(),
		}
		exec := &base.ExecutionWrapper{Execution: e}

		pbInput, err := base.ConvertToStructpb(tc.inputTicket)
		pbInput.Fields["contact-id-or-email"] = structpb.NewStringValue(tc.inputContactId)

		c.Assert(err, qt.IsNil)

		res, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbInput})
		c.Assert(err, qt.IsNil)

		resString := res[0].Fields["ticket-id"].GetStringValue()

		c.Check(resString, qt.Equals, tc.wantResp)

	})
}
