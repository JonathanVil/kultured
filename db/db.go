package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if _, err = db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		return nil, err
	}
	if err := initSchema(db); err != nil {
		return nil, err
	}
	if err := applyMigrations(db); err != nil {
		return nil, err
	}
	return db, nil
}

// initSchema creates the base tables for fresh installs and sets up the
// migrations tracker. Column renames and structural changes are handled by
// numbered migration files instead.
func initSchema(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS batches (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			name           TEXT NOT NULL,
			started_at     TEXT NOT NULL,
			tea_type       TEXT NOT NULL,
			sugar_g        REAL NOT NULL,
			volume_l       REAL NOT NULL,
			scoby_weight_g REAL NOT NULL,
			stage          TEXT NOT NULL DEFAULT 'f1',
			notes          TEXT,
			created_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
		)`,
		`CREATE TABLE IF NOT EXISTS readings (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			batch_id    INTEGER NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
			recorded_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
			gravity     REAL,
			temp_c      REAL,
			taste_notes TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			name       TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
		)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	// Ignored: these fail silently if columns already exist.
	db.Exec(`ALTER TABLE batches ADD COLUMN tea_g REAL NOT NULL DEFAULT 0`)
	db.Exec(`ALTER TABLE batches ADD COLUMN steep_min REAL NOT NULL DEFAULT 0`)
	return nil
}

func applyMigrations(db *sql.DB) error {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".sql")

		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE name = ?`, name).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		data, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return err
		}
		if err := runMigration(db, name, string(data)); err != nil {
			return err
		}
	}
	return nil
}

func runMigration(db *sql.DB, name, sqlStr string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("migration %s: begin: %w", name, err)
	}
	defer tx.Rollback()

	for _, stmt := range strings.Split(sqlStr, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
	}
	if _, err := tx.Exec(`INSERT INTO schema_migrations (name) VALUES (?)`, name); err != nil {
		return fmt.Errorf("migration %s: record: %w", name, err)
	}
	return tx.Commit()
}
