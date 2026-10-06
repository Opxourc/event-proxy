package database

import (
	"database/sql"
	"fmt"
)

// Open connects to the database and creates the tables required by the
// endpoints and events features.
func Open(driver string, source string) (*sql.DB, error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if _, err := db.Exec(
		`CREATE TABLE IF NOT EXISTS endpoints (
			id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			url TEXT NOT NULL UNIQUE
		);
		
		CREATE TABLE IF NOT EXISTS events (
			id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			source_url TEXT NOT NULL,
			target_url TEXT NOT NULL,
			method TEXT NOT NULL,
			status TEXT NOT NULL,
			type TEXT NOT NULL
		);`,
	); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("initialize database schema: %w", err)
	}

	return db, nil
}
