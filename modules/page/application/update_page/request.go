package updatepage

import "github.com/google/uuid"

type UpdatePageRequest struct {
	Name string   `json:"name"`
	URL  string   `json:"url"`
	Tags []string `json:"tags"`
}

type UpdatePageResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	URL  string    `json:"url"`
}
