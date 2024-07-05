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
	OwnerId    string `json:"ownerId,omitempty"`
	DealName   string `json:"dealName"`
	Pipeline   string `json:"pipeline"`
	DealStage  string `json:"dealStage"`
	Amount     string `json:"amount,omitempty"`
	DealType   string `json:"dealType,omitempty"`
	CloseDate  string `json:"closeDate,omitempty"`
	CreateDate string `json:"createDate"`
	DealId     string `json:"dealId"`
}
