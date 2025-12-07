package errors

import "fmt"

// InvalidEmailError is returned when email data is invalid
type InvalidEmailError struct {
	Message string
}

func (e *InvalidEmailError) Error() string {
	return fmt.Sprintf("invalid email: %s", e.Message)
}

// EmailNotFoundError is returned when email is not found
type EmailNotFoundError struct {
	EmailID string
}

func (e *EmailNotFoundError) Error() string {
	return fmt.Sprintf("email not found: %s", e.EmailID)
}

// SendingFailedError is returned when email sending fails
type SendingFailedError struct {
	Reason string
}

func (e *SendingFailedError) Error() string {
	return fmt.Sprintf("sending failed: %s", e.Reason)
}
