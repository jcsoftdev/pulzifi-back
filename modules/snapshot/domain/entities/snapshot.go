package entities

import "time"

type SnapshotRequest struct {
	PageID     string `json:"page_id"`
	URL        string `json:"url"`
	SchemaName string `json:"schema_name"`
}

type SnapshotResult struct {
	PageID       string    `json:"page_id"`
	URL          string    `json:"url"`
	SchemaName   string    `json:"schema_name"`
	ImageURL     string    `json:"image_url"`
	HTMLURL      string    `json:"html_url,omitempty"`
	TextURL      string    `json:"text_url,omitempty"`
	ImageHash    string    `json:"image_hash,omitempty"`
	HTMLHash     string    `json:"html_hash,omitempty"`
	TextHash     string    `json:"text_hash,omitempty"`
	ContentHash  string    `json:"content_hash,omitempty"`
	Status       string    `json:"status"` // "success", "failed"
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
