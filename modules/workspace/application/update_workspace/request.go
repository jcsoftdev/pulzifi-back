package updateworkspace

type UpdateWorkspaceRequest struct {
	Name *string   `json:"name,omitempty"`
	Type *string   `json:"type,omitempty"`
	Tags *[]string `json:"tags,omitempty"`
}
