package createcheckoutsession

// Response is the output DTO for the CreateCheckoutSession use case.
type Response struct {
	CheckoutURL string `json:"checkout_url"`
}
