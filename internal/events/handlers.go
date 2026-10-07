package events

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
)

// Handler translates incoming event requests to service calls and relays the
// successful upstream response back to the caller.
type Handler struct {
	service Service
}

// NewHandler constructs an event handler for the given service.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// SendEvent decodes an event request and relays the upstream response.
func (handler *Handler) SendEvent(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Method string          `json:"method"`
		URL    string          `json:"url"`
		Data   json.RawMessage `json:"data"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
		log.Printf("%s: %v", Errors["Log"]["InvalidRequestBody"], err)
		http.Error(w, Errors["HTTP"]["InvalidRequestBody"], http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		log.Printf("%s: trailing or malformed JSON", Errors["Log"]["InvalidRequestBody"])
		http.Error(w, Errors["HTTP"]["InvalidRequestBody"], http.StatusBadRequest)
		return
	}

	response, err := handler.service.Send(r.Context(), Request{
		Method: body.Method,
		URL:    body.URL,
		Data:   body.Data,
	})
	if err != nil {
		handler.writeServiceError(w, err)
		return
	}

	copyResponseHeaders(w.Header(), response.Header)
	w.WriteHeader(response.StatusCode)
	if _, err := w.Write(response.Body); err != nil {
		log.Printf("%s: %v", Errors["Log"]["WriteResponseFailure"], err)
	}
}

func (handler *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrEndpointNotRegistered):
		log.Printf("%s: %v", Errors["Log"]["EndpointNotRegistered"], err)
		http.Error(w, Errors["HTTP"]["EndpointNotRegistered"], http.StatusBadRequest)
	case errors.Is(err, ErrInvalidMethod):
		http.Error(w, Errors["HTTP"]["InvalidMethod"], http.StatusBadRequest)
	case errors.Is(err, ErrInvalidURL):
		http.Error(w, Errors["HTTP"]["InvalidURL"], http.StatusBadRequest)
	case errors.Is(err, ErrMissingData):
		http.Error(w, Errors["HTTP"]["MissingData"], http.StatusBadRequest)
	case errors.Is(err, ErrInvalidData):
		http.Error(w, Errors["HTTP"]["InvalidData"], http.StatusBadRequest)
	case errors.Is(err, ErrDelivery):
		log.Printf("%s: %v", Errors["Log"]["DeliveryFailure"], err)
		http.Error(w, Errors["HTTP"]["DeliveryFailure"], http.StatusBadGateway)
	default:
		log.Printf("%s: %v", Errors["Log"]["ProcessingFailure"], err)
		http.Error(w, Errors["HTTP"]["ProcessingFailure"], http.StatusInternalServerError)
	}
}

func copyResponseHeaders(destination, source http.Header) {
	hopByHop := map[string]struct{}{
		"Connection":          {},
		"Keep-Alive":          {},
		"Proxy-Authenticate":  {},
		"Proxy-Authorization": {},
		"Te":                  {},
		"Trailer":             {},
		"Transfer-Encoding":   {},
		"Upgrade":             {},
	}
	for _, value := range source.Values("Connection") {
		for _, name := range strings.Split(value, ",") {
			hopByHop[http.CanonicalHeaderKey(strings.TrimSpace(name))] = struct{}{}
		}
	}
	for name, values := range source {
		if _, excluded := hopByHop[http.CanonicalHeaderKey(name)]; excluded {
			continue
		}
		for _, value := range values {
			destination.Add(name, value)
		}
	}
}
