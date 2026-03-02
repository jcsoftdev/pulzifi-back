# Skill: Code Review

## Purpose

Perform a structured code review on staged or committed changes, checking for adherence to project conventions, potential bugs, security issues, and architectural violations.

## When to Apply

- Before creating a pull request
- When the user asks to review changes, a diff, or a specific file
- After implementing a feature to self-check quality

## Inputs

- Git diff (staged, unstaged, or between commits/branches)
- Optionally, specific file paths to focus on

## Process

1. **Gather changes**: Run `git diff --staged` or `git diff <base>..HEAD` to collect the changeset
2. **Check architecture rules**:
   - Domain layer has no infrastructure imports
   - No cross-module imports between `modules/`
   - Repository interfaces defined in `domain/repositories/`, implementations in `infrastructure/persistence/`
   - Use cases isolated in `application/{use_case}/` directories
3. **Check coding conventions**:
   - Package naming: directory `create_check` → `package createcheck`
   - Tenant-aware repos use `middleware.GetSetSearchPathSQL(tenant)` before queries
   - `context.WithTimeout` used for external calls in goroutines
   - No business logic in `shared/` packages
4. **Check for bugs**:
   - SQL injection via string concatenation (use parameterized queries)
   - Missing error handling (Go errors must be checked)
   - Resource leaks (unclosed DB rows, HTTP response bodies)
   - Race conditions in concurrent code (shared state without mutex)
5. **Check security**:
   - No secrets or credentials in code
   - Input validation at HTTP handler level
   - Proper authorization checks (middleware applied to routes)
6. **Check tests**:
   - New use cases have corresponding `handler_test.go`
   - Tests use in-memory repository implementations, not real DB
7. **Frontend checks** (if applicable):
   - Components follow Atomic Design (atoms, molecules, organisms)
   - API calls go through `packages/services/` layer
   - Tenant extraction uses `shared-http` factory

## Output

A structured review with sections:
- **Summary**: One-line description of the changes
- **Issues**: Bugs, security problems, or convention violations (with file:line references)
- **Suggestions**: Improvements that are optional but recommended
- **Verdict**: Approve / Request Changes
