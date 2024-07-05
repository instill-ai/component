package hubspot

type ContactInfoHSFormat struct {
	FirstName      string `json:"firstname,omitempty"`
	LastName       string `json:"lastname,omitempty"`
	Email          string `json:"email,omitempty"`
	PhoneNumber    string `json:"phone,omitempty"`
	Company        string `json:"company,omitempty"`
	OwnerId        string `json:"hubspot_owner_id,omitempty"`
	JobTitle       string `json:"jobtitle,omitempty"`
	LifecycleStage string `json:"lifecyclestage,omitempty"`
	LeadStatus     string `json:"hs_lead_status,omitempty"`
	ContactId      string `json:"hs_object_id"`
}

type ContactInfoTaskFormat struct {
	FirstName      string `json:"firstName,omitempty"`
	LastName       string `json:"lastName,omitempty"`
	Email          string `json:"email,omitempty"`
	PhoneNumber    string `json:"phoneNumber,omitempty"`
	Company        string `json:"company,omitempty"`
	OwnerId        string `json:"ownerId,omitempty"`
	JobTitle       string `json:"jobTitle,omitempty"`
	LifecycleStage string `json:"lifecycleStage,omitempty"`
	LeadStatus     string `json:"leadStatus,omitempty"`
	ContactId      string `json:"contactId"`
}

// use to get contact ID from email
type ContactIDHSFormat struct {
	ContactId string `json:"hs_object_id,omitempty"`
}
