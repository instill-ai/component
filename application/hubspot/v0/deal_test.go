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
	arbitraryDealID := "12345678900"

	fakeDealInfo := deal.(*TaskCreateDealReq)

	fakeDealInfo.DealID = arbitraryDealID

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
		e.execute = e.GetDeal
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
		inputContactID string //used to associate contact with deal
		wantResp       string
	}{
		name: "ok - create deal",
		inputDeal: TaskCreateDealInput{
			DealName:  "Test Creating Deal",
			Pipeline:  "default",
			DealStage: "contractsent",
		},
		inputContactID: "32027696539",
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
		e.execute = e.CreateDeal
		exec := &base.ExecutionWrapper{Execution: e}

		pbInput, err := base.ConvertToStructpb(tc.inputDeal)
		pbInput.Fields["contact-id-or-email"] = structpb.NewStringValue(tc.inputContactID)

		c.Assert(err, qt.IsNil)

		res, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbInput})
		c.Assert(err, qt.IsNil)

		resString := res[0].Fields["deal-id"].GetStringValue()

		c.Check(resString, qt.Equals, tc.wantResp)

	})
}
