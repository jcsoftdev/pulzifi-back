package valueobjects

import "fmt"

// Platform represents a supported social media platform.
type Platform string

const (
	PlatformInstagram Platform = "instagram"
	PlatformTikTok    Platform = "tiktok"
	PlatformFacebook  Platform = "facebook"
)

// Validate returns an error if the platform is not one of the supported values.
func (p Platform) Validate() error {
	switch p {
	case PlatformInstagram, PlatformTikTok, PlatformFacebook:
		return nil
	default:
		return fmt.Errorf("unsupported platform %q: must be one of instagram, tiktok, facebook", string(p))
	}
}

// String returns the string representation of the platform.
func (p Platform) String() string {
	return string(p)
}
