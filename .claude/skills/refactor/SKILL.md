# Skill: Refactor

## Purpose

Safely restructure code while preserving behavior. Applies project-specific patterns (hexagonal architecture, vertical slicing, multi-tenancy) to ensure refactored code stays consistent with the codebase.

## When to Apply

- When code violates the dependency rule (domain → application → infrastructure)
- When a use case handler grows too large and needs splitting
- When extracting shared logic into `shared/` or domain services
- When consolidating duplicate code across modules
- When the user explicitly requests refactoring

## Inputs

- Target file(s) or module to refactor
- Specific concern (e.g., "extract service", "split handler", "fix layer violation")

## Process

1. **Read and understand**: Read the target code and its callers/dependents
2. **Identify violations**:
   - Cross-layer imports (e.g., domain importing infrastructure packages)
   - Cross-module imports (modules should never import each other)
   - Business logic in `shared/` (should be in domain/services)
   - Fat handlers doing too much (should delegate to domain services)
3. **Plan the refactor**:
   - Identify what moves where
   - Ensure no import cycles
   - Preserve all public interfaces (or update all callers)
4. **Execute**:
   - Move/rename files
   - Update imports
   - Adjust package names (directory `create_check` → `package createcheck`)
5. **Verify**:
   - `go build ./...` compiles
   - `go vet ./...` passes
   - Existing tests pass: `go test ./modules/{name}/...`

## Output

- Modified files with clean boundaries
- Summary of what moved and why
- Confirmation that build and tests pass

## Constraints

- Never break the dependency rule: domain ← application ← infrastructure
- Never add cross-module imports; use gRPC or events instead
- Keep `shared/` as technical utilities only
- Preserve existing test coverage
