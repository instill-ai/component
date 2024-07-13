package hubspot

import (
	"strconv"
)

type DealInfoHSFormat struct {
	OwnerId    string `json:"hubspot_owner_id,omitempty"`
	DealName   string `json:"dealname"`
	Pipeline   string `json:"pipeline"`
	DealStage  string `json:"dealstage"`
	Amount     string `json:"amount,omitempty"`
	DealType   string `json:"dealtype,omitempty"`
	CloseDate  string `json:"closedate,omitempty"`
	CreateDate string `json:"createdate"`
	DealId     string `json:"hs_object_id"`
}

type DealInfoTaskFormat struct {
	OwnerId    string  `json:"owner-id,omitempty"`
	DealName   string  `json:"deal-name"`
	Pipeline   string  `json:"pipeline"`
	DealStage  string  `json:"deal-stage"`
	Amount     float64 `json:"amount,omitempty"`
	DealType   string  `json:"deal-type,omitempty"`
	CloseDate  string  `json:"close-date,omitempty"`
	CreateDate string  `json:"create-date"`
	DealId     string  `json:"deal-id"`
}

func DealConvertToTaskFormat(deal *DealInfoHSFormat) (*DealInfoTaskFormat, error) {

	var amount float64

	if deal.Amount != "" {
		var err error
		amount, err = strconv.ParseFloat(deal.Amount, 64)

		if err != nil {
			return nil, err
		}
	}

	ret := &DealInfoTaskFormat{
		OwnerId:    deal.OwnerId,
		DealName:   deal.DealName,
		Pipeline:   deal.Pipeline,
		DealStage:  deal.DealStage,
		Amount:     amount,
		DealType:   deal.DealType,
		CloseDate:  deal.CloseDate,
		CreateDate: deal.CreateDate,
	}

	return ret, nil

}

func DealConvertToHSFormat(deal *DealInfoTaskFormat) *DealInfoHSFormat {

	var amount string
	if deal.Amount != 0 {
		amount = strconv.FormatFloat(deal.Amount, 'f', -1, 64)
	}

	ret := &DealInfoHSFormat{
		DealName:  deal.DealName,
		Pipeline:  deal.Pipeline,
		DealStage: deal.DealStage,
		Amount:    amount,
		DealType:  deal.DealType,
		CloseDate: deal.CloseDate,
	}

	return ret

}
