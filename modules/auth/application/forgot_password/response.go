package forgotpassword

// Response is the forgot password response (always success to prevent enumeration).
type Response struct {
	Message string `json:"message"`
}
