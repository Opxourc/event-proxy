package endpoints

import "context"

// Service is the endpoint use-case API used by the HTTP handlers.
type Service interface {
	// GetEndpoints returns all registered endpoint URLs.
	GetEndpoints(ctx context.Context) ([]string, error)
	// CreateEndpoint validates and registers an endpoint URL.
	CreateEndpoint(ctx context.Context, rawURL string) error
	// DeleteEndpoint removes an endpoint URL from the registry.
	DeleteEndpoint(ctx context.Context, rawURL string) error
}

// Repository stores registered endpoint URLs.
type Repository interface {
	// GetEndpoints loads all registered endpoint URLs.
	GetEndpoints(ctx context.Context) ([]string, error)
	// CreateEndpoint stores an endpoint URL.
	CreateEndpoint(ctx context.Context, rawURL string) error
	// DeleteEndpoint removes an endpoint URL.
	DeleteEndpoint(ctx context.Context, rawURL string) error
}

// Checker verifies that an endpoint can be reached before it is registered.
type Checker interface {
	// Check reports whether an endpoint can be reached.
	Check(ctx context.Context, rawURL string) error
}
