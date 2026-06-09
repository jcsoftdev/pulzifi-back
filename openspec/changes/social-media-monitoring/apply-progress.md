# Apply Progress — social-media-monitoring

## Batches Completed
- Batch 1: DB migrations + config
- Batch 2: Domain Layer (complete)
- Batch 3: Application Use Cases (complete)
- Batch 4: Infrastructure (complete — postgres repos, Apify client+mapper, media store, scheduler, HTTP module)

## Tasks Done

### Batch 1
- [x] 1.1 — Config vars added to `shared/config/config.go`; TDD test written in `config_social_test.go` (RED→GREEN verified). Note: `.env.example` is permission-blocked by the sandbox — user must append the social section manually (see below).
- [x] 1.2 — Public migration `000029_plans_social_limits` (up + down)
- [x] 1.3 — Tenant migration `000024_create_social_profiles` (up + down)
- [x] 1.4 — Tenant migration `000025_create_social_snapshots_and_changes` (up + down)
- [x] 1.5 — Tenant migration `000026_create_social_check_usage` (up + down)

### Batch 2
- [x] 2.1 — `value_objects/platform.go` + `platform_test.go` (Platform enum, Validate(), String(), 3 valid values, 2 invalid; 7 subtests GREEN)
- [x] 2.2 — `value_objects/change_type.go` (7 ChangeType constants: new_post, removed_post, caption_edited, bio_changed, followers_changed, avatar_changed, display_name_changed)
- [x] 2.3 — `entities/profile_data.go` (ProfileData + Post structs, provider-agnostic)
- [x] 2.4 — `entities/profile.go` (SocialProfile with all design §3 fields + NewSocialProfile constructor)
- [x] 2.5 — `entities/snapshot.go` (SocialSnapshot, SnapshotStatus enum success|failed, NewSuccessSnapshot + NewFailedSnapshot constructors)
- [x] 2.6 — `entities/change.go` (SocialChange, ChangeSummary, TextDiff, CaptionDiff, NewSocialChange constructor)
- [x] 2.7 — `errors/errors.go` (ErrProfileNotFound, ErrQuotaExceeded, ErrPlatformNotSupported, ErrProfileLimitReached, ErrProfileAlreadyExists, ErrFetchFailed)
- [x] 2.8 — Repository interfaces: `profile_repository.go` (Create, GetByID, ListByWorkspace, Update, Delete, CountActiveByWorkspace, ListDue), `snapshot_repository.go` (Save, GetLatestByProfile, GetByID), `change_repository.go` (Save, List w/ ChangeFilter, GetByID)
- [x] 2.9 — Service port interfaces: `fetcher.go` (SocialFetcher.FetchProfile), `media_store.go` (MediaStore.Store), `quota.go` (CheckQuota.Consume + Compensate, QuotaResult)
- [x] 2.10 — **RED** `differ_test.go` — 11 tests covering all 7 ChangeType scenarios + no-change + multi-change + pure-function determinism (build failed before differ.go existed)
- [x] 2.11 — **GREEN** `differ.go` — pure Diff(prev, next ProfileData) ([]ChangeType, ChangeSummary) — all 11 tests pass
- [x] 2.12 — Manual mocks: `repositories/mocks/` (profile, snapshot, change) + `services/mocks/` (fetcher, media_store, quota)

## Files Touched

