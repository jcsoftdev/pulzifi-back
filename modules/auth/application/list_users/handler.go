package listusers

import (
	"context"

	"github.com/google/uuid"
)

const maxPageSize = 50

// AdminUserRow is the read model returned by the AdminUserReader port.
// It embeds the is_super_admin flag computed by the persistence layer,
// avoiding an N+1 role lookup per user.
type AdminUserRow struct {
	ID            uuid.UUID
	Email         string
	FirstName     string
	LastName      string
	IsSuperAdmin  bool
	Status        string
	EmailVerified bool
	OrgCount      int
}

// ListFilter carries the inbound filters for the list query.
type ListFilter struct {
	Search string
	OrgID  string // optional; restrict to members of this org
	Status string // optional; "approved" | "suspended" | "trial_expired"
}

// AdminUserReader is the narrow port used by this use case.
// Implemented by the persistence layer; also satisfied by test fakes.
type AdminUserReader interface {
	ListUsers(ctx context.Context, filter ListFilter, limit, offset int) (rows []AdminUserRow, total int, err error)
}

// Request carries the inbound query parameters.
type Request struct {
	Search   string
	Page     int
	PageSize int
	OrgID    string
	Status   string
}

// UserDTO is the per-user shape in the JSON response.
type UserDTO struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	IsSuperAdmin  bool   `json:"isSuperAdmin"`
	Status        string `json:"status"`
	EmailVerified bool   `json:"emailVerified"`
	OrgCount      int    `json:"orgCount"`
}

// Response is the JSON body returned on success.
type Response struct {
	Users    []UserDTO `json:"users"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"pageSize"`
}

// Handler orchestrates the list-users use case.
type Handler struct {
	reader AdminUserReader
}

// NewHandler creates a Handler backed by the given reader.
func NewHandler(reader AdminUserReader) *Handler {
	return &Handler{reader: reader}
}

// Handle validates + normalises input, delegates to the reader, and returns
// the paginated response DTO.
func (h *Handler) Handle(ctx context.Context, req Request) (*Response, error) {
	// Normalise page (minimum 1).
	page := req.Page
	if page < 1 {
		page = 1
	}

	// Clamp pageSize.
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 5
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	offset := (page - 1) * pageSize

	rows, total, err := h.reader.ListUsers(ctx, ListFilter{
		Search: req.Search,
		OrgID:  req.OrgID,
		Status: req.Status,
	}, pageSize, offset)
	if err != nil {
		return nil, err
	}

	dtos := make([]UserDTO, 0, len(rows))
	for _, r := range rows {
		dtos = append(dtos, UserDTO{
			ID:            r.ID.String(),
			Email:         r.Email,
			FirstName:     r.FirstName,
			LastName:      r.LastName,
			IsSuperAdmin:  r.IsSuperAdmin,
			Status:        r.Status,
			EmailVerified: r.EmailVerified,
			OrgCount:      r.OrgCount,
		})
	}

	return &Response{
		Users:    dtos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
