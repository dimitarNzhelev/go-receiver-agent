package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-sql-driver/mysql"
)

// AlertHandler processes incoming alert requests.
func AlertHandler(w http.ResponseWriter, r *http.Request) {
	var payload AlertmanagerPayload

	// Decode the JSON payload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		log.Printf("JSON decoding error: %v", err)
		return
	}
	defer r.Body.Close()

	// Process each alert asynchronously
	for _, alert := range payload.Alerts {
		go processAlert(alert)
	}

	// Respond to Alertmanager
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "success"}); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}

var dorisClient *DorisClient

// NB! this init function is called when the package is loaded
func init() {
	log.Println("Initializing Doris client...")
	mysql.SetLogger(log.New(os.Stdout, "[mysql] ", log.Ldate|log.Ltime|log.Lshortfile))

	client, err := NewDorisClient()
	if err != nil {
		log.Fatalf("Failed to initialize Doris client: %v", err)
	}
	dorisClient = client
	log.Println("Doris client initialized successfully")
}

// processAlert handles individual alert and proccesses it.
func processAlert(alert Alert) {
	log.Printf("Processing alert: %s", alert.Labels["alertname"])

	// Save alert to Apache Doris
	if err := dorisClient.SaveAlert(alert); err != nil {
		log.Printf("Failed to save alert to Doris: %v", err)
		return
	}

	log.Printf("Completed processing alert: %s", alert.Labels["alertname"])
}
