package router

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// ModuleRegisterer defines the interface that each module must implement
// to register its HTTP routes and gRPC services
type ModuleRegisterer interface {
	// RegisterHTTPRoutes registers all HTTP routes for the module
	// The router is a chi.Router (e.g., /api/v1)
	RegisterHTTPRoutes(router chi.Router)

	// ModuleName returns the name of the module for logging
	ModuleName() string
}

// Registry holds all registered modules
type Registry struct {
	modules []ModuleRegisterer
	logger  *zap.Logger
}

// NewRegistry creates a new module registry
func NewRegistry(logger *zap.Logger) *Registry {
	return &Registry{
		modules: make([]ModuleRegisterer, 0),
		logger:  logger,
	}
}

// Register adds a module to the registry
func (r *Registry) Register(module ModuleRegisterer) {
	r.modules = append(r.modules, module)
	r.logger.Info("Module registered", zap.String("module", module.ModuleName()))
}

// RegisterAll registers all modules with the given router
func (r *Registry) RegisterAll(router chi.Router) {
	for _, module := range r.modules {
		r.logger.Info("Registering HTTP routes", zap.String("module", module.ModuleName()))
		module.RegisterHTTPRoutes(router)
	}
	r.logger.Info("All modules registered successfully", zap.Int("count", len(r.modules)))
}

// GetModules returns all registered modules
func (r *Registry) GetModules() []ModuleRegisterer {
	return r.modules
}

// Count returns the number of registered modules
func (r *Registry) Count() int {
	return len(r.modules)
}
