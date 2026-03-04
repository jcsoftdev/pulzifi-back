package main

import (
	"context"
	"database/sql"
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
	var (
		dbURL    = flag.String("db", defaultDBURL(), "Database URL")
		cmd      = flag.String("cmd", "up", "Command to run: up, down, version")
		steps    = flag.Int("steps", 0, "Number of steps to migrate (optional)")
		scope    = flag.String("scope", "all", "Scope of migration: all, public, tenant")
		tenantID = flag.String("tenant", "", "Specific tenant schema to migrate (optional)")
	)
	flag.Parse()

	if *dbURL == "" {
		log.Fatal("Database URL is required")
	}

	// 1. Connect to Database
	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database")

	// 2. Run Public Migrations
	if *scope == "all" || *scope == "public" {
		log.Println("=== Running Public Migrations ===")
		if err := runMigration(db, "public", "public", *cmd, *steps); err != nil {
			log.Fatalf("Public migration failed: %v", err)
		}
	}

	// 3. Run Tenant Migrations
	if *scope == "all" || *scope == "tenant" {
		log.Println("=== Running Tenant Migrations ===")

		var tenants []string
		if *tenantID != "" {
			tenants = []string{*tenantID}
		} else {
			// Fetch all tenant schemas
			rows, err := db.Query("SELECT schema_name FROM organizations WHERE deleted_at IS NULL")
			if err != nil {
				log.Fatalf("Failed to fetch tenants: %v", err)
			}
			defer rows.Close()

			for rows.Next() {
				var schema string
				if err := rows.Scan(&schema); err != nil {
					log.Printf("Warning: failed to scan tenant schema: %v", err)
					continue
				}
				tenants = append(tenants, schema)
			}
		}

		for _, tenant := range tenants {
			log.Printf("Migrating tenant: %s", tenant)
			if err := ensureTenantSchemaExists(db, tenant); err != nil {
				log.Printf("Error ensuring tenant schema %s: %v", tenant, err)
				continue
			}
			// For tenant migrations, we must set the search path to the tenant schema
			// We do this by passing a configured driver instance or ensuring the connection uses the right path
			if err := runMigration(db, tenant, "tenant", *cmd, *steps); err != nil {
				log.Printf("Error migrating tenant %s: %v", tenant, err)
				// Decide whether to fail hard or continue. For now, we log and continue.
			}
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
	var (
		driver database.Driver
		conn   *sql.Conn
		err    error
	)

	if migrationDirName == "tenant" {
		conn, err = db.Conn(context.Background())
		if err != nil {
			return fmt.Errorf("failed to get tenant db connection: %w", err)
		}
		defer conn.Close()

		if _, err := conn.ExecContext(context.Background(), fmt.Sprintf("SET search_path TO %s, public", pq.QuoteIdentifier(schemaName))); err != nil {
			return fmt.Errorf("failed to set search_path for tenant %s: %w", schemaName, err)
		}

		driver, err = postgres.WithConnection(context.Background(), conn, &postgres.Config{
			SchemaName:      schemaName,
			MigrationsTable: "schema_migrations",
		})
		if err != nil {
			return fmt.Errorf("failed to create tenant driver: %w", err)
		}
	} else {
		driver, err = postgres.WithInstance(db, &postgres.Config{
			SchemaName: schemaName,
			// Using x-migrations-table to track migrations per schema independently
			MigrationsTable: "schema_migrations",
		})
		if err != nil {
			return fmt.Errorf("failed to create driver: %w", err)
		}
	}

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
	versionBefore, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		log.Printf("[%s] Failed to get version: %v", schemaName, err)
	} else if err == migrate.ErrNilVersion {
		log.Printf("[%s] Current version: none (fresh schema)", schemaName)
	} else {
		label := nameOf[versionBefore]
		if label != "" {
			log.Printf("[%s] Current version: %d — %s", schemaName, versionBefore, label)
		} else {
			log.Printf("[%s] Current version: %d", schemaName, versionBefore)
		}
		if dirty {
			log.Printf("[%s] WARNING: schema is dirty; manual intervention may be required", schemaName)
		}
	}

	// Show pending migrations when running 'up'
	if command == "up" || command == "version" {
		var pending []migrationFile
		for _, f := range allMigrations {
			if err == migrate.ErrNilVersion || f.version > versionBefore {
				pending = append(pending, f)
			}
		}
		if command == "version" {
			log.Printf("[%s] Pending migrations (%d):", schemaName, len(pending))
			for _, f := range pending {
				log.Printf("[%s]   → %06d  %s", schemaName, f.version, f.name)
			}
			return nil
		}
		if len(pending) == 0 {
			log.Printf("[%s] No changes", schemaName)
		} else {
			log.Printf("[%s] Pending migrations (%d):", schemaName, len(pending))
			for _, f := range pending {
				log.Printf("[%s]   → %06d  %s", schemaName, f.version, f.name)
			}
		}
	}

	switch command {
	case "up":
		if steps > 0 {
			err = m.Steps(steps)
		} else {
			err = m.Up()
		}
	case "down":
		if steps > 0 {
			err = m.Steps(-steps)
		} else {
			err = m.Down()
		}
	case "force":
		err = m.Force(steps)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}

	if err != nil {
		if err == migrate.ErrNoChange {
			// Already reported above for 'up'
			if command != "up" {
				log.Printf("[%s] No changes", schemaName)
			}
			return nil
		}
		return err
	}

	// Report what changed
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
	return nil
}
