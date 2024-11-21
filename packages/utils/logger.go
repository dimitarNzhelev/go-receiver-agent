package utils

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs the details of each HTTP request and response.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		// Proceed to the next middleware/handler
		next.ServeHTTP(w, r)
		// Log after response is served
		log.Printf(
			"%s %s %s %s %dms",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			r.UserAgent(),
			time.Since(startTime).Milliseconds(),
		)
	})
}
