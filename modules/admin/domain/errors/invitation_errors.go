package errors

import "errors"

var (
	ErrCannotInviteEmail        = errors.New("invitation: cannot invite this email")
	ErrInvitationNotFound       = errors.New("invitation: not found or expired")
	ErrDailyCapExceeded         = errors.New("invitation: daily cap exceeded")
	ErrInvitationAlreadyDecided = errors.New("invitation: already accepted or revoked")
	ErrSubdomainTaken           = errors.New("invitation: subdomain already taken")
	ErrInvalidSubdomain         = errors.New("invitation: invalid subdomain")
	ErrSchemaProvisioning       = errors.New("invitation: schema provisioning failed")
	ErrSessionIssueFailed       = errors.New("invitation: session issuance failed")
)
