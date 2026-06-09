# Tasks: Social Media Monitoring (Phase 1 — Instagram MVP)

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 2,800–3,500 |
| 400-line budget risk | High |
| Chained PRs recommended | No (user chose single-pr with size:exception) |
| Suggested split | Single PR — size:exception approved |
| Delivery strategy | single-pr |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: High

### Suggested Work Units (informational only — single PR)

| Unit | Goal | Notes |
|------|------|-------|
| 1 | DB migrations + config | Foundation; everything depends on it |
| 2 | Domain layer + differ | Pure Go, no I/O; fully testable in isolation |
| 3 | Application use cases (TDD) | RED→GREEN per use case; in-memory repos |
| 4 | Infrastructure (postgres, apify, storage, scheduler) | Implements domain ports |
| 5 | Wiring + server/worker registration | Composes everything; no logic here |
| 6 | Frontend feature slice | Parallel to Unit 4–5 once service types are known |

---

## Batch 1: Config + Database Migrations (sequential foundation)

- [x] 1.1 Add `SOCIAL_ENABLED`, `APIFY_TOKEN`, `APIFY_ACTOR_INSTAGRAM`, `SOCIAL_POSTS_PER_CHECK` to `shared/config/config.go` and `.env.example`. Satisfies: REQ-FLAG-01, REQ-FLAG-03.
- [x] 1.2 Run `./tools/scripts/new-migration.sh public 000029_plans_social_limits`. Write `.up.sql` (`ALTER TABLE plans ADD COLUMN social_profiles_limit INT, social_checks_per_day INT`) and `.down.sql`. Satisfies: REQ-DB-04.
- [x] 1.3 Run `./tools/scripts/new-migration.sh tenant 000024_create_social_profiles`. Write `.up.sql` (table + unique constraint + `idx_social_profiles_due` partial index) and `.down.sql`. Satisfies: REQ-DB-01.
- [x] 1.4 Run `./tools/scripts/new-migration.sh tenant 000025_create_social_snapshots_and_changes`. Write `.up.sql` (both tables + indexes, `ON DELETE CASCADE`) and `.down.sql`. Satisfies: REQ-DB-02.
- [x] 1.5 Run `./tools/scripts/new-migration.sh tenant 000026_create_social_check_usage`. Write `.up.sql` (`usage_date DATE PRIMARY KEY`) and `.down.sql`. Satisfies: REQ-DB-03, REQ-QUOTA-CONSUME-06.

---

## Batch 2: Domain Layer (parallel within batch after Batch 1)

- [x] 2.1 Create `modules/social/domain/value_objects/platform.go` — `Platform` enum (`instagram|tiktok|facebook`), `Validate()`. Write test `platform_test.go`. Satisfies: REQ-PROFILE-01.
- [x] 2.2 Create `modules/social/domain/value_objects/change_type.go` — `ChangeType` enum (7 types). Satisfies: REQ-DIFF-01.
- [x] 2.3 Create `modules/social/domain/entities/profile_data.go` — `ProfileData`, `Post` structs (provider-agnostic). Satisfies: REQ-CHECK-02.
- [x] 2.4 Create `modules/social/domain/entities/profile.go` — `SocialProfile` entity with all fields from design §3. Satisfies: REQ-PROFILE-05, REQ-PROFILE-07, REQ-PROFILE-09.
- [x] 2.5 Create `modules/social/domain/entities/snapshot.go` — `SocialSnapshot`, status enum `success|failed`. Satisfies: REQ-CHECK-05, REQ-FAIL-02.
- [x] 2.6 Create `modules/social/domain/entities/change.go` — `SocialChange` with `[]ChangeType`, `Summary` (JSONB-compatible struct). Satisfies: REQ-CHECK-06, REQ-DIFF-03, REQ-DIFF-04, REQ-DIFF-05.
- [x] 2.7 Create `modules/social/domain/errors/errors.go` — `ErrProfileNotFound`, `ErrQuotaExceeded`, `ErrPlatformNotSupported`, `ErrProfileLimitReached`, `ErrProfileAlreadyExists`, `ErrFetchFailed`. Satisfies: REQ-PROFILE-03, REQ-PROFILE-04.
- [x] 2.8 Create port interfaces: `domain/repositories/profile_repository.go`, `snapshot_repository.go`, `change_repository.go`. Satisfies: REQ-PROFILE-11, REQ-FEED-01.
- [x] 2.9 Create port interfaces: `domain/services/fetcher.go` (`SocialFetcher`), `media_store.go` (`MediaStore`), `quota.go` (`CheckQuota — Consume + Compensate`). Satisfies: REQ-CHECK-01, REQ-CHECK-02, REQ-CHECK-03, REQ-QUOTA-CONSUME-02.
- [x] 2.10 **RED** Write `domain/services/differ_test.go` covering all 7 `ChangeType` scenarios + pure-function contract (no I/O). Satisfies: REQ-DIFF-01 through REQ-DIFF-06.
- [x] 2.11 **GREEN** Create `domain/services/differ.go` — `Diff(prev, next ProfileData) ([]ChangeType, Summary)` — pure function. Pass all tests from 2.10.
- [x] 2.12 Create `domain/repositories/mocks/` and `domain/services/mocks/` using interface-matching stubs (manual mocks, no codegen). Satisfies: REQ-WIRING-05 (enables unit tests without infra).

