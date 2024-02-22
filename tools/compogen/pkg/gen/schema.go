package gen

type property struct {
	Description string `json:"description" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Type        string `json:"type" validate:"required"`
	Order       *int   `json:"instillUIOrder" validate:"required"`
}

type objectSchema struct {
	Properties map[string]property `json:"properties" validate:"gt=0,dive"`
	Required   []string            `json:"required"`
}
