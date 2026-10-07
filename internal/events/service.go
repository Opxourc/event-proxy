package events

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const sendAttempts = 3

var (
	ErrEndpointNotRegistered = errors.New("endpoint is not registered")
	ErrInvalidMethod         = errors.New("invalid HTTP method")
	ErrInvalidURL            = errors.New("invalid endpoint URL")
	ErrMissingData           = errors.New("event data is required")
	ErrInvalidData           = errors.New("event data must be valid JSON")
	ErrDelivery              = errors.New("event delivery failed")
)

// EventService validates and records an event before forwarding it, then records
// whether delivery succeeded.
type EventService struct {
	repository Repository
	sender     Sender
}

var _ Service = (*EventService)(nil)

// NewService constructs an event service from its repository and sender.
func NewService(repository Repository, sender Sender) *EventService {
	return &EventService{repository: repository, sender: sender}
}

// Send validates the request, records it, delivers it, and updates its status.
func (service *EventService) Send(ctx context.Context, request Request) (Response, error) {
	if err := validateRequest(request); err != nil {
		return Response{}, err
	}

	registered, err := service.repository.IsEndpointRegistered(ctx, request.URL)
	if err != nil {
		return Response{}, fmt.Errorf("check endpoint registration: %w", err)
	}
	if !registered {
		return Response{}, ErrEndpointNotRegistered
	}

	eventID, err := service.repository.CreateEvent(ctx, Record{
		Method:    request.Method,
		TargetURL: request.URL,
		Status:    "pending",
		Type:      "application/json",
	})
	if err != nil {
		return Response{}, fmt.Errorf("record event: %w", err)
	}

	response, deliveryErr := service.deliver(ctx, request)
	if deliveryErr != nil {
		updateCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if err := service.repository.UpdateEventStatus(updateCtx, eventID, "failed"); err != nil {
			return Response{}, fmt.Errorf("delivery failed: %v; record failed status: %w", deliveryErr, err)
		}
		return Response{}, deliveryErr
	}

	status := fmt.Sprintf("delivered: %d", response.StatusCode)
	if err := service.repository.UpdateEventStatus(ctx, eventID, status); err != nil {
		return Response{}, fmt.Errorf("record successful delivery: %w", err)
	}
	return response, nil
}

func (service *EventService) deliver(ctx context.Context, request Request) (Response, error) {
	var lastErr error
	for attempt := 0; attempt < sendAttempts; attempt++ {
		response, err := service.sender.Send(ctx, request.Method, request.URL, request.Data)
		if err == nil {
			return response, nil
		}
		lastErr = err

		if ctx.Err() != nil {
			lastErr = ctx.Err()
			break
		}
		if attempt+1 < sendAttempts {
			timer := time.NewTimer(time.Duration(attempt+1) * 100 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				lastErr = ctx.Err()
				attempt = sendAttempts
			case <-timer.C:
			}
		}
	}
	return Response{}, fmt.Errorf("%w after %d attempts: %v", ErrDelivery, sendAttempts, lastErr)
}

func validateRequest(request Request) error {
	if !validMethod(request.Method) {
		return ErrInvalidMethod
	}
	parsedURL, err := url.ParseRequestURI(request.URL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" || parsedURL.User != nil {
		return ErrInvalidURL
	}
	if len(request.Data) == 0 {
		return ErrMissingData
	}
	if !json.Valid(request.Data) {
		return ErrInvalidData
	}
	return nil
}

func validMethod(method string) bool {
	if method == "" {
		return false
	}
	for _, char := range method {
		if char <= 32 || char >= 127 || strings.ContainsRune("()<>@,;:\\\"/[]?={} \t", char) {
			return false
		}
	}
	return true
}

type httpSender struct {
	client *http.Client
}

// NewHTTPSender returns a sender that forwards JSON data and captures the
// upstream status, headers, and body.
func NewHTTPSender() Sender {
	return &httpSender{client: &http.Client{Timeout: 30 * time.Second}}
}

func (sender *httpSender) Send(ctx context.Context, method, rawURL string, payload []byte) (Response, error) {
	request, err := http.NewRequestWithContext(ctx, method, rawURL, bytes.NewReader(payload))
	if err != nil {
		return Response{}, err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := sender.client.Do(request)
	if err != nil {
		return Response{}, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return Response{}, err
	}
	return Response{
		StatusCode: response.StatusCode,
		Header:     response.Header.Clone(),
		Body:       body,
	}, nil
}

var _ Sender = (*httpSender)(nil)
