package endpoints

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Opxourc/event-proxy/internal/utility"
)

const (
	checkAttempts = 3
	checkTimeout  = 3 * time.Second
)

// EndpointService validates endpoint requests, checks reachability, and delegates
// persistence to the endpoint repository.
type EndpointService struct {
	repository Repository
	checker    Checker
}

var _ Service = (*EndpointService)(nil)

// NewService constructs an endpoint service from its repository and checker.
func NewService(repository Repository, checker Checker) *EndpointService {
	return &EndpointService{repository: repository, checker: checker}
}

// GetEndpoints returns the registered endpoint URLs.
func (service *EndpointService) GetEndpoints(ctx context.Context) ([]string, error) {
	endpoints, err := service.repository.GetEndpoints(ctx)
	if err != nil {
		return nil, fmt.Errorf("get endpoints: %w", err)
	}
	if endpoints == nil {
		return []string{}, nil
	}
	return endpoints, nil
}

// CreateEndpoint checks an endpoint's URL and reachability before registering it.
func (service *EndpointService) CreateEndpoint(ctx context.Context, rawURL string) error {
	if rawURL == "" {
		return ErrMissingEndpointURL
	}
	if !validEndpointURL(rawURL) {
		return ErrInvalidEndpointURL
	}
	if err := service.checker.Check(ctx, rawURL); err != nil {
		return fmt.Errorf("%w: %w", ErrEndpointUnreachable, err)
	}
	if err := service.repository.CreateEndpoint(ctx, rawURL); err != nil {
		return fmt.Errorf("create endpoint: %w", err)
	}
	return nil
}

// DeleteEndpoint removes a registered endpoint URL.
func (service *EndpointService) DeleteEndpoint(ctx context.Context, rawURL string) error {
	if rawURL == "" {
		return ErrMissingEndpointURL
	}
	if err := service.repository.DeleteEndpoint(ctx, rawURL); err != nil {
		return fmt.Errorf("delete endpoint: %w", err)
	}
	return nil
}

func validEndpointURL(rawURL string) bool {
	parsedURL, err := url.ParseRequestURI(rawURL)
	return err == nil &&
		(parsedURL.Scheme == "http" || parsedURL.Scheme == "https") &&
		parsedURL.Host != "" &&
		parsedURL.User == nil
}

type httpChecker struct {
	client *http.Client
}

// NewHTTPChecker returns a checker that probes an endpoint with HEAD requests.
func NewHTTPChecker() Checker {
	return &httpChecker{client: &http.Client{Timeout: checkTimeout}}
}

func (checker *httpChecker) Check(ctx context.Context, rawURL string) error {
	var lastErr error

	for attempt := 0; attempt < checkAttempts; attempt++ {
		if err := utility.PingURL(ctx, rawURL, checker.client); err == nil {
			return nil
		} else {
			lastErr = err
			if ctx.Err() != nil {
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("endpoint check failed after %d attempts: %w", checkAttempts, lastErr)
}

var _ Checker = (*httpChecker)(nil)
