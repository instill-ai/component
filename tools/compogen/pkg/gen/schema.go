package gen

type property struct {
	Description string `json:"description" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Order       *int   `json:"instillUIOrder" validate:"required"`

	Type string `json:"type"`

	// If Type is array, Items defines the element type.
	Items struct {
		Type       string              `json:"type"`
		Properties map[string]property `json:"properties" validate:"omitempty,dive"`
	} `json:"items"`

	Properties map[string]property `json:"properties" validate:"omitempty,dive"`

	OneOf []objectSchema `json:"oneOf,omitempty"`
	Const string         `json:"const,omitempty"`

	Deprecated bool `json:"deprecated"`
}

type objectSchema struct {
	Properties map[string]property `json:"properties" validate:"gt=0,dive"`
	Required   []string            `json:"required"`
}
