package socialwiring

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	runcheck "github.com/jcsoftdev/pulzifi-back/modules/social/application/run_check"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/services"
	socialapify "github.com/jcsoftdev/pulzifi-back/modules/social/infrastructure/apify"
	socialpostgres "github.com/jcsoftdev/pulzifi-back/modules/social/infrastructure/persistence/postgres"
	socialscheduler "github.com/jcsoftdev/pulzifi-back/modules/social/infrastructure/scheduler"
	socialstorage "github.com/jcsoftdev/pulzifi-back/modules/social/infrastructure/storage"
	snapshotstorage "github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/storage"
	"github.com/jcsoftdev/pulzifi-back/shared/ai"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	socialpubsub "github.com/jcsoftdev/pulzifi-back/shared/pubsub"
	"go.uber.org/zap"
)

// Compile-time interface checks — verifies wiring satisfies scheduler ports.
var (
	_ socialscheduler.CheckHandlerFactory = (*TenantHandlerFactory)(nil)
)

// TenantHandlerFactory implements scheduler.CheckHandlerFactory.
// It constructs a fully-wired run_check.Handler for each tenant by composing
// all infrastructure and cross-module dependencies (REQ-WIRING-04).
//
// Pattern mirrors cmd/wiring/integration/tenant_repo_factory.go for repos,
// and cmd/wiring/monitoring/snapshot_worker.go for the object storage init.
type TenantHandlerFactory struct {
	db            *sql.DB
	bus           eventbus.MessageBus
	cfg           *config.Config
	fetcher       services.SocialFetcher
	mediaStore    services.MediaStore
	orgLookupInst *orgLookup
	broker        *socialpubsub.MemorySocialBroker
	emailProvider emailservices.EmailProvider
	aiClient      completer
}

// sseBrokerAdapter adapts *MemorySocialBroker to services.SocialSSEPublisher
// (uuid-keyed Publish), delegating to the string-keyed broker.
type sseBrokerAdapter struct{ b *socialpubsub.MemorySocialBroker }

func (a sseBrokerAdapter) Publish(profileID uuid.UUID, payload []byte) {
	a.b.Publish(profileID.String(), payload)
}

// NewTenantHandlerFactory builds a TenantHandlerFactory from shared dependencies.
// The Apify fetcher and MediaStore are singletons — they are safe for concurrent
// use and do not hold tenant-specific state.
func NewTenantHandlerFactory(db *sql.DB, bus eventbus.MessageBus, cfg *config.Config, emailProvider emailservices.EmailProvider) *TenantHandlerFactory {
	fetcher := socialapify.NewClient(
		cfg.ApifyToken,
		cfg.ApifyActorInstagram,
		cfg.ApifyActorTikTok,
		cfg.ApifyActorFacebook,
	)

	// Build object storage uploader (same provider as the snapshot module).
	objectStorage, err := snapshotstorage.NewObjectStorage(cfg)
	if err != nil {
		logger.Warn("social wiring: object storage unavailable — media will not be persisted",
			zap.Error(err))
		objectStorage = nil
	}

	var mediaStore services.MediaStore
	if objectStorage != nil {
		mediaStore = socialstorage.NewMediaStore(objectStorage)
	} else {
		mediaStore = &noopMediaStore{}
	}

	broker := socialpubsub.NewSocialBroker()

	var aiClient completer
	if cfg.OpenRouterAPIKey != "" {
		aiClient = ai.NewOpenRouterClient(cfg.OpenRouterAPIKey, cfg.OpenRouterModel)
	}

	return &TenantHandlerFactory{
		db:            db,
		bus:           bus,
		cfg:           cfg,
		fetcher:       fetcher,
		mediaStore:    mediaStore,
		orgLookupInst: newOrgLookup(db),
		broker:        broker,
		emailProvider: emailProvider,
		aiClient:      aiClient,
	}
}

// Broker returns the shared SocialBroker (passed to the HTTP module for SSE).
func (f *TenantHandlerFactory) Broker() *socialpubsub.MemorySocialBroker { return f.broker }

