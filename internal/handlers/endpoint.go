package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Opxourc/event-proxy/internal/repository"
)

// GetEndpoints gets stored endpoints from the database and writes them to the client.
func GetEndpoints(repo *repository.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		endpoints, err := repo.GetEndpoints(r.Context())
		if err != nil {
			http.Error(w, "could not retrieve endpoints", http.StatusInternalServerError)
			return
		}
		if endpoints == nil {
			endpoints = []string{}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(endpoints); err != nil {
			http.Error(w, "could not encode endpoints", http.StatusInternalServerError)
		}
	}
}

// CreateEndpoint takes the requested URL for creation and attempts to register it.
func CreateEndpoint(repo *repository.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid endpoint request body", http.StatusBadRequest)
			return
		}
		if request.URL == "" {
			http.Error(w, "endpoint URL is required", http.StatusBadRequest)
			return
		}

		if err := repo.CreateEndpoint(r.Context(), request.URL); err != nil {
			http.Error(w, "could not create endpoint", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

// DeleteEndpoint deletes an already existing provided URL.
func DeleteEndpoint(repo *repository.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Query().Get("url")
		if url == "" {
			http.Error(w, "endpoint URL is required", http.StatusBadRequest)
			return
		}

		if err := repo.DeleteEndpoint(r.Context(), url); err != nil {
			http.Error(w, "could not delete endpoint", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
