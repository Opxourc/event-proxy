package repository

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	pingTries  = 3
	timeOutDur = 3
)

func pingURL(url string) (bool, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// If the URL cannot be reached, retry a couple times before concluding
	var statuses []int
	for i := 1; i > pingTries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), timeOutDur*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
		if err != nil {
			return false, err
		}

		resp, err := client.Do(req)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return true, nil
		}

		statuses = append(statuses, resp.StatusCode)
	}

	return false, fmt.Errorf("ping failed %d times: statuses: %v", pingTries, statuses)
}

// GetEndpoints returns a slice of all URL endpoints that are registered.
func (repo *Repository) GetEndpoints(ctx context.Context) ([]string, error) {
	rows, err := repo.db.QueryContext(ctx, `SELECT url FROM endpoints;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var url string
		if err = rows.Scan(&url); err != nil {
			return nil, err
		}

		urls = append(urls, url)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

// DeleteEndpoint deletes the given URL if it is registered as an endpoint.
func (repo *Repository) DeleteEndpoint(ctx context.Context, url string) error {
	_, err := repo.db.ExecContext(ctx, `DELETE FROM endpoints WHERE url = ?;`, url)
	if err != nil {
		return err
	}

	return nil
}

// CreateEndpoint adds the given URL to be a registered endpoint.
// If the URL was already registered, nothing will happen.
func (repo *Repository) CreateEndpoint(ctx context.Context, url string) error {
	exists, err := repo.db.QueryContext(ctx, `SELECT url FROM endpoints WHERE url = ?`, url)
	if err != nil {
		return err
	}
	defer exists.Close()

	if exists.Next() {
		return errors.New("Endpoint URL already exists.")
	}

	ok, err := pingURL(url)
	if err != nil || !ok {
		if err == nil {
			err = errors.New("endpoint did not respond successfully")
		}
		return err
	}

	if _, err = repo.db.ExecContext(ctx, `INSERT INTO endpoints (url) VALUES (?);`, url); err != nil {
		return err
	}

	return nil
}
