package configLoader

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

// TODO: refactor configLoader

type definition struct {
	Custom           bool        `json:"custom"`
	DocumentationUrl string      `json:"documentation_url"`
	Icon             string      `json:"icon"`
	IconUrl          string      `json:"icon_url"`
	Id               string      `json:"id"`
	Public           bool        `json:"public"`
	Title            string      `json:"title"`
	Tombstone        bool        `json:"tombstone"`
	Uid              string      `json:"uid"`
	Spec             interface{} `json:"spec"`
	VendorAttributes interface{} `json:"vendor_attributes"`
}

func (c *ConfigLoader) Load(vendorName string, connectorType connectorPB.ConnectorType, definitionsJson []byte) ([]*connectorPB.ConnectorDefinition, error) {

	defs := []*connectorPB.ConnectorDefinition{}
	var defJsonArr []definition
	err := json.Unmarshal(definitionsJson, &defJsonArr)
	if err != nil {
		panic(err)
	}

	for _, defJson := range defJsonArr {

		defJsonBytes, err := json.Marshal(defJson)
		if err != nil {
			return nil, err
		}

		def := &connectorPB.ConnectorDefinition{}

		err = protojson.Unmarshal(defJsonBytes, def)
		if err != nil {
			return nil, err
		}
		def.Name = fmt.Sprintf("connector-definitions/%s", def.Id)
		def.Vendor = vendorName
		def.Type = connectorType

		defs = append(defs, def)

	}
	return defs, nil
}
