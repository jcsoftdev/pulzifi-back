package services_test

import (
	"testing"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/services"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"
)

// baseProfile returns a ProfileData suitable for use as a baseline snapshot.
func baseProfile() entities.ProfileData {
	return entities.ProfileData{
		Bio:            "Hello world",
		DisplayName:    "NASA",
		AvatarURL:      "https://storage.example.com/avatar1.jpg",
		FollowersCount: 1000,
		FollowingCount: 50,
		PostsCount:     10,
		Posts: []entities.Post{
			{ExternalID: "post-1", Caption: "First post", PostedAt: time.Now().Add(-24 * time.Hour)},
			{ExternalID: "post-2", Caption: "Second post", PostedAt: time.Now().Add(-12 * time.Hour)},
		},
	}
}

// hasChangeType checks whether a specific ChangeType is in the slice.
func hasChangeType(types []valueobjects.ChangeType, target valueobjects.ChangeType) bool {
	for _, ct := range types {
		if ct == target {
			return true
		}
	}
	return false
}

// --- no change ---

func TestDiff_NoChange(t *testing.T) {
	prev := baseProfile()
	next := baseProfile() // identical

	changeTypes, summary := services.Diff(prev, next)

	if len(changeTypes) != 0 {
		t.Errorf("expected 0 change types, got %d: %v", len(changeTypes), changeTypes)
	}
	_ = summary // summary fields should all be zero/nil
}

// --- new_post ---

func TestDiff_NewPost(t *testing.T) {
	prev := baseProfile()
	next := baseProfile()
	next.Posts = append(next.Posts, entities.Post{
		ExternalID: "post-3",
		Caption:    "Brand new post",
		PostedAt:   time.Now(),
	})

	changeTypes, summary := services.Diff(prev, next)

	if !hasChangeType(changeTypes, valueobjects.ChangeTypeNewPost) {
		t.Errorf("expected new_post in change types, got %v", changeTypes)
	}
	if len(summary.NewPosts) != 1 || summary.NewPosts[0] != "post-3" {
		t.Errorf("expected summary.NewPosts = [post-3], got %v", summary.NewPosts)
	}
}

// --- removed_post ---

func TestDiff_RemovedPost(t *testing.T) {
	prev := baseProfile()
	next := baseProfile()
	// Remove post-2 from the next snapshot
	next.Posts = []entities.Post{
		{ExternalID: "post-1", Caption: "First post", PostedAt: time.Now().Add(-24 * time.Hour)},
	}

	changeTypes, summary := services.Diff(prev, next)

	if !hasChangeType(changeTypes, valueobjects.ChangeTypeRemovedPost) {
		t.Errorf("expected removed_post in change types, got %v", changeTypes)
	}
	if len(summary.RemovedPosts) != 1 || summary.RemovedPosts[0] != "post-2" {
		t.Errorf("expected summary.RemovedPosts = [post-2], got %v", summary.RemovedPosts)
	}
}

// --- caption_edited ---

func TestDiff_CaptionEdited(t *testing.T) {
	prev := baseProfile()
	next := baseProfile()
	next.Posts[0].Caption = "First post (updated)"

	changeTypes, summary := services.Diff(prev, next)

	if !hasChangeType(changeTypes, valueobjects.ChangeTypeCaptionEdited) {
		t.Errorf("expected caption_edited in change types, got %v", changeTypes)
	}
	if len(summary.EditedCaptions) != 1 {
		t.Fatalf("expected 1 edited caption, got %d", len(summary.EditedCaptions))
	}
	ec := summary.EditedCaptions[0]
	if ec.ExternalID != "post-1" {
		t.Errorf("expected ExternalID=post-1, got %q", ec.ExternalID)
	}
	if ec.From != "First post" {
		t.Errorf("expected From=%q, got %q", "First post", ec.From)
	}
	if ec.To != "First post (updated)" {
		t.Errorf("expected To=%q, got %q", "First post (updated)", ec.To)
	}
}

// --- bio_changed ---

func TestDiff_BioChanged(t *testing.T) {
	prev := baseProfile()
	next := baseProfile()
	next.Bio = "Hello universe"

	changeTypes, summary := services.Diff(prev, next)

	if !hasChangeType(changeTypes, valueobjects.ChangeTypeBioChanged) {
		t.Errorf("expected bio_changed in change types, got %v", changeTypes)
	}
	if summary.Bio == nil {
		t.Fatal("expected summary.Bio to be set, got nil")
	}
	if summary.Bio.From != "Hello world" {
		t.Errorf("Bio.From = %q, want %q", summary.Bio.From, "Hello world")
	}
	if summary.Bio.To != "Hello universe" {
		t.Errorf("Bio.To = %q, want %q", summary.Bio.To, "Hello universe")
	}
}

// --- followers_changed ---

