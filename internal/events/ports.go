package events

import (
	"context"
	"net/http"
)

// Service is the event-delivery use-case API used by the HTTP handler.
type Service interface {
	// Send validates and delivers an event to a registered endpoint.
	Send(ctx context.Context, request Request) (Response, error)
}

// Repository checks endpoint registration and records event delivery status.
type Repository interface {
	// IsEndpointRegistered reports whether the URL is registered.
	IsEndpointRegistered(ctx context.Context, rawURL string) (bool, error)
	// CreateEvent stores a pending event record and returns its ID.
	CreateEvent(ctx context.Context, record Record) (int64, error)
	// UpdateEventStatus stores the final delivery status.
	UpdateEventStatus(ctx context.Context, id int64, status string) error
}

// Sender forwards event data to the registered endpoint.
type Sender interface {
	// Send forwards the event payload and returns the upstream response.
	Send(ctx context.Context, method, rawURL string, payload []byte) (Response, error)
}

// Request contains the HTTP method, destination URL, and JSON event payload.
type Request struct {
	Method string
	URL    string
	Data   []byte
}

// Response contains the upstream HTTP response returned after event delivery.
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// Record contains the event fields persisted while delivery is processed.
type Record struct {
	Method    string
	TargetURL string
	Status    string
	Type      string
}
