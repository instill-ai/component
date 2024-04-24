package gen

type property struct {
	Description string `json:"description" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Order       *int   `json:"instillUIOrder" validate:"required"`

	Type string `json:"type"`

	// If Type is array, Items defines the element type.
	Items struct {
		Type string `json:"type"`
	} `json:"items"`
}

type objectSchema struct {
	Required []string `json:"required"`

	// TODO we could validate gt=0 here to avoid empty objects. At this moment
	// there's a connector (Instill Model) that requires this, but if we
	// overcome that limitation nonempty objects should be enforced.
	Properties map[string]property `json:"properties" validate:"dive"`
}
