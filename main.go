package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type Response struct {
	Status string `json:"status"`
}

func writeJSONResponse(w http.ResponseWriter, r *http.Request, status int, payload interface{}) {
	// Set CORS headers for pre-flight requests
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSONResponse(w, r, http.StatusOK, Response{Status: "ok"})
}

func getSecureHandler() (func(w http.ResponseWriter, r *http.Request), error) {
	authToken, exists := os.LookupEnv("AUTH_TOKEN")
	if !exists {
		return nil, fmt.Errorf("AUTH_TOKEN env variable not found")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != authToken {
			writeJSONResponse(w, r, http.StatusUnauthorized, Response{Status: "unauthorized"})
			return
		}
		writeJSONResponse(w, r, http.StatusOK, Response{Status: "secured"})
	}, nil
}

func setupRouter() (*http.ServeMux, error) {
	router := http.NewServeMux()

	secureHandler, err := getSecureHandler()
	if err != nil {
		return nil, err
	}

	router.HandleFunc("/health", healthHandler)
	router.HandleFunc("/secure", secureHandler)

	return router, nil
}

func main() {
	// Only load .env if it exists
	if _, err := os.Stat(".env"); err == nil {
		log.Printf("Loading .env")
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
