package kubernetes

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// getClientset creates a Kubernetes clientset
func getClientset() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

// PodsGETHandler returns a list of pods
func PodsGETHandler(w http.ResponseWriter, r *http.Request) {
	clientset, err := getClientset()
	if err != nil {
		http.Error(w, "Failed to create Kubernetes client", http.StatusInternalServerError)
		log.Printf("Failed to create Kubernetes client: %v", err)
		return
	}

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

// NodesGETHandler returns a list of nodes
func NodesGETHandler(w http.ResponseWriter, r *http.Request) {
	clientset, err := getClientset()
	if err != nil {
		http.Error(w, "Failed to create Kubernetes client", http.StatusInternalServerError)
		log.Printf("Failed to create Kubernetes client: %v", err)
		return
	}

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

// NamespacesGETHandler returns a list of namespaces
func NamespacesGETHandler(w http.ResponseWriter, r *http.Request) {
	clientset, err := getClientset()
	if err != nil {
		http.Error(w, "Failed to create Kubernetes client", http.StatusInternalServerError)
		log.Printf("Failed to create Kubernetes client: %v", err)
		return
	}

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

// DeploymentsGETHandler returns a list of deployments
func DeploymentsGETHandler(w http.ResponseWriter, r *http.Request) {
	clientset, err := getClientset()
	if err != nil {
		http.Error(w, "Failed to create Kubernetes client", http.StatusInternalServerError)
		log.Printf("Failed to create Kubernetes client: %v", err)
		return
	}

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
