package endpoints

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

// Handler translates endpoint HTTP requests into service calls.
type Handler struct {
	service Service
}

// NewHandler constructs an endpoint handler for the given service.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// GetEndpoints writes the registered endpoint URLs as JSON.
func (handler *Handler) GetEndpoints(w http.ResponseWriter, r *http.Request) {
	endpoints, err := handler.service.GetEndpoints(r.Context())
	if err != nil {
		log.Printf("%s: %v", Errors["Log"]["GetEndpointsFailure"], err)
		http.Error(w, Errors["HTTP"]["GetEndpointsFailure"], http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(endpoints); err != nil {
		log.Printf("%s: %v", Errors["Log"]["EncodeEndpointsFailure"], err)
		http.Error(w, Errors["HTTP"]["EncodeEndpointsFailure"], http.StatusInternalServerError)
	}
}

// CreateEndpoint decodes and registers an endpoint URL from the request body.
func (handler *Handler) CreateEndpoint(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL string `json:"url"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
		log.Printf("%s: %v", Errors["Log"]["InvalidEndpointRequest"], err)
		http.Error(w, Errors["HTTP"]["InvalidEndpointRequest"], http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		log.Printf("%s: trailing or malformed JSON", Errors["Log"]["InvalidEndpointRequest"])
		http.Error(w, Errors["HTTP"]["InvalidEndpointRequest"], http.StatusBadRequest)
		return
	}

	if err := handler.service.CreateEndpoint(r.Context(), body.URL); err != nil {
		switch {
		case errors.Is(err, ErrMissingEndpointURL):
			http.Error(w, Errors["HTTP"]["MissingEndpointURL"], http.StatusBadRequest)
		case errors.Is(err, ErrInvalidEndpointURL):
			http.Error(w, Errors["HTTP"]["InvalidEndpointURL"], http.StatusBadRequest)
		case errors.Is(err, ErrEndpointUnreachable):
			log.Printf("%s: %v", Errors["Log"]["EndpointUnreachable"], err)
			http.Error(w, Errors["HTTP"]["EndpointUnreachable"], http.StatusBadGateway)
		case errors.Is(err, ErrEndpointAlreadyRegistered):
			log.Printf("%s: %v", Errors["Log"]["EndpointAlreadyRegistered"], err)
			http.Error(w, Errors["HTTP"]["EndpointAlreadyRegistered"], http.StatusConflict)
		default:
			log.Printf("%s: %v", Errors["Log"]["CreateEndpointFailure"], err)
			http.Error(w, Errors["HTTP"]["CreateEndpointFailure"], http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// DeleteEndpoint removes the URL supplied in the request's url query parameter.
func (handler *Handler) DeleteEndpoint(w http.ResponseWriter, r *http.Request) {
	rawURL := r.URL.Query().Get("url")
	if err := handler.service.DeleteEndpoint(r.Context(), rawURL); err != nil {
		if errors.Is(err, ErrMissingEndpointURL) {
			http.Error(w, Errors["HTTP"]["MissingEndpointURL"], http.StatusBadRequest)
			return
		}
		log.Printf("%s: %v", Errors["Log"]["DeleteEndpointFailure"], err)
		http.Error(w, Errors["HTTP"]["DeleteEndpointFailure"], http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