### Batch 1
| File | Action |
|------|--------|
| `shared/config/config.go` | Added `SocialEnabled`, `ApifyToken`, `ApifyActorInstagram`, `ApifyActorTikTok`, `ApifyActorFacebook`, `SocialPostsPerCheck` fields + Load() bindings |
| `shared/config/config_social_test.go` | New — TDD tests for social config defaults and env overrides |
| `shared/database/migrations/public/000029_plans_social_limits.up.sql` | New — ALTER TABLE plans ADD social_profiles_limit + social_checks_per_day + seed values |
| `shared/database/migrations/public/000029_plans_social_limits.down.sql` | New — DROP COLUMN rollback |
| `shared/database/migrations/tenant/000024_create_social_profiles.up.sql` | New — CREATE TABLE social_profiles + unique constraint + partial index |
| `shared/database/migrations/tenant/000024_create_social_profiles.down.sql` | New — DROP TABLE rollback |
| `shared/database/migrations/tenant/000025_create_social_snapshots_and_changes.up.sql` | New — CREATE TABLE social_snapshots + social_changes + indexes |
| `shared/database/migrations/tenant/000025_create_social_snapshots_and_changes.down.sql` | New — DROP TABLE rollback (changes first, then snapshots) |
| `shared/database/migrations/tenant/000026_create_social_check_usage.up.sql` | New — CREATE TABLE social_check_usage (DATE PRIMARY KEY) |
| `shared/database/migrations/tenant/000026_create_social_check_usage.down.sql` | New — DROP TABLE rollback |

### Batch 2
| File | Action |
|------|--------|
| `modules/social/domain/value_objects/platform.go` | New — Platform enum (instagram\|tiktok\|facebook), Validate(), String() |
| `modules/social/domain/value_objects/platform_test.go` | New — TDD tests (7 subtests) |
| `modules/social/domain/value_objects/change_type.go` | New — ChangeType enum (7 constants) |
| `modules/social/domain/entities/profile_data.go` | New — ProfileData, Post structs |
| `modules/social/domain/entities/profile.go` | New — SocialProfile entity + NewSocialProfile constructor |
| `modules/social/domain/entities/snapshot.go` | New — SocialSnapshot + SnapshotStatus enum + constructors |
| `modules/social/domain/entities/change.go` | New — SocialChange + ChangeSummary + TextDiff + CaptionDiff + NewSocialChange |
| `modules/social/domain/errors/errors.go` | New — 6 domain error sentinel values |
| `modules/social/domain/repositories/profile_repository.go` | New — ProfileRepository interface (7 methods) |
| `modules/social/domain/repositories/snapshot_repository.go` | New — SnapshotRepository interface (3 methods) |
| `modules/social/domain/repositories/change_repository.go` | New — ChangeRepository interface + ChangeFilter struct |
| `modules/social/domain/repositories/mocks/profile_repository_mock.go` | New — hand-rolled mock |
| `modules/social/domain/repositories/mocks/snapshot_repository_mock.go` | New — hand-rolled mock |
| `modules/social/domain/repositories/mocks/change_repository_mock.go` | New — hand-rolled mock |
| `modules/social/domain/services/fetcher.go` | New — SocialFetcher port interface |
| `modules/social/domain/services/media_store.go` | New — MediaStore port interface |
| `modules/social/domain/services/quota.go` | New — CheckQuota port interface + QuotaResult |
| `modules/social/domain/services/differ.go` | New — Diff() pure function |
| `modules/social/domain/services/differ_test.go` | New — 11 TDD tests (RED→GREEN) |
| `modules/social/domain/services/mocks/fetcher_mock.go` | New — hand-rolled mock |
| `modules/social/domain/services/mocks/media_store_mock.go` | New — hand-rolled mock |
| `modules/social/domain/services/mocks/quota_mock.go` | New — hand-rolled mock |

## Verification Results

### Batch 1
- `go test ./shared/config/...` — PASS (all 15 tests including new social tests)
- `go build ./...` — PASS (no compile errors)
- Migration file pairs: all present, no gaps/duplicates in numbering

### Batch 2
- `go test ./modules/social/...` — PASS (18 tests: 11 differ + 7 platform)
- `go build ./...` — PASS
- `./tools/scripts/check-architecture.sh` — PASS (520 files scanned, 0 violations)
- TDD RED→GREEN confirmed: differ_test.go failed before differ.go was created; platform_test.go confirmed via test-first write

