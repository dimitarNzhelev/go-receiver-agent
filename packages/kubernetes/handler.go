package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	clientset     *kubernetes.Clientset
	dynamicClient dynamic.Interface
)

func InitKubernetesClients() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	dynamicClient, err = dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	return nil
}

// GET /pods
// PodsGETHandler returns a list of pods
func PodsGETHandler(w http.ResponseWriter, r *http.Request) {
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		http.Error(w, "Failed to get pods", http.StatusInternalServerError)
		log.Printf("Failed to get pods: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pods.Items); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}

// GET /nodes
// NodesGETHandler returns a list of nodes
func NodesGETHandler(w http.ResponseWriter, r *http.Request) {
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		http.Error(w, "Failed to get nodes", http.StatusInternalServerError)
		log.Printf("Failed to get nodes: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(nodes.Items); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}

// GET /namespaces
// NamespacesGETHandler returns a list of namespaces
func NamespacesGETHandler(w http.ResponseWriter, r *http.Request) {
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		http.Error(w, "Failed to get namespaces", http.StatusInternalServerError)
		log.Printf("Failed to get namespaces: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(namespaces.Items); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}

// GET /deployments
// DeploymentsGETHandler returns a list of deployments
func DeploymentsGETHandler(w http.ResponseWriter, r *http.Request) {
	deployments, err := clientset.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		http.Error(w, "Failed to get deployments", http.StatusInternalServerError)
		log.Printf("Failed to get deployments: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(deployments.Items); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}

var prometheusRuleGVR = schema.GroupVersionResource{
	Group:    "monitoring.coreos.com",
	Version:  "v1",
	Resource: "prometheusrules",
}

// GET /rules
// GetRulesHandler fetches all PrometheusRule objects in the cluster
func GetRulesHandler(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		http.Error(w, "Namespace query parameter is required", http.StatusBadRequest)
		return
	}

	// Handling all namespaces case by setting namespace to empty string
	if namespace == "all" {
		namespace = ""
	}

	rules, err := dynamicClient.Resource(prometheusRuleGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch PrometheusRule objects: %v", err), http.StatusInternalServerError)
		log.Printf("Error fetching PrometheusRule objects: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(rules.Items); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Printf("JSON encoding error: %v", err)
	}
}

// POST /rules
// CreateRuleHandler creates a new PrometheusRule object
func CreateRuleHandler(w http.ResponseWriter, r *http.Request) {

	var rule map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		log.Printf("JSON decoding error: %v", err)
		return
	}
	defer r.Body.Close()

	metadata, ok := rule["metadata"].(map[string]interface{})
	if !ok || metadata["namespace"] == "" {
		http.Error(w, "Namespace is required in metadata", http.StatusBadRequest)
		return
	}

	namespace, ok := metadata["namespace"].(string)
	if !ok || namespace == "" {
		http.Error(w, "Namespace must be a non-empty string", http.StatusBadRequest)
		return
	}

	createdRule, err := dynamicClient.Resource(prometheusRuleGVR).Namespace(namespace).Create(
		context.TODO(),
		&unstructured.Unstructured{Object: rule},
		metav1.CreateOptions{},
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create PrometheusRule object: %v", err), http.StatusInternalServerError)
		log.Printf("Error creating PrometheusRule: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(createdRule); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Printf("JSON encoding error: %v", err)
	}
}

// PUT /rules/{id}
// UpdateRuleHandler updates an existing PrometheusRule object
func UpdateRuleHandler(w http.ResponseWriter, r *http.Request) {

	id := strings.TrimPrefix(r.URL.Path, "/rules/")
	if id == "" {
		http.Error(w, "Rule ID is required in the URL", http.StatusBadRequest)
		return
	}

	var rule map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		log.Printf("JSON decoding error: %v", err)
		return
	}
	defer r.Body.Close()

	metadata, ok := rule["metadata"].(map[string]interface{})
	if !ok || metadata["namespace"] == "" {
		http.Error(w, "Namespace is required in metadata", http.StatusBadRequest)
		return
	}

	namespace, ok := metadata["namespace"].(string)
	if !ok || namespace == "" {
		http.Error(w, "Namespace must be a non-empty string", http.StatusBadRequest)
		return
	}

	updatedRule, err := dynamicClient.Resource(prometheusRuleGVR).Namespace(namespace).Update(
		context.TODO(),
		&unstructured.Unstructured{Object: rule},
		metav1.UpdateOptions{},
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update PrometheusRule object: %v", err), http.StatusInternalServerError)
		log.Printf("Error updating PrometheusRule: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedRule); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Printf("JSON encoding error: %v", err)
	}
}

// DELETE /rules/{id}
// DeleteRuleHandler deletes a specific PrometheusRule object by its name
func DeleteRuleHandler(w http.ResponseWriter, r *http.Request) {

	id := strings.TrimPrefix(r.URL.Path, "/rules/")
	if id == "" {
		http.Error(w, "Rule ID is required in the URL", http.StatusBadRequest)
		return
	}

	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		http.Error(w, "Namespace query parameter is required", http.StatusBadRequest)
		return
	}

	err := dynamicClient.Resource(prometheusRuleGVR).Namespace(namespace).Delete(context.TODO(), id, metav1.DeleteOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete PrometheusRule object: %v", err), http.StatusInternalServerError)
		log.Printf("Error deleting PrometheusRule: %v", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
