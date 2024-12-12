package alertmanager

import (
	"encoding/json"
	"fmt"
	"log"
	"main/packages/database"
	"main/packages/models"
	"net/http"
)

var dorisClient *database.DorisClient

func init() {
	log.Println("Initializing Doris client...")

	client, err := database.NewDorisClient()
	if err != nil {
		log.Fatalf("Failed to initialize Doris client: %v", err)
	}

	log.Println("Doris client initialized successfully")

	err = client.CreateTableIfNotExists()
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	dorisClient = client
	log.Println("Doris client initialized successfully")
}

// AlertPOSTHandler processes incoming alert requests.
func AlertPOSTHandler(w http.ResponseWriter, r *http.Request) {
	var payload models.AlertmanagerPayload

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

// AlertGETHandler returns all alerts from Apache Doris
func AlertGETHandler(w http.ResponseWriter, r *http.Request) {
	alerts, err := dorisClient.GetAlerts()
	if err != nil {
		http.Error(w, "Failed to retrieve alerts", http.StatusInternalServerError)
		log.Printf("Failed to retrieve alerts: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(alerts); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}

// processAlert handles individual alert and processes it.
func processAlert(alert models.Alert) {
	log.Printf("Processing alert: %s", alert.Labels["alertname"])
	// Save alert to Apache Doris
	if err := dorisClient.SaveAlert(alert); err != nil {
		log.Printf("Failed to save alert to Doris: %v", err)
		return
	}

	log.Printf("Completed processing alert: %s", alert.Labels["alertname"])
}

func AlertRulesGETHandler(w http.ResponseWriter, r *http.Request) {
	alertRules, err := getAlertRulesFromAlertmanager()
	if err != nil {
		http.Error(w, "Failed to retrieve alert rules", http.StatusInternalServerError)
		log.Printf("Failed to retrieve alert rules: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(alertRules); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}

// getAlertRulesFromAlertmanager fetches alert rules from Alertmanager
func getAlertRulesFromAlertmanager() ([]models.AlertRuleGroup, error) {
	alertmanagerURL := "http://localhost:9090/api/v1/rules"

	resp, err := http.Get(alertmanagerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rules from Alertmanager: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Define the structure matching the API response
	var result struct {
		Status string `json:"status"`
		Data   struct {
			Groups []models.AlertRuleGroup `json:"groups"`
		} `json:"data"`
	}

	// Decode the response body
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %v", err)
	}

	return result.Data.Groups, nil
}
