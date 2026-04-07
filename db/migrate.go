package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
)
//The Go embed package restricts paths to be within the same directory or subdirectories. 
// Going up a directory with .. is explicitly forbidden — hence, the migration folder is 
// placed inside the db package directory for embedding.

//The Following line is not a comment — it’s a directive to the Go compiler 
// to embed the contents of the specified file into the variable that follows it.

//go:embed migrations/001_create_patients.up.sql
var migration001 string


func Migrate(db *sql.DB) error {
	// Ensure the migrations tracking table exists first
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version     TEXT PRIMARY KEY,
		applied_at  TIMESTAMPTZ NOT NULL DEFAULT now()
	)`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	migrations := []struct {
		version string
		sql     string
	}{
		{"001_create_patients", migration001},
	}

	for _, m := range migrations {
		var exists bool
		err := db.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", m.version,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", m.version, err)
		}
		if exists {
			log.Printf("Migration %s already applied, skipping", m.version)
			continue
		}

		log.Printf("Applying migration %s...", m.version)
		if _, err := db.Exec(m.sql); err != nil {
			return fmt.Errorf("apply migration %s: %w", m.version, err)
		}
		if _, err := db.Exec(
			"INSERT INTO schema_migrations (version) VALUES ($1)", m.version,
		); err != nil {
			return fmt.Errorf("record migration %s: %w", m.version, err)
		}
		log.Printf("Migration %s applied successfully", m.version)
	}

	return nil
}
