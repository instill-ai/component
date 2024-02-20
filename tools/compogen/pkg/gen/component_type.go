package gen

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

// ComponentSubtype holds the possible subtypes of a component (e.g.
// "operator", "AI connector", "data connector") and implements several helper
// methods.
type ComponentSubtype string

const (
	cstOperator            ComponentSubtype = "operator"
	cstAIConnector         ComponentSubtype = "AI connector"
	cstBlockchainConnector ComponentSubtype = "blockchain connector"
	cstDataConnector       ComponentSubtype = "data connector"
)

var toComponentSubtype = map[string]ComponentSubtype{
	"CONNECTOR_TYPE_AI":         cstAIConnector,
	"CONNECTOR_TYPE_BLOCKCHAIN": cstBlockchainConnector,
	"CONNECTOR_TYPE_DATA":       cstDataConnector,
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
