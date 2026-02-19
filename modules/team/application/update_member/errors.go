package updatemember

import "errors"

var (
	ErrMemberNotFound        = errors.New("member not found")
	ErrCannotUpdateOwnerRole = errors.New("cannot change role of organization owner")
)
