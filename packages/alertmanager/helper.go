package alertmanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"main/packages/models"
	"net/http"
	"regexp"
)

func FetchFiringAlerts() ([]models.AlertPrometheus, error) {
	resp, err := http.Get("http://localhost:9090/api/v1/alerts")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			Alerts []models.AlertPrometheus `json:"alerts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var firingAlerts []models.AlertPrometheus
	for _, alert := range result.Data.Alerts {
		if alert.State == "firing" {
			firingAlerts = append(firingAlerts, alert)
		}
	}
	return firingAlerts, nil
}

func FetchSilencedAlerts() ([]models.Silence, error) {
	resp, err := http.Get("http://localhost:9093/api/v1/silences")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []models.Silence `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func IsSilenced(alert models.AlertPrometheus, silences []models.Silence) bool {
	for _, silence := range silences {
		if silence.Status.State == "active" {
			matched := true
			for _, matcher := range silence.Matchers {
				val, ok := alert.Labels[matcher.Name]
				if matcher.IsEqual {
					if !ok || val != matcher.Value {
						// fmt.Println("alert: ", alert.Labels["alertname"], "for label: ", matcher.Name, "values: ", val, "checked: ", matcher.Value, "ok: ", ok)
						matched = false
						break
					}
				} else if matcher.IsRegex {
					re, err := regexp.Compile(matcher.Value)
					if err != nil || !re.MatchString(val) {
						matched = false
						break
					}
				}
			}
			if matched {
				return true
			}
		}
	}
	return false
}

func DeleteSilence(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://localhost:9093/api/v1/silence/%s", id), nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	return nil
}

func CreateSilence(silence models.Silence) error {
	// Marshal the silence struct into JSON
	body, err := json.Marshal(silence)
	if err != nil {
		return err
	}

	// Print the JSON body for debugging
	fmt.Println("Request Body:", string(body))

	// Create an HTTP POST request with the JSON body
	req, err := http.NewRequest("POST", "http://localhost:9093/api/v1/silences", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	// Set content type to JSON
	req.Header.Set("Content-Type", "application/json")

	// Use the default HTTP client to send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	// Check if the response status code indicates success
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("failed to create silence, status code: %d, response: %s", resp.StatusCode, respBody)
	}

	return nil
}

// processAlert handles individual alert and processes it.
func ProcessAlert(alert models.Alert) {
	// Save alert to Apache Doris
	if err := dorisClient.SaveAlert(alert); err != nil {
		log.Printf("Failed to save alert to Doris: %v\n %s", err, alert.Labels["alertname"])
		return
	}

	log.Printf("Completed processing alert: %s", alert.Labels["alertname"])
}
