// Package handler owns the HTTP side of the calculator: decoding the JSON
// request, validating it, dispatching to the calculator package, and shaping the
// JSON response.
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
// bytes; 1 MiB is far more than any legitimate request needs and stops a
// deliberately huge body from being read into memory before it is rejected.
const maxBodyBytes = 1 << 20

// request holds each field as raw JSON, decoded one at a time later, so the
// handler can tell an absent field from one sent with the wrong type and return
// the specific message for each ("operation is required" vs "must be a string").
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

// apiError is a validation failure carrying the HTTP status and message to send.
type apiError struct {
	status  int
	message string
}

func (e *apiError) Error() string { return e.message }

func badRequest(format string, args ...any) *apiError {
	return &apiError{status: http.StatusBadRequest, message: fmt.Sprintf(format, args...)}
}

// Calculate handles POST /calculate. Routing (path and method) is done in
// cmd/server.
func Calculate(w http.ResponseWriter, r *http.Request) {
	// An unexpected panic becomes a JSON 500 rather than a dropped connection.
	defer func() {
		if p := recover(); p != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		}
	}()

	// Once the body passes maxBodyBytes, the next read on it returns an
	// *http.MaxBytesError, which compute turns into a 413.
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	result, apiErr := compute(r.Body)
	if apiErr != nil {
		writeJSON(w, apiErr.status, errorResponse{Error: apiErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, successResponse{Result: result})
}

// compute runs the decode → validate → calculate pipeline. The checks are
// ordered so the response to any given bad request is predictable.
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

	// Presence before value, so "b is required" wins over "b must be a number".
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
		// A request with no defined answer is the client's mistake, not a server fault.
		return 0, badRequest("%s", err.Error())
	}
	return result, nil
}

// absent reports whether a raw JSON value was omitted or sent as null.
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

// parseOperand decodes one operand. Decoding via json.Number keeps two error
// cases distinct: a non-number value ("x", true) is "must be a number", while a
// literal outside the float64 range (1e400 parses to ±Inf) is "must be a finite
// number" — unmarshalling 1e400 straight into a float64 would report it as the
// former.
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

// WriteError writes an error in the API's wire format, {"error": "..."}. Exported
// so cmd/server returns its 404/405 in the same shape as every other error.
func WriteError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
