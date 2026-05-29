package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
)

const migrationsBaseDir = "shared/database/migrations"

func defaultDBURL() string {
	_ = godotenv.Load()
	host := getenv("DB_HOST", "localhost")
	port := getenv("DB_PORT", "5434")
	user := getenv("DB_USER", "pulzifi_user")
	password := getenv("DB_PASSWORD", "pulzifi_password")
	name := getenv("DB_NAME", "pulzifi")
	sslmode := getenv("DB_SSLMODE", "disable")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, name, sslmode)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

var validSchemaName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run parses flags, connects to the database, and runs the requested
// migrations. It returns an error instead of calling log.Fatal directly so
// that deferred cleanup (db.Close, rows.Close) runs before the process exits.
func run() error {
	var (
		dbURL    = flag.String("db", defaultDBURL(), "Database URL")
		cmd      = flag.String("cmd", "up", "Command to run: up, down, version")
		steps    = flag.Int("steps", 0, "Number of steps to migrate (optional)")
		scope    = flag.String("scope", "all", "Scope of migration: all, public, tenant")
		tenantID = flag.String("tenant", "", "Specific tenant schema to migrate (optional)")
	)
	flag.Parse()

	if *dbURL == "" {
		return errors.New("database URL is required")
	}

	// 1. Connect to Database
	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	log.Println("Connected to database")

	// 2. Run Public Migrations
	if *scope == "all" || *scope == "public" {
		log.Println("=== Running Public Migrations ===")
		if err := runMigration(db, "public", "public", *cmd, *steps); err != nil {
			return fmt.Errorf("public migration failed: %w", err)
		}
	}

	// 3. Run Tenant Migrations
	if *scope == "all" || *scope == "tenant" {
		log.Println("=== Running Tenant Migrations ===")

		tenants, err := gatherTenants(db, *tenantID)
		if err != nil {
			return err
		}

		migrateTenants(db, tenants, *cmd, *steps)
	}

	return nil
}

// gatherTenants returns the list of tenant schemas to migrate. When tenantID is
// non-empty only that schema is returned; otherwise all non-deleted tenant
// schemas are read from the organizations table.
func gatherTenants(db *sql.DB, tenantID string) ([]string, error) {
	if tenantID != "" {
		return []string{tenantID}, nil
	}

	rows, err := db.Query("SELECT schema_name FROM organizations WHERE deleted_at IS NULL")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tenants: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tenants []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			log.Printf("Warning: failed to scan tenant schema: %v", err)
			continue
		}
		tenants = append(tenants, schema)
	}
	return tenants, nil
}

// migrateTenants ensures each tenant schema exists and runs its migrations,
// logging and continuing on per-tenant errors rather than failing hard.
func migrateTenants(db *sql.DB, tenants []string, cmd string, steps int) {
	for _, tenant := range tenants {
		log.Printf("Migrating tenant: %s", tenant)
		if err := ensureTenantSchemaExists(db, tenant); err != nil {
			log.Printf("Error ensuring tenant schema %s: %v", tenant, err)
			continue
		}
		// For tenant migrations, we must set the search path to the tenant schema
		// We do this by passing a configured driver instance or ensuring the connection uses the right path
		if err := runMigration(db, tenant, "tenant", cmd, steps); err != nil {
			log.Printf("Error migrating tenant %s: %v", tenant, err)
			// Decide whether to fail hard or continue. For now, we log and continue.
		}
	}
}

func ensureTenantSchemaExists(db *sql.DB, schemaName string) error {
	if !validSchemaName.MatchString(schemaName) {
		return fmt.Errorf("invalid schema name: %s", schemaName)
	}

	query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", pq.QuoteIdentifier(schemaName))
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to ensure schema exists: %w", err)
	}

	return nil
}

// migrationFile holds metadata parsed from an .up.sql filename.
type migrationFile struct {
	version uint
	name    string
}

// parseMigrationDir scans a directory for *.up.sql files and returns a sorted
// slice of migrationFile, mapping each version number to its human-readable name.
func parseMigrationDir(dir string) ([]migrationFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`^(\d+)_(.+)\.up\.sql$`)
	var files []migrationFile
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := re.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		v, _ := strconv.ParseUint(m[1], 10, 64)
		// Convert snake_case to readable label, e.g. "init_tenant_schema" → "init tenant schema"
		label := strings.ReplaceAll(m[2], "_", " ")
		files = append(files, migrationFile{version: uint(v), name: label})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].version < files[j].version })
	return files, nil
}

func runMigration(db *sql.DB, schemaName, migrationDirName, command string, steps int) error {
	driver, cleanup, err := buildMigrationDriver(db, schemaName, migrationDirName)
	if err != nil {
		return err
	}
	defer cleanup()

	cwd, _ := os.Getwd()
	migDir := filepath.Join(cwd, migrationsBaseDir, migrationDirName)
	sourceURL := fmt.Sprintf("file://%s", migDir)

	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Parse migration files for rich logging
	allMigrations, _ := parseMigrationDir(migDir)
	nameOf := make(map[uint]string, len(allMigrations))
	for _, f := range allMigrations {
		nameOf[f.version] = f.name
	}

	// Check current version before running
	versionBefore, versionErr := logCurrentVersion(m, schemaName, nameOf)

	// Show pending migrations when running 'up' or 'version'.
	// For 'version' this is terminal — it reports and returns without migrating.
	if command == "up" || command == "version" {
		if done := reportPendingMigrations(command, schemaName, allMigrations, versionBefore, versionErr); done {
			return nil
		}
	}

	if err := executeMigration(m, command, steps); err != nil {
		if err == migrate.ErrNoChange {
			// Already reported above for 'up'
			if command != "up" {
				log.Printf("[%s] No changes", schemaName)
			}
			return nil
		}
		return err
	}

	reportAppliedMigrations(m, schemaName, allMigrations, versionBefore)
	return nil
}

