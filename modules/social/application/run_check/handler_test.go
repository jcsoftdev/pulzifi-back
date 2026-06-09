package runcheck_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/social/domain/errors"
	repomocks "github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories/mocks"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/services"
	svcmocks "github.com/jcsoftdev/pulzifi-back/modules/social/domain/services/mocks"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"
	"github.com/jcsoftdev/pulzifi-back/modules/social/infrastructure/persistence/memory"

	runcheck "github.com/jcsoftdev/pulzifi-back/modules/social/application/run_check"
)

func makeProfile(id uuid.UUID, consecutiveFailures int) *entities.SocialProfile {
	now := time.Now().UTC()
	return &entities.SocialProfile{
		ID:                   id,
		WorkspaceID:          uuid.New(),
		Platform:             valueobjects.PlatformInstagram,
		Handle:               "nasa",
		IsActive:             true,
		CheckIntervalMinutes: 720,
		NextCheckAt:          &now,
		ConsecutiveFailures:  consecutiveFailures,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

func makeProfileData() *entities.ProfileData {
	return &entities.ProfileData{
		Bio:            "We explore the universe",
		DisplayName:    "NASA",
		AvatarURL:      "https://storage.example.com/avatar.jpg",
		FollowersCount: 10000,
		FollowingCount: 50,
		PostsCount:     200,
		Posts: []entities.Post{
			{
				ExternalID:     "post-1",
				Caption:        "New mission!",
				MediaURL:       "https://cdn.instagram.com/post1.jpg",
				StoredMediaURL: "",
				PostedAt:       time.Now().UTC(),
			},
		},
	}
}

func makeHandler(
	profileRepo *repomocks.MockProfileRepository,
	snapshotRepo *repomocks.MockSnapshotRepository,
	changeRepo *repomocks.MockChangeRepository,
	quota services.CheckQuota,
	fetcher *svcmocks.MockSocialFetcher,
	mediaStore *svcmocks.MockMediaStore,
	alertCreator *svcmocks.MockAlertCreator,
	postsPerCheck int,
	checksPerDay int,
) *runcheck.Handler {
	return runcheck.NewHandler(
		profileRepo,
		snapshotRepo,
		changeRepo,
		quota,
		fetcher,
		mediaStore,
		alertCreator,
		postsPerCheck,
		checksPerDay,
	)
}

// --- Test: baseline (no previous snapshot) ---

func TestRunCheck_Baseline(t *testing.T) {
	profileID := uuid.New()
	profile := makeProfile(profileID, 0)

	profileRepo := &repomocks.MockProfileRepository{GetByIDResult: profile}
	snapshotRepo := &repomocks.MockSnapshotRepository{
		GetLatestByProfileResult: nil, // no previous snapshot
	}
	changeRepo := &repomocks.MockChangeRepository{}
	quota := memory.NewCheckQuota()
	fetcher := &svcmocks.MockSocialFetcher{FetchProfileResult: makeProfileData()}
	mediaStore := &svcmocks.MockMediaStore{StoreResult: "https://stored.example.com/post1.jpg"}
	alertCreator := &svcmocks.MockAlertCreator{}

	h := makeHandler(profileRepo, snapshotRepo, changeRepo, quota, fetcher, mediaStore, alertCreator, 5, 120)

	resp, err := h.Handle(context.Background(), profileID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	// Snapshot persisted
	if snapshotRepo.SaveCalls != 1 {
		t.Errorf("expected 1 snapshot save, got %d", snapshotRepo.SaveCalls)
	}
	// No change created on baseline
	if changeRepo.SaveCalls != 0 {
		t.Errorf("expected 0 change saves on baseline, got %d", changeRepo.SaveCalls)
	}
	// No alert on baseline
	if alertCreator.CreateAlertCalls != 0 {
		t.Errorf("expected 0 alert calls on baseline, got %d", alertCreator.CreateAlertCalls)
	}
	// Quota consumed
	if quota.UsedToday() != 1 {
		t.Errorf("expected 1 quota consumed, got %d", quota.UsedToday())
	}
	// Profile updated
	if profileRepo.UpdateCalls != 1 {
		t.Errorf("expected 1 profile update, got %d", profileRepo.UpdateCalls)
	}
	// Fetcher called
	if fetcher.FetchProfileCalls != 1 {
		t.Errorf("expected 1 fetcher call, got %d", fetcher.FetchProfileCalls)
	}
	// Media stored for new post
	if mediaStore.StoreCalls != 1 {
		t.Errorf("expected 1 media store call, got %d", mediaStore.StoreCalls)
	}
}

// --- Test: change detected ---

func TestRunCheck_ChangeDetected(t *testing.T) {
	profileID := uuid.New()
	profile := makeProfile(profileID, 0)

	prevData := &entities.ProfileData{
		Bio:            "Old bio",
		FollowersCount: 9000,
		Posts:          []entities.Post{{ExternalID: "post-old"}},
	}
	prevSnapshot := &entities.SocialSnapshot{
		ID:        uuid.New(),
		ProfileID: profileID,
		Status:    entities.SnapshotStatusSuccess,
		Data:      prevData,
	}

	newData := &entities.ProfileData{
		Bio:            "New bio",
		FollowersCount: 10000,
		Posts: []entities.Post{
			{ExternalID: "post-old"},
			{ExternalID: "post-new", MediaURL: "https://cdn.ig.com/new.jpg"},
		},
	}

	profileRepo := &repomocks.MockProfileRepository{GetByIDResult: profile}
	snapshotRepo := &repomocks.MockSnapshotRepository{GetLatestByProfileResult: prevSnapshot}
	changeRepo := &repomocks.MockChangeRepository{}
	quota := memory.NewCheckQuota()
	fetcher := &svcmocks.MockSocialFetcher{FetchProfileResult: newData}
	mediaStore := &svcmocks.MockMediaStore{StoreResult: "https://stored.example.com/new.jpg"}
	alertCreator := &svcmocks.MockAlertCreator{}

	h := makeHandler(profileRepo, snapshotRepo, changeRepo, quota, fetcher, mediaStore, alertCreator, 5, 120)

	resp, err := h.Handle(context.Background(), profileID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Change created
	if changeRepo.SaveCalls != 1 {
		t.Errorf("expected 1 change save, got %d", changeRepo.SaveCalls)
	}
	// Alert fired
	if alertCreator.CreateAlertCalls != 1 {
		t.Errorf("expected 1 alert call, got %d", alertCreator.CreateAlertCalls)
	}
	// Snapshot persisted
	if snapshotRepo.SaveCalls != 1 {
		t.Errorf("expected 1 snapshot save, got %d", snapshotRepo.SaveCalls)
	}
	if resp.ChangeCreated {
		// good
	}
	_ = resp
}

// --- Test: no change ---

func TestRunCheck_NoChange(t *testing.T) {
	profileID := uuid.New()
	profile := makeProfile(profileID, 0)

	sameData := &entities.ProfileData{
		Bio:            "Same bio",
		FollowersCount: 10000,
		Posts:          []entities.Post{{ExternalID: "post-1"}},
	}
	prevSnapshot := &entities.SocialSnapshot{
		ID:        uuid.New(),
		ProfileID: profileID,
		Status:    entities.SnapshotStatusSuccess,
		Data:      sameData,
	}

	profileRepo := &repomocks.MockProfileRepository{GetByIDResult: profile}
	snapshotRepo := &repomocks.MockSnapshotRepository{GetLatestByProfileResult: prevSnapshot}
	changeRepo := &repomocks.MockChangeRepository{}
	quota := memory.NewCheckQuota()
	fetcher := &svcmocks.MockSocialFetcher{FetchProfileResult: sameData}
	mediaStore := &svcmocks.MockMediaStore{StoreResult: "https://stored.example.com/p.jpg"}
	alertCreator := &svcmocks.MockAlertCreator{}

	h := makeHandler(profileRepo, snapshotRepo, changeRepo, quota, fetcher, mediaStore, alertCreator, 5, 120)

	resp, err := h.Handle(context.Background(), profileID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Snapshot still persisted
	if snapshotRepo.SaveCalls != 1 {
		t.Errorf("expected 1 snapshot save, got %d", snapshotRepo.SaveCalls)
	}
	// No change, no alert
	if changeRepo.SaveCalls != 0 {
		t.Errorf("expected 0 change saves, got %d", changeRepo.SaveCalls)
	}
	if alertCreator.CreateAlertCalls != 0 {
		t.Errorf("expected 0 alert calls, got %d", alertCreator.CreateAlertCalls)
	}
	if resp.ChangeCreated {
		t.Error("expected no change created")
	}
}

// --- Test: quota exhausted ---

func TestRunCheck_QuotaExhausted(t *testing.T) {
	profileID := uuid.New()
	profile := makeProfile(profileID, 0)

	// Use a mock quota that always returns exhausted
	quotaMock := &svcmocks.MockCheckQuota{
		ConsumeResult: services.QuotaResult{Allowed: false, Used: 2, Limit: 2},
		ConsumeErr:    domainerrors.ErrQuotaExceeded,
	}

	profileRepo := &repomocks.MockProfileRepository{GetByIDResult: profile}
	snapshotRepo := &repomocks.MockSnapshotRepository{}
	changeRepo := &repomocks.MockChangeRepository{}
	fetcher := &svcmocks.MockSocialFetcher{}
	mediaStore := &svcmocks.MockMediaStore{}
	alertCreator := &svcmocks.MockAlertCreator{}

	h := makeHandler(profileRepo, snapshotRepo, changeRepo, quotaMock, fetcher, mediaStore, alertCreator, 5, 2)

	_, err := h.Handle(context.Background(), profileID)

	if !errors.Is(err, domainerrors.ErrQuotaExceeded) {
		t.Errorf("expected ErrQuotaExceeded, got %v", err)
	}
	// Fetcher must NOT be called
	if fetcher.FetchProfileCalls != 0 {
		t.Errorf("fetcher should not be called when quota exhausted, got %d calls", fetcher.FetchProfileCalls)
	}
	// No snapshot persisted
	if snapshotRepo.SaveCalls != 0 {
		t.Errorf("no snapshot should be saved on quota exhaustion, got %d", snapshotRepo.SaveCalls)
	}
}

// --- Test: provider 5xx — quota compensated, failed snapshot ---

func TestRunCheck_Provider5xx(t *testing.T) {
	profileID := uuid.New()
	profile := makeProfile(profileID, 0)

	quota := memory.NewCheckQuota()
	fetcher := &svcmocks.MockSocialFetcher{
		FetchProfileErr: domainerrors.ErrFetchFailed,
	}

	profileRepo := &repomocks.MockProfileRepository{GetByIDResult: profile}
	snapshotRepo := &repomocks.MockSnapshotRepository{GetLatestByProfileResult: nil}
	changeRepo := &repomocks.MockChangeRepository{}
	mediaStore := &svcmocks.MockMediaStore{}
	alertCreator := &svcmocks.MockAlertCreator{}

	h := makeHandler(profileRepo, snapshotRepo, changeRepo, quota, fetcher, mediaStore, alertCreator, 5, 120)

	_, err := h.Handle(context.Background(), profileID)

	// Error returned (fetch failed)
	if err == nil {
		t.Fatal("expected error on provider 5xx, got nil")
	}
	// Quota was consumed then compensated back
	if quota.UsedToday() != 0 {
		t.Errorf("quota should be compensated back to 0 after 5xx, got %d", quota.UsedToday())
	}
	// Failed snapshot persisted
	if snapshotRepo.SaveCalls != 1 {
		t.Errorf("expected 1 failed snapshot save, got %d", snapshotRepo.SaveCalls)
	}
	// Profile updated (consecutive_failures incremented, backoff applied)
	if profileRepo.UpdateCalls != 1 {
		t.Errorf("expected 1 profile update, got %d", profileRepo.UpdateCalls)
	}
	// No change, no alert for the failed check
	if changeRepo.SaveCalls != 0 {
		t.Errorf("expected 0 change saves on failure, got %d", changeRepo.SaveCalls)
	}
}

// --- Test: consecutive failures reach threshold — profile deactivated ---

func TestRunCheck_ConsecutiveFailuresThreshold(t *testing.T) {
	profileID := uuid.New()
	// consecutive_failures = 4; this check will make it 5 → deactivate
	profile := makeProfile(profileID, 4)

	quota := memory.NewCheckQuota()
	fetcher := &svcmocks.MockSocialFetcher{
		FetchProfileErr: domainerrors.ErrFetchFailed,
	}
	profileRepo := &repomocks.MockProfileRepository{GetByIDResult: profile}
	snapshotRepo := &repomocks.MockSnapshotRepository{GetLatestByProfileResult: nil}
	changeRepo := &repomocks.MockChangeRepository{}
	mediaStore := &svcmocks.MockMediaStore{}
	alertCreator := &svcmocks.MockAlertCreator{}

	h := makeHandler(profileRepo, snapshotRepo, changeRepo, quota, fetcher, mediaStore, alertCreator, 5, 120)

	_, _ = h.Handle(context.Background(), profileID)

	// Profile must be deactivated
	updatedProfile := profileRepo.LastUpdatedProfile
	if updatedProfile == nil {
		t.Fatal("expected profile update, got nil")
	}
	if updatedProfile.IsActive {
		t.Error("profile should be deactivated after 5 consecutive failures")
	}
	if updatedProfile.ConsecutiveFailures != 5 {
		t.Errorf("expected consecutive_failures=5, got %d", updatedProfile.ConsecutiveFailures)
	}
	// Suspension alert must fire
	if alertCreator.CreateAlertCalls != 1 {
		t.Errorf("expected 1 suspension alert, got %d", alertCreator.CreateAlertCalls)
	}
	if !alertCreator.LastAlertPayload.Suspended {
		t.Error("alert payload should mark Suspended=true")
	}
}

// --- Test: backoff tiers ---

func TestRunCheck_BackoffTiers(t *testing.T) {
	tests := []struct {
		name                string
		consecutiveBefore   int
		expectedBackoffMins int
	}{
		{"first failure → 5min", 0, 5},
		{"second failure → 15min", 1, 15},
		{"third failure → 60min", 2, 60},
		{"fourth failure → normal interval", 3, 720}, // uses profile.CheckIntervalMinutes
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileID := uuid.New()
			profile := makeProfile(profileID, tt.consecutiveBefore)

			quota := memory.NewCheckQuota()
			fetcher := &svcmocks.MockSocialFetcher{
				FetchProfileErr: domainerrors.ErrFetchFailed,
			}
			profileRepo := &repomocks.MockProfileRepository{GetByIDResult: profile}
			snapshotRepo := &repomocks.MockSnapshotRepository{}
			changeRepo := &repomocks.MockChangeRepository{}
			mediaStore := &svcmocks.MockMediaStore{}
			alertCreator := &svcmocks.MockAlertCreator{}

			h := makeHandler(profileRepo, snapshotRepo, changeRepo, quota, fetcher, mediaStore, alertCreator, 5, 120)

			before := time.Now().UTC()
			_, _ = h.Handle(context.Background(), profileID)

			updated := profileRepo.LastUpdatedProfile
			if updated == nil {
				t.Fatal("expected profile update")
			}
			if updated.NextCheckAt == nil {
				t.Fatal("expected NextCheckAt to be set after failure backoff")
			}
			expectedMinDuration := time.Duration(tt.expectedBackoffMins-1) * time.Minute
			expectedMaxDuration := time.Duration(tt.expectedBackoffMins+1) * time.Minute
			actual := updated.NextCheckAt.Sub(before)
			if actual < expectedMinDuration || actual > expectedMaxDuration {
				t.Errorf("backoff out of range: expected ~%d min, got %v", tt.expectedBackoffMins, actual)
			}
		})
	}
}
