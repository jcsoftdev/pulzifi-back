# Skill: Release

## Purpose

Prepare and execute a release by validating the codebase, building artifacts, and creating a tagged release with changelog.

## When to Apply

- When the user asks to prepare a release or cut a new version
- Before deploying to production

## Inputs

- Version number (semver: major.minor.patch) or version bump type (patch/minor/major)
- Optional: release notes or summary of changes

## Process

1. **Pre-flight checks**:
   - Working tree is clean: `git status` shows no uncommitted changes
   - All tests pass: `go test ./...`
   - Go code compiles: `go build ./...`
   - Go vet passes: `go vet ./...`
   - Frontend builds: `cd frontend && bun run build`
   - Frontend type-check: `cd frontend && bun run type-check`
2. **Build artifacts**:
   - `make build` (produces `./bin/api`)
   - Docker images if applicable:
     - `docker build -f Dockerfile.monolith.all-in-one -t pulzifi:<version> .`
3. **Generate changelog**:
   - Collect commits since last tag: `git log <last-tag>..HEAD --oneline`
   - Group by type (feat, fix, refactor, docs)
4. **Create release**:
   - Tag: `git tag -a v<version> -m "Release v<version>"`
   - Push tag: `git push origin v<version>`
5. **Database migrations**:
   - List any new migration files since last release
   - Flag if migrations require manual intervention (data migrations, destructive changes)

## Output

- Git tag created and pushed
- Changelog summary
- Build artifacts confirmed
- Migration notes (if any new migrations)

## Constraints

- Never release with failing tests
- Never skip the build verification step
- Always tag from the main branch
- Include migration warnings prominently if schema changes exist
