package errors

import "fmt"

// UserError represents user-related errors
type UserError struct {
	Code    string
	Message string
}

func (e UserError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewUserError creates a new user error
func NewUserError(code, message string) UserError {
	return UserError{Code: code, Message: message}
}

// Common user errors
var (
	ErrUserNotFound      = UserError{Code: "USER_NOT_FOUND", Message: "user not found"}
	ErrUserAlreadyExists = UserError{Code: "USER_ALREADY_EXISTS", Message: "user already exists"}
	ErrInvalidEmail      = UserError{Code: "INVALID_EMAIL", Message: "invalid email format"}
	ErrInvalidPassword   = UserError{Code: "INVALID_PASSWORD", Message: "invalid password"}
	ErrWeakPassword      = UserError{Code: "WEAK_PASSWORD", Message: "password is too weak"}
	ErrUnauthorized      = UserError{Code: "UNAUTHORIZED", Message: "unauthorized"}
	ErrUserNotApproved   = UserError{Code: "USER_NOT_APPROVED", Message: "account is pending approval"}
	ErrUserRejected      = UserError{Code: "USER_REJECTED", Message: "account has been rejected"}
)
