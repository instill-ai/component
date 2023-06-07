package configLoader

import (
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/datatypes"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

// TODO: refactor configLoader

func (c *ConfigLoader) Load(definitionsYaml []byte, specsYaml []byte, connDefs interface{}) {

	defs := []*connectorPB.ConnectorDefinition{}
	dockerImageSpecs := []*connectorPB.DockerImageSpec{}

	if jsonSliceMap, err := ProcessJSONSliceMap(definitionsYaml); err == nil {
		if err := UnmarshalConnectorPB(jsonSliceMap, connDefs); err != nil {
			c.Logger.Error(err.Error())
		}
		if err := UnmarshalConnectorPB(jsonSliceMap, &defs); err != nil {
			c.Logger.Error(err.Error())
		}
	} else {
		c.Logger.Error(err.Error())
	}

	if jsonSliceMap, err := ProcessJSONSliceMap(specsYaml); err == nil {
		if err := UnmarshalConnectorPB(jsonSliceMap, &dockerImageSpecs); err != nil {
			c.Logger.Error(err.Error())
		}
	} else {
		c.Logger.Error(err.Error())
	}

	for idx, def := range defs {
		var imgTag string
		if def.GetDockerImageTag() != "" {
			imgTag = ":" + def.GetDockerImageTag()
		} else {
			imgTag = def.GetDockerImageTag()
		}
		if spec, err := DindDockerImageSpec(def.GetDockerRepository()+imgTag, &dockerImageSpecs); err != nil {
			c.Logger.Error(err.Error())
		} else {
			switch connDefs := connDefs.(type) {
			case *[]*connectorPB.DestinationConnectorDefinition:
				if err := c.CreateDestinationConnectorDefinition((*connDefs)[idx], def, spec); err != nil {
					c.Logger.Error(err.Error())
				}
			case *[]*connectorPB.SourceConnectorDefinition:
				if err := c.CreateSourceConnectorDefinition((*connDefs)[idx], def, spec); err != nil {
					c.Logger.Error(err.Error())
				}
			}

		}
	}
}

func (c *ConfigLoader) CreateDestinationConnectorDefinition(dstConnDef *connectorPB.DestinationConnectorDefinition, connDef *connectorPB.ConnectorDefinition, spec datatypes.JSON) error {

	if dstConnDef.GetId() == "" {
		dstConnDef.Id = connDef.GetDockerRepository()[strings.LastIndex(connDef.GetDockerRepository(), "/")+1:]
	}

	dstConnDef.ConnectorDefinition = connDef
	dstConnDef.GetConnectorDefinition().CreateTime = &timestamppb.Timestamp{}
	dstConnDef.GetConnectorDefinition().UpdateTime = &timestamppb.Timestamp{}
	dstConnDef.GetConnectorDefinition().Tombstone = false
	dstConnDef.GetConnectorDefinition().Public = true
	dstConnDef.GetConnectorDefinition().Custom = false

	dstConnDef.GetConnectorDefinition().Spec = &connectorPB.Spec{}
	if err := protojson.Unmarshal(spec, dstConnDef.GetConnectorDefinition().Spec); err != nil {
		c.Logger.Error(err.Error())
	}

	if dstConnDef.GetConnectorDefinition().GetResourceRequirements() == nil {
		dstConnDef.GetConnectorDefinition().ResourceRequirements = &structpb.Struct{}
	}

	// Validate JSON Schema
	if err := ValidateJSONSchema(c.DstConnDefJSONSchema, dstConnDef, true); err != nil {
		return err
	}

	return nil
}

func (c *ConfigLoader) CreateSourceConnectorDefinition(srcConnDef *connectorPB.SourceConnectorDefinition, connDef *connectorPB.ConnectorDefinition, spec datatypes.JSON) error {

	if srcConnDef.GetId() == "" {
		srcConnDef.Id = connDef.GetDockerRepository()[strings.LastIndex(connDef.GetDockerRepository(), "/")+1:]
	}

	srcConnDef.ConnectorDefinition = connDef
	srcConnDef.GetConnectorDefinition().CreateTime = &timestamppb.Timestamp{}
	srcConnDef.GetConnectorDefinition().UpdateTime = &timestamppb.Timestamp{}
	srcConnDef.GetConnectorDefinition().Tombstone = false
	srcConnDef.GetConnectorDefinition().Public = true
	srcConnDef.GetConnectorDefinition().Custom = false

	srcConnDef.GetConnectorDefinition().Spec = &connectorPB.Spec{}
	if err := protojson.Unmarshal(spec, srcConnDef.GetConnectorDefinition().Spec); err != nil {
		c.Logger.Error(err.Error())
	}
	if srcConnDef.GetConnectorDefinition().GetResourceRequirements() == nil {
		srcConnDef.GetConnectorDefinition().ResourceRequirements = &structpb.Struct{}
	}

	// Validate JSON Schema
	if err := ValidateJSONSchema(c.SrcConnDefJSONSchema, srcConnDef, true); err != nil {
		return err
	}

	return nil
}
