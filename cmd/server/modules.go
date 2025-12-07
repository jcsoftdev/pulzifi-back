package main

import (
	"database/sql"

	alert "github.com/jcsoftdev/pulzifi-back/modules/alert/infrastructure/http"
	auth "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/http"
	insight "github.com/jcsoftdev/pulzifi-back/modules/insight/infrastructure/http"
	integration "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/http"
	monitoring "github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/http"
	organization "github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/http"
	page "github.com/jcsoftdev/pulzifi-back/modules/page/infrastructure/http"
	report "github.com/jcsoftdev/pulzifi-back/modules/report/infrastructure/http"
	usage "github.com/jcsoftdev/pulzifi-back/modules/usage/infrastructure/http"
	workspace "github.com/jcsoftdev/pulzifi-back/modules/workspace/infrastructure/http"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

// registerAllModulesInternal registers all 10 modules for the monolith
func registerAllModulesInternal(registry *router.Registry, db *sql.DB) {
	moduleInstances := []struct {
		name   string
		module router.ModuleRegisterer
	}{
		{"Auth", auth.NewModule()},
		{"Organization", organization.NewModule()},
		{"Workspace", workspace.NewModule()},
		{"Page", page.NewModule()},
		{"Alert", alert.NewModuleWithDB(db)},
		{"Monitoring", monitoring.NewModule()},
		{"Integration", integration.NewModule()},
		{"Insight", insight.NewModule()},
		{"Report", report.NewModule()},
		{"Usage", usage.NewModule()},
	}

	logger.Info("Registering all modules", zap.Int("count", len(moduleInstances)))

	for _, m := range moduleInstances {
		registry.Register(m.module)
		logger.Info("Registered module", zap.String("module", m.name))
	}

	logger.Info("All modules registered successfully", zap.Int("total", registry.Count()))
}
