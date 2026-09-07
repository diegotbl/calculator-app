// Command server runs the calculator HTTP API: a single POST /calculate
// endpoint. This file owns process concerns (listen address, timeouts) and
// routing; everything about the request itself lives in internal/handler.
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

	// The zero-value http.Server has no timeouts at all, which lets a slow or
	// idle client hold a connection open indefinitely. These are conservative
	// values for an endpoint that only ever does arithmetic.
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

// port returns the PORT environment variable, falling back to 8080
// (DECISIONS.md T5).
func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return defaultPort
}

// router dispatches every incoming request. It checks path and method by hand
// rather than using ServeMux's method patterns because those answer a bad path
// or method with net/http's plain-text default body; this API returns JSON for
// every error, including these two.
func router() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != calculatePath {
			handler.WriteError(w, http.StatusNotFound, "not found")
			return
		}
		if r.Method != http.MethodPost {
			// RFC 9110 requires Allow on a 405.
			w.Header().Set("Allow", http.MethodPost)
			handler.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		handler.Calculate(w, r)
	})
}
