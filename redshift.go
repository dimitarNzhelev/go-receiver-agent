package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type RedshiftClient struct {
	db *sql.DB
}

// NewRedshiftClient creates a new AWS Redshift client
func NewRedshiftClient() (*RedshiftClient, error) {
	// host := "dzhelev-test-cluster-redshift.cs6ir1smkzux.eu-central-1.redshift.amazonaws.com"
	// port := "5439"
	// user := "adminuser"
	// password := "StrongPassword123"
	// database := "alertsdb"
	// sslmode := "disable"

	host := os.Getenv("REDSHIFT_HOST")
	port := os.Getenv("REDSHIFT_PORT")
	user := os.Getenv("REDSHIFT_USER")
	password := os.Getenv("REDSHIFT_PASSWORD")
	database := os.Getenv("REDSHIFT_DATABASE")
	sslmode := os.Getenv("REDSHIFT_SSLMODE")

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
func (c *RedshiftClient) SaveAlert(alert Alert) error {
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
		return fmt.Errorf("failed to save alert: %v", err)
	}

	return nil
}

func (c *RedshiftClient) createTableIfNotExists() error {
	query := `
        CREATE TABLE IF NOT EXISTS alerts (
			id BIGINT IDENTITY(1,1) PRIMARY KEY,
			status VARCHAR(50) NOT NULL,
			alert_name VARCHAR(255) NOT NULL,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP,
			generator_url VARCHAR(512),
			fingerprint VARCHAR(255) UNIQUE,
			labels VARCHAR(255),
			annotations VARCHAR(255)
		);
    `

	_, err := c.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	return nil
}