---

## Batch 3: Application Use Cases — TDD (sequential within each use case; use cases parallel to each other)

All use cases in `modules/social/application/`. Each directory: `handler.go`, `request.go`, `response.go`, `handler_test.go`. Tests use in-memory repos from Batch 4.1.

- [x] 3.1 Scaffold module with `./tools/scripts/new-module.sh social`. Remove stubs that don't apply; ensure package layout matches design §3. Satisfies: REQ-WIRING-05.
- [x] 3.2 **RED** Write `create_profile/handler_test.go` covering: happy path (201), duplicate (409), plan limit (422), invalid platform (400), invalid interval (400), interval-exceeds-plan (422). Satisfies: REQ-PROFILE-01 through REQ-PROFILE-07, REQ-QUOTA-INTERVAL-01 through REQ-QUOTA-INTERVAL-04.
- [x] 3.3 **GREEN** Implement `create_profile/` handler — validates platform, handle, presets, plan limits, interval vs plan capacity; sets `next_check_at=now()`. Pass 3.2 tests.
- [x] 3.4 **RED** Write `list_profiles/handler_test.go` — scoped to workspace + tenant. Satisfies: REQ-PROFILE-11, REQ-HTTP-04.
- [x] 3.5 **GREEN** Implement `list_profiles/` handler. Pass 3.4 tests.
- [x] 3.6 **RED** Write `get_profile/handler_test.go` — returns profile + latest snapshot; 404 on cross-tenant. Satisfies: REQ-PROFILE-12, REQ-HTTP-05.
- [x] 3.7 **GREEN** Implement `get_profile/` handler. Pass 3.6 tests.
- [x] 3.8 **RED** Write `update_profile/handler_test.go` — interval preset validation, `is_active=false` removes from scheduler, `is_active=true` resets `next_check_at`. Satisfies: REQ-PROFILE-08, REQ-PROFILE-09, REQ-HTTP-06.
- [x] 3.9 **GREEN** Implement `update_profile/` handler. Pass 3.8 tests.
- [x] 3.10 **RED** Write `delete_profile/handler_test.go` — 404 on not found. Satisfies: REQ-PROFILE-10, REQ-HTTP-07.
- [x] 3.11 **GREEN** Implement `delete_profile/` handler. Pass 3.10 tests.
- [x] 3.12 **RED** Write `run_check/handler_test.go` covering: baseline (no prev snapshot, no change created), change detected (diff → change → alert), no change (snapshot only), quota exhausted (stop before fetch), provider 5xx (compensate quota, failed snapshot, backoff), consecutive_failures=5 (deactivate + suspension alert). Satisfies: REQ-CHECK-01 through REQ-CHECK-10, REQ-FAIL-01 through REQ-FAIL-06, REQ-QUOTA-CONSUME-01 through REQ-QUOTA-CONSUME-05.
- [x] 3.13 **GREEN** Implement `run_check/` handler — full flow: Consume → Fetch → MediaStore → Snapshot → Diff → Change → Alert → Reschedule; quota compensation on 5xx; exponential backoff tiers; failure threshold deactivation. Pass 3.12 tests.
- [x] 3.14 **RED** Write `list_changes/handler_test.go` — filter by profile_id or workspace_id, descending order, pagination. Satisfies: REQ-FEED-01, REQ-FEED-02, REQ-HTTP-09.
- [x] 3.15 **GREEN** Implement `list_changes/` handler. Pass 3.14 tests.
- [x] 3.16 **RED** Write `get_change/handler_test.go` — full before/after ProfileData, 404 cross-tenant. Satisfies: REQ-FEED-03, REQ-HTTP-10.
- [x] 3.17 **GREEN** Implement `get_change/` handler. Pass 3.16 tests.
- [x] 3.18 **RED** Write `get_quota_status/handler_test.go` — checks_used, checks_limit (null for unlimited), resets_at; read-only. Satisfies: REQ-QUOTA-STATUS-01 through REQ-QUOTA-STATUS-05.
- [x] 3.19 **GREEN** Implement `get_quota_status/` handler. Pass 3.18 tests.

