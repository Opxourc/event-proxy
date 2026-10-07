package endpoints

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type handlerServiceStub struct {
	endpoints  []string
	createdURL string
	deletedURL string
	createErr  error
}

func (service *handlerServiceStub) GetEndpoints(context.Context) ([]string, error) {
	return service.endpoints, nil
}

func (service *handlerServiceStub) CreateEndpoint(_ context.Context, rawURL string) error {
	service.createdURL = rawURL
	return service.createErr
}

func (service *handlerServiceStub) DeleteEndpoint(_ context.Context, rawURL string) error {
	service.deletedURL = rawURL
	return nil
}

func TestCreateEndpointDecodesRequestAndCallsService(t *testing.T) {
	service := &handlerServiceStub{}
	handler := NewHandler(service)
	request := httptest.NewRequest(
		http.MethodPost,
		"/endpoints/",
		strings.NewReader(`{"url":"https://example.com/events"}`),
	)
	recorder := httptest.NewRecorder()

	handler.CreateEndpoint(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("HTTP status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if service.createdURL != "https://example.com/events" {
		t.Fatalf("service received URL %q", service.createdURL)
	}
}

func TestCreateEndpointRejectsMalformedAndTrailingJSON(t *testing.T) {
	for _, body := range []string{`{"url":`, `{"url":"https://example.com"} {}`} {
		service := &handlerServiceStub{}
		handler := NewHandler(service)
		request := httptest.NewRequest(http.MethodPost, "/endpoints/", strings.NewReader(body))
		recorder := httptest.NewRecorder()

		handler.CreateEndpoint(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("body %q returned status %d, want %d", body, recorder.Code, http.StatusBadRequest)
		}
		if service.createdURL != "" {
			t.Errorf("body %q reached service with URL %q", body, service.createdURL)
		}
	}
}

func TestGetEndpointsReturnsJSONList(t *testing.T) {
	handler := NewHandler(&handlerServiceStub{endpoints: []string{"https://example.com"}})
	request := httptest.NewRequest(http.MethodGet, "/endpoints/", nil)
	recorder := httptest.NewRecorder()

	handler.GetEndpoints(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("HTTP status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got, want := recorder.Body.String(), "[\"https://example.com\"]\n"; got != want {
		t.Fatalf("response body = %q, want %q", got, want)
	}
}
