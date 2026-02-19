package invitemember

import "errors"

var (
	ErrUserNotFound      = errors.New("user with that email not found")
	ErrAlreadyMember     = errors.New("user is already a member of this organization")
	ErrOrganizationNotFound = errors.New("organization not found")
)