// buildMigrationDriver creates the golang-migrate database driver for the given
// schema. For tenant schemas it pins a dedicated connection with the search_path
// set; the returned cleanup func closes that connection (no-op for public).
func buildMigrationDriver(db *sql.DB, schemaName, migrationDirName string) (database.Driver, func(), error) {
	noop := func() {}

	if migrationDirName != "tenant" {
		driver, err := postgres.WithInstance(db, &postgres.Config{
			SchemaName: schemaName,
			// Using x-migrations-table to track migrations per schema independently
			MigrationsTable: "schema_migrations",
		})
		if err != nil {
			return nil, noop, fmt.Errorf("failed to create driver: %w", err)
		}
		return driver, noop, nil
	}

	conn, err := db.Conn(context.Background())
	if err != nil {
		return nil, noop, fmt.Errorf("failed to get tenant db connection: %w", err)
	}
	cleanup := func() { _ = conn.Close() }

	if _, err := conn.ExecContext(context.Background(), fmt.Sprintf("SET search_path TO %s, public", pq.QuoteIdentifier(schemaName))); err != nil {
		cleanup()
		return nil, noop, fmt.Errorf("failed to set search_path for tenant %s: %w", schemaName, err)
	}

	driver, err := postgres.WithConnection(context.Background(), conn, &postgres.Config{
		SchemaName:      schemaName,
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		cleanup()
		return nil, noop, fmt.Errorf("failed to create tenant driver: %w", err)
	}
	return driver, cleanup, nil
}

// logCurrentVersion logs the schema's current migration version and returns the
// version number plus the raw Version() error (so callers can detect
// migrate.ErrNilVersion for pending-migration calculations).
func logCurrentVersion(m *migrate.Migrate, schemaName string, nameOf map[uint]string) (uint, error) {
	versionBefore, dirty, err := m.Version()
	switch {
	case err != nil && err != migrate.ErrNilVersion:
		log.Printf("[%s] Failed to get version: %v", schemaName, err)
	case err == migrate.ErrNilVersion:
		log.Printf("[%s] Current version: none (fresh schema)", schemaName)
	default:
		if label := nameOf[versionBefore]; label != "" {
			log.Printf("[%s] Current version: %d — %s", schemaName, versionBefore, label)
		} else {
			log.Printf("[%s] Current version: %d", schemaName, versionBefore)
		}
		if dirty {
			log.Printf("[%s] WARNING: schema is dirty; manual intervention may be required", schemaName)
		}
	}
	return versionBefore, err
}

// reportPendingMigrations logs the migrations that have not yet been applied.
// It returns true when the call is terminal (command == "version"), signalling
// the caller to stop before running any migration.
func reportPendingMigrations(command, schemaName string, allMigrations []migrationFile, versionBefore uint, versionErr error) bool {
	var pending []migrationFile
	for _, f := range allMigrations {
		if versionErr == migrate.ErrNilVersion || f.version > versionBefore {
			pending = append(pending, f)
		}
	}

	if command == "version" {
		log.Printf("[%s] Pending migrations (%d):", schemaName, len(pending))
		for _, f := range pending {
			log.Printf("[%s]   → %06d  %s", schemaName, f.version, f.name)
		}
		return true
	}

	if len(pending) == 0 {
		log.Printf("[%s] No changes", schemaName)
	} else {
		log.Printf("[%s] Pending migrations (%d):", schemaName, len(pending))
		for _, f := range pending {
			log.Printf("[%s]   → %06d  %s", schemaName, f.version, f.name)
		}
	}
	return false
}

// executeMigration runs the requested migrate command (up/down/force).
func executeMigration(m *migrate.Migrate, command string, steps int) error {
	switch command {
	case "up":
		if steps > 0 {
			return m.Steps(steps)
		}
		return m.Up()
	case "down":
		if steps > 0 {
			return m.Steps(-steps)
		}
		return m.Down()
	case "force":
		return m.Force(steps)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// reportAppliedMigrations logs the migrations applied during the run and the
// resulting schema version.
func reportAppliedMigrations(m *migrate.Migrate, schemaName string, allMigrations []migrationFile, versionBefore uint) {
	versionAfter, _, _ := m.Version()
	if versionAfter != versionBefore {
		log.Printf("[%s] Applied migrations %d → %d:", schemaName, versionBefore, versionAfter)
		for _, f := range allMigrations {
			if f.version > versionBefore && f.version <= versionAfter {
				log.Printf("[%s]   ✓ %06d  %s", schemaName, f.version, f.name)
			}
		}
	}
	log.Printf("[%s] Migration complete (version: %d)", schemaName, versionAfter)
}
