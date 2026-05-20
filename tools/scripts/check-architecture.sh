#!/usr/bin/env bash
# Verify Pulzifi architecture boundaries.
#
# This is intentionally import-based: it catches compile-time coupling that makes
# modules hard to split, test, or scale independently.
#
# Defaults:
#   - production Go files only
#   - fail on violations
#
# Options:
#   --include-tests   also scan *_test.go files
#   --warn-only       print violations but exit 0
#
# Env alternatives:
#   ARCH_INCLUDE_TESTS=1
#   ARCH_WARN_ONLY=1

set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
INCLUDE_TESTS="${ARCH_INCLUDE_TESTS:-0}"
WARN_ONLY="${ARCH_WARN_ONLY:-0}"

for arg in "$@"; do
  case "$arg" in
    --include-tests) INCLUDE_TESTS=1 ;;
    --warn-only) WARN_ONLY=1 ;;
    -h|--help)
      sed -n '1,24p' "$0"
      exit 0
      ;;
    *)
      echo "Unknown argument: $arg" >&2
      exit 2
      ;;
  esac
done

python3 - "$PROJECT_DIR" "$INCLUDE_TESTS" "$WARN_ONLY" <<'PY'
from __future__ import annotations

import re
import subprocess
import sys
from collections import defaultdict
from pathlib import Path

project = Path(sys.argv[1]).resolve()
include_tests = sys.argv[2] == "1"
warn_only = sys.argv[3] == "1"

mod_line = (project / "go.mod").read_text(encoding="utf-8").splitlines()[0]
module_path = mod_line.split()[1]
modules_import_root = f"{module_path}/modules/"

# Modules that are intentionally not standard hexagonal Go modules.
SKIP_MODULES = {"infra", "api-docs"}

# Wiring/composition roots are allowed to couple modules because they assemble
# adapters at process startup instead of baking dependencies into modules.
ALLOWED_COUPLING_PATH_PARTS = {
    ("cmd", "wiring"),
    ("cmd", "server"),
    ("cmd", "worker"),
    ("cmd", "migrate"),
}

DOMAIN_FORBIDDEN_IMPORT_PREFIXES = (
    "database/sql",
    "net/http",
    "github.com/go-chi/chi",
    "github.com/jackc/pgx",
    "github.com/lib/pq",
    "github.com/redis/",
    "github.com/aws/",
    "github.com/stripe/",
)

APPLICATION_FORBIDDEN_IMPORT_PREFIXES = DOMAIN_FORBIDDEN_IMPORT_PREFIXES + (
    "github.com/resend/",
)

IMPORT_BLOCK_RE = re.compile(r'import\s*\((.*?)\)', re.S)
SINGLE_IMPORT_RE = re.compile(r'import\s+(?:[\w.]+\s+)?"([^"]+)"')
BLOCK_IMPORT_RE = re.compile(r'(?:[\w.]+\s+)?"([^"]+)"')


def rel(path: Path) -> str:
    return str(path.relative_to(project))


def strip_comments(src: str) -> str:
    # Good enough for import declarations and avoids false positives from
    # commented imports. Go import paths cannot contain comment delimiters.
    src = re.sub(r'/\*.*?\*/', '', src, flags=re.S)
    src = re.sub(r'//.*', '', src)
    return src


def parse_imports(path: Path) -> list[str]:
    src = strip_comments(path.read_text(encoding="utf-8", errors="ignore"))
    imports: list[str] = []
    for block in IMPORT_BLOCK_RE.findall(src):
        imports.extend(BLOCK_IMPORT_RE.findall(block))
    imports.extend(SINGLE_IMPORT_RE.findall(src))
    return imports


def module_name_for_file(path: Path) -> str | None:
    try:
        parts = path.relative_to(project).parts
    except ValueError:
        return None
    if len(parts) >= 3 and parts[0] == "modules":
        return parts[1]
    return None


def imported_module(import_path: str) -> str | None:
    if not import_path.startswith(modules_import_root):
        return None
    rest = import_path[len(modules_import_root):]
    return rest.split("/", 1)[0] if rest else None


def is_layer_file(path: Path, layer: str) -> bool:
    parts = path.relative_to(project).parts
    return len(parts) >= 4 and parts[0] == "modules" and parts[2] == layer


