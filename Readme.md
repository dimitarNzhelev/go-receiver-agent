# Kubernetes Monitoring and Alert Management Agent

A Go-based agent that provides a unified interface for managing Kubernetes resources, Prometheus alerts, and alert storage using Apache Doris. The agent implements the Command pattern for handling different operations and includes security measures through token-based authentication.

## Features

- **Kubernetes Resource Management**
  - List pods, nodes, namespaces, and deployments
  - View, create, update, and delete PrometheusRule resources
  - Real-time interaction with the Kubernetes cluster

- **Alert Management**
  - Receive and process alerts from Alertmanager
  - Store alerts in Apache Doris for persistence
  - Retrieve historical alerts

- **Security**
  - Token-based authentication for all endpoints
  - Secure communication with Kubernetes cluster
  - Environment variable configuration for sensitive data

## Prerequisites

- Go 1.21 or higher
- Access to a Kubernetes cluster
- Apache Doris instance
- Prometheus and Alertmanager setup

## Configuration

The agent uses environment variables for configuration:

```env
# Server Configuration
PORT=5000
AUTH_TOKEN=your_secret_token

# Apache Doris Configuration
DORIS_HOST=your_doris_host
DORIS_PORT=9030
DORIS_USER=your_username
DORIS_PASSWORD=your_password
DORIS_DATABASE=your_database
```

## API Endpoints
### Kubernetes Resources
GET /pods - List all pods

GET /nodes - List all nodes

GET /namespaces - List all namespaces

GET /deployments - List all deployments

### PrometheusRules Management
GET /rules - List all PrometheusRules

POST /rules - Create a new PrometheusRule

PUT /rules/{id} - Update an existing PrometheusRule

DELETE /rules/{id} - Delete a PrometheusRule

### Alert Management
POST /alerts - Receive alerts from Alertmanager

GET /alerts - Retrieve stored alerts

## Architecture
- The agent follows the Command pattern for handling different operations:

- Each endpoint handler acts as a command

- Middleware for logging and authentication

- Asynchronous alert processing
 
- Connection pooling for database operations

## Error Handling
The agent implements comprehensive error handling:

- Database connection errors
 
- Invalid JSON payloads
 
- Authentication failures
 
- Kubernetes API errors