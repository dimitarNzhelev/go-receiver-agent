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

// AlertRule represents a single alert rule.
type AlertRule struct {
	State          string            `json:"state"`
	Name           string            `json:"name"`
	Query          string            `json:"query"`
	Duration       int               `json:"duration"`
	Labels         map[string]string `json:"labels"`
	Health         string            `json:"health"`
	EvaluationTime float64           `json:"evaluationTime"`
	LastEvaluation time.Time         `json:"lastEvaluation"`
	Type           string            `json:"type"`
}

// AlertRuleGroup represents a group of alert rules.
type AlertRuleGroup struct {
	Name  string      `json:"name"`
	File  string      `json:"file"`
	Rules []AlertRule `json:"rules"`
}
