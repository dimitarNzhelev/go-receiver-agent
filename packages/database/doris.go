package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"main/packages/config"
	"main/packages/models"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type DorisClient struct {
	db *sql.DB
}

// FOR TESTING
// host := "192.168.1.111"
// port := "9030"
// user := "dzhelev"
// password := "dzhelev@123"
// database := "dzhelev_db"
// NewDorisClient creates a new Apache Doris client
func NewDorisClient() (*DorisClient, error) {

	host := config.GetEnv("DORIS_HOST", "localhost")
	port := config.GetEnv("DORIS_PORT", "9030")
	user := config.GetEnv("DORIS_USER", "dzhelev")
	password := config.GetEnv("DORIS_PASSWORD", "dzhelev@123")
	database := config.GetEnv("DORIS_DATABASE", "dzhelev_db")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?timeout=5s&readTimeout=5s&writeTimeout=5s&tls=false&allowNativePasswords=true",
		user, password, host, port, database)

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

	// Check if an alert with the same fingerprint exists
	checkQuery := fmt.Sprintf("SELECT COUNT(*) FROM alerts WHERE fingerprint = '%s'", alert.Fingerprint)
	var count int
	err = c.db.QueryRow(checkQuery).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for existing alert: %v", err)
	}

	if count > 0 {
		// Update existing alert
		updateQuery := fmt.Sprintf(`
			UPDATE alerts
			SET
				status = '%s',
				alert_name = '%s',
				start_time = '%s',
				end_time = '%s',
				generator_url = '%s',
				labels = '%s',
				annotations = '%s'
			WHERE fingerprint = '%s'
		`,
			alert.Status,
			alert.Labels["alertname"],
			alert.StartsAt.Format("2006-01-02 15:04:05"),
			alert.EndsAt.Format("2006-01-02 15:04:05"),
			alert.GeneratorURL,
			string(labelsStr),
			string(annotationsStr),
			alert.Fingerprint,
		)

		log.Printf("Executing update query: %s", updateQuery)
		_, err = c.db.Exec(updateQuery)
		if err != nil {
			return fmt.Errorf("failed to update alert: %v", err)
		}
		log.Printf("Updated alert with fingerprint: %s", alert.Fingerprint)
	} else {
		// Insert new alert
		insertQuery := fmt.Sprintf(`
			INSERT INTO alerts (
				fingerprint,
				status,
				alert_name,
				start_time,
				end_time,
				generator_url,
				labels,
				annotations
			) VALUES ('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s')
		`,
			alert.Fingerprint,
			alert.Status,
			alert.Labels["alertname"],
			alert.StartsAt.Format("2006-01-02 15:04:05"),
			alert.EndsAt.Format("2006-01-02 15:04:05"),
			alert.GeneratorURL,
			string(labelsStr),
			string(annotationsStr),
		)

		log.Printf("Executing insert query: %s", insertQuery)
		_, err = c.db.Exec(insertQuery)
		if err != nil {
			return fmt.Errorf("failed to insert alert: %v", err)
		}
		log.Printf("Inserted new alert with fingerprint: %s", alert.Fingerprint)
	}

	return nil
}

func (c *DorisClient) CreateTableIfNotExists() error {
	query := `
		CREATE TABLE IF NOT EXISTS alerts (
			fingerprint CHAR(255) NOT NULL,
			status STRING NOT NULL,
			alert_name STRING NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			generator_url STRING,
			labels STRING,
			annotations STRING
		)
		UNIQUE KEY (fingerprint)
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

func (c *DorisClient) GetAlerts() ([]models.AlertResponse, error) {
	query := `
		SELECT fingerprint, status, alert_name, start_time, end_time, generator_url, labels, annotations
		FROM alerts
	`
	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve alerts: %v", err)
	}
	defer rows.Close()

	var alerts []models.AlertResponse
	for rows.Next() {
		var alert models.AlertResponse
		var labels, annotations []byte
		var startTime, endTime string

		err := rows.Scan(
			&alert.Fingerprint,
			&alert.Status,
			&alert.Name,
			&startTime,
			&endTime,
			&alert.GeneratorURL,
			&labels,
			&annotations,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert row: %v", err)
		}

		// Parse start and end times
		if alert.StartsAt, err = time.Parse("2006-01-02 15:04:05", startTime); err != nil {
			return nil, fmt.Errorf("failed to parse start_time: %v", err)
		}
		if endTime != "" {
			if alert.EndsAt, err = time.Parse("2006-01-02 15:04:05", endTime); err != nil {
				return nil, fmt.Errorf("failed to parse end_time: %v", err)
			}
		}

		// Unmarshal labels JSON into a map
		if len(labels) > 0 {
			if err := json.Unmarshal(labels, &alert.Labels); err != nil {
				return nil, fmt.Errorf("failed to parse labels JSON: %v", err)
			}
		}

		alert.Annotations = string(annotations)

		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %v", err)
	}

	return alerts, nil
}
