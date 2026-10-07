package main

import (
	"log"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/Opxourc/event-proxy/internal/endpoints"
	"github.com/Opxourc/event-proxy/internal/events"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Application serves the HTTP API using the configured handlers and address.
type Application struct {
	endpointHandler *endpoints.Handler
	eventHandler    *events.Handler
	address         string
	router          *chi.Mux
}

// NewApplication constructs an application with endpoint and event handlers.
func NewApplication(address string, endpointHandler *endpoints.Handler, eventHandler *events.Handler) *Application {
	return &Application{
		endpointHandler: endpointHandler,
		eventHandler:    eventHandler,
		address:         address,
		router:          nil,
	}
}

// MountRoutes configures the application's HTTP routes and middleware.
func (app *Application) MountRoutes() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Endpoint routes
	r.Route("/endpoints", func(r chi.Router) {
		r.Get("/", app.endpointHandler.GetEndpoints)
		r.Post("/", app.endpointHandler.CreateEndpoint)
		r.Delete("/", app.endpointHandler.DeleteEndpoint)
	})

	// Event routes
	r.Route("/event", func(r chi.Router) {
		r.Post("/", app.eventHandler.SendEvent)
	})

	app.router = r
}

// ListenAndServe starts the HTTP server on the application's configured address.
func (app *Application) ListenAndServe() {
	if app.router == nil {
		slog.Warn("Attempted to call ListenAndServe for application when no router was created.")
		return
	}

	listener, err := net.Listen("tcp", app.address)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Server listening at http://localhost%s\n", app.address)

	if err := http.Serve(listener, app.router); err != nil {
		log.Fatal(err)
	}

	log.Println("Server shutdown.")
}
