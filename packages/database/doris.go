package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"main/packages/models"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type DorisClient struct {
	db *sql.DB
}

// FOR TESTING

// NewDorisClient creates a new Apache Doris client
func NewDorisClient() (*DorisClient, error) {

	// host := config.GetEnv("DORIS_HOST", "localhost")
	// port := config.GetEnv("DORIS_PORT", "9030")
	// user := config.GetEnv("DORIS_USER", "root")
	// password := config.GetEnv("DORIS_PASSWORD", "root")
	// database := config.GetEnv("DORIS_DATABASE", "test_database")
	host := "192.168.1.111"
	port := "9030"
	user := "dzhelev"
	password := "dzhelev@123"
	database := "dzhelev_db"

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

// SaveAlert saves an alert to Apache Doris (also must check if its working)
func (c *DorisClient) SaveAlert(alert models.Alert) error {
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

// MUST CHECK if its working
func (c *DorisClient) CreateTableIfNotExists() error {
	query := `
        CREATE TABLE IF NOT EXISTS alerts (
            fingerprint VARCHAR(255) NOT NULL,
            id BIGINT NOT NULL,
            status VARCHAR(255) NOT NULL,
            alert_name VARCHAR(255) NOT NULL,
            start_time DATETIME NOT NULL,
            end_time DATETIME,
            generator_url VARCHAR(1024),
            labels STRING,
            annotations STRING
        )
        UNIQUE KEY(fingerprint, id)
        DISTRIBUTED BY HASH(fingerprint) BUCKETS 10
        PROPERTIES (
            "replication_num" = "1"
        );
    `

	_, err := c.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	return nil
}

// also must check if its working
func (c *DorisClient) GetAlerts() ([]models.AlertResponse, error) {
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
		if len(labels) > 0 {
			if err := json.Unmarshal(labels, &alert.Labels); err != nil {
				return nil, fmt.Errorf("failed to parse labels JSON: %v", err)
			}
		}
		if len(annotations) > 0 {
			if err := json.Unmarshal(annotations, &alert.Annotations); err != nil {
				return nil, fmt.Errorf("failed to parse annotations JSON: %v", err)
			}
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}
