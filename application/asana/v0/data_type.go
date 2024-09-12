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

type RawLike struct {
	LikeGID string `json:"gid"`
	User    User   `json:"user"`
}

type Job struct {
	GID        string  `json:"gid" api:"gid"`
	NewTask    Task    `json:"task" api:"new_task.name"`
	NewProject Project `json:"project" api:"new_project.name"`
}

type SimpleProject struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

type Goal struct {
	GID       string `json:"gid" api:"gid"`
	Name      string `json:"name" api:"name"`
	Owner     User   `json:"owner" api:"owner.name"`
	Notes     string `json:"notes" api:"notes"`
	HTMLNotes string `json:"html-notes" api:"html_notes"`
	DueOn     string `json:"due-on" api:"due_on"`
	StartOn   string `json:"start-on" api:"start_on"`
	Liked     bool   `json:"liked" api:"liked"`
	Likes     []Like `json:"likes" api:"likes.user.name"`
}

type TaskParent struct {
	GID             string `json:"gid"`
	Name            string `json:"name"`
	ResourceSubtype string `json:"resource-subtype"`
	CreatedBy       User   `json:"created-by"`
}
type Task struct {
	GID             string          `json:"gid" api:"gid"`
	Name            string          `json:"name" api:"name"`
	Notes           string          `json:"notes" api:"notes"`
	HTMLNotes       string          `json:"html-notes" api:"html_notes"`
	Projects        []SimpleProject `json:"projects" api:"projects.name"`
	DueOn           string          `json:"due-on" api:"due_on"`
	StartOn         string          `json:"start-on" api:"start_on"`
	Liked           bool            `json:"liked" api:"liked"`
	Likes           []Like          `json:"likes" api:"likes.user.name"`
	ApprovalStatus  string          `json:"approval-status" api:"approval_status"`
	ResourceSubtype string          `json:"resource-subtype" api:"resource_subtype"`
	Completed       bool            `json:"completed" api:"completed"`
	Assignee        string          `json:"assignee" api:"assignee.name"`
	Parent          string          `json:"parent" api:"parent.name"`
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
	CreatedBy           User              `json:"created-by" api:"created_by"`
	CurrentStatus       map[string]string `json:"current-status" api:"current_status"`
	CustomFields        map[string]string `json:"custom-fields" api:"custom_fields"`
	CustomFieldSettings map[string]string `json:"custom-field-settings" api:"custom_field_settings"`
}
