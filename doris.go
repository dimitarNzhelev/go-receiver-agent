package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type DorisClient struct {
	db *sql.DB
}

// FOR TESTING
// host := "192.168.1.111"
// port := "9030"
// user := "testUser"
// password := "testPass"
// database := "test_database"

// NewDorisClient creates a new Apache Doris client
func NewDorisClient() (*DorisClient, error) {

	host := os.Getenv("DORIS_HOST")
	port := os.Getenv("DORIS_PORT")
	user := os.Getenv("DORIS_USER")
	password := os.Getenv("DORIS_PASSWORD")
	database := os.Getenv("DORIS_DATABASE")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?timeout=5s&readTimeout=5s&writeTimeout=5s&tls=false&allowNativePasswords=true",
		user, password, host, port, database)
	fmt.Printf("Attempting to connect with DSN: %s\n", dsn)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %v", err)
	}

	// Disable connection pooling
	db.SetMaxIdleConns(0)
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(time.Second * 10)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping Doris: %v", err)
	}

	return &DorisClient{db: db}, nil
}

// Close closes the database connection
func (c *DorisClient) Close() error {
	return c.db.Close()
}

// SaveAlert saves an alert to Apache Doris
func (c *DorisClient) SaveAlert(alert Alert) error {
	// Convert maps to JSON strings
	labelsStr, err := json.Marshal(alert.Labels)
	if err != nil {
		return fmt.Errorf("failed to marshal labels: %v", err)
	}

	annotationsStr, err := json.Marshal(alert.Annotations)
	if err != nil {
		return fmt.Errorf("failed to marshal annotations: %v", err)
	}

	query := fmt.Sprintf(`
        INSERT INTO alerts (
            status,
            alert_name,
            start_time,
            end_time,
            generator_url,
            fingerprint,
            labels,
            annotations
        ) VALUES ('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s')
    `,
		alert.Status,
		alert.Labels["alertname"],
		alert.StartsAt,
		alert.EndsAt,
		alert.GeneratorURL,
		alert.Fingerprint,
		string(labelsStr),
		string(annotationsStr),
	)

	log.Printf("Executing query: %s", query)

	_, err = c.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to save alert: %v", err)
	}

	return nil
}
