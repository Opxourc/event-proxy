package repository

import (
	"database/sql"
	"log"
)

type Repository struct {
	db *sql.DB
}

// New constructs a new Repository object.
// For the database, if the tables needed don't already exist, this function will create them.
func New(db *sql.DB) *Repository {
	repo := &Repository{
		db: db,
	}

	// Before returning, ensure that the database has the tables it needs
	if _, err := db.Exec(
		`CREATE TABLE IF NOT EXISTS endpoints (
			id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			url TEXT NOT NULL,
		);
		
		CREATE TABLE IF NOT EXISTS events (
			id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			source_url TEXT NOT NULL,
			target_url TEXT NOT NULL,
			method TEXT NOT NULL,
			status TEXT NOT NULL,
			type TEXT NOT NULL,
		);`,
	); err != nil {
		log.Fatalf("Failed to create DB table: %v\n", err)
	}

	return repo
}
