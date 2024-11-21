package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"main/packages/config"
	"main/packages/models"

	_ "github.com/lib/pq"
)

type RedshiftClient struct {
	db *sql.DB
}

// NewRedshiftClient creates a new AWS Redshift client
func NewRedshiftClient() (*RedshiftClient, error) {

	host := config.GetEnv("REDSHIFT_HOST", "")
	port := config.GetEnv("REDSHIFT_PORT", "")
	user := config.GetEnv("REDSHIFT_USER", "")
	password := config.GetEnv("REDSHIFT_PASSWORD", "")
	database := config.GetEnv("REDSHIFT_DATABASE", "")
	sslmode := config.GetEnv("REDSHIFT_SSLMODE", "")

	if host == "" || port == "" || user == "" || password == "" || database == "" || sslmode == "" {
		return nil, fmt.Errorf("missing required Redshift configuration")
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, database, sslmode)
	fmt.Printf("Attempting to connect with DSN: %s\n", dsn)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %v", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping Redshift: %v", err)
	}

	return &RedshiftClient{db: db}, nil
}

// Close closes the database connection
func (c *RedshiftClient) Close() error {
	return c.db.Close()
}

// SaveAlert saves an alert to AWS Redshift
func (c *RedshiftClient) SaveAlert(alert models.Alert) error {
	// Convert maps to JSON strings
	labelsStr, err := json.Marshal(alert.Labels)
	if err != nil {
		return fmt.Errorf("failed to marshal labels: %v", err)
	}

	annotationsStr, err := json.Marshal(alert.Annotations)
	if err != nil {
		return fmt.Errorf("failed to marshal annotations: %v", err)
	}

	query := `
        INSERT INTO alerts (
            status,
            alert_name,
            start_time,
            end_time,
            generator_url,
            fingerprint,
            labels,
            annotations
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

	log.Printf("Executing query: %s", query)

	_, err = c.db.Exec(query,
		alert.Status,
		alert.Labels["alertname"],
		alert.StartsAt,
		alert.EndsAt,
		alert.GeneratorURL,
		alert.Fingerprint,
		string(labelsStr),
		string(annotationsStr),
	)
	if err != nil {
		if err.Error() == "pq: Value too long for character type" {
			log.Printf("Value too long for character type. Alert details: Status=%s, AlertName=%s, StartsAt=%s, EndsAt=%s, GeneratorURL=%s, Fingerprint=%s, Labels=%s, Annotations=%s",
				alert.Status,
				alert.Labels["alertname"],
				alert.StartsAt,
				alert.EndsAt,
				alert.GeneratorURL,
				alert.Fingerprint,
				string(labelsStr),
				string(annotationsStr))
		}
		return fmt.Errorf("failed to save alert: %v", err)
	}

	return nil
}

func (c *RedshiftClient) CreateTableIfNotExists() error {
	query := `
        CREATE TABLE IF NOT EXISTS alerts (
			id BIGINT IDENTITY(1,1) PRIMARY KEY,
			status VARCHAR(255) NOT NULL,
			alert_name VARCHAR(255) NOT NULL,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP,
			generator_url VARCHAR(1024),
			fingerprint VARCHAR(255) UNIQUE,
			labels VARCHAR(4096),
			annotations VARCHAR(4096)
		);
    `

	_, err := c.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	return nil
}

// GetAlerts returns all alerts from AWS Redshift
func (c *RedshiftClient) GetAlerts() ([]models.AlertResponse, error) {
	query := `SELECT id, alert_name, status, labels, annotations, start_time, end_time, generator_url, fingerprint FROM alerts`
	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve alerts: %v", err)
	}
	defer rows.Close()

	var alerts []models.AlertResponse
	for rows.Next() {
		var alert models.AlertResponse
		var labels, annotations []byte
		err := rows.Scan(&alert.Id, &alert.Name, &alert.Status, &labels, &annotations, &alert.StartsAt, &alert.EndsAt, &alert.GeneratorURL, &alert.Fingerprint)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert row: %v", err)
		}

		// Unmarshal JSON fields into maps
		if err := json.Unmarshal(labels, &alert.Labels); err != nil {
			return nil, fmt.Errorf("failed to parse labels JSON: %v", err)
		}
		if err := json.Unmarshal(annotations, &alert.Annotations); err != nil {
			return nil, fmt.Errorf("failed to parse annotations JSON: %v", err)
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}
