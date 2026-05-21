package forgotpassword

// Request carries the email for password reset.
type Request struct {
	Email       string
	FrontendURL string
}
