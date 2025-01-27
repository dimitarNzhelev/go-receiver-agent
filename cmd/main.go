package main

import (
	"context"
	"log"
	"main/packages/alertmanager"
	"main/packages/config"
	"main/packages/kubernetes"
	"main/packages/utils"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	port := config.GetEnv("PORT", "5000")
	token := config.GetEnv("AUTH_TOKEN", "secret")

	if err := kubernetes.InitKubernetesClients(); err != nil {
		log.Fatalf("Failed to initialize Kubernetes clients: %v", err)
	}

	router := http.NewServeMux()

	router.Handle("GET /pods", utils.LoggingMiddleware(utils.AuthenticationMiddleware(http.HandlerFunc(kubernetes.PodsGETHandler), token)))
	router.Handle("GET /namespaces", utils.LoggingMiddleware(utils.AuthenticationMiddleware(http.HandlerFunc(kubernetes.NamespacesGETHandler), token)))
	router.Handle("GET /nodes", utils.LoggingMiddleware(utils.AuthenticationMiddleware(http.HandlerFunc(kubernetes.NodesGETHandler), token)))
	router.Handle("GET /deployments", utils.LoggingMiddleware(utils.AuthenticationMiddleware(http.HandlerFunc(kubernetes.DeploymentsGETHandler), token)))

	router.Handle("GET /alerts", utils.LoggingMiddleware(utils.AuthenticationMiddleware(http.HandlerFunc(alertmanager.AlertGETHandler), token)))
	router.Handle("POST /alerts", utils.LoggingMiddleware(utils.AuthenticationMiddleware(http.HandlerFunc(alertmanager.AlertPOSTHandler), token)))

	router.Handle("GET /alerts/firing", utils.LoggingMiddleware(utils.AuthenticationMiddleware(http.HandlerFunc(alertmanager.AlertFiringGETHandler), token)))

	router.Handle("GET /alerts/silences", utils.LoggingMiddleware(utils.AuthenticationMiddleware(http.HandlerFunc(alertmanager.AlertSilencesGETHandler), token)))
	router.Handle("POST /alerts/silences", utils.LoggingMiddleware(utils.AuthenticationMiddleware(http.HandlerFunc(alertmanager.AlertSilencesPOSTHandler), token)))
	router.Handle("DELETE /alerts/silences/{id}", utils.LoggingMiddleware(utils.AuthenticationMiddleware(http.HandlerFunc(alertmanager.AlertSilencesDELETEHandler), token)))

	router.Handle("GET /rules", utils.LoggingMiddleware(utils.AuthenticationMiddleware(http.HandlerFunc(kubernetes.GetRulesHandler), token)))
	router.Handle("POST /rules", utils.LoggingMiddleware(utils.AuthenticationMiddleware(http.HandlerFunc(kubernetes.CreateRuleHandler), token)))
	router.Handle("PUT /rules/{id}", utils.LoggingMiddleware(utils.AuthenticationMiddleware(http.HandlerFunc(kubernetes.UpdateRuleHandler), token)))
	router.Handle("DELETE /rules/{id}", utils.LoggingMiddleware(utils.AuthenticationMiddleware(http.HandlerFunc(kubernetes.DeleteRuleHandler), token)))

	// Define the server configuration
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Create a channel to listen for OS interrupt signals (like Ctrl+C)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Start the server in a separate goroutine
	go func() {
		log.Printf("Server is running on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on port %s: %v", port, err)
		}
	}()

	// Block until an interrupt signal is received
	<-stop

	// Create a context with a 5-second timeout for the shutdown process
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Shutting down the server...")
	// Attempt to gracefully shut down the server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Println("Server exited properly")
}
