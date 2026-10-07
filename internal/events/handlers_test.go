package events_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Opxourc/event-proxy/internal/events"
)

type serviceStub struct {
	request events.Request
}

func (service *serviceStub) Send(_ context.Context, request events.Request) (events.Response, error) {
	service.request = request
	return events.Response{
		StatusCode: http.StatusAccepted,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       []byte(`{"accepted":true}`),
	}, nil
}

func TestSendEventPassesDecodedRequestAndRelaysResponse(t *testing.T) {
	service := &serviceStub{}
	handler := events.NewHandler(service)
	request := httptest.NewRequest(http.MethodPost, "/event/", strings.NewReader(
		`{"method":"PATCH","url":"https://example.com/events","data":{"id":7}}`,
	))
	recorder := httptest.NewRecorder()

	handler.SendEvent(recorder, request)

	if service.request.Method != http.MethodPatch || service.request.URL != "https://example.com/events" {
		t.Fatalf("service request = %#v", service.request)
	}
	var data map[string]int
	if err := json.Unmarshal(service.request.Data, &data); err != nil {
		t.Fatalf("event data is invalid JSON: %v", err)
	}
	if data["id"] != 7 {
		t.Fatalf("event data = %s", service.request.Data)
	}
	if recorder.Code != http.StatusAccepted || recorder.Body.String() != `{"accepted":true}` {
		t.Fatalf("HTTP response = %d %q", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type = %q", recorder.Header().Get("Content-Type"))
	}
}

func TestSendEventRejectsTrailingJSON(t *testing.T) {
	handler := events.NewHandler(&serviceStub{})
	request := httptest.NewRequest(http.MethodPost, "/event/", strings.NewReader(
		`{"method":"POST","url":"https://example.com","data":{}} {}`,
	))
	recorder := httptest.NewRecorder()

	handler.SendEvent(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("HTTP status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}
