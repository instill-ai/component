package hubspot

import "strconv"

type CompanyInfoHSFormat struct {
	OwnerId       string `json:"hubspot_owner_id,omitempty"`
	CompanyName   string `json:"name,omitempty"`
	CompanyDomain string `json:"domain,omitempty"`
	Description   string `json:"description,omitempty"`
	PhoneNumber   string `json:"phone,omitempty"`
	Industry      string `json:"industry,omitempty"`
	CompanyType   string `json:"type,omitempty"`
	City          string `json:"city,omitempty"`
	State         string `json:"state,omitempty"`
	Country       string `json:"country,omitempty"`
	PostalCode    string `json:"zip,omitempty"`
	TimeZone      string `json:"timezone,omitempty"`
	AnnualRevenue string `json:"annualrevenue,omitempty"`
	TotalRevenue  string `json:"totalrevenue,omitempty"`
	LinkedinPage  string `json:"linkedin_company_page,omitempty"`
	CompanyId     string `json:"hs_object_id"`
}

type CompanyInfoTaskFormat struct {
	OwnerId       string  `json:"owner-id,omitempty"`
	CompanyName   string  `json:"company-name,omitempty"`
	CompanyDomain string  `json:"company-domain,omitempty"`
	Description   string  `json:"description,omitempty"`
	PhoneNumber   string  `json:"phone-number,omitempty"`
	Industry      string  `json:"industry,omitempty"`
	CompanyType   string  `json:"company-type,omitempty"`
	City          string  `json:"city,omitempty"`
	State         string  `json:"state,omitempty"`
	Country       string  `json:"country,omitempty"`
	PostalCode    string  `json:"postal-code,omitempty"`
	TimeZone      string  `json:"time-zone,omitempty"`
	AnnualRevenue float64 `json:"annual-revenue,omitempty"`
	TotalRevenue  float64 `json:"total-revenue,omitempty"`
	LinkedinPage  string  `json:"linkedin-page,omitempty"`
	CompanyId     string  `json:"company-id"`
}

func CompanyConvertToTaskFormat(company *CompanyInfoHSFormat) (*CompanyInfoTaskFormat, error) {
	var annualRevenue, totalRevenue float64

	if company.AnnualRevenue != "" {
		var err error
		annualRevenue, err = strconv.ParseFloat(company.AnnualRevenue, 64)

		if err != nil {
			return nil, err
		}
	}

	if company.TotalRevenue != "" {
		var err error
		totalRevenue, err = strconv.ParseFloat(company.TotalRevenue, 64)

		if err != nil {
			return nil, err
		}
	}

	ret := &CompanyInfoTaskFormat{
		CompanyName:   company.CompanyName,
		CompanyDomain: company.CompanyDomain,
		Description:   company.Description,
		PhoneNumber:   company.PhoneNumber,
		Industry:      company.Industry,
		CompanyType:   company.CompanyType,
		City:          company.City,
		State:         company.State,
		Country:       company.Country,
		PostalCode:    company.PostalCode,
		TimeZone:      company.TimeZone,
		AnnualRevenue: annualRevenue,
		TotalRevenue:  totalRevenue,
		LinkedinPage:  company.LinkedinPage,
	}

	return ret, nil
}

func CompanyConvertToHSFormat(company *CompanyInfoTaskFormat) *CompanyInfoHSFormat {
	var annualRevenue string
	if company.AnnualRevenue != 0 {
		annualRevenue = strconv.FormatFloat(company.AnnualRevenue, 'f', -1, 64)
	}

	ret := &CompanyInfoHSFormat{
		CompanyName:   company.CompanyName,
		CompanyDomain: company.CompanyDomain,
		Description:   company.Description,
		PhoneNumber:   company.PhoneNumber,
		Industry:      company.Industry,
		CompanyType:   company.CompanyType,
		City:          company.City,
		State:         company.State,
		Country:       company.Country,
		PostalCode:    company.PostalCode,
		TimeZone:      company.TimeZone,
		AnnualRevenue: annualRevenue,
		LinkedinPage:  company.LinkedinPage,
	}

	return ret
}
