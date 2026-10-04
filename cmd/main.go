package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Opxourc/event-proxy/cmd/api"
	"github.com/Opxourc/event-proxy/internal/repository"
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
	db, err := sql.Open("pgx", url)
	if err != nil {
		log.Fatal(err)
	}

	app := api.New(":8080", repository.New(db))
	app.MountRoutes()
	app.ListenAndServe()
}
