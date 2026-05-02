package deliveryworker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
)

// RepoFactory returns tenant-scoped repository instances.
// Matches dispatch_event.RepoFactory and cmd/wiring/integration.TenantRepoFactory.
type RepoFactory interface {
	DestinationRepoForTenant(tenant string) repositories.DestinationRepository
	DeliveryRepoForTenant(tenant string) repositories.DeliveryRepository
}

// Config holds tuning parameters for the delivery worker.
type Config struct {
	PollInterval   time.Duration
	PoolSize       int
	MaxAttempts    int
	ClaimBatchSize int // default 50
	// Synchronous disables goroutine dispatch — for deterministic testing only.
	Synchronous bool
}

// Worker polls the database for pending deliveries and sends them via the provider registry.
type Worker struct {
	db       *sql.DB
	factory  RepoFactory
	intRepo  repositories.IntegrationRepository
	registry services.ProviderRegistry
	builder  *services.PayloadBuilder
	cfg      Config
}

// New constructs a Worker with sane defaults for any zero-value Config fields.
func New(
	db *sql.DB,
	factory RepoFactory,
	intRepo repositories.IntegrationRepository,
	registry services.ProviderRegistry,
	builder *services.PayloadBuilder,
	cfg Config,
) *Worker {
	if cfg.ClaimBatchSize == 0 {
		cfg.ClaimBatchSize = 50
	}
	if cfg.PoolSize == 0 {
		cfg.PoolSize = 10
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 5 * time.Second
	}
	if cfg.MaxAttempts == 0 {
		cfg.MaxAttempts = 5
	}
	return &Worker{
		db:       db,
		factory:  factory,
		intRepo:  intRepo,
		registry: registry,
		builder:  builder,
		cfg:      cfg,
	}
}

// Run starts the polling loop. Blocks until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.cfg.PollInterval)
	defer ticker.Stop()
	sem := make(chan struct{}, w.cfg.PoolSize)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			w.tick(ctx, sem)
		}
	}
}

// TickForTest exposes tick for integration tests — do not call in production code.
func (w *Worker) TickForTest(ctx context.Context, sem chan struct{}) {
	w.tick(ctx, sem)
}

// tick is one poll cycle: list tenants, claim pending deliveries, process each.
func (w *Worker) tick(ctx context.Context, sem chan struct{}) {
	tenants, err := w.listTenants(ctx)
	if err != nil {
		logger.ErrorWithContext(ctx, "delivery worker: list tenants", zap.Error(err))
		return
	}

	for _, tenant := range tenants {
		delRepo := w.factory.DeliveryRepoForTenant(tenant)
		destRepo := w.factory.DestinationRepoForTenant(tenant)

		deliveries, err := delRepo.ClaimPending(ctx, w.cfg.ClaimBatchSize, time.Now())
		if err != nil {
			logger.ErrorWithContext(ctx, "delivery worker: claim pending",
				zap.Error(err), zap.String("tenant", tenant))
			continue
		}

		for _, d := range deliveries {
			if w.cfg.Synchronous {
				// Run inline — used in tests for deterministic assertions.
				w.process(ctx, tenant, d, delRepo, destRepo)
				continue
			}
			sem <- struct{}{} // block until a pool slot is free
			go func(d *entities.Delivery) {
				defer func() { <-sem }()
				w.process(ctx, tenant, d, delRepo, destRepo)
			}(d)
		}
	}
}