def allowed_coupling_file(path: Path) -> bool:
    parts = path.relative_to(project).parts
    return any(parts[: len(prefix)] == prefix for prefix in ALLOWED_COUPLING_PATH_PARTS)


def has_prefix(import_path: str, prefixes: tuple[str, ...]) -> bool:
    return any(import_path == p.rstrip("/") or import_path.startswith(p) for p in prefixes)


violations: list[tuple[str, str, str]] = []
counts = defaultdict(int)

def tracked_go_files() -> list[Path]:
    try:
        result = subprocess.run(
            ["git", "-C", str(project), "ls-files", "*.go"],
            check=True,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
        )
        return [project / line for line in result.stdout.splitlines() if line]
    except (subprocess.CalledProcessError, FileNotFoundError):
        # Fallback for source archives or environments without git.
        roots = [project / "cmd", project / "modules", project / "shared"]
        return [p for root in roots if root.exists() for p in root.rglob("*.go")]


files = sorted(tracked_go_files())
if not include_tests:
    files = [p for p in files if not p.name.endswith("_test.go")]

for path in files:
    r = rel(path)
    if "/vendor/" in f"/{r}/":
        continue

    current_module = module_name_for_file(path)
    if current_module in SKIP_MODULES:
        continue

    imports = parse_imports(path)

    for imp in imports:
        target_module = imported_module(imp)

        # Rule 1: modules cannot import other modules. Use cmd/wiring or shared
        # interfaces/adapters for cross-module orchestration.
        if (
            current_module
            and target_module
            and target_module != current_module
            and not allowed_coupling_file(path)
        ):
            violations.append((
                "cross-module-import",
                r,
                f"module {current_module!r} imports module {target_module!r}: {imp}",
            ))
            counts["cross-module-import"] += 1

        # Rule 2: domain is pure business logic. It must not depend on app,
        # infra, transport, database, queues, payment SDKs, or other modules.
        if is_layer_file(path, "domain"):
            if f"/application" in imp or f"/infrastructure" in imp:
                violations.append(("domain-layer-direction", r, f"domain imports non-domain layer: {imp}"))
                counts["domain-layer-direction"] += 1
            elif target_module and target_module != current_module:
                violations.append(("domain-cross-module", r, f"domain imports another module: {imp}"))
                counts["domain-cross-module"] += 1
            elif has_prefix(imp, DOMAIN_FORBIDDEN_IMPORT_PREFIXES):
                violations.append(("domain-infrastructure-import", r, f"domain imports infrastructure dependency: {imp}"))
                counts["domain-infrastructure-import"] += 1

        # Rule 3: application orchestrates use cases via ports/interfaces. It
        # must not import infrastructure details.
        if is_layer_file(path, "application"):
            if f"/infrastructure" in imp:
                violations.append(("application-infrastructure-import", r, f"application imports infrastructure: {imp}"))
                counts["application-infrastructure-import"] += 1
            elif has_prefix(imp, APPLICATION_FORBIDDEN_IMPORT_PREFIXES):
                violations.append(("application-transport-or-db-import", r, f"application imports transport/db/sdk dependency: {imp}"))
                counts["application-transport-or-db-import"] += 1

        # Rule 4: shared is technical utility code only. It must not depend on
        # business modules.
        if r.startswith("shared/") and target_module:
            violations.append(("shared-imports-module", r, f"shared imports module {target_module!r}: {imp}"))
            counts["shared-imports-module"] += 1


print("Checking architecture rules...")
print(f"  project: {project}")
print(f"  module path: {module_path}")
print(f"  files scanned: {len(files)}" + (" (including tests)" if include_tests else ""))
print("")

if violations:
    print("FAILED: architecture violations found")
    print("")
    for rule, count in sorted(counts.items()):
        print(f"  {rule}: {count}")
    print("")
    for rule, path, message in violations:
        print(f"{path}: [{rule}] {message}")
    print("")
    print("Allowed pattern: put cross-module adapters in cmd/wiring/* or depend on shared interfaces, not another module's implementation.")
    if warn_only:
        print("WARN_ONLY enabled: exiting 0 despite violations.")
        raise SystemExit(0)
    raise SystemExit(1)

print("OK: architecture import boundaries pass.")
PY
