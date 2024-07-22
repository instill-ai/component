package hubspot

import (
	"strconv"

	hubspot "github.com/belong-inc/go-hubspot"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

// Get Deal

type TaskGetDealInput struct {
	DealId string `json:"deal-id"`
}

type TaskGetDealResp struct {
	OwnerId    string `json:"hubspot_owner_id,omitempty"`
	DealName   string `json:"dealname"`
	Pipeline   string `json:"pipeline"`
	DealStage  string `json:"dealstage"`
	Amount     string `json:"amount,omitempty"`
	DealType   string `json:"dealtype,omitempty"`
	CloseDate  string `json:"closedate,omitempty"`
	CreateDate string `json:"createdate"`
}

type TaskGetDealOutput struct {
	OwnerId              string   `json:"owner-id,omitempty"`
	DealName             string   `json:"deal-name"`
	Pipeline             string   `json:"pipeline"`
	DealStage            string   `json:"deal-stage"`
	Amount               float64  `json:"amount,omitempty"`
	DealType             string   `json:"deal-type,omitempty"`
	CreateDate           string   `json:"create-date"`
	CloseDate            string   `json:"close-date,omitempty"`
	AssociatedContactIds []string `json:"associated-contact-ids,omitempty"`
}

func (e *execution) GetDeal(input *structpb.Struct) (*structpb.Struct, error) {
	inputStruct := TaskGetDealInput{}

	err := base.ConvertFromStructpb(input, &inputStruct)

	if err != nil {
		return nil, err
	}

	// get deal information

	res, err := e.client.CRM.Deal.Get(inputStruct.DealId, &TaskGetDealResp{}, &hubspot.RequestQueryOption{Associations: []string{"contacts"}})

	if err != nil {
		return nil, err
	}

	dealInfo := res.Properties.(*TaskGetDealResp)

	// get contacts associated with deal

	dealContactAssociation := res.Associations.Contacts.Results
	dealContactList := make([]string, len(dealContactAssociation))

	for index, value := range dealContactAssociation {
		dealContactList[index] = value.ID
	}

	// convert to outputStruct

	var amount float64

	if dealInfo.Amount != "" {
		var err error
		amount, err = strconv.ParseFloat(dealInfo.Amount, 64)

		if err != nil {
			return nil, err
		}
	}

	outputStruct := TaskGetDealOutput{
		OwnerId:              dealInfo.OwnerId,
		DealName:             dealInfo.DealName,
		Pipeline:             dealInfo.Pipeline,
		DealStage:            dealInfo.DealStage,
		Amount:               amount,
		DealType:             dealInfo.DealType,
		CreateDate:           dealInfo.CreateDate,
		CloseDate:            dealInfo.CloseDate,
		AssociatedContactIds: dealContactList,
	}

	output, err := base.ConvertToStructpb(outputStruct)

	if err != nil {
		return nil, err
	}

	return output, nil
}

// Create Deal

type TaskCreateDealInput struct {
	OwnerId                   string   `json:"owner-id"`
	DealName                  string   `json:"deal-name"`
	Pipeline                  string   `json:"pipeline"`
	DealStage                 string   `json:"deal-stage"`
	Amount                    float64  `json:"amount"`
	DealType                  string   `json:"deal-type"`
	CloseDate                 string   `json:"close-date"`
	CreateContactsAssociation []string `json:"create-contacts-association"`
}

type TaskCreateDealReq struct {
	OwnerId   string `json:"hubspot_owner_id,omitempty"`
	DealName  string `json:"dealname"`
	Pipeline  string `json:"pipeline"`
	DealStage string `json:"dealstage"`
	Amount    string `json:"amount,omitempty"`
	DealType  string `json:"dealtype,omitempty"`
	CloseDate string `json:"closedate,omitempty"`
	DealId    string `json:"hs_object_id"`
}

type TaskCreateDealOutput struct {
	DealId string `json:"deal-id"`
}

func (e *execution) CreateDeal(input *structpb.Struct) (*structpb.Struct, error) {

	inputStruct := TaskCreateDealInput{}
	err := base.ConvertFromStructpb(input, &inputStruct)

	if err != nil {
		return nil, err
	}

	var amount string
	if inputStruct.Amount != 0 {
		amount = strconv.FormatFloat(inputStruct.Amount, 'f', -1, 64)
	}

	req := TaskCreateDealReq{
		OwnerId:   inputStruct.OwnerId,
		DealName:  inputStruct.DealName,
		Pipeline:  inputStruct.Pipeline,
		DealStage: inputStruct.DealStage,
		Amount:    amount,
		DealType:  inputStruct.DealType,
		CloseDate: inputStruct.CloseDate,
	}

	res, err := e.client.CRM.Deal.Create(&req)

	if err != nil {
		return nil, err
	}

	// get deal Id
	dealId := res.Properties.(*TaskCreateDealReq).DealId

	outputStruct := TaskCreateDealOutput{DealId: dealId}

	output, err := base.ConvertToStructpb(outputStruct)

	if err != nil {
		return nil, err
	}

	// This section is for creating associations (deal -> object)
	if len(inputStruct.CreateContactsAssociation) != 0 {
		err := CreateAssociation(&outputStruct.DealId, &inputStruct.CreateContactsAssociation, "deal", "contact", e)

		if err != nil {
			return nil, err
		}
	}

	return output, nil
}
