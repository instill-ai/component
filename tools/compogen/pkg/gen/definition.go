package gen

type property struct {
	Description string `json:"description" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Type        string `json:"type" validate:"required"`
	Order       *int   `json:"instillUIOrder" validate:"required"`
}

type resourceSpec struct {
	Properties map[string]property `json:"properties" validate:"gt=0,dive"`
	Required   []string            `json:"required"`
}

type spec struct {
	ResourceSpecification *resourceSpec `json:"resource_specification" validate:"omitnil"`
}

type definition struct {
	ID             string   `json:"id" validate:"required"`
	Title          string   `json:"title" validate:"required"`
	Description    string   `json:"description" validate:"required"`
	Version        string   `json:"version" validate:"required,semver"`
	AvailableTasks []string `json:"available_tasks" validate:"gt=0"`
	SourceURL      string   `json:"source_url" validate:"url"`

	Public        bool   `json:"public"`
	Type          string `json:"type"`
	Prerequisites string `json:"prerequisites"`

	Spec spec `json:"spec" validate:"omitempty"`
}

// This struct is used to validate the definitions schema.
type definitions struct {
	// Definitions is an array for legacy reasons: Airbyte used to have several
	// definitions. These were merged into one but the structure remained. It
	// should be refactored to remove the array nesting in the future.
	Definitions []definition `validate:"len=1,dive"`
}
