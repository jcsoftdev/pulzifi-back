package getuserdetail

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrUserNotFound is returned when no user matches the given id.
var ErrUserNotFound = errors.New("user not found")

// Membership is one org membership of the user.
type Membership struct {
	OrgID            uuid.UUID
	OrgName          string
	Subdomain        string
	Role             string
	InvitationStatus string
	IsOwner          bool
}

// UserDetail is the full read model for a single user.
type UserDetail struct {
	ID            uuid.UUID
	Email         string
	FirstName     string
	LastName      string
	Status        string
	EmailVerified bool
	IsSuperAdmin  bool
	Memberships   []Membership
}

// Reader is the port implemented by the persistence layer.
type Reader interface {
	GetUserDetail(ctx context.Context, id uuid.UUID) (*UserDetail, error)
}

// MembershipDTO is the JSON shape of a membership.
type MembershipDTO struct {
	OrgID            string `json:"orgId"`
	OrgName          string `json:"orgName"`
	Subdomain        string `json:"subdomain"`
	Role             string `json:"role"`
	InvitationStatus string `json:"invitationStatus"`
	IsOwner          bool   `json:"isOwner"`
}

// Response is the JSON body returned on success.
type Response struct {
	ID            string          `json:"id"`
	Email         string          `json:"email"`
	FirstName     string          `json:"firstName"`
	LastName      string          `json:"lastName"`
	Status        string          `json:"status"`
	EmailVerified bool            `json:"emailVerified"`
	IsSuperAdmin  bool            `json:"isSuperAdmin"`
	Memberships   []MembershipDTO `json:"memberships"`
}

// Handler orchestrates the get_user_detail use case.
type Handler struct{ reader Reader }

// NewHandler creates a new Handler using the given Reader port.
func NewHandler(reader Reader) *Handler { return &Handler{reader: reader} }

// Handle fetches the user detail and maps it to the response DTO.
func (h *Handler) Handle(ctx context.Context, id uuid.UUID) (*Response, error) {
	d, err := h.reader.GetUserDetail(ctx, id)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, ErrUserNotFound
	}
	ms := make([]MembershipDTO, 0, len(d.Memberships))
	for _, m := range d.Memberships {
		ms = append(ms, MembershipDTO{
			OrgID:            m.OrgID.String(),
			OrgName:          m.OrgName,
			Subdomain:        m.Subdomain,
			Role:             m.Role,
			InvitationStatus: m.InvitationStatus,
			IsOwner:          m.IsOwner,
		})
	}
	return &Response{
		ID:            d.ID.String(),
		Email:         d.Email,
		FirstName:     d.FirstName,
		LastName:      d.LastName,
		Status:        d.Status,
		EmailVerified: d.EmailVerified,
		IsSuperAdmin:  d.IsSuperAdmin,
		Memberships:   ms,
	}, nil
}
