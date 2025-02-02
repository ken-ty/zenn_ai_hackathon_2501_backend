// Sample run-helloworld is a minimal Cloud Run service.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ken-ty/zenn_ai_hackathon_2501_backend/service"
)

func main() {
	log.Print("starting server...")
	// ヘルスチェック
	http.HandleFunc("/health", service.HealthCheckHandler)

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	// Start HTTP server.
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
