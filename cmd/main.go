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

	mux := http.NewServeMux()

	mux.Handle("GET /pods", utils.LoggingMiddleware(utils.AuthenticationMiddleware(http.HandlerFunc(kubernetes.PodsGETHandler), token)))
	mux.Handle("GET /namespaces", utils.LoggingMiddleware(utils.AuthenticationMiddleware(http.HandlerFunc(kubernetes.NamespacesGETHandler), token)))
	mux.Handle("GET /nodes", utils.LoggingMiddleware(utils.AuthenticationMiddleware(http.HandlerFunc(kubernetes.NodesGETHandler), token)))
	mux.Handle("GET /deployments", utils.LoggingMiddleware(utils.AuthenticationMiddleware(http.HandlerFunc(kubernetes.DeploymentsGETHandler), token)))
	mux.Handle("GET /alerts", utils.LoggingMiddleware(utils.AuthenticationMiddleware(http.HandlerFunc(alertmanager.AlertGETHandler), token)))
	mux.Handle("POST /alerts", utils.LoggingMiddleware(utils.AuthenticationMiddleware(http.HandlerFunc(alertmanager.AlertPOSTHandler), token)))
	mux.Handle("GET /alerts/rules", utils.LoggingMiddleware(utils.AuthenticationMiddleware(http.HandlerFunc(alertmanager.AlertRulesGETHandler), token)))

	// Define the server configuration
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
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
