package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"calculator-app/backend/internal/handler"
)

// post runs one request through the handler and returns the recorder.
func post(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/calculate", strings.NewReader(body))
	rec := httptest.NewRecorder()
	handler.Calculate(rec, req)
	return rec
}

func TestCalculateSuccess(t *testing.T) {
	tests := []struct {
		name string
		body string
		want float64
	}{
		{"add", `{"operation":"add","a":2,"b":3}`, 5},
		{"subtract", `{"operation":"subtract","a":10,"b":4}`, 6},
		{"multiply", `{"operation":"multiply","a":6,"b":7}`, 42},
		{"divide", `{"operation":"divide","a":20,"b":4}`, 5},
		{"power", `{"operation":"power","a":2,"b":10}`, 1024},
		{"percentage", `{"operation":"percentage","a":15,"b":200}`, 30},
		{"sqrt without b", `{"operation":"sqrt","a":144}`, 12},
		{"sqrt ignores b", `{"operation":"sqrt","a":9,"b":999}`, 3},
		{"zero operands are valid, not 'missing'", `{"operation":"add","a":0,"b":0}`, 0},
		{"unknown fields are ignored", `{"operation":"add","a":1,"b":2,"extra":true}`, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := post(t, tt.body)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want %q", got, "application/json")
			}

			var body struct {
				Result float64 `json:"result"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decoding response %s: %v", rec.Body.String(), err)
			}
			if body.Result != tt.want {
				t.Errorf("result = %v, want %v", body.Result, tt.want)
			}
		})
	}
}

func TestCalculateErrors(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		// Decoding.
		{"empty body", ``, http.StatusBadRequest, "request body is empty"},
		{"whitespace-only body", "  \n ", http.StatusBadRequest, "request body is empty"},
		{"malformed JSON", `{"operation":`, http.StatusBadRequest, "invalid JSON in request body"},
		{"operation is not a string", `{"operation":5,"a":1,"b":2}`, http.StatusBadRequest, "operation must be a string"},
		{"operand a is a string", `{"operation":"add","a":"x","b":2}`, http.StatusBadRequest, `operand "a" must be a number`},
		{"operand b is a bool", `{"operation":"add","a":1,"b":true}`, http.StatusBadRequest, `operand "b" must be a number`},

		// Semantic validation, in the order the handler checks.
		{"operation missing", `{"a":1,"b":2}`, http.StatusBadRequest, "operation is required"},
		{"operation empty", `{"operation":"","a":1,"b":2}`, http.StatusBadRequest, "operation is required"},
		{"operation null", `{"operation":null,"a":1,"b":2}`, http.StatusBadRequest, "operation is required"},
		{"unknown operation", `{"operation":"cube","a":2,"b":3}`, http.StatusBadRequest, `unknown operation "cube"`},
		{"operand a missing", `{"operation":"add","b":2}`, http.StatusBadRequest, `operand "a" is required`},
		{"operand a null", `{"operation":"add","a":null,"b":2}`, http.StatusBadRequest, `operand "a" is required`},
		{"operand b missing for binary op", `{"operation":"add","a":1}`, http.StatusBadRequest, `operand "b" is required for operation "add"`},
		{"operand b missing for divide", `{"operation":"divide","a":1}`, http.StatusBadRequest, `operand "b" is required for operation "divide"`},
		{"operand a out of float64 range", `{"operation":"add","a":1e400,"b":1}`, http.StatusBadRequest, `operand "a" must be a finite number`},
		{"operand b out of float64 range", `{"operation":"add","a":1,"b":-1e400}`, http.StatusBadRequest, `operand "b" must be a finite number`},

		// Undefined answers and non-finite results, surfaced from the
		// calculator package as 400s.
		{"division by zero", `{"operation":"divide","a":1,"b":0}`, http.StatusBadRequest, "division by zero"},
		{"sqrt of a negative", `{"operation":"sqrt","a":-4}`, http.StatusBadRequest, "square root of a negative number is undefined"},
		{"overflow to +Inf", `{"operation":"multiply","a":1e308,"b":10}`, http.StatusBadRequest, "result is not a finite number"},
		{"power producing NaN", `{"operation":"power","a":-2,"b":0.5}`, http.StatusBadRequest, "result is not a finite number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := post(t, tt.body)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
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

// A body past the 1 MiB cap is rejected as 413 before it is fully read, rather
// than being buffered into memory and then failing as invalid JSON.
func TestCalculateBodyTooLarge(t *testing.T) {
	// Valid JSON prefix followed by more than 1 MiB of digits, so the request is
	// rejected for its size, not its syntax.
	body := `{"operation":"add","a":1,"b":` + strings.Repeat("1", (1<<20)+1) + `}`
	rec := post(t, body)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
	var resp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding response %s: %v", rec.Body.String(), err)
	}
	if resp.Error != "request body too large" {
		t.Errorf("error = %q, want %q", resp.Error, "request body too large")
	}
}

// WriteError is what cmd/server uses for its 404/405, so the routing layer's
// errors come out in the same wire format as the handler's own.
func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	handler.WriteError(rec, http.StatusNotFound, "not found")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decoding response %s: %v", rec.Body.String(), err)
	}
	if body.Error != "not found" {
		t.Errorf("error = %q, want %q", body.Error, "not found")
	}
}

// failingReader stands in for a connection that dies mid-body.
type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("connection reset") }

// panickingReader forces the handler's recover path.
type panickingReader struct{}

func (panickingReader) Read([]byte) (int, error) { panic("boom") }

func TestCalculateUnreadableBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/calculate", failingReader{})
	rec := httptest.NewRecorder()
	handler.Calculate(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// A panic anywhere in the handler must become a JSON 500 rather than a dropped
// connection.
func TestCalculateRecoversFromPanic(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/calculate", panickingReader{})
	rec := httptest.NewRecorder()

	handler.Calculate(rec, req) // must not propagate the panic

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decoding response %s: %v", rec.Body.String(), err)
	}
	if body.Error != "internal server error" {
		t.Errorf("error = %q, want %q", body.Error, "internal server error")
	}
}

// A failing request must never produce a "result" key, and a successful one must
// never produce an "error" key — the two response shapes are disjoint.
func TestResponseShapesAreDisjoint(t *testing.T) {
	ok := post(t, `{"operation":"add","a":1,"b":2}`)
	if strings.Contains(ok.Body.String(), `"error"`) {
		t.Errorf("success response contains an error key: %s", ok.Body.String())
	}

	bad := post(t, `{"operation":"divide","a":1,"b":0}`)
	if strings.Contains(bad.Body.String(), `"result"`) {
		t.Errorf("error response contains a result key: %s", bad.Body.String())
	}
}
