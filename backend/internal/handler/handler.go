// Package handler owns the HTTP side of the calculator: decoding the JSON
// request, validating it, dispatching to the calculator package, and shaping the
// JSON response. The exact status codes and messages are specified in
// DECISIONS.md § Error responses; this file implements that table.
package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"

	"calculator-app/backend/internal/calculator"
)

// maxBodyBytes caps the request body. A calculation payload is a few dozen
// bytes; 1 MiB is far more than any legitimate request needs and keeps a
// deliberately huge body from being read into memory before it is rejected.
const maxBodyBytes = 1 << 20

// request is decoded in two stages. The fields are json.RawMessage (the raw
// bytes of each JSON value, decoded later) rather than string/float64 so we can
// tell "field absent" from "field present but the wrong type" and return the
// specific message the spec asks for — e.g. `operation must be a string` vs
// `operation is required`.
type request struct {
	Operation json.RawMessage `json:"operation"`
	A         json.RawMessage `json:"a"`
	B         json.RawMessage `json:"b"`
}

type successResponse struct {
	Result float64 `json:"result"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// apiError is a validation failure carrying the HTTP status and client-facing
// message to send. It implements error so it can flow through normal return
// values.
type apiError struct {
	status  int
	message string
}

func (e *apiError) Error() string { return e.message }

func badRequest(format string, args ...any) *apiError {
	return &apiError{status: http.StatusBadRequest, message: fmt.Sprintf(format, args...)}
}

// Calculate handles POST /calculate. Routing (path and method) is done in
// cmd/server; by the time we get here the request is a POST to the right path.
func Calculate(w http.ResponseWriter, r *http.Request) {
	// Defence in depth: an unexpected panic becomes a 500 with a JSON body
	// rather than a dropped connection (DECISIONS.md § Error responses).
	defer func() {
		if p := recover(); p != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		}
	}()

	// Cap the body before anything reads it. Once the limit is passed, the next
	// Read on r.Body returns an *http.MaxBytesError, which compute turns into a
	// 413.
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	result, apiErr := compute(r.Body)
	if apiErr != nil {
		writeJSON(w, apiErr.status, errorResponse{Error: apiErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, successResponse{Result: result})
}

// compute does the decode → validate → calculate pipeline. It is separated from
// the HTTP plumbing so the ordering of checks (which the spec pins down) is easy
// to read top to bottom.
func compute(body io.Reader) (float64, *apiError) {
	raw, err := io.ReadAll(body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return 0, &apiError{status: http.StatusRequestEntityTooLarge, message: "request body too large"}
		}
		return 0, badRequest("invalid JSON in request body")
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return 0, badRequest("request body is empty")
	}

	var req request
	if err := json.Unmarshal(raw, &req); err != nil {
		return 0, badRequest("invalid JSON in request body")
	}

	op, apiErr := parseOperation(req.Operation)
	if apiErr != nil {
		return 0, apiErr
	}
	if !calculator.Supported(op) {
		return 0, badRequest("unknown operation %q", op)
	}

	// Presence checks first (spec order), then value checks.
	if absent(req.A) {
		return 0, badRequest(`operand "a" is required`)
	}
	needB := calculator.RequiresB(op)
	if needB && absent(req.B) {
		return 0, badRequest("operand %q is required for operation %q", "b", op)
	}

	a, apiErr := parseOperand(req.A, "a")
	if apiErr != nil {
		return 0, apiErr
	}
	var b float64
	if needB {
		b, apiErr = parseOperand(req.B, "b")
		if apiErr != nil {
			return 0, apiErr
		}
	}

	result, err := calculator.Calculate(op, a, b)
	if err != nil {
		// Every calculator error (divide by zero, sqrt of negative, non-finite
		// result) is a client error, not a server fault — DECISIONS.md B4/B5.
		return 0, badRequest("%s", err.Error())
	}
	return result, nil
}

// absent reports whether a raw JSON value was omitted or sent explicitly as
// null. Both mean "no value" for our purposes.
func absent(raw json.RawMessage) bool {
	return len(bytes.TrimSpace(raw)) == 0 || string(bytes.TrimSpace(raw)) == "null"
}

func parseOperation(raw json.RawMessage) (string, *apiError) {
	if absent(raw) {
		return "", badRequest("operation is required")
	}
	var op string
	if err := json.Unmarshal(raw, &op); err != nil {
		return "", badRequest("operation must be a string")
	}
	if op == "" {
		return "", badRequest("operation is required")
	}
	return op, nil
}

// parseOperand decodes one operand. A non-number JSON value (`"x"`, true) is
// `must be a number`; a numeric literal outside the float64 range (1e400) parses
// to ±Inf and is `must be a finite number` — two distinct messages per the spec.
func parseOperand(raw json.RawMessage, name string) (float64, *apiError) {
	var num json.Number
	if err := json.Unmarshal(raw, &num); err != nil {
		return 0, badRequest("operand %q must be a number", name)
	}
	f, err := strconv.ParseFloat(num.String(), 64)
	if err != nil || math.IsInf(f, 0) || math.IsNaN(f) {
		return 0, badRequest("operand %q must be a finite number", name)
	}
	return f, nil
}

// WriteError writes an error in the API's wire format, {"error": "..."}. It is
// exported so the routing layer in cmd/server can return its 404/405 in exactly
// the same shape as every other error — the wire format is defined once, here.
func WriteError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// A trailing newline is conventional for JSON API responses and keeps
	// curl output tidy. An encode error here is unrecoverable (headers are
	// already sent), so there is nothing useful to do with it.
	_ = json.NewEncoder(w).Encode(body)
}
