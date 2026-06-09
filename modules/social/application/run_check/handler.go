package runcheck

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/social/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/services"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"
)

const (
	// maxConsecutiveFailures is the threshold after which a profile is deactivated.
	maxConsecutiveFailures = 5
)

// Handler handles the run_check use case — the core social monitoring flow.
type Handler struct {
	profiles      repositories.ProfileRepository
	snapshots     repositories.SnapshotRepository
	changes       repositories.ChangeRepository
	quota         services.CheckQuota
	fetcher       services.SocialFetcher
	mediaStore    services.MediaStore
	alertCreator  services.AlertCreator
	planLimits    services.PlanLimits // optional; when non-nil, limit is resolved per-plan at runtime
	postsPerCheck int
	checksPerDay  int // 0 = unlimited; overridden at runtime when planLimits is set
}

// NewHandler creates a new Handler with a pre-resolved checksPerDay limit.
// Used by the scheduler path where the limit is resolved once at factory time.
func NewHandler(
	profiles repositories.ProfileRepository,
	snapshots repositories.SnapshotRepository,
	changes repositories.ChangeRepository,
	quota services.CheckQuota,
	fetcher services.SocialFetcher,
	mediaStore services.MediaStore,
	alertCreator services.AlertCreator,
	postsPerCheck int,
	checksPerDay int,
) *Handler {
	return &Handler{
		profiles:      profiles,
		snapshots:     snapshots,
		changes:       changes,
		quota:         quota,
		fetcher:       fetcher,
		mediaStore:    mediaStore,
		alertCreator:  alertCreator,
		postsPerCheck: postsPerCheck,
		checksPerDay:  checksPerDay,
	}
}

// NewHandlerWithPlanLimits creates a Handler that resolves the daily check limit
// from the PlanLimits port at call time (using the profile's workspaceID).
// Used by the HTTP manual-check path where no pre-resolved limit is available.
func NewHandlerWithPlanLimits(
	profiles repositories.ProfileRepository,
	snapshots repositories.SnapshotRepository,
	changes repositories.ChangeRepository,
	quota services.CheckQuota,
	fetcher services.SocialFetcher,
	mediaStore services.MediaStore,
	alertCreator services.AlertCreator,
	planLimits services.PlanLimits,
	postsPerCheck int,
) *Handler {
	return &Handler{
		profiles:      profiles,
		snapshots:     snapshots,
		changes:       changes,
		quota:         quota,
		fetcher:       fetcher,
		mediaStore:    mediaStore,
		alertCreator:  alertCreator,
		planLimits:    planLimits,
		postsPerCheck: postsPerCheck,
		checksPerDay:  0, // resolved at call time via planLimits
	}
}

