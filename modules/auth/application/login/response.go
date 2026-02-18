package login

type Response struct {
	SessionID string  `json:"session_id"`
	ExpiresIn int64   `json:"expires_in"`
	Tenant    *string `json:"tenant,omitempty"`
}
