package api

import (
	"log"
	"log/slog"
	"net"
	"net/http"
	"time"

	httpadapter "github.com/Opxourc/event-proxy/internal/adapters/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Application is the representation of a app instance with it's own handlers, address, and router.
type Application struct {
	endpointHandler *httpadapter.EndpointHandler
	eventHandler    *httpadapter.EventHandler
	address         string
	router          *chi.Mux
}

// New constructs a new Application object.
func New(address string, endpointHandler *httpadapter.EndpointHandler, eventHandler *httpadapter.EventHandler) *Application {
	return &Application{
		endpointHandler: endpointHandler,
		eventHandler:    eventHandler,
		address:         address,
		router:          nil,
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

	log.Printf("Server listening at http://localhost%s\n", app.address)

	if err := http.Serve(listener, app.router); err != nil {
		log.Fatal(err)
	}

	log.Println("Server shutdown.")
}
