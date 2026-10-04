package api

import (
	"log"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/Opxourc/event-proxy/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Application struct {
	repo    *repository.Repository
	address string
	router  *chi.Mux
}

// New constructs a new Application object.
func New(address string, repo *repository.Repository) *Application {
	return &Application{
		repo:    repo,
		address: address,
		router:  nil,
	}
}

// MountRoutes sets up the route paths that the API will listen to.
func (app *Application) MountRoutes() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Endpoint routes
	r.Route("/endpoints", func(r chi.Router) {

	})

	// Event routes
	r.Route("/event", func(r chi.Router) {

	})

	app.router = r
}

// ListenAndServe will start up the API and begin listening to requests that are coming in.
func (app *Application) ListenAndServe() {
	if app.router == nil {
		slog.Warn("Attempted to call ListenAndServe for application when no router was created.")
		return
	}

	listener, err := net.Listen("tcp", app.address)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("server listening at http://localhost%s\n", app.address)

	if err := http.Serve(listener, app.router); err != nil {
		log.Fatal(err)
	}
}
