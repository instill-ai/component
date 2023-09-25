package configLoader

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1alpha"
)

// TODO: share data structure for connector and operator as much as possible

type connectorDefinition struct {
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

type operatorDefinition struct {
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
}

func (c *ConfigLoader) LoadConnector(vendorName string, connectorType connectorPB.ConnectorType, definitionsJson []byte) ([]*connectorPB.ConnectorDefinition, error) {

	defs := []*connectorPB.ConnectorDefinition{}
	var defJsonArr []connectorDefinition
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

func (c *ConfigLoader) LoadOperator(definitionsJson []byte) ([]*pipelinePB.OperatorDefinition, error) {

	defs := []*pipelinePB.OperatorDefinition{}
	var defJsonArr []operatorDefinition
	err := json.Unmarshal(definitionsJson, &defJsonArr)
	if err != nil {
		panic(err)
	}

	for _, defJson := range defJsonArr {

		defJsonBytes, err := json.Marshal(defJson)
		if err != nil {
			return nil, err
		}

		def := &pipelinePB.OperatorDefinition{}

		err = protojson.Unmarshal(defJsonBytes, def)
		if err != nil {
			return nil, err
		}
		def.Name = fmt.Sprintf("operator-definitions/%s", def.Id)

		defs = append(defs, def)

	}
	return defs, nil
}
