package models

import "time"

// Alert represents a single alert.
type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

type Silence struct {
	ID        string    `json:"id"`
	Matchers  []Matcher `json:"matchers"`
	Status    Status    `json:"status"`
	StartsAt  string    `json:"startsAt"`
	EndsAt    string    `json:"endsAt"`
	UpdatedAt string    `json:"updatedAt"`
	CreatedBy string    `json:"createdBy"`
	Comment   string    `json:"comment"`
}
type Status struct {
	State string `json:"state"`
}

type Matcher struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	IsRegex bool   `json:"isRegex"`
	IsEqual bool   `json:"isEqual"`
}

type AlertPrometheus struct {
	State       string            `json:"state"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	ActiveAt    time.Time         `json:"activeAt"`
	Value       string            `json:"value"`
}

type AlertResponse struct {
	Name         string            `json:"alert_name"`
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  string            `json:"annotations"`
	StartsAt     time.Time         `json:"start_time"`
	EndsAt       time.Time         `json:"end_time"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

// AlertmanagerPayload represents the payload sent by Alertmanager.
type AlertmanagerPayload struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"`
	Alerts            []Alert           `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	TruncatedAlerts   int               `json:"truncatedAlerts,omitempty"`
}
