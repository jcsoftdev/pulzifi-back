package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

const (
	defaultDBURL      = "postgres://pulzifi_user:pulzifi_password@localhost:5434/pulzifi?sslmode=disable"
	migrationsBaseDir = "shared/database/migrations"
)

func main() {
	var (
		dbURL    = flag.String("db", defaultDBURL, "Database URL")
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
			// For tenant migrations, we must set the search path to the tenant schema
			// We do this by passing a configured driver instance or ensuring the connection uses the right path
			if err := runMigration(db, tenant, "tenant", *cmd, *steps); err != nil {
				log.Printf("Error migrating tenant %s: %v", tenant, err)
				// Decide whether to fail hard or continue. For now, we log and continue.
			}
		}
	}
}

func runMigration(db *sql.DB, schemaName, migrationDirName, command string, steps int) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{
		SchemaName: schemaName,
		// Using x-migrations-table to track migrations per schema independently
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return fmt.Errorf("failed to create driver: %w", err)
	}

	cwd, _ := os.Getwd()
	sourceURL := fmt.Sprintf("file://%s/%s/%s", cwd, migrationsBaseDir, migrationDirName)

	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Check current version
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		log.Printf("[%s] Failed to get version: %v", schemaName, err)
	} else if err == migrate.ErrNilVersion {
		log.Printf("[%s] Current version: None", schemaName)
	} else {
		log.Printf("[%s] Current version: %d (dirty: %v)", schemaName, version, dirty)
	}

	if dirty {
		log.Printf("[%s] Database is dirty. Forcing version cleanup...", schemaName)
		// Option: m.Force(int(version))
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
	case "version":
		// Already printed above
		return nil
	default:
		return fmt.Errorf("unknown command: %s", command)
	}

	if err != nil {
		if err == migrate.ErrNoChange {
			log.Printf("[%s] No changes", schemaName)
			return nil
		}
		return err
	}

	log.Printf("[%s] Migration success", schemaName)
	return nil
}
