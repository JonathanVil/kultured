package db

import (
    "database/sql"
    _ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
    db, err := sql.Open("sqlite", path)
    if err != nil {
        return nil, err
    }

    if err := db.Ping(); err != nil {
        return nil, err
    }

    _, err = db.Exec(`PRAGMA foreign_keys = ON`)
    if err != nil {
        return nil, err
    }

    if err := migrate(db); err != nil {
        return nil, err
    }

    return db, nil
}

func migrate(db *sql.DB) error {
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS batches (
            id              INTEGER PRIMARY KEY AUTOINCREMENT,
            name            TEXT NOT NULL,
            started_at      TEXT NOT NULL,
            tea_type        TEXT NOT NULL,
            sugar_g         REAL NOT NULL,
            volume_l        REAL NOT NULL,
            scoby_weight_g  REAL NOT NULL,
            stage           TEXT NOT NULL DEFAULT 'f1',
            notes           TEXT,
            created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
        )
    `)
    return err
}
