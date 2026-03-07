# Router Package (`shared/router/`)

Module registration system for the monolith server.

## Files

- `registry.go` — Module registerer interface and registry

## Exported API

### Interfaces
- `ModuleRegisterer` — Required for every module:
  - `RegisterHTTPRoutes(router chi.Router)` — Registers all HTTP routes
  - `ModuleName() string` — Returns module name for logging

### Types
- `Registry` — Module registry with logging

### Functions
- `NewRegistry(logger *zap.Logger) *Registry` — Creates empty registry

### Methods (`*Registry`)
- `Register(module ModuleRegisterer)` — Adds module, logs registration
- `RegisterAll(router chi.Router)` — Calls `RegisterHTTPRoutes` on all modules, logs total count
- `GetModules() []ModuleRegisterer` — Returns registered modules
- `Count() int` — Returns count

## Usage

Each module's `infrastructure/http/module.go` implements `ModuleRegisterer`. The server bootstrap in `cmd/server/modules.go` creates a registry and registers all 14 modules.

## Implementation Pattern

```go
// In each module's infrastructure/http/module.go:
type Module struct { /* dependencies */ }

func (m *Module) RegisterHTTPRoutes(router chi.Router) {
    router.Route("/resource", func(r chi.Router) {
        r.Get("/", m.listHandler)
        r.Post("/", m.createHandler)
    })
}

func (m *Module) ModuleName() string { return "resource" }
```
