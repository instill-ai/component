package gen

type spec struct {
	ResourceSpecification *objectSchema `json:"resource_specification" validate:"omitnil"`
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