// Handle executes a single check cycle for the given profile.
//
// Flow (REQ-CHECK-01 through REQ-CHECK-10, REQ-FAIL-01 through REQ-FAIL-06):
//  1. Consume quota FIRST — returns ErrQuotaExceeded when pool is empty.
//  2. Fetch live profile data from the provider (SocialFetcher).
//  3. On fetch failure: compensate quota + persist failed snapshot + backoff.
//  4. Store new media assets (MediaStore) for new posts.
//  5. Persist success snapshot.
//  6. If previous snapshot exists, diff → persist SocialChange + alert.
//  7. Reschedule next_check_at.
func (h *Handler) Handle(ctx context.Context, profileID uuid.UUID) (*Response, error) {
	// Load profile
	profile, err := h.profiles.GetByID(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("loading profile: %w", err)
	}

	// Resolve the effective daily limit.
	// When planLimits is set (HTTP path), resolve the real limit per-plan at call
	// time — this prevents the pre-wired checksPerDay=0 from being treated as
	// unlimited (REQ-QUOTA-CONSUME-01/02/03). The scheduler path pre-resolves the
	// limit at factory time (via GetChecksPerDayByTenant) and leaves planLimits nil.
	checksPerDay := h.checksPerDay
	if h.planLimits != nil {
		resolved, limErr := h.planLimits.GetChecksPerDay(ctx, profile.WorkspaceID.String())
		if limErr != nil {
			return nil, fmt.Errorf("resolving plan limit: %w", limErr)
		}
		// -1 = feature disabled; treat as 0-remaining limit → ErrQuotaExceeded path.
		if resolved == -1 {
			return nil, domainerrors.ErrQuotaExceeded
		}
		checksPerDay = resolved
	}

	// --- Step 1: Consume quota (REQ-CHECK-01, REQ-QUOTA-CONSUME-03) ---
	now := time.Now().UTC()
	result, err := h.quota.Consume(ctx, now, checksPerDay)
	if err != nil {
		return nil, err // ErrQuotaExceeded — do not proceed
	}
	_ = result

	// --- Step 2: Fetch live data ---
	data, fetchErr := h.fetcher.FetchProfile(ctx, profile.Platform, profile.Handle, h.postsPerCheck)
	if fetchErr != nil {
		// --- Failure handling (REQ-FAIL-01 through REQ-FAIL-06) ---
		return h.handleFetchFailure(ctx, profile, fetchErr, now)
	}

	// --- Step 3: Store new media for new posts (REQ-CHECK-03) ---
	prevSnapshot, err := h.snapshots.GetLatestByProfile(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("loading latest snapshot: %w", err)
	}

	if err := h.storeNewMedia(ctx, profile.ID, data, prevSnapshot); err != nil {
		// Media store failures are logged but not fatal — snapshot still captured.
		_ = err
	}

	// --- Step 4: Persist success snapshot (REQ-CHECK-05) ---
	snapshot := entities.NewSuccessSnapshot(profileID, data)
	if err := h.snapshots.Save(ctx, snapshot); err != nil {
		return nil, fmt.Errorf("saving snapshot: %w", err)
	}

	// --- Step 5: Diff vs previous + persist change + alert (REQ-CHECK-04, REQ-CHECK-06, REQ-CHECK-07) ---
	resp := &Response{
		SnapshotID:  snapshot.ID,
		Status:      string(snapshot.Status),
		NextCheckAt: profile.NextCheckAt,
		Data:        data,
	}

	if prevSnapshot != nil && prevSnapshot.Data != nil {
		// REQ-CHECK-04: diff only when a previous snapshot exists
		changeTypes, summary := services.Diff(*prevSnapshot.Data, *data)
		if len(changeTypes) > 0 {
			// REQ-CHECK-06: persist change
			change := entities.NewSocialChange(
				profileID,
				prevSnapshot.ID,
				snapshot.ID,
				changeTypes,
				summary,
			)
			if err := h.changes.Save(ctx, change); err != nil {
				return nil, fmt.Errorf("saving change: %w", err)
			}
			resp.ChangeCreated = true
			resp.ChangeID = &change.ID

			// REQ-CHECK-07: alert
			alertPayload := buildChangeAlert(profile, changeTypes, summary)
			_ = h.alertCreator.CreateAlert(ctx, alertPayload) // best-effort
		}
	}
	// REQ-CHECK-10: baseline (no previous snapshot) — no change, no alert

	// --- Step 6: Reschedule + reset failure counter (REQ-CHECK-08, REQ-CHECK-09) ---
	nextCheck := now.Add(time.Duration(profile.CheckIntervalMinutes) * time.Minute)
	profile.NextCheckAt = &nextCheck
	profile.LastCheckedAt = &now
	profile.ConsecutiveFailures = 0
	profile.UpdatedAt = now

	if err := h.profiles.Update(ctx, profile); err != nil {
		return nil, fmt.Errorf("updating profile: %w", err)
	}
	resp.NextCheckAt = profile.NextCheckAt

	return resp, nil
}

