package config

import (
	"os"
	"testing"
)

// TestLoad_SocialDefaults verifies that social media monitoring config vars
// default to the documented values when no env vars are set.
func TestLoad_SocialDefaults(t *testing.T) {
	env := requiredEnv()
	env["JWT_SECRET"] = "dev-secret"
	setEnvMap(t, env)
	for _, k := range []string{
		"SOCIAL_ENABLED",
		"APIFY_TOKEN",
		"APIFY_ACTOR_INSTAGRAM",
		"APIFY_ACTOR_TIKTOK",
		"APIFY_ACTOR_FACEBOOK",
		"SOCIAL_POSTS_PER_CHECK",
	} {
		os.Unsetenv(k)
	}

	cfg := Load()

	if cfg.SocialEnabled {
		t.Error("SocialEnabled default = true, want false")
	}
	if cfg.ApifyToken != "" {
		t.Errorf("ApifyToken default = %q, want empty string", cfg.ApifyToken)
	}
	if cfg.ApifyActorInstagram != "apify~instagram-profile-scraper" {
		t.Errorf("ApifyActorInstagram default = %q, want %q", cfg.ApifyActorInstagram, "apify~instagram-profile-scraper")
	}
	if cfg.ApifyActorTikTok != "" {
		t.Errorf("ApifyActorTikTok default = %q, want empty string", cfg.ApifyActorTikTok)
	}
	if cfg.ApifyActorFacebook != "" {
		t.Errorf("ApifyActorFacebook default = %q, want empty string", cfg.ApifyActorFacebook)
	}
	if cfg.SocialPostsPerCheck != 5 {
		t.Errorf("SocialPostsPerCheck default = %d, want 5", cfg.SocialPostsPerCheck)
	}
}

// TestLoad_SocialFromEnv verifies that all social config vars can be overridden.
func TestLoad_SocialFromEnv(t *testing.T) {
	env := requiredEnv()
	env["JWT_SECRET"] = "dev-secret"
	env["SOCIAL_ENABLED"] = "true"
	env["APIFY_TOKEN"] = "test-token"
	env["APIFY_ACTOR_INSTAGRAM"] = "custom~instagram-actor"
	env["APIFY_ACTOR_TIKTOK"] = "custom~tiktok-actor"
	env["APIFY_ACTOR_FACEBOOK"] = "custom~facebook-actor"
	env["SOCIAL_POSTS_PER_CHECK"] = "10"
	setEnvMap(t, env)

	cfg := Load()

	if !cfg.SocialEnabled {
		t.Error("SocialEnabled = false, want true")
	}
	if cfg.ApifyToken != "test-token" {
		t.Errorf("ApifyToken = %q, want %q", cfg.ApifyToken, "test-token")
	}
	if cfg.ApifyActorInstagram != "custom~instagram-actor" {
		t.Errorf("ApifyActorInstagram = %q, want %q", cfg.ApifyActorInstagram, "custom~instagram-actor")
	}
	if cfg.ApifyActorTikTok != "custom~tiktok-actor" {
		t.Errorf("ApifyActorTikTok = %q, want %q", cfg.ApifyActorTikTok, "custom~tiktok-actor")
	}
	if cfg.ApifyActorFacebook != "custom~facebook-actor" {
		t.Errorf("ApifyActorFacebook = %q, want %q", cfg.ApifyActorFacebook, "custom~facebook-actor")
	}
	if cfg.SocialPostsPerCheck != 10 {
		t.Errorf("SocialPostsPerCheck = %d, want 10", cfg.SocialPostsPerCheck)
	}
}
