package website

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

type ScrapeSitemapInput struct {
	URL string `json:"url"`
}

type ScrapeSitemapOutput struct {
	List []SiteInformation `json:"list"`
}

type SiteInformation struct {
	Loc string `json:"loc"`
	// Follow ISO 8601 format
	LastModifiedTime string  `json:"lastmod"`
	ChangeFrequency  string  `json:"changefreq,omitempty"`
	Priority         float64 `json:"priority,omitempty"`
}

type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	Urls    []URL    `xml:"url"`
}

type URL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod"`
	ChangeFreq string `xml:"changefreq"`
	Priority   string `xml:"priority"`
}

func ScrapeSitemap(input *structpb.Struct) (*structpb.Struct, error) {

	inputStruct := ScrapeSitemapInput{}
	err := base.ConvertFromStructpb(input, &inputStruct)

	if err != nil {
		return nil, fmt.Errorf("failed to convert input to struct: %v", err)
	}

	resp, err := http.Get(inputStruct.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the response body: %v", err)
	}

	var urlSet URLSet
	err = xml.Unmarshal(body, &urlSet)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %v", err)
	}

	list := []SiteInformation{}
	for _, url := range urlSet.Urls {
		priority, err := strconv.ParseFloat(url.Priority, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse priority: %v", err)
		}

		list = append(list, SiteInformation{
			Loc:              url.Loc,
			LastModifiedTime: url.LastMod,
			ChangeFrequency:  url.ChangeFreq,
			Priority:         priority,
		})
	}

	output := ScrapeSitemapOutput{
		List: list,
	}

	outputStruct, err := base.ConvertToStructpb(output)

	if err != nil {
		return nil, fmt.Errorf("failed to convert output to struct: %v", err)
	}
	return outputStruct, nil
}
