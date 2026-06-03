package saveonboardingprofile

// Response is returned on a successful save.
type Response struct {
	OnboardingCompleted bool `json:"onboarding_completed"`
}
