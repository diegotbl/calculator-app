package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouter(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantError  string // empty means a success response is expected
		wantAllow  string
	}{
		{
			name:       "POST /calculate reaches the handler",
			method:     http.MethodPost,
			path:       calculatePath,
			body:       `{"operation":"add","a":2,"b":3}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "handler errors still pass through",
			method:     http.MethodPost,
			path:       calculatePath,
			body:       `{"operation":"divide","a":1,"b":0}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "division by zero",
		},
		{
			name:       "unknown path",
			method:     http.MethodPost,
			path:       "/nope",
			wantStatus: http.StatusNotFound,
			wantError:  "not found",
		},
		{
			name:       "root path",
			method:     http.MethodGet,
			path:       "/",
			wantStatus: http.StatusNotFound,
			wantError:  "not found",
		},
		{
			name:       "path is matched exactly, not by prefix",
			method:     http.MethodPost,
			path:       calculatePath + "/extra",
			wantStatus: http.StatusNotFound,
			wantError:  "not found",
		},
		{
			name:       "GET on the right path",
			method:     http.MethodGet,
			path:       calculatePath,
			wantStatus: http.StatusMethodNotAllowed,
			wantError:  "method not allowed",
			wantAllow:  http.MethodPost,
		},
		{
			name:       "PUT on the right path",
			method:     http.MethodPut,
			path:       calculatePath,
			body:       `{"operation":"add","a":1,"b":2}`,
			wantStatus: http.StatusMethodNotAllowed,
			wantError:  "method not allowed",
			wantAllow:  http.MethodPost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			router().ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if got := rec.Header().Get("Allow"); got != tt.wantAllow {
				t.Errorf("Allow header = %q, want %q", got, tt.wantAllow)
			}

			// Every response, success or failure, is JSON.
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want %q", got, "application/json")
			}
			var body struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decoding response %s: %v", rec.Body.String(), err)
			}
			if body.Error != tt.wantError {
				t.Errorf("error = %q, want %q", body.Error, tt.wantError)
			}
		})
	}
}

func TestPort(t *testing.T) {
	t.Setenv("PORT", "9999")
	if got := port(); got != "9999" {
		t.Errorf("port() = %q, want %q", got, "9999")
	}

	// An unset or empty PORT falls back to the default.
	t.Setenv("PORT", "")
	if got := port(); got != defaultPort {
		t.Errorf("port() with empty PORT = %q, want %q", got, defaultPort)
	}
}
