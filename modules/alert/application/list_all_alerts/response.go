package listallalerts

type AlertItem struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	PageID      string `json:"page_id"`
	CheckID     string `json:"check_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Read        bool   `json:"read"`
	CreatedAt   string `json:"created_at"`
	PageName    string `json:"page_name"`
	PageURL     string `json:"page_url"`
}

type ListAllAlertsResponse struct {
	Data  []AlertItem `json:"data"`
	Total int         `json:"total"`
}
