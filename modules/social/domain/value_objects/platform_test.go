package valueobjects_test

import (
	"testing"

	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"
)

func TestPlatformValidate(t *testing.T) {
	t.Run("instagram is valid", func(t *testing.T) {
		p := valueobjects.Platform("instagram")
		if err := p.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("tiktok is valid", func(t *testing.T) {
		p := valueobjects.Platform("tiktok")
		if err := p.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("facebook is valid", func(t *testing.T) {
		p := valueobjects.Platform("facebook")
		if err := p.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("empty string is invalid", func(t *testing.T) {
		p := valueobjects.Platform("")
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty platform, got nil")
		}
	})

	t.Run("unknown platform is invalid", func(t *testing.T) {
		p := valueobjects.Platform("twitter")
		if err := p.Validate(); err == nil {
			t.Error("expected error for unknown platform, got nil")
		}
	})

	t.Run("constants match expected values", func(t *testing.T) {
		if valueobjects.PlatformInstagram != "instagram" {
			t.Errorf("PlatformInstagram = %q, want %q", valueobjects.PlatformInstagram, "instagram")
		}
		if valueobjects.PlatformTikTok != "tiktok" {
			t.Errorf("PlatformTikTok = %q, want %q", valueobjects.PlatformTikTok, "tiktok")
		}
		if valueobjects.PlatformFacebook != "facebook" {
			t.Errorf("PlatformFacebook = %q, want %q", valueobjects.PlatformFacebook, "facebook")
		}
	})

	t.Run("String returns underlying value", func(t *testing.T) {
		p := valueobjects.PlatformInstagram
		if p.String() != "instagram" {
			t.Errorf("String() = %q, want %q", p.String(), "instagram")
		}
	})
}