func TestDiff_FollowersChanged(t *testing.T) {
	prev := baseProfile()
	next := baseProfile()
	next.FollowersCount = 1120 // +120

	changeTypes, summary := services.Diff(prev, next)

	if !hasChangeType(changeTypes, valueobjects.ChangeTypeFollowersChanged) {
		t.Errorf("expected followers_changed in change types, got %v", changeTypes)
	}
	if summary.FollowersDelta != 120 {
		t.Errorf("FollowersDelta = %d, want 120", summary.FollowersDelta)
	}
}

func TestDiff_FollowersUnchanged_NoEvent(t *testing.T) {
	prev := baseProfile()
	next := baseProfile()
	// same follower count — no event expected

	changeTypes, _ := services.Diff(prev, next)

	if hasChangeType(changeTypes, valueobjects.ChangeTypeFollowersChanged) {
		t.Error("expected no followers_changed when count is identical")
	}
}

// --- avatar_changed ---
// Avatar URL comparison is disabled: Instagram CDN URLs are ephemeral (token in URL
// changes every Apify call). Until we hash the downloaded image we skip this signal.

func TestDiff_AvatarURL_NotDetected(t *testing.T) {
	prev := baseProfile()
	next := baseProfile()
	next.AvatarURL = "https://cdn.instagram.com/avatar2.jpg?se=<new-token>"

	changeTypes, summary := services.Diff(prev, next)

	if hasChangeType(changeTypes, valueobjects.ChangeTypeAvatarChanged) {
		t.Errorf("avatar_changed must NOT be emitted (ephemeral CDN URL); got %v", changeTypes)
	}
	if summary.AvatarChanged {
		t.Error("summary.AvatarChanged must be false")
	}
}

// --- display_name_changed ---

func TestDiff_DisplayNameChanged(t *testing.T) {
	prev := baseProfile()
	next := baseProfile()
	next.DisplayName = "NASA Official"

	changeTypes, summary := services.Diff(prev, next)

	if !hasChangeType(changeTypes, valueobjects.ChangeTypeDisplayNameChanged) {
		t.Errorf("expected display_name_changed in change types, got %v", changeTypes)
	}
	if summary.DisplayName == nil {
		t.Fatal("expected summary.DisplayName to be set, got nil")
	}
	if summary.DisplayName.From != "NASA" {
		t.Errorf("DisplayName.From = %q, want %q", summary.DisplayName.From, "NASA")
	}
	if summary.DisplayName.To != "NASA Official" {
		t.Errorf("DisplayName.To = %q, want %q", summary.DisplayName.To, "NASA Official")
	}
}

// --- pure function contract ---

func TestDiff_PureFunction_NoSideEffects(t *testing.T) {
	// Calling Diff multiple times with the same inputs must yield the same output.
	prev := baseProfile()
	next := baseProfile()
	next.Bio = "Different bio"
	next.FollowersCount = 2000

	types1, summary1 := services.Diff(prev, next)
	types2, summary2 := services.Diff(prev, next)

	if len(types1) != len(types2) {
		t.Errorf("non-deterministic: got %d types on first call, %d on second", len(types1), len(types2))
	}
	if summary1.FollowersDelta != summary2.FollowersDelta {
		t.Errorf("non-deterministic FollowersDelta: %d vs %d", summary1.FollowersDelta, summary2.FollowersDelta)
	}
}

// --- multiple changes in one diff ---

func TestDiff_MultipleChanges(t *testing.T) {
	prev := baseProfile()
	next := baseProfile()
	next.Bio = "New bio"
	next.FollowersCount = 1500
	next.Posts = append(next.Posts, entities.Post{ExternalID: "post-99", Caption: "New"})
	next.DisplayName = "NASA Space"
	next.AvatarURL = "https://storage.example.com/new-avatar.jpg"

	changeTypes, summary := services.Diff(prev, next)

	expectedTypes := []valueobjects.ChangeType{
		valueobjects.ChangeTypeBioChanged,
		valueobjects.ChangeTypeFollowersChanged,
		valueobjects.ChangeTypeNewPost,
		valueobjects.ChangeTypeDisplayNameChanged,
		// avatar_changed omitted — CDN URL comparison disabled (ephemeral tokens)
	}
	for _, expected := range expectedTypes {
		if !hasChangeType(changeTypes, expected) {
			t.Errorf("expected %q in change types %v", expected, changeTypes)
		}
	}

	if summary.Bio == nil {
		t.Error("expected Bio diff in summary")
	}
	if summary.FollowersDelta != 500 {
		t.Errorf("FollowersDelta = %d, want 500", summary.FollowersDelta)
	}
	if len(summary.NewPosts) != 1 {
		t.Errorf("expected 1 new post, got %d", len(summary.NewPosts))
	}
}
