package configLoader

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

// TODO: refactor configLoader

type definition struct {
	Custom           bool        `json:"custom"`
	DocumentationUrl string      `json:"documentationUrl"`
	Icon             string      `json:"icon"`
	IconUrl          string      `json:"iconUrl"`
	Id               string      `json:"id"`
	Public           bool        `json:"public"`
	Title            string      `json:"title"`
	Tombstone        bool        `json:"tombstone"`
	Uid              string      `json:"uid"`
	Spec             interface{} `json:"spec"`
	VendorAttributes interface{} `json:"vendorAttributes"`
}

func (c *ConfigLoader) Load(vendorName string, definitionsJson []byte, connDefs interface{}) {
	var defJsonArr []definition
	err := json.Unmarshal(definitionsJson, &defJsonArr)
	if err != nil {
		panic(err)
	}

	for _, defJson := range defJsonArr {

		specBytes, err := json.Marshal(defJson.Spec)
		if err != nil {
			panic(err)
		}
		specStruct := &connectorPB.Spec{}
		err = protojson.Unmarshal(specBytes, specStruct)
		if err != nil {
			panic(err)
		}
		vendorAttributesBytes, err := json.Marshal(defJson.VendorAttributes)
		if err != nil {
			panic(err)
		}
		vendorAttributesStruct := &structpb.Struct{}
		err = protojson.Unmarshal(vendorAttributesBytes, vendorAttributesStruct)
		if err != nil {
			panic(err)
		}

		def := connectorPB.ConnectorDefinition{
			Custom:           defJson.Custom,
			DocumentationUrl: defJson.DocumentationUrl,
			Icon:             defJson.Icon,
			IconUrl:          defJson.IconUrl,
			Public:           defJson.Public,
			Title:            defJson.Title,
			Tombstone:        defJson.Tombstone,
			Vendor:           vendorName,
			VendorAttributes: vendorAttributesStruct,
			Spec:             specStruct,
		}
		switch connDefs := connDefs.(type) {
		case *[]*connectorPB.DestinationConnectorDefinition:
			*connDefs = append((*connDefs), &connectorPB.DestinationConnectorDefinition{
				Id:                  defJson.Id,
				Uid:                 defJson.Uid,
				Name:                fmt.Sprintf("destination-connectors/%s", defJson.Id),
				ConnectorDefinition: &def,
			})

		case *[]*connectorPB.SourceConnectorDefinition:
			*connDefs = append((*connDefs), &connectorPB.SourceConnectorDefinition{
				Id:                  defJson.Id,
				Uid:                 defJson.Uid,
				Name:                fmt.Sprintf("source-connectors/%s", defJson.Id),
				ConnectorDefinition: &def,
			})
		}
	}
	// TODO: validate jsonschema for Spec
}
