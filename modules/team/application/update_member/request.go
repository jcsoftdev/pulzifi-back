package updatemember

type UpdateMemberRequest struct {
	Role string `json:"role"` // "MEMBER", "ADMIN"
}
