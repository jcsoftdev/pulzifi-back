package errors

import "fmt"

// AdminError represents admin-related errors
type AdminError struct {
	Code    string
	Message string
}

func (e AdminError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

var (
	ErrRegistrationRequestNotFound = AdminError{Code: "REGISTRATION_REQUEST_NOT_FOUND", Message: "registration request not found"}
	ErrAlreadyReviewed             = AdminError{Code: "ALREADY_REVIEWED", Message: "registration request has already been reviewed"}
)
