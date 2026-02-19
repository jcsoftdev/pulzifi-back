package removemember

import "errors"

var (
	ErrMemberNotFound   = errors.New("member not found")
	ErrCannotRemoveOwner = errors.New("cannot remove the organization owner")
	ErrCannotRemoveSelf  = errors.New("cannot remove yourself from the organization")
)
