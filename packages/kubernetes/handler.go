package kubernetes

import (
	"encoding/json"
	"log"
	"net/http"
)

// PodsGETHandler returns a list of pods
func PodsGETHandler(w http.ResponseWriter, r *http.Request) {
	// Implement logic to get pods
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode([]string{"pod1", "pod2"}); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}

// NodesGETHandler returns a list of nodes
func NodesGETHandler(w http.ResponseWriter, r *http.Request) {
	// Implement logic to get nodes
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode([]string{"node1", "node2"}); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}

// NamespacesGETHandler returns a list of namespaces
func NamespacesGETHandler(w http.ResponseWriter, r *http.Request) {
	// Implement logic to get namespaces
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode([]string{"namespace1", "namespace2"}); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}

// DeploymentsGETHandler returns a list of deployments
func DeploymentsGETHandler(w http.ResponseWriter, r *http.Request) {
	// Implement logic to get deployments
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode([]string{"deployment1", "deployment2"}); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}