---

## Batch 4: Infrastructure (parallel sub-tasks after Batch 2; sequential within each component)

- [x] 4.1 Create `infrastructure/persistence/memory/` in-memory implementations for all three repos (profile, snapshot, change) + in-memory `CheckQuota`. Used by Batch 3 tests. Satisfies: test isolation requirement.
- [x] 4.2 Create `infrastructure/persistence/postgres/profile_repo.go` — tenant-aware (`tenant string` param, `middleware.GetSetSearchPathSQL`); all CRUD + `ListDue` (`FOR UPDATE SKIP LOCKED`). Satisfies: REQ-DB-01, REQ-SCHED-02.
- [x] 4.3 Create `infrastructure/persistence/postgres/snapshot_repo.go` — `Save`, `GetLatestByProfile`, `GetByID`. Satisfies: REQ-CHECK-05, REQ-FEED-03.
- [x] 4.4 Create `infrastructure/persistence/postgres/change_repo.go` — `Save`, `ListByProfile`, `ListByWorkspace` (paginated, desc order), `GetByID`. Satisfies: REQ-FEED-01, REQ-FEED-02.
- [x] 4.5 Create `infrastructure/persistence/postgres/check_quota_repo.go` — atomic upsert SQL (`ON CONFLICT DO UPDATE WHERE checks_used < $limit RETURNING checks_used`) + compensation decrement. Satisfies: REQ-QUOTA-CONSUME-01, REQ-QUOTA-CONSUME-04, REQ-QUOTA-CONSUME-05.
- [x] 4.6 Create `infrastructure/apify/client.go` — `POST /v2/acts/{actor}/run-sync-get-dataset-items?token=...` with `resultsLimit`, `context.WithTimeout` (90s), one retry on 5xx. Satisfies: REQ-CHECK-02, REQ-FAIL-01.
- [x] 4.7 Create `infrastructure/apify/instagram_mapper.go` — normalizes Apify actor JSON output → `ProfileData`. Satisfies: REQ-CHECK-02.
- [x] 4.8 Create `infrastructure/storage/media_store.go` — `MediaStore` impl over MinIO/Cloudinary (reuses `shared/config` `OBJECT_STORAGE_PROVIDER`). Downloads source URL, uploads under `social/{profileID}/{key}`. Satisfies: REQ-CHECK-03, D4.
- [x] 4.9 Create `infrastructure/scheduler/scheduler.go` — 30s poll, per-tenant due-profile query, bounded worker pool dispatching `run_check`; `ErrQuotaExceeded` → set `next_check_at` to next UTC midnight. Gated by `SOCIAL_ENABLED`. Satisfies: REQ-SCHED-01 through REQ-SCHED-06.
- [x] 4.10 Create `infrastructure/http/module.go` — `ModuleRegisterer` with all 9 routes; gated by `SOCIAL_ENABLED` (no routes registered when false). Satisfies: REQ-FLAG-01, REQ-HTTP-01 through REQ-HTTP-10.

---

## Batch 5: Wiring + Registration (sequential; after Batches 3–4)

- [x] 5.1 Create `cmd/wiring/social/alert_creator.go` — implements `social.AlertCreator` port via alert module (mirrors `cmd/wiring/snapshot` pattern). Satisfies: REQ-WIRING-02, REQ-CHECK-07, REQ-FAIL-05.
- [x] 5.2 Create `cmd/wiring/social/plan_lookup.go` — implements `social.PlanLimits` (`GetProfilesLimit`, `GetChecksPerDay`) via raw SQL on `public.plans` + `public.organization_plans`. Satisfies: REQ-WIRING-03, REQ-QUOTA-PLAN-01 through REQ-QUOTA-PLAN-03.
- [x] 5.3 Create `cmd/wiring/social/tenant_repo_factory.go` — per-tenant repo factory for the worker scheduler (mirrors `cmd/wiring/integration` pattern). Satisfies: REQ-WIRING-04.
- [x] 5.4 Register social module in `cmd/server/modules.go` gated by `SOCIAL_ENABLED` (mirrors `BILLING_ENABLED` pattern). Satisfies: REQ-FLAG-01, REQ-FLAG-02.
- [x] 5.5 Register social scheduler in `cmd/worker/` and in the monolith when `ENABLE_WORKERS=true`. Satisfies: REQ-SCHED-05.
- [x] 5.6 Run `./tools/scripts/check-architecture.sh` — verify no cross-module imports from `modules/social/`. Satisfies: REQ-WIRING-01, REQ-WIRING-05.

