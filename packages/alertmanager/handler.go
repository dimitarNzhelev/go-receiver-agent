package alertmanager

import (
	"encoding/json"
	"fmt"
	"log"
	"main/packages/database"
	"main/packages/models"
	"net/http"
	"strings"
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
		http.Error(w, ErrorInvalidJSONPayload.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Process each alert asynchronously
	for _, alert := range payload.Alerts {
		go ProcessAlert(alert)
	}

	// Respond to Alertmanager
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "success"}); err != nil {
		http.Error(w, ErrorInvalidJSONPayload.Error(), http.StatusUnprocessableEntity)
	}
}

// AlertGETHandler returns all alerts from Apache Doris
func AlertGETHandler(w http.ResponseWriter, r *http.Request) {
	alerts, err := dorisClient.GetAlerts()
	if err != nil {
		log.Printf("Failed to retrieve alerts: %v", err)
		http.Error(w, ErrorFailedToRetrieve.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(alerts); err != nil {
		log.Printf("JSON encoding error: %v", err)
		http.Error(w, ErrorJSONEncoding.Error(), http.StatusInternalServerError)
	}
}

func AlertFiringGETHandler(w http.ResponseWriter, r *http.Request) {
	firingAlerts, err := FetchFiringAlerts()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching firing alerts: %v", err), http.StatusInternalServerError)
		return
	}

	silences, err := FetchSilencedAlerts()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching silences: %v", err), http.StatusInternalServerError)
		return
	}

	var unsilencedAlerts []models.AlertPrometheus
	for _, alert := range firingAlerts {
		if !IsSilenced(alert, silences) {
			unsilencedAlerts = append(unsilencedAlerts, alert)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(unsilencedAlerts)
}

func SilencesGETHandler(w http.ResponseWriter, r *http.Request) {
	silences, err := FetchSilencedAlerts()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching silences: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(silences)
}

func AlertSilencesGETHandler(w http.ResponseWriter, r *http.Request) {
	firingAlerts, err := FetchFiringAlerts()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching firing alerts: %v", err), http.StatusInternalServerError)
		return
	}

	silences, err := FetchSilencedAlerts()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching silences: %v", err), http.StatusInternalServerError)
		return
	}

	var silencedAlerts []models.AlertPrometheus
	for _, alert := range firingAlerts {
		if IsSilenced(alert, silences) {
			silencedAlerts = append(silencedAlerts, alert)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(silencedAlerts)
}

func AlertSilencesPOSTHandler(w http.ResponseWriter, r *http.Request) {
	var silence models.Silence
	err := json.NewDecoder(r.Body).Decode(&silence)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error decoding silence: %v", err), http.StatusBadRequest)
		return
	}

	err = CreateSilence(silence)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating silence: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func AlertSilencesDELETEHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	id := pathParts[len(pathParts)-1]

	if id == "" {
		http.Error(w, ErrorSilenceIDNotFound.Error(), http.StatusBadRequest)
		return
	}

	err := DeleteSilence(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting silence: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
