package gen

import (
	pb "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

// ComponentType defines the type of a component (e.g. operator, connector).
// This will condition how the component definition is parsed and how the
// documentation is generated.
type ComponentType string

const (
	// ComponentTypeConnector ...
	ComponentTypeConnector ComponentType = "connector"
	// ComponentTypeOperator ...
	ComponentTypeOperator ComponentType = "operator"
)

// HasConnectionConfig determines whether a component type requires the creation
// of a component resource before making it available in pipelines.
func (ct ComponentType) HasConnectionConfig() bool {
	return ct == ComponentTypeConnector
}

// ComponentSubtype holds the possible subtypes of a component (e.g.
// "operator", "AI connector", "data connector") and implements several helper
// methods.
type ComponentSubtype string

const (
	cstOperator             ComponentSubtype = "operator"
	cstAIConnector          ComponentSubtype = "AI connector"
	cstApplicationConnector ComponentSubtype = "application connector"
	cstDataConnector        ComponentSubtype = "data connector"
)

var toComponentSubtype = map[string]ComponentSubtype{
	pb.ConnectorType_CONNECTOR_TYPE_AI.String():          cstAIConnector,
	pb.ConnectorType_CONNECTOR_TYPE_APPLICATION.String(): cstApplicationConnector,
	pb.ConnectorType_CONNECTOR_TYPE_DATA.String():        cstDataConnector,
}

var modifiesArticle = map[ComponentSubtype]bool{
	cstOperator:    true,
	cstAIConnector: true,
}

// IndefiniteArticle returns the correct indefinite article (in English) for a
// component subtype, e.g., "an" (operator), "an" (AI connector), "a" (data
// connector).
func (ct ComponentSubtype) IndefiniteArticle() string {
	if modifiesArticle[ct] {
		return "an"
	}

	return "a"
}
