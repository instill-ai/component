package gen

// This struct is used to validate the tasks schema.
// type taskMap struct {
// 	Tasks map[string]task `validate:gt=0,dive`
// }

type task struct {
	Description string `json:"instillShortDescription"`
	Title       string `json:"title"`

	// Some `tasks.json` files have a "$defs" key that isn't a task object. We
	// can only validate fields if present.
	Input  *objectSchema `json:"input" validate:"omitnil"`
	Output *objectSchema `json:"output" validate:"omitnil"`
}