### Batch 3 + Task 4.1
- `go test ./modules/social/...` — PASS (all 9 use case packages + domain packages: 47+ test cases)
- `go test -race ./modules/social/...` — PASS (race detector clean)
- `go build ./...` — PASS (540 files compiled)
- `./tools/scripts/check-architecture.sh` — PASS (540 files scanned, 0 violations)
- TDD RED→GREEN confirmed for all 9 use cases: each handler_test.go written before handler.go

## Manual Step Required
The `.env.example` file is permission-blocked in the CI sandbox. Append this section manually:

```
# SOCIAL MEDIA MONITORING (Apify-backed — Phase 1: Instagram)
# SOCIAL_ENABLED=false disables module registration and the scheduler entirely
# (same pattern as BILLING_ENABLED). Set to true to activate social monitoring.
SOCIAL_ENABLED=false
APIFY_TOKEN=                                   # Apify API token — https://console.apify.com/account/integrations
APIFY_ACTOR_INSTAGRAM=apidojo~instagram-scraper  # pay-per-result, ~$0.50/1k posts
APIFY_ACTOR_TIKTOK=                            # phase 2
APIFY_ACTOR_FACEBOOK=                          # phase 2
SOCIAL_POSTS_PER_CHECK=5                       # posts fetched per check (cost control)
```

### Batch 3 + Task 4.1
| File | Action |
|------|--------|
| `modules/social/domain/services/plan_limits.go` | New — PlanLimits port interface |
| `modules/social/domain/services/alert_creator.go` | New — AlertCreator port + AlertPayload |
| `modules/social/domain/services/mocks/plan_limits_mock.go` | New — hand-rolled mock |
| `modules/social/domain/services/mocks/alert_creator_mock.go` | New — hand-rolled mock |
| `modules/social/domain/repositories/mocks/profile_repository_mock.go` | Modified — added LastUpdatedProfile |
| `modules/social/infrastructure/persistence/memory/profile_repo.go` | New — in-memory ProfileRepository |
| `modules/social/infrastructure/persistence/memory/snapshot_repo.go` | New — in-memory SnapshotRepository |
| `modules/social/infrastructure/persistence/memory/change_repo.go` | New — in-memory ChangeRepository |
| `modules/social/infrastructure/persistence/memory/check_quota.go` | New — in-memory CheckQuota |
| `modules/social/infrastructure/http/module.go` | New — stub package (full impl Batch 4) |
| `modules/social/application/create_profile/handler.go` | New — platform/handle/preset/plan validation |
| `modules/social/application/create_profile/request.go` | New |
| `modules/social/application/create_profile/response.go` | New |
| `modules/social/application/create_profile/handler_test.go` | New — 7 test cases (RED→GREEN) |
| `modules/social/application/list_profiles/handler.go` | New |
| `modules/social/application/list_profiles/response.go` | New |
| `modules/social/application/list_profiles/handler_test.go` | New — 3 test cases (RED→GREEN) |
| `modules/social/application/get_profile/handler.go` | New |
| `modules/social/application/get_profile/response.go` | New |
| `modules/social/application/get_profile/handler_test.go` | New — 3 test cases (RED→GREEN) |
| `modules/social/application/update_profile/handler.go` | New — partial update, is_active scheduler semantics |
| `modules/social/application/update_profile/request.go` | New |
| `modules/social/application/update_profile/response.go` | New |
| `modules/social/application/update_profile/handler_test.go` | New — 6 test cases (RED→GREEN) |
| `modules/social/application/delete_profile/handler.go` | New |
| `modules/social/application/delete_profile/handler_test.go` | New — 3 test cases (RED→GREEN) |
| `modules/social/application/run_check/handler.go` | New — full check cycle with quota compensation, backoff, deactivation |
| `modules/social/application/run_check/response.go` | New |
| `modules/social/application/run_check/handler_test.go` | New — 7 test cases (RED→GREEN) |
| `modules/social/application/list_changes/handler.go` | New |
| `modules/social/application/list_changes/request.go` | New |
| `modules/social/application/list_changes/response.go` | New |
| `modules/social/application/list_changes/handler_test.go` | New — 5 test cases (RED→GREEN) |
| `modules/social/application/get_change/handler.go` | New |
| `modules/social/application/get_change/response.go` | New |
| `modules/social/application/get_change/handler_test.go` | New — 2 test cases (RED→GREEN) |
| `modules/social/application/get_quota_status/handler.go` | New — read-only, nil limit for unlimited |
| `modules/social/application/get_quota_status/response.go` | New |
| `modules/social/application/get_quota_status/handler_test.go` | New — 3 test cases (RED→GREEN) |

## Commits Made
- `0d0c69a` — `✨ feat(migrations): add social media monitoring DB schema`
- `1062372` — `✨ feat(config): add social media monitoring config vars`
- `5484ae8` — `✨ feat(social): add domain layer — entities, value objects, ports, differ (TDD)`
- `b29e1ef` — `✨ feat(social): add in-memory repos, AlertCreator and PlanLimits ports`
- `0a88cf6` — `✨ feat(social): add application use cases — TDD RED→GREEN (Batch 3)`

### Batch 3 + Task 4.1
- [x] 3.1 — Directory structure manually created (module dir already existed; script would have errored)
- [x] 3.2 — **RED** `create_profile/handler_test.go` — 7 test cases (happy path, invalid platform, empty handle, invalid preset, plan limit, duplicate, unlimited plan, feature disabled)
- [x] 3.3 — **GREEN** `create_profile/handler.go + request.go + response.go` — all validation passes
- [x] 3.4 — **RED** `list_profiles/handler_test.go` — 3 test cases (scoped list, empty, repo error)
- [x] 3.5 — **GREEN** `list_profiles/handler.go + response.go`
- [x] 3.6 — **RED** `get_profile/handler_test.go` — 3 test cases (with snapshot, nil snapshot, 404)
- [x] 3.7 — **GREEN** `get_profile/handler.go + response.go`
- [x] 3.8 — **RED** `update_profile/handler_test.go` — 5 test cases (interval update, deactivate, re-activate, invalid preset, interval exceeds plan, 404)
- [x] 3.9 — **GREEN** `update_profile/handler.go + request.go + response.go`
- [x] 3.10 — **RED** `delete_profile/handler_test.go` — 3 test cases (happy, 404, repo error)
- [x] 3.11 — **GREEN** `delete_profile/handler.go`
- [x] 3.12 — **RED** `run_check/handler_test.go` — 7 test cases (baseline, change detected, no change, quota exhausted, provider 5xx with compensation, consecutive failures threshold, backoff tiers)
- [x] 3.13 — **GREEN** `run_check/handler.go + response.go` — full flow with quota compensation, backoff, deactivation
- [x] 3.14 — **RED** `list_changes/handler_test.go` — 5 test cases (profile filter, workspace filter, pagination, empty, precedence)
- [x] 3.15 — **GREEN** `list_changes/handler.go + request.go + response.go`
- [x] 3.16 — **RED** `get_change/handler_test.go` — 2 test cases (with before/after data, 404)
- [x] 3.17 — **GREEN** `get_change/handler.go + response.go`
- [x] 3.18 — **RED** `get_quota_status/handler_test.go` — 3 test cases (limited, unlimited, read-only)
- [x] 3.19 — **GREEN** `get_quota_status/handler.go + response.go`
- [x] 4.1 — `infrastructure/persistence/memory/` — ProfileRepo, SnapshotRepo, ChangeRepo, CheckQuota

New domain ports added:
- `domain/services/plan_limits.go` — PlanLimits interface (GetProfilesLimit, GetChecksPerDay, GetChecksUsedToday)
- `domain/services/alert_creator.go` — AlertCreator interface + AlertPayload
- `domain/services/mocks/plan_limits_mock.go` — hand-rolled mock
- `domain/services/mocks/alert_creator_mock.go` — hand-rolled mock
- `domain/repositories/mocks/profile_repository_mock.go` — added LastUpdatedProfile field

