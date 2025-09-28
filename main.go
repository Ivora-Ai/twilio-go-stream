package main

import (
	"log"
	"net/http"
	"os"
	"twilio-go-stream/handler"

	"github.com/joho/godotenv"
)

// loadEnv loads environment variables from .env file
func loadEnv() {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	// Load environment variables
	loadEnv()

	// Get configuration from environment variables
	publicURL := getEnv("PUBLIC_URL", "twilio-go-stream-804663264218.us-central1.run.app")
	sttProvider := getEnv("STT_PROVIDER", "deepgram")
	ttsProvider := getEnv("TTS_PROVIDER", "deepgram")

	// Log the providers being used
	log.Printf("Using STT provider: %s", sttProvider)
	log.Printf("Using TTS provider: %s", ttsProvider)

	// Initialize handler
	handlers := handler.New(publicURL, sttProvider, ttsProvider)
	handlers.SetRoutes()

	// Start HTTP server
	port := getEnv("PORT", "8080")
	log.Printf("Server started on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
