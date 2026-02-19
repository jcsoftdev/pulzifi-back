package invitemember

type InviteMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"` // "MEMBER", "ADMIN"
}
