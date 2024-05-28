package archetypeai

// fileQueryParams holds a query about an file. It is used as the input in
// e.g. video description or image summarization tasks.
type fileQueryParams struct {
	Query   string   `json:"query"`
	FileIDs []string `json:"file_ids"`
}

// summarizeOutput is used to return the output of a TASK_SUMMARIZE execution.
type summarizeOutput struct {
	Response string `json:"response"`
}

const (
	statusCompleted = "completed"
	statusFailed    = "failed"
)

// summarizeResp holds the response from the Archetype AI API call.
type summarizeResp struct {
	QueryID  string `json:"query_id"`
	Status   string `json:"status"`
	Response struct {
		ProcessedText string `json:"processed_text"`
	} `json:"response"`
}

type frameDescription struct {
	Timestamp   float32 `json:"timestamp"`
	FrameID     uint64  `json:"frame_id"`
	Description string  `json:"description"`
}

// describeResp holds the response from the Archetype AI API call.
type describeResp struct {
	QueryID  string             `json:"query_id"`
	Status   string             `json:"status"`
	Response []frameDescription `json:"response"`
}

// summarizeOutput is used to return the output of a TASK_DESCRIBE execution.
type describeOutput struct {
	Descriptions []frameDescription `json:"descriptions"`
}

// uploadFileParams holds the input of a file upload task.
type uploadFileParams struct {
	File string `json:"file"`
}

// uploadFileOutput is used to return the output of a file TASK_UPLOAD_FILE
// execution.
type uploadFileOutput struct {
	FileID string `json:"file_id"`
}

// uploadFileResp holds the response from the Archetype AI API call.
type uploadFileResp struct {
	uploadFileOutput

	IsValid bool     `json:"is_valid"`
	Errors  []string `json:"errors"`
}
