package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Opxourc/event-proxy/internal/database"
	"github.com/Opxourc/event-proxy/internal/endpoints"
	"github.com/Opxourc/event-proxy/internal/events"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	fmt.Println("PostgreSQL connection URL:")
	fmt.Println("Use a local/development database (e.g. localhost), not a production database.")
	fmt.Print("> ")

	url, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	url = strings.TrimSpace(url)

	fmt.Println("Connecting to PostgreSQL...")
	db, err := database.Open("pgx", url)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create repositories, services, and handlers
	// These components will then communicate with one another

	endpointRepository := endpoints.NewRepository(db)
	eventRepository := events.NewRepository(db)

	endpointService := endpoints.NewService(endpointRepository, endpoints.NewHTTPChecker())
	eventService := events.NewService(eventRepository, events.NewHTTPSender())

	endpointHandler := endpoints.NewHandler(endpointService)
	eventHandler := events.NewHandler(eventService)

	app := NewApplication(":8080", endpointHandler, eventHandler)
	app.MountRoutes()
	app.ListenAndServe()
}