// Fetcher returns the shared SocialFetcher singleton (safe for concurrent use).
// Exposed so cmd/server/modules.go can pass the same instance to the HTTP module.
func (f *TenantHandlerFactory) Fetcher() services.SocialFetcher { return f.fetcher }

// MediaStore returns the shared MediaStore singleton.
// Exposed so cmd/server/modules.go can pass the same instance to the HTTP module.
func (f *TenantHandlerFactory) MediaStore() services.MediaStore { return f.mediaStore }

// NewHandler constructs a fully-wired run_check.Handler for the given tenant.
// Called once per scheduler tick for each tenant with due profiles (per design §3).
func (f *TenantHandlerFactory) NewHandler(tenant string) *runcheck.Handler {
	profileRepo := socialpostgres.NewProfilePostgresRepository(f.db, tenant)
	snapshotRepo := socialpostgres.NewSnapshotPostgresRepository(f.db, tenant)
	changeRepo := socialpostgres.NewChangePostgresRepository(f.db, tenant)
	quotaRepo := socialpostgres.NewCheckQuotaPostgresRepository(f.db, tenant)
	alertCreator := NewAlertCreator(f.bus, f.orgLookupInst, tenant)

	// Look up the org's daily check limit for quota enforcement.
	// This is best-effort: unknown tenants / no active plan → 0 (unlimited).
	checksPerDay, _ := GetChecksPerDayByTenant(context.Background(), f.db, tenant)

	h := runcheck.NewHandler(
		profileRepo,
		snapshotRepo,
		changeRepo,
		quotaRepo,
		f.fetcher,
		f.mediaStore,
		alertCreator,
		f.cfg.SocialPostsPerCheck,
		checksPerDay,
	)

	h.SetAlertRepo(socialpostgres.NewAlertPostgresRepository(f.db, tenant))
	h.SetSSEPublisher(sseBrokerAdapter{b: f.broker})
	if f.emailProvider != nil {
		h.SetEmailer(NewEmailer(f.db, f.emailProvider, tenant, f.cfg.AppDomain))
	}
	if f.aiClient != nil {
		h.SetInsightGenerator(NewInsightGenerator(f.aiClient, f.db, tenant))
	}
	return h
}

// TenantRepoFactory satisfies scheduler.TenantRepoFactory for the scheduled
// run_check path (repo access without full handler wiring).
// NOTE: The scheduler uses CheckHandlerFactory directly in the current design;
// this type is kept for future use (e.g. quota status in the worker).
type TenantRepoFactory struct{ db *sql.DB }

// NewTenantRepoFactory builds a repo factory for direct repository access.
func NewTenantRepoFactory(db *sql.DB) *TenantRepoFactory {
	return &TenantRepoFactory{db: db}
}

// ProfileRepo returns a tenant-scoped ProfileRepository.
func (f *TenantRepoFactory) ProfileRepo(tenant string) repositories.ProfileRepository {
	return socialpostgres.NewProfilePostgresRepository(f.db, tenant)
}

// SnapshotRepo returns a tenant-scoped SnapshotRepository.
func (f *TenantRepoFactory) SnapshotRepo(tenant string) repositories.SnapshotRepository {
	return socialpostgres.NewSnapshotPostgresRepository(f.db, tenant)
}

// ChangeRepo returns a tenant-scoped ChangeRepository.
func (f *TenantRepoFactory) ChangeRepo(tenant string) repositories.ChangeRepository {
	return socialpostgres.NewChangePostgresRepository(f.db, tenant)
}

// CheckQuota returns a tenant-scoped CheckQuota implementation.
func (f *TenantRepoFactory) CheckQuota(tenant string) services.CheckQuota {
	return socialpostgres.NewCheckQuotaPostgresRepository(f.db, tenant)
}

// ── noopMediaStore ────────────────────────────────────────────────────────────

// noopMediaStore is a fallback MediaStore used when object storage is
// unavailable. Store returns the original source URL unchanged so the
// snapshot is still captured (with CDN URLs that may expire).
type noopMediaStore struct{}

func (n *noopMediaStore) Store(_ context.Context, sourceURL, _ string) (string, error) {
	return sourceURL, nil // best-effort: return CDN URL as fallback
}
