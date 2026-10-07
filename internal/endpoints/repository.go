package endpoints

import (
	"context"
	"database/sql"
	"fmt"
)

// SQLRepository stores registered endpoints in PostgreSQL.
type SQLRepository struct {
	db *sql.DB
}

// NewRepository constructs a SQL-backed endpoint repository.
func NewRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

var _ Repository = (*SQLRepository)(nil)

// GetEndpoints loads registered endpoint URLs in registration order.
func (repository *SQLRepository) GetEndpoints(ctx context.Context) ([]string, error) {
	rows, err := repository.db.QueryContext(ctx, `SELECT url FROM endpoints ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("query endpoints: %w", err)
	}
	defer rows.Close()

	endpoints := make([]string, 0)
	for rows.Next() {
		var rawURL string
		if err := rows.Scan(&rawURL); err != nil {
			return nil, fmt.Errorf("scan endpoint: %w", err)
		}
		endpoints = append(endpoints, rawURL)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate endpoints: %w", err)
	}
	return endpoints, nil
}

// CreateEndpoint stores a URL unless it is already registered.
func (repository *SQLRepository) CreateEndpoint(ctx context.Context, rawURL string) error {
	result, err := repository.db.ExecContext(ctx, `INSERT INTO endpoints (url) VALUES ($1) ON CONFLICT (url) DO NOTHING`, rawURL)
	if err != nil {
		return fmt.Errorf("insert endpoint: %w", err)
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check endpoint insert result: %w", err)
	}
	if inserted == 0 {
		return ErrEndpointAlreadyRegistered
	}
	return nil
}

// DeleteEndpoint removes the matching endpoint URL.
func (repository *SQLRepository) DeleteEndpoint(ctx context.Context, rawURL string) error {
	if _, err := repository.db.ExecContext(ctx, `DELETE FROM endpoints WHERE url = $1`, rawURL); err != nil {
		return fmt.Errorf("delete endpoint: %w", err)
	}
	return nil
}
