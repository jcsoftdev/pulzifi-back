package invitetoplatform

import "time"

type Response struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	ExpiresAt     time.Time `json:"expires_at"`
	EmailDelivery string    `json:"email_delivery"` // "sent" | "failed"
}
