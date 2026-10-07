package events

import (
	"context"
	"database/sql"
	"fmt"
)

// SQLRepository checks endpoint registration and persists event delivery state.
// SQLRepository stores event records and queries endpoint registrations.
type SQLRepository struct {
	db *sql.DB
}

// NewRepository constructs a SQL-backed event repository.
func NewRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

var _ Repository = (*SQLRepository)(nil)

// IsEndpointRegistered reports whether the URL is in the endpoint registry.
func (repository *SQLRepository) IsEndpointRegistered(ctx context.Context, rawURL string) (bool, error) {
	var registered bool
	if err := repository.db.QueryRowContext(
		ctx,
		`SELECT EXISTS (SELECT 1 FROM endpoints WHERE url = $1)`,
		rawURL,
	).Scan(&registered); err != nil {
		return false, fmt.Errorf("query endpoint registration: %w", err)
	}
	return registered, nil
}

// CreateEvent inserts a pending event record and returns its database ID.
func (repository *SQLRepository) CreateEvent(ctx context.Context, record Record) (int64, error) {
	var id int64
	err := repository.db.QueryRowContext(
		ctx,
		`INSERT INTO events (source_url, target_url, method, status, type)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		"event-proxy",
		record.TargetURL,
		record.Method,
		record.Status,
		record.Type,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert event record: %w", err)
	}
	return id, nil
}

// UpdateEventStatus replaces the stored status for an event.
func (repository *SQLRepository) UpdateEventStatus(ctx context.Context, id int64, status string) error {
	if _, err := repository.db.ExecContext(
		ctx,
		`UPDATE events SET status = $1 WHERE id = $2`,
		status,
		id,
	); err != nil {
		return fmt.Errorf("update event status: %w", err)
	}
	return nil
}
