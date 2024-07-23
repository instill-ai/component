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

// mockClient is in contact_test.go

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
	arbitraryCompanyID := "99999999999"

	fakeCompanyInfo := company.(*TaskCreateCompanyReq)

	fakeCompanyInfo.CompanyID = arbitraryCompanyID

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
		e.execute = e.GetCompany
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
		inputContactID string //used to associate contact with company
		wantResp       string
	}{
		name: "ok - create company",
		inputCompany: TaskCreateCompanyInput{
			CompanyName:   "Fake Company",
			CompanyDomain: "fakecompany.com",
			AnnualRevenue: 5000000,
		},
		inputContactID: "32027696539",
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
		e.execute = e.CreateCompany
		exec := &base.ExecutionWrapper{Execution: e}

		pbInput, err := base.ConvertToStructpb(tc.inputCompany)
		pbInput.Fields["contact-id-or-email"] = structpb.NewStringValue(tc.inputContactID)

		c.Assert(err, qt.IsNil)

		res, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbInput})
		c.Assert(err, qt.IsNil)

		resString := res[0].Fields["company-id"].GetStringValue()

		c.Check(resString, qt.Equals, tc.wantResp)

	})
}
