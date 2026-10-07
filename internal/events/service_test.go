package events

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
)

type repositoryStub struct {
	registered bool
	record     Record
	statuses   []string
}

func (repo *repositoryStub) IsEndpointRegistered(context.Context, string) (bool, error) {
	return repo.registered, nil
}

func (repo *repositoryStub) CreateEvent(_ context.Context, record Record) (int64, error) {
	repo.record = record
	return 42, nil
}

func (repo *repositoryStub) UpdateEventStatus(_ context.Context, _ int64, status string) error {
	repo.statuses = append(repo.statuses, status)
	return nil
}

type senderStub struct {
	calls   int
	err     error
	failFor int
	method  string
	url     string
	payload []byte
}

func (sender *senderStub) Send(_ context.Context, method, url string, payload []byte) (Response, error) {
	sender.calls++
	sender.method = method
	sender.url = url
	sender.payload = append([]byte(nil), payload...)
	if sender.err != nil && sender.calls <= sender.failFor {
		return Response{}, sender.err
	}
	return Response{
		StatusCode: http.StatusAccepted,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       []byte(`{"received":true}`),
	}, nil
}

func TestSendChecksRegistrationAndForwardsEvent(t *testing.T) {
	repo := &repositoryStub{registered: true}
	sender := &senderStub{}
	service := NewService(repo, sender)
	payload := []byte(`{"message":"hello"}`)

	response, err := service.Send(context.Background(), Request{
		Method: http.MethodPost,
		URL:    "https://example.com/events",
		Data:   payload,
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if sender.calls != 1 || sender.method != http.MethodPost || sender.url != "https://example.com/events" || !reflect.DeepEqual(sender.payload, payload) {
		t.Fatalf("Send() did not forward the original event: %#v", sender)
	}
	if response.StatusCode != http.StatusAccepted || string(response.Body) != `{"received":true}` {
		t.Fatalf("Send() response = %#v", response)
	}
	if repo.record.Status != "pending" || !reflect.DeepEqual(repo.statuses, []string{"delivered: 202"}) {
		t.Fatalf("event status record = %#v, updates = %v", repo.record, repo.statuses)
	}
}

func TestSendRejectsUnregisteredEndpoint(t *testing.T) {
	repo := &repositoryStub{}
	sender := &senderStub{}
	service := NewService(repo, sender)

	_, err := service.Send(context.Background(), Request{
		Method: http.MethodPost,
		URL:    "https://example.com/events",
		Data:   []byte(`{}`),
	})
	if !errors.Is(err, ErrEndpointNotRegistered) {
		t.Fatalf("Send() error = %v, want %v", err, ErrEndpointNotRegistered)
	}
	if sender.calls != 0 {
		t.Fatalf("Send() called sender %d times for unregistered endpoint", sender.calls)
	}
}

func TestSendRetriesFailedDelivery(t *testing.T) {
	repo := &repositoryStub{registered: true}
	sender := &senderStub{err: errors.New("connection refused"), failFor: sendAttempts}
	service := NewService(repo, sender)

	_, err := service.Send(context.Background(), Request{
		Method: http.MethodPost,
		URL:    "https://example.com/events",
		Data:   []byte(`{"value":1}`),
	})
	if !errors.Is(err, ErrDelivery) {
		t.Fatalf("Send() error = %v, want %v", err, ErrDelivery)
	}
	if sender.calls != sendAttempts {
		t.Fatalf("Send() made %d attempts, want %d", sender.calls, sendAttempts)
	}
	if !reflect.DeepEqual(repo.statuses, []string{"failed"}) {
		t.Fatalf("event status updates = %v, want [failed]", repo.statuses)
	}
}

func TestSendSucceedsAfterRetry(t *testing.T) {
	repo := &repositoryStub{registered: true}
	sender := &senderStub{err: errors.New("temporary connection failure"), failFor: 1}
	service := NewService(repo, sender)

	response, err := service.Send(context.Background(), Request{
		Method: http.MethodPost,
		URL:    "https://example.com/events",
		Data:   []byte(`{"value":1}`),
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if sender.calls != 2 || response.StatusCode != http.StatusAccepted {
		t.Fatalf("Send() made %d attempts and returned status %d", sender.calls, response.StatusCode)
	}
	if !reflect.DeepEqual(repo.statuses, []string{"delivered: 202"}) {
		t.Fatalf("event status updates = %v, want [delivered: 202]", repo.statuses)
	}
}

func TestSendValidatesMethodURLAndData(t *testing.T) {
	tests := []struct {
		name    string
		request Request
		wantErr error
	}{
		{"invalid method", Request{Method: "BAD METHOD", URL: "https://example.com", Data: []byte(`{}`)}, ErrInvalidMethod},
		{"invalid URL", Request{Method: http.MethodPost, URL: "file:///tmp/data", Data: []byte(`{}`)}, ErrInvalidURL},
		{"missing data", Request{Method: http.MethodPost, URL: "https://example.com"}, ErrMissingData},
		{"invalid JSON data", Request{Method: http.MethodPost, URL: "https://example.com", Data: []byte(`{`)}, ErrInvalidData},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(&repositoryStub{registered: true}, &senderStub{})
			_, err := service.Send(context.Background(), test.request)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Send() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}
