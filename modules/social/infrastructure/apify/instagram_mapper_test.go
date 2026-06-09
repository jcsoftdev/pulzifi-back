package apify

import (
	"encoding/json"
	"testing"
	"time"
)

// fixture represents a realistic (but fictional) Apify Instagram actor response item.
// Based on apidojo~instagram-scraper output schema.
const fixtureInstagramItem = `{
	"username": "nasa",
	"fullName": "NASA",
	"biography": "Explore the universe.",
	"profilePicUrl": "https://cdn.instagram.com/nasa/avatar.jpg",
	"followersCount": 96000000,
	"followsCount": 80,
	"postsCount": 4200,
	"latestPosts": [
		{
			"id": "post001",
			"caption": "Beautiful planet Earth 🌍",
			"displayUrl": "https://cdn.instagram.com/posts/post001.jpg",
			"timestamp": "2026-01-15T10:00:00.000Z",
			"likesCount": 500000,
			"commentsCount": 2000
		},
		{
			"id": "post002",
			"caption": "Mars surface closeup",
			"displayUrl": "https://cdn.instagram.com/posts/post002.jpg",
			"timestamp": "2026-01-14T08:00:00.000Z",
			"likesCount": 300000,
			"commentsCount": 1500
		}
	]
}`

func TestMapInstagramItems_HappyPath(t *testing.T) {
	item := json.RawMessage(fixtureInstagramItem)
	data, err := mapInstagramItems([]json.RawMessage{item})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.Bio != "Explore the universe." {
		t.Errorf("Bio = %q, want %q", data.Bio, "Explore the universe.")
	}
	if data.DisplayName != "NASA" {
		t.Errorf("DisplayName = %q, want %q", data.DisplayName, "NASA")
	}
	if data.AvatarURL != "https://cdn.instagram.com/nasa/avatar.jpg" {
		t.Errorf("AvatarURL = %q, want %q", data.AvatarURL, "https://cdn.instagram.com/nasa/avatar.jpg")
	}
	if data.FollowersCount != 96000000 {
		t.Errorf("FollowersCount = %d, want %d", data.FollowersCount, 96000000)
	}
	if data.FollowingCount != 80 {
		t.Errorf("FollowingCount = %d, want %d", data.FollowingCount, 80)
	}
	if data.PostsCount != 4200 {
		t.Errorf("PostsCount = %d, want %d", data.PostsCount, 4200)
	}
	if len(data.Posts) != 2 {
		t.Fatalf("Posts count = %d, want 2", len(data.Posts))
	}

	p0 := data.Posts[0]
	if p0.ExternalID != "post001" {
		t.Errorf("Posts[0].ExternalID = %q, want %q", p0.ExternalID, "post001")
	}
	if p0.Caption != "Beautiful planet Earth 🌍" {
		t.Errorf("Posts[0].Caption = %q, want %q", p0.Caption, "Beautiful planet Earth 🌍")
	}
	if p0.MediaURL != "https://cdn.instagram.com/posts/post001.jpg" {
		t.Errorf("Posts[0].MediaURL = %q, want %q", p0.MediaURL, "https://cdn.instagram.com/posts/post001.jpg")
	}
	if p0.LikesCount != 500000 {
		t.Errorf("Posts[0].LikesCount = %d, want %d", p0.LikesCount, 500000)
	}
	if p0.CommentsCount != 2000 {
		t.Errorf("Posts[0].CommentsCount = %d, want %d", p0.CommentsCount, 2000)
	}
	expectedTime, _ := time.Parse(time.RFC3339, "2026-01-15T10:00:00Z")
	if !p0.PostedAt.Equal(expectedTime) {
		t.Errorf("Posts[0].PostedAt = %v, want %v", p0.PostedAt, expectedTime)
	}
}

func TestMapInstagramItems_EmptyItems(t *testing.T) {
	_, err := mapInstagramItems([]json.RawMessage{})
	if err == nil {
		t.Error("expected error for empty items, got nil")
	}
}

func TestMapInstagramItems_MultipleItems_UsesFirst(t *testing.T) {
	// Multiple items: mapper uses only the first (profile object).
	second := json.RawMessage(`{"username":"other","fullName":"Other","biography":"other","followersCount":1,"followsCount":0,"postsCount":0,"latestPosts":[]}`)
	data, err := mapInstagramItems([]json.RawMessage{json.RawMessage(fixtureInstagramItem), second})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.DisplayName != "NASA" {
		t.Errorf("expected first item's DisplayName NASA, got %q", data.DisplayName)
	}
}

func TestMapInstagramItems_PostsListNil(t *testing.T) {
	// Actor may return a profile with no latestPosts array (private or new account).
	noPostsItem := json.RawMessage(`{
		"username": "private_acc",
		"fullName": "Private",
		"biography": "private",
		"profilePicUrl": "",
		"followersCount": 0,
		"followsCount": 0,
		"postsCount": 0
	}`)
	data, err := mapInstagramItems([]json.RawMessage{noPostsItem})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data.Posts) != 0 {
		t.Errorf("expected 0 posts for nil latestPosts, got %d", len(data.Posts))
	}
}

func TestMapInstagramItems_InvalidJSON(t *testing.T) {
	_, err := mapInstagramItems([]json.RawMessage{json.RawMessage(`not-valid-json`)})
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}