---

## Batch 6: Frontend Feature Slice (parallel with Batch 4–5 once service types are stable)

- [ ] 6.1 Create `packages/services/src/social.service.ts` — 13th service; all 9 API endpoints; tenant-aware client from `@workspace/shared-http`. Satisfies: REQ-FE-09, REQ-QUOTA-FE-03.
- [ ] 6.2 Create `features/social/domain/` — `SocialProfile`, `SocialChange`, `Platform`, `ChangeType` TypeScript types. Satisfies: REQ-FE-04, REQ-FE-07.
- [ ] 6.3 Create `features/social/application/` — hooks `useSocialProfiles`, `useSocialChanges`, `useSocialQuota`. `useSocialQuota` refetches after manual check response. Satisfies: REQ-FE-02, REQ-QUOTA-FE-03.
- [ ] 6.4 Create `features/social/ui/platform-icon.tsx` — IG/TikTok/FB icons (Phase 1: Instagram). Satisfies: REQ-FE-04, REQ-FE-05, REQ-FE-11.
- [ ] 6.5 Create `features/social/ui/social-quota-badge.tsx` — `{used}/{limit} checks today`; unlimited variant; exhausted warning color; add-profile button unaffected. Satisfies: REQ-QUOTA-FE-01, REQ-QUOTA-FE-02, REQ-QUOTA-FE-04.
- [ ] 6.6 Create `features/social/ui/social-profile-card.tsx` and `social-profile-grid.tsx` — avatar (stored URL), handle, platform badge, followers, time-since. Satisfies: REQ-FE-04, REQ-FE-11.
- [ ] 6.7 Create `features/social/ui/add-social-profile-dialog.tsx` — platform select (Instagram enabled), handle input, interval preset select; client-side validation. Satisfies: REQ-FE-03.
- [ ] 6.8 Modify workspace detail page (`(main)/workspaces/[id]/page.tsx`) — add `Pages | Social` tab split; Pages tab content UNTOUCHED; Social tab renders grid + quota badge + add button. Satisfies: REQ-FE-01, REQ-FE-02, REQ-FE-08.
- [ ] 6.9 Create `features/social/ui/changes/social-change-card.tsx` — per-`change_type` renderer: `new_post` (thumb + caption + date), `bio_changed` (inline diff), `followers_changed` (delta badge), others (summary string). Satisfies: REQ-FE-07.
- [ ] 6.10 Create `features/social/ui/changes/social-change-timeline.tsx`. Satisfies: REQ-FE-06, REQ-FE-10.
- [x] 6.11 Modify changes view (existing) — render `SocialChange` items in the same feed with platform icon + `Social` badge; clicking opens `social-change-card` instead of screenshot slider. Satisfies: REQ-FE-05, REQ-FE-06.
- [ ] 6.12 Create profile detail route `(main)/workspaces/[id]/social/[profileId]/page.tsx` — header (avatar, handle, platform badge, followers), latest posts grid, change timeline. Satisfies: REQ-FE-10.

---

## Batch 7: Quality Gate (sequential; after all batches)

- [ ] 7.1 Run `go test -race ./modules/social/...` — all unit tests pass (RED→GREEN from Batch 3). Satisfies: all REQ-CHECK-*, REQ-FAIL-*, REQ-DIFF-*, REQ-QUOTA-CONSUME-* acceptance scenarios.
- [ ] 7.2 Run `go test -race ./cmd/wiring/social/...` — wiring adapters covered. Satisfies: REQ-WIRING-01 through REQ-WIRING-05.
- [ ] 7.3 Run `make check-arch` (`./tools/scripts/check-architecture.sh`) — zero cross-module import violations. Satisfies: REQ-WIRING-01.
- [ ] 7.4 Run `bun run type-check` from `frontend/` — zero TypeScript errors. Satisfies: REQ-FE-09, REQ-FE-11.
- [ ] 7.5 Run `bun run lint:fix` (Biome) from `frontend/` — zero lint errors. Satisfies: REQ-FE-11.
- [ ] 7.6 Run `make swagger` — regenerate Swagger docs; verify social routes appear. Satisfies: REQ-HTTP-01 through REQ-HTTP-10.
- [ ] 7.7 Run `COVERAGE_FLOOR=15 ./tools/scripts/coverage-gate.sh c.out` — coverage floor passes.