// handleFetchFailure handles the error path: compensate quota, persist failed
// snapshot, apply backoff, and deactivate profile if threshold reached.
func (h *Handler) handleFetchFailure(
	ctx context.Context,
	profile *entities.SocialProfile,
	fetchErr error,
	now time.Time,
) (*Response, error) {
	// REQ-QUOTA-CONSUME-04: compensate quota (best-effort, REQ-QUOTA-CONSUME-05)
	if compErr := h.quota.Compensate(ctx, now); compErr != nil {
		// Log only — do not surface as check failure (REQ-QUOTA-CONSUME-05)
		_ = compErr
	}

	// REQ-FAIL-02: persist failed snapshot
	snapshot := entities.NewFailedSnapshot(profile.ID, fetchErr.Error())
	_ = h.snapshots.Save(ctx, snapshot) // best-effort

	// REQ-FAIL-04: increment consecutive failures
	profile.ConsecutiveFailures++

	// REQ-FAIL-03: exponential backoff
	backoffMins := backoffMinutes(profile.ConsecutiveFailures, profile.CheckIntervalMinutes)
	nextCheck := now.Add(time.Duration(backoffMins) * time.Minute)
	profile.NextCheckAt = &nextCheck
	profile.LastCheckedAt = &now
	profile.UpdatedAt = now

	// REQ-FAIL-05: deactivate after threshold
	if profile.ConsecutiveFailures >= maxConsecutiveFailures {
		profile.IsActive = false
		profile.NextCheckAt = nil
		// Suspension alert
		_ = h.alertCreator.CreateAlert(ctx, services.AlertPayload{
			ProfileID: profile.ID,
			Handle:    profile.Handle,
			Platform:  profile.Platform,
			Suspended: true,
			Summary: fmt.Sprintf(
				"monitoring suspended for @%s after %d consecutive failures",
				profile.Handle, profile.ConsecutiveFailures,
			),
		})
	}

	if err := h.profiles.Update(ctx, profile); err != nil {
		return nil, fmt.Errorf("updating profile after failure: %w", err)
	}

	return nil, domainerrors.ErrFetchFailed
}

// storeNewMedia downloads and re-uploads media for posts that are new
// relative to the previous snapshot (REQ-CHECK-03).
func (h *Handler) storeNewMedia(
	ctx context.Context,
	profileID uuid.UUID,
	data *entities.ProfileData,
	prevSnapshot *entities.SocialSnapshot,
) error {
	prevPostIDs := make(map[string]bool)
	if prevSnapshot != nil && prevSnapshot.Data != nil {
		for _, p := range prevSnapshot.Data.Posts {
			prevPostIDs[p.ExternalID] = true
		}
	}

	for i, post := range data.Posts {
		isNew := !prevPostIDs[post.ExternalID]
		if !isNew || post.MediaURL == "" {
			continue
		}
		key := fmt.Sprintf("social/%s/posts/%s", profileID, post.ExternalID)
		storedURL, err := h.mediaStore.Store(ctx, post.MediaURL, key)
		if err != nil {
			continue // best-effort; CDN URL stays as fallback
		}
		data.Posts[i].StoredMediaURL = storedURL
	}
	return nil
}

// backoffMinutes returns the next-check delay in minutes based on the failure tier.
// Tier 1 (failures=1) → 5min, Tier 2 → 15min, Tier 3 → 60min, Tier 4+ → normal interval.
func backoffMinutes(consecutiveFailures, checkIntervalMinutes int) int {
	switch consecutiveFailures {
	case 1:
		return 5
	case 2:
		return 15
	case 3:
		return 60
	default:
		return checkIntervalMinutes
	}
}

// buildChangeAlert constructs an AlertPayload from detected changes.
func buildChangeAlert(
	profile *entities.SocialProfile,
	changeTypes []valueobjects.ChangeType,
	summary entities.ChangeSummary,
) services.AlertPayload {
	_ = summary
	return services.AlertPayload{
		ProfileID:   profile.ID,
		Handle:      profile.Handle,
		Platform:    profile.Platform,
		ChangeTypes: changeTypes,
		Summary:     fmt.Sprintf("changes detected on @%s", profile.Handle),
		Suspended:   false,
	}
}
