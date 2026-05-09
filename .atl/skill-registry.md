# Skill Registry — pulzifi-back

Generated: 2026-05-07. Auto-resolved by SDD orchestrator. Sub-agents receive matching compact rules in their prompt; do not load this file directly.

## Project Conventions

- **CLAUDE.md** (project root) — primary instructions: Go + Next.js commands, hexagonal architecture, multi-tenant by subdomain, validation script.
- **frontend/CLAUDE.md** — frontend-specific conventions if present.
- **frontend/biome.json** — lint/format rules (2-space indent, lineWidth 100, useConst warn).
- **modules/<name>/CLAUDE.md** — per-module hexagonal architecture rules (auto-scaffolded by `./tools/scripts/new-module.sh`).
- **./tools/scripts/check-architecture.sh** — enforces hexagonal boundaries.
- **./tools/scripts/validate-build.sh** — pre-commit gate (build + vet + tests + types + arch).

## User Skills

Listed by trigger context. Sub-agents inject only matching compact rules.

### Code review / quality
- `code-review` — review patterns, comment style
- `refactor` — refactoring approaches
- `release` — release workflow

### React / Next.js (50 PatternsDev skills under .claude/skills/)
- `react-2026`, `react-composition-2026`, `react-data-fetching`, `react-render-optimization`
- `react-server-components`, `react-selective-hydration`, `progressive-hydration`
- `client-side-rendering`, `incremental-static-rendering`, `route-based`, `islands-architecture`
- `hooks-pattern`, `hoc-pattern`, `provider-pattern`, `presentational-container-pattern`, `render-props-pattern`
- `command-pattern`, `compound-pattern`, `factory-pattern`, `flyweight-pattern`, `mediator-pattern`, `mixin-pattern`, `module-pattern`, `observer-pattern`, `prototype-pattern`, `proxy-pattern`, `singleton-pattern`

### Performance
- `bundle-splitting`, `dynamic-import`, `import-on-interaction`, `import-on-visibility`
- `prefetch`, `preload`, `prpl`, `loading-sequence`, `compression`
- `js-performance-patterns`, `ai-ui-patterns`

### User-level (17 at ~/.claude/skills/)
- See `~/.claude/skills/` — includes `go-testing`, `skill-creator`, plus SDD skills and superpowers.

## Compact Rules (auto-resolved per delegation)

For each sub-agent launch, the SDD orchestrator matches skills by:
- **File extensions/paths the sub-agent will touch** (e.g. `*.tsx` → react-* skills; `*.go` → go-testing; `frontend/**` → biome rules + Next.js patterns)
- **Task context** (review → code-review; release → release; refactor → refactor)

Sub-agents receive matching compact rules pre-digested as `## Project Standards (auto-resolved)` injected at the top of their task prompt. Do NOT read SKILL.md files in sub-agents.

## Re-resolution

If sub-agent reports `skill_resolution: fallback-*`, the orchestrator's cache was lost. Re-read this file and re-inject before next delegation.
