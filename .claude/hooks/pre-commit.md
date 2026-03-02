# Hook: Pre-Commit Checks

## Purpose

Guardrails and automation checks to run before committing code changes.

## Checks

### Go Backend
1. **Build verification**: `go build ./...` — ensures code compiles
2. **Vet**: `go vet ./...` — catches common mistakes
3. **Tests**: `go test ./...` — all tests pass
4. **Architecture rule**: No cross-module imports (modules/ directories must not import each other)
5. **Dependency rule**: Domain layer must not import infrastructure packages

### Frontend
1. **Type check**: `cd frontend && bun run type-check`
2. **Lint**: `cd frontend && bun run lint:fix`

### General
1. **No secrets**: Check for hardcoded API keys, passwords, or tokens in staged files
2. **Migration ordering**: New migration files use the next sequential number
3. **Package naming**: Directory names with underscores use concatenated package names (e.g., `create_check` → `package createcheck`)

## When to Run

- Before every `git commit`
- Can be skipped with `--no-verify` for WIP commits (not recommended)