## Next Batch
**Batch 5: Wiring + Server Registration**

Tasks remaining: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6
Then: Batch 7 (quality gate)

### Batch 4 Tasks (all done)
- [x] 4.1 — `infrastructure/persistence/memory/` — ProfileRepo, SnapshotRepo, ChangeRepo, CheckQuota (done in Batch 3)
- [x] 4.2 — `infrastructure/persistence/postgres/profile_repo.go` — tenant-aware CRUD + ListDue (FOR UPDATE SKIP LOCKED)
- [x] 4.3 — `infrastructure/persistence/postgres/snapshot_repo.go` — Save, GetLatestByProfile, GetByID; Data as JSONB
- [x] 4.4 — `infrastructure/persistence/postgres/change_repo.go` — Save, List (profile+workspace filter), GetByID; change_types as TEXT[]
- [x] 4.5 — `infrastructure/persistence/postgres/check_quota_repo.go` — atomic ON CONFLICT WHERE upsert + compensate + GetUsageForDate
- [x] 4.6 — `infrastructure/apify/client.go` — SocialFetcher impl; 90s timeout; one retry on 5xx; per-platform adapter resolution
- [x] 4.7 — `infrastructure/apify/instagram_mapper.go` — TDD RED→GREEN; 5 tests; maps actor JSON → ProfileData
- [x] 4.8 — `infrastructure/storage/media_store.go` — MediaStore impl over MinIO/Cloudinary; download+reupload pattern
- [x] 4.9 — `infrastructure/scheduler/scheduler.go` — 30s poll; ListDue; bounded pool (10); quota exhausted → push to next midnight; SOCIAL_ENABLED gate
- [x] 4.10 — `infrastructure/http/module.go` — ModuleRegisterer; 9 routes; SOCIAL_ENABLED gate; domain error → HTTP status mapping

### Batch 4 Files Created
| File | Action |
|------|--------|
| `modules/social/infrastructure/persistence/postgres/profile_repo.go` | New — ProfilePostgresRepository (tenant-aware, database.WithTenant) |
| `modules/social/infrastructure/persistence/postgres/snapshot_repo.go` | New — SnapshotPostgresRepository (JSONB data serialization) |
| `modules/social/infrastructure/persistence/postgres/change_repo.go` | New — ChangePostgresRepository (pq.Array for TEXT[]) |
| `modules/social/infrastructure/persistence/postgres/check_quota_repo.go` | New — CheckQuotaPostgresRepository (atomic upsert + compensate + GetUsageForDate) |
| `modules/social/infrastructure/apify/client.go` | New — Apify SocialFetcher (90s timeout, one retry on 5xx) |
| `modules/social/infrastructure/apify/instagram_mapper.go` | New — Instagram actor JSON → ProfileData |
| `modules/social/infrastructure/apify/instagram_mapper_test.go` | New — 5 TDD tests (RED→GREEN) |
| `modules/social/infrastructure/storage/media_store.go` | New — MediaStore over MinIO/Cloudinary |
| `modules/social/infrastructure/scheduler/scheduler.go` | New — 30s poll scheduler; 10-worker pool; quota → midnight push |
| `modules/social/infrastructure/http/module.go` | New — ModuleRegisterer with all 9 routes gated by SOCIAL_ENABLED |

### Batch 4 Commits
- `31d0e34` — `✨ feat(social): add postgres repos — profile, snapshot, change, check quota (tenant-aware)`
- `53f2dfa` — `✨ feat(social): add Apify client + Instagram mapper (TDD RED→GREEN)`
- `b41510e` — `✨ feat(social): add media store, scheduler, and HTTP module`

### Batch 4 Verification
- `go build ./...` — PASS (569 files)
- `go test ./modules/social/...` — PASS (all packages + 5 new mapper tests)
- `go test -race ./modules/social/...` — PASS (race detector clean)
- `./tools/scripts/check-architecture.sh` — PASS (569 files scanned, 0 violations)
