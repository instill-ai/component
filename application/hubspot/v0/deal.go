package hubspot

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
	OwnerId    string `json:"owner-id,omitempty"`
	DealName   string `json:"deal-name"`
	Pipeline   string `json:"pipeline"`
	DealStage  string `json:"deal-stage"`
	Amount     string `json:"amount,omitempty"`
	DealType   string `json:"deal-type,omitempty"`
	CloseDate  string `json:"close-date,omitempty"`
	CreateDate string `json:"create-date"`
	DealId     string `json:"deal-id"`
}
