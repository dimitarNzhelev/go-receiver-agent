package utils

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type PromQLValidationResponse struct {
	Valid  bool   `json:"valid"`
	Error  string `json:"error,omitempty"`
	Query  string `json:"query"`
}

func PromQLValidationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body
	var requestBody struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if requestBody.Query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	// Get Prometheus URL from environment
	prometheusUrl := config.GetEnv("PROMETHEUS_URL", "http://localhost:9090")
	
	// Create the validation URL
	validateURL, err := url.Parse(prometheusURL + "/api/v1/query")
	if err != nil {
		http.Error(w, "Invalid Prometheus URL", http.StatusInternalServerError)
		return
	}

	// Add query parameters
	params := url.Values{}
	params.Add("query", requestBody.Query)
	validateURL.RawQuery = params.Encode()

	// Make the request to Prometheus
	resp, err := http.Get(validateURL.String())
	if err != nil {
		http.Error(w, "Failed to validate query", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Parse the response
	var prometheusResponse struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
		} `json:"data"`
		ErrorType string `json:"errorType,omitempty"`
		Error     string `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&prometheusResponse); err != nil {
		http.Error(w, "Failed to parse Prometheus response", http.StatusInternalServerError)
		return
	}

	response := PromQLValidationResponse{
		Query: requestBody.Query,
		Valid: prometheusResponse.Status == "success",
	}

	if !response.Valid {
		response.Error = prometheusResponse.Error
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
} 