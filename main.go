package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Response represents the JSON response structure
type Response struct {
	Status string `json:"status"`
}

// writeJSONResponse is a helper function to write JSON responses
func writeJSONResponse(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// healthHandler handles the /health route
func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSONResponse(w, http.StatusOK, Response{Status: "ok"})
}

func getSecureHandler() (func(w http.ResponseWriter, r *http.Request), error) {
	authToken, exists := os.LookupEnv("AUTH_TOKEN")
	if !exists {
		return nil, fmt.Errorf("AUTH_TOKEN env variable not found")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != authToken {
			writeJSONResponse(w, http.StatusUnauthorized, Response{Status: "unauthorized"})
			return
		}
		writeJSONResponse(w, http.StatusOK, Response{Status: "secured"})
	}, nil
}

// setupRouter sets up the routes and returns the router
func setupRouter() (*http.ServeMux, error) {
	router := http.NewServeMux()

	secureHandler, err := getSecureHandler()
	if err != nil {
		return nil, err
	}

	// Register routes
	router.HandleFunc("/health", healthHandler)
	router.HandleFunc("/secure", secureHandler)

	return router, nil
}

func main() {
	if isLocal, err := strconv.ParseBool(os.Getenv("IS_LOCAL")); !isLocal || err != nil {
		if err := godotenv.Load(); err != nil {
			log.Fatalf("Error loading .env: %v", err)
		}
	}

	router, err := setupRouter()
	if err != nil {
		log.Fatal(err)
	}

	// Start the server
	port := ":8888"
	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(port, router))
}
