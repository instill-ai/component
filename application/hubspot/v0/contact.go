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
	FirstName      string `json:"first-name,omitempty"`
	LastName       string `json:"last-name,omitempty"`
	Email          string `json:"email,omitempty"`
	PhoneNumber    string `json:"phone-number,omitempty"`
	Company        string `json:"company,omitempty"`
	OwnerId        string `json:"owner-id,omitempty"`
	JobTitle       string `json:"job-title,omitempty"`
	LifecycleStage string `json:"lifecycle-stage,omitempty"`
	LeadStatus     string `json:"lead-status,omitempty"`
	ContactId      string `json:"contact-id"`
}

// use to get contact ID from email
type ContactIDHSFormat struct {
	ContactId string `json:"hs_object_id,omitempty"`
}
