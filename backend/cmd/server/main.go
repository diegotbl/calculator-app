// Command server runs the calculator HTTP API: a single POST /calculate
// endpoint. This file owns process setup and routing; the request itself is
// handled in internal/handler.
package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"calculator-app/backend/internal/handler"
)

const (
	defaultPort   = "8080"
	calculatePath = "/calculate"
)

func main() {
	addr := ":" + port()

	// The zero-value http.Server has no timeouts, which lets a slow or idle
	// client hold a connection open indefinitely.
	srv := &http.Server{
		Addr:              addr,
		Handler:           router(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("calculator API listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server stopped: %v", err)
	}
}

// port returns $PORT, or 8080 if it is unset.
func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return defaultPort
}

// router checks path and method by hand rather than with ServeMux's method
// patterns so that a bad path or method gets a JSON error body like every other
// response, not net/http's plain-text default.
func router() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != calculatePath {
			handler.WriteError(w, http.StatusNotFound, "not found")
			return
		}
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			handler.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		handler.Calculate(w, r)
	})
}
