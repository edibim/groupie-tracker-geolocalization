package main

import (
	"log"
	"net/http"

	"groupie-tracker-geolocalization/handlers"
)

// main starts the web server.
// This is the entry point used when running go run .
func main() {
	server := newServer(":8080")

	log.Println("Server running at http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// newServer builds the HTTP server with all routes attached.
func newServer(addr string) *http.Server {
	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