// process handles a single delivery row end-to-end.
func (w *Worker) process(
	parentCtx context.Context,
	tenant string,
	d *entities.Delivery,
	delRepo repositories.DeliveryRepository,
	destRepo repositories.DestinationRepository,
) {
	ctx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
	defer cancel()

	// Load destination.
	dest, err := destRepo.GetByID(ctx, d.DestinationID)
	if err != nil || dest == nil {
		markFailedOrDead(ctx, delRepo, d, w.cfg.MaxAttempts,
			fmt.Errorf("destination missing: %v", err), nil)
		return
	}

	// Load integration (optional — destinations without an integration_id are valid,
	// e.g. email destinations that rely only on Target config).
	var integ *entities.Integration
	if dest.IntegrationID != nil {
		integ, err = w.intRepo.GetByID(ctx, *dest.IntegrationID)
		if err != nil {
			markFailedOrDead(ctx, delRepo, d, w.cfg.MaxAttempts,
				fmt.Errorf("integration load: %w", err), nil)
			return
		}
		if integ == nil || integ.DeletedAt != nil {
			// Integration has been disconnected; mark dead, no retry.
			_ = delRepo.MarkDead(ctx, d.ID, "integration disconnected",
				appendAttempt(d.AttemptHistory, nil, "integration disconnected", w.cfg.MaxAttempts))
			return
		}

		// Token refresh check — only when a refresh token is available and
		// the access token is about to expire.
		if integ.TokenExpiresAt != nil && integ.RefreshToken != "" &&
			time.Until(*integ.TokenExpiresAt) < time.Minute {

			client, ok := w.registry.Get(integ.ServiceType)
			if ok {
				newTok, refreshErr := client.RefreshAccessToken(ctx, integ.RefreshToken)
				if refreshErr != nil || newTok == nil {
					integ.Status = entities.IntegrationExpired
					_ = w.intRepo.Update(ctx, integ)
					_ = delRepo.MarkDead(ctx, d.ID, "token refresh failed",
						appendAttempt(d.AttemptHistory, nil, "token refresh failed", w.cfg.MaxAttempts))
					return
				}
				integ.AccessToken = newTok.AccessToken
				if newTok.RefreshToken != "" {
					integ.RefreshToken = newTok.RefreshToken
				}
				integ.TokenExpiresAt = newTok.ExpiresAt
				_ = w.intRepo.Update(ctx, integ)
			}
		}
	}

	// Build notification payload from the stored event data.
	rawData, _ := json.Marshal(d.EventPayload) // safe even when EventPayload is nil
	payload, err := w.builder.Build(eventbus.DomainEvent{
		Type: d.EventType,
		Data: rawData,
	})
	if err != nil {
		markFailedOrDead(ctx, delRepo, d, w.cfg.MaxAttempts, err, nil)
		return
	}

	// Resolve provider client.
	client, ok := w.registry.Get(dest.ServiceType)
	if !ok {
		noProviderMsg := "no provider for service: " + dest.ServiceType
		_ = delRepo.MarkDead(ctx, d.ID, noProviderMsg,
			appendAttempt(d.AttemptHistory, nil, noProviderMsg, w.cfg.MaxAttempts))
		return
	}

	// Send.
	result, sendErr := client.Send(ctx, integ, dest, payload)
	if sendErr != nil {
		// 401 → integration token invalid; mark integration expired and delivery dead.
		if result != nil && result.Code == 401 {
			if integ != nil {
				integ.Status = entities.IntegrationExpired
				_ = w.intRepo.Update(ctx, integ)
			}
			_ = delRepo.MarkDead(ctx, d.ID, sendErr.Error(),
				appendAttempt(d.AttemptHistory, &result.Code, sendErr.Error(), w.cfg.MaxAttempts))
			return
		}
		// Transient error → retry with backoff or mark dead if max attempts exceeded.
		var code *int
		if result != nil {
			code = &result.Code
		}
		markFailedOrDead(ctx, delRepo, d, w.cfg.MaxAttempts, sendErr, code)
		return
	}

	code := 200
	body := ""
	if result != nil {
		code = result.Code
		body = result.BodySnip
	}
	_ = delRepo.MarkDelivered(ctx, d.ID, code, body)
}

// listTenants returns all active tenant schema names from public.organizations.
func (w *Worker) listTenants(ctx context.Context) ([]string, error) {
	rows, err := w.db.QueryContext(ctx,
		`SELECT schema_name FROM public.organizations WHERE schema_name IS NOT NULL AND deleted_at IS NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// appendAttempt returns a new history slice with the new attempt appended,
// trimmed to the most recent maxLen entries (oldest dropped).
// Phase 2: caps unbounded growth on misconfigured destinations.
func appendAttempt(history []entities.Attempt, code *int, errMsg string, maxLen int) []entities.Attempt {
	next := append(history, entities.Attempt{At: time.Now().UTC(), Code: code, Error: errMsg})
	if maxLen > 0 && len(next) > maxLen {
		next = next[len(next)-maxLen:]
	}
	return next
}

// markFailedOrDead reschedules a delivery for retry or marks it dead when max attempts is reached.
func markFailedOrDead(
	ctx context.Context,
	delRepo repositories.DeliveryRepository,
	d *entities.Delivery,
	maxAttempts int,
	err error,
	code *int,
) {
	history := appendAttempt(d.AttemptHistory, code, err.Error(), maxAttempts)
	if d.Attempts >= maxAttempts {
		_ = delRepo.MarkDead(ctx, d.ID, err.Error(), history)
		return
	}
	_ = delRepo.MarkFailed(ctx, d.ID, code, "", err.Error(), time.Now().Add(backoff(d.Attempts)), history)
}

// backoff returns the wait duration before the next delivery attempt.
// Schedule: 5s, 30s, 2m, 10m, 1h — indexed by attempts-1, capped at last entry.
func backoff(attempts int) time.Duration {
	s := []time.Duration{
		5 * time.Second,
		30 * time.Second,
		2 * time.Minute,
		10 * time.Minute,
		time.Hour,
	}
	if attempts <= 0 {
		return s[0]
	}
	if attempts > len(s) {
		return s[len(s)-1]
	}
	return s[attempts-1]
}
