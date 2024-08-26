package asana

type User struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

type Like struct {
	LikeGID  string `json:"like-gid"`
	UserGID  string `json:"user-gid"`
	UserName string `json:"name"`
}

type SimpleProject struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

type Goal struct {
	GID       string `json:"gid" api:"gid"`
	Name      string `json:"name" api:"name"`
	Owner     User   `json:"owner" api:"owner"`
	Notes     string `json:"notes" api:"notes"`
	HTMLNotes string `json:"html-notes" api:"html_notes"`
	DueOn     string `json:"due-on" api:"due_on"`
	StartOn   string `json:"start-on" api:"start_on"`
	Liked     bool   `json:"liked" api:"liked"`
	Likes     []Like `json:"likes" api:"likes"`
}

type Task struct {
	GID       string          `json:"gid" api:"gid"`
	Name      string          `json:"name" api:"name"`
	Notes     string          `json:"notes" api:"notes"`
	HTMLNotes string          `json:"html-notes" api:"html_notes"`
	Projects  []SimpleProject `json:"projects" api:"projects"`
	DueOn     string          `json:"due-on" api:"due_on"`
	StartOn   string          `json:"start-on" api:"start_on"`
	Liked     bool            `json:"liked" api:"liked"`
	Likes     []Like          `json:"likes" api:"likes"`
}

type Project struct {
	GID                 string            `json:"gid" api:"gid"`
	Name                string            `json:"name" api:"name"`
	Owner               User              `json:"owner" api:"owner"`
	Notes               string            `json:"notes" api:"notes"`
	HTMLNotes           string            `json:"html-notes" api:"html_notes"`
	DueOn               string            `json:"due-on" api:"due_on"`
	StartOn             string            `json:"start-on" api:"start_on"`
	Completed           bool              `json:"completed" api:"completed"`
	Color               string            `json:"color" api:"color"`
	PrivacySetting      string            `json:"privacy-setting" api:"privacy_setting"`
	Archived            bool              `json:"archived" api:"archived"`
	CompletedBy         User              `json:"completed-by" api:"completed_by"`
	CurrentStatus       map[string]string `json:"current-status" api:"current_status"`
	CustomFields        map[string]string `json:"custom-fields" api:"custom_fields"`
	CustomFieldSettings map[string]string `json:"custom-field-settings" api:"custom_field_settings"`
	ModifiedAt          string            `json:"modified-at" api:"modified_at"`
}

type Portfolio struct {
	GID                 string            `json:"gid" api:"gid"`
	Name                string            `json:"name" api:"name"`
	Owner               User              `json:"owner" api:"owner"`
	DueOn               string            `json:"due-on" api:"due_on"`
	StartOn             string            `json:"start-on" api:"start_on"`
	Color               string            `json:"color" api:"color"`
	Public              bool              `json:"public" api:"public"`
	CompletedBy         User              `json:"completed-by" api:"completed_by"`
	CurrentStatus       map[string]string `json:"current-status" api:"current_status"`
	CustomFields        map[string]string `json:"custom-fields" api:"custom_fields"`
	CustomFieldSettings map[string]string `json:"custom-field-settings" api:"custom_field_settings"`
}
