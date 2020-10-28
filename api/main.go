package main

import (
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Output(2, "No PORT environment variable found, using 8000.")
		port = ":8000"
	}

	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	db := NewPostgreSQL()
	defer db.Close()

	app := API{&db}
	http.HandleFunc("/", app.PostHandler)
	log.Fatal(http.ListenAndServe(port, nil))
}
