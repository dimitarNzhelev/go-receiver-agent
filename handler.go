package main

import (
	"encoding/json"
	"log"
	"net/http"
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

var redshiftClient *RedshiftClient

// NB! this init function is called when the package is loaded
func init() {
	log.Println("Initializing Redshift client...")

	client, err := NewRedshiftClient()
	if err != nil {
		log.Fatalf("Failed to initialize Redshift client: %v", err)
	}
	redshiftClient = client
	log.Println("Redshift client initialized successfully")

	err_ := redshiftClient.createTableIfNotExists()

	if err_ != nil {
		log.Fatalf("Failed to create table: %v", err_)
	}

}

// processAlert handles individual alert and proccesses it.
func processAlert(alert Alert) {
	log.Printf("Processing alert: %s", alert.Labels["alertname"])

	// Save alert to Apache Doris
	if err := redshiftClient.SaveAlert(alert); err != nil {
		log.Printf("Failed to save alert to Doris: %v", err)
		return
	}

	log.Printf("Completed processing alert: %s", alert.Labels["alertname"])
}
