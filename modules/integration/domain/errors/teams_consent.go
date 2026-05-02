package errors

import "fmt"

// ErrConsentRequired is returned (or wrapped by callers) when the OAuth
// callback indicates the user's tenant admin has not yet approved the
// app. Callers (HTTP layer) should redirect the user to AdminConsentURL
// with a friendly modal.
type ErrConsentRequired struct {
	AdminConsentURL string
}

func (e *ErrConsentRequired) Error() string {
	return fmt.Sprintf("microsoft teams: tenant admin consent required: %s", e.AdminConsentURL)
}
