package main

import (
	"context"
	"database/sql"

	invitetoplatform "github.com/jcsoftdev/pulzifi-back/modules/admin/application/invite_to_platform"
	listinvitations "github.com/jcsoftdev/pulzifi-back/modules/admin/application/list_invitations"
	resendinvitation "github.com/jcsoftdev/pulzifi-back/modules/admin/application/resend_invitation"
	revokeinvitation "github.com/jcsoftdev/pulzifi-back/modules/admin/application/revoke_invitation"
	admin "github.com/jcsoftdev/pulzifi-back/modules/admin/infrastructure/http"
	adminpersistence "github.com/jcsoftdev/pulzifi-back/modules/admin/infrastructure/persistence"
	alert "github.com/jcsoftdev/pulzifi-back/modules/alert/infrastructure/http"
	acceptinvitation "github.com/jcsoftdev/pulzifi-back/modules/auth/application/accept_invitation"
	getinvitation "github.com/jcsoftdev/pulzifi-back/modules/auth/application/get_invitation"
	auth "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/http"
	authpersistence "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/persistence"
	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/services"
	dashboard "github.com/jcsoftdev/pulzifi-back/modules/dashboard/infrastructure/http"
	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	email "github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/http"
	emailproviders "github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/providers"
	insight "github.com/jcsoftdev/pulzifi-back/modules/insight/infrastructure/http"
	integration "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/http"
	monitoring "github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/http"
	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
	organization "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/http"
	orgmessaging "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/messaging"
	orgpersistence "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/persistence"
	page "github.com/jcsoftdev/pulzifi-back/modules/page/infrastructure/http"
	report "github.com/jcsoftdev/pulzifi-back/modules/report/infrastructure/http"
	snapshotextractor "github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/extractor"
	team "github.com/jcsoftdev/pulzifi-back/modules/team/infrastructure/http"
	usage "github.com/jcsoftdev/pulzifi-back/modules/usage/infrastructure/http"
	workspace "github.com/jcsoftdev/pulzifi-back/modules/workspace/infrastructure/http"
	"github.com/jcsoftdev/pulzifi-back/shared/bff"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/noncestore"
	"github.com/jcsoftdev/pulzifi-back/shared/pubsub"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

// createEmailProvider creates the Resend email provider.
func createEmailProvider(cfg *config.Config) emailservices.EmailProvider {
	return emailproviders.NewResendProvider(cfg.ResendAPIKey, cfg.EmailFromAddress, cfg.EmailFromName)
}

// invitationDailyCapPerInviter is the per-inviter rolling 24h cap on platform
// invitations. Hard-coded for now; can be promoted to config when needed.
const invitationDailyCapPerInviter = 50

// invitationDailyCapGlobal is the platform-wide rolling 24h cap on platform
// invitations. Hard-coded for now; can be promoted to config when needed.
const invitationDailyCapGlobal = 500

// invitationInviterDisplayName is the "From" name surfaced inside the
// invitation email template. Sourcing this from the authenticated SUPER_ADMIN
// user is a future improvement; for now we use a generic platform label.
const invitationInviterDisplayName = "Pulzifi Admin"

func registerAllModulesInternal(registry *router.Registry, db *sql.DB, eventBus *eventbus.EventBus, enableWorkers bool) *bff.Handler {
	cfg := config.Load()

	userRepo := authpersistence.NewUserPostgresRepository(db)
	roleRepo := authpersistence.NewRolePostgresRepository(db)
	permRepo := authpersistence.NewPermissionPostgresRepository(db)
	refreshTokenRepo := authpersistence.NewRefreshTokenPostgresRepository(db)
	orgRepo := orgpersistence.NewOrganizationPostgresRepository(db)

	regReqRepo := adminpersistence.NewRegistrationRequestPostgresRepository(db)
	invRepo := adminpersistence.NewInvitationPostgres(db)
	orgService := orgservices.NewOrganizationService()

	authService := authservices.NewBcryptAuthService(userRepo, permRepo)
	jwtService := authservices.NewJWTService(cfg.JWTSecret, cfg.JWTExpiration, cfg.JWTRefreshExpiration, roleRepo, permRepo, userRepo, refreshTokenRepo)
	cookieSecure := cfg.Environment == "production"

	// Create email provider (shared across modules)
	emailProvider := createEmailProvider(cfg)

	// Invitation use cases — admin (create/list/revoke/resend) + auth (get/accept).
	// The use case packages declare a local Emailer interface satisfied
	// structurally by emailservices.EmailProvider (same Send signature).
	inviteHandler := invitetoplatform.New(
		invRepo,
		emailProvider,
		invitationInviterDisplayName,
		cfg.FrontendURL,
		invitationDailyCapPerInviter,
		invitationDailyCapGlobal,
	)
	listInvitationsHandler := listinvitations.New(invRepo)
	revokeHandler := revokeinvitation.New(invRepo)
	resendHandler := resendinvitation.New(
		invRepo,
		emailProvider,
		invitationInviterDisplayName,
		cfg.FrontendURL,
	)

	getInvitationHandler := getinvitation.New(invRepo, userRepo)
	provisionFunc := func(schema string) error {
		return database.ProvisionTenantSchema(db, schema)
	}
	acceptHandler := acceptinvitation.New(invRepo, authService, provisionFunc)

	// Create auth module and set global middleware.
	// The BFF handler is constructed below from this module's handlers and
	// then injected back via SetBFFHandler — this avoids a constructor cycle.
	authModule := auth.NewModule(auth.ModuleDeps{
		UserRepo:                userRepo,
		RefreshTokenRepo:        refreshTokenRepo,
		RoleRepo:                roleRepo,
		PermRepo:                permRepo,
		RegReqRepo:              regReqRepo,
		OrgRepo:                 orgRepo,
		OrgService:              orgService,
		AuthService:             authService,
		TokenService:            jwtService,
		CookieDomain:            cfg.CookieDomain,
		CookieSecure:            cookieSecure,
		FrontendURL:             cfg.FrontendURL,
		EmailProvider:           emailProvider,
		EventBus:                eventBus,
		DB:                      db,
		GetInvitationHandler:    getInvitationHandler,
		AcceptInvitationHandler: acceptHandler,
		// BFFHandler is injected after construction (see SetBFFHandler below).
	})
	authMod := authModule.(*auth.Module)
	authMiddleware := authMod.AuthMiddleware()

	// Set global middleware for all modules
	middleware.SetAuthMiddleware(authMiddleware)
	middleware.SetOrganizationMiddleware(middleware.NewOrganizationMiddleware(db))

	moduleInstances := []struct {
		name   string
		module router.ModuleRegisterer
	}{
		{"Auth", authModule},
		{"Admin", admin.NewModule(admin.ModuleDeps{
			DB:                     db,
			RegReqRepo:             regReqRepo,
			UserRepo:               userRepo,
			OrgRepo:                orgRepo,
			OrgService:             orgService,
			AuthMiddleware:         authMiddleware,
			EmailProvider:          emailProvider,
			FrontendURL:            cfg.FrontendURL,
			InviteHandler:          inviteHandler,
			ListInvitationsHandler: listInvitationsHandler,
			RevokeHandler:          revokeHandler,
			ResendHandler:          resendHandler,
		})},
		{"Email", email.NewModule(emailProvider)},
		{"Organization", organization.NewModule(orgRepo)},
		{"Workspace", workspace.NewModuleWithDB(db)},
		{"Page", page.NewModuleWithExtractor(db, snapshotextractor.NewHTTPClient(cfg.ExtractorURL))},
		{"Alert", alert.NewModuleWithDB(db)},
		{"Monitoring", monitoring.NewModuleWithDB(db, eventBus, emailProvider, cfg.FrontendURL)},
		{"Integration", integration.NewModuleWithDB(db)},
		{"Insight", insight.NewModuleWithDB(db, pubsub.NewInsightBroker())},
		{"Report", report.NewModuleWithDB(db)},
		{"Usage", usage.NewModuleWithDB(db)},
		{"Dashboard", dashboard.NewModuleWithDB(db)},
		{"Team", team.NewModuleWithDB(db, emailProvider, cfg.FrontendURL)},
	}

	logger.Info("Registering all modules", zap.Int("count", len(moduleInstances)))

	for _, m := range moduleInstances {
		registry.Register(m.module)
		logger.Info("Registered module", zap.String("module", m.name))

		// Special handling for Monitoring module to start background processes if enabled
		if m.name == "Monitoring" && enableWorkers {
			if monModule, ok := m.module.(*monitoring.Module); ok {
				monModule.StartBackgroundProcesses()
				logger.Info("Started background processes for Monitoring module")
			}
		}
	}

	// Start organization event subscriber in background
	orgSubscriber := orgmessaging.NewSubscriber(eventBus, db)
	go func() {
		orgSubscriber.ListenToEvents(context.Background())
	}()
	logger.Info("Started organization event subscriber")

	logger.Info("All modules registered successfully", zap.Int("total", registry.Count()))

	// Create BFF auth handler using extracted auth module handlers
	bffHandler := bff.NewHandler(bff.HandlerDeps{
		LoginHandler:   authMod.LoginHandler(),
		LogoutHandler:  authMod.LogoutHandler(),
		RefreshHandler: authMod.RefreshHandler(),
		TokenService:   authMod.TokenService(),
		NonceStore:     noncestore.New(),
		CookieDomain:   authMod.CookieDomain(),
		CookieSecure:   authMod.CookieSecure(),
		Logger:         logger.Logger,
	})

	// Inject the BFF handler back into the auth module so the
	// /api/v1/auth/invitations/{token}/accept handler can issue a session.
	authMod.SetBFFHandler(bffHandler)

	return bffHandler
}
