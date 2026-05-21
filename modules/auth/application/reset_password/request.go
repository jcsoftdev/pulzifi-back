package resetpassword

// Request carries the token and new password.
type Request struct {
	Token       string
	NewPassword string
}
