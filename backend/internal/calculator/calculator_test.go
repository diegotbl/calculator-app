// Black-box tests (package calculator_test): they exercise the package through
// its exported API only, the same way the handler package uses it.
package calculator_test

import (
	"errors"
	"testing"

	"calculator-app/backend/internal/calculator"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		name    string
		op      string
		a, b    float64
		want    float64
		wantErr error
	}{
		// Happy paths, one or more per operation.
		{"add", "add", 2, 3, 5, nil},
		{"add negatives", "add", -2, -3, -5, nil},
		{"subtract", "subtract", 10, 4, 6, nil},
		{"subtract into negative", "subtract", 3, 10, -7, nil},
		{"multiply", "multiply", 6, 7, 42, nil},
		{"multiply by zero", "multiply", 5, 0, 0, nil},
		{"divide", "divide", 20, 4, 5, nil},
		{"divide with fractional result", "divide", 1, 8, 0.125, nil},
		{"power", "power", 2, 10, 1024, nil},
		{"power to the zero", "power", 123, 0, 1, nil},
		{"power negative exponent", "power", 2, -2, 0.25, nil},
		{"sqrt", "sqrt", 144, 0, 12, nil},
		{"sqrt of zero", "sqrt", 0, 0, 0, nil},
		{"sqrt ignores b", "sqrt", 9, 999, 3, nil},
		{"percentage", "percentage", 15, 200, 30, nil},
		{"percentage of zero", "percentage", 50, 0, 0, nil},

		// Mathematically undefined requests → error (DECISIONS.md B4).
		{"divide by zero", "divide", 1, 0, 0, calculator.ErrDivideByZero},
		{"sqrt of negative", "sqrt", -4, 0, 0, calculator.ErrSqrtNegative},

		// Unknown operation.
		{"unknown operation", "cube", 2, 0, 0, calculator.ErrUnknownOperation},

		// Non-finite results → error (DECISIONS.md B5).
		{"multiply overflow to +Inf", "multiply", 1e308, 10, 0, calculator.ErrNotFinite},
		{"add overflow to +Inf", "add", 1e308, 1e308, 0, calculator.ErrNotFinite},
		{"power overflow to +Inf", "power", 10, 400, 0, calculator.ErrNotFinite},
		{"power NaN: negative base, fractional exponent", "power", -2, 0.5, 0, calculator.ErrNotFinite},
		{"zero to a negative power is +Inf", "power", 0, -1, 0, calculator.ErrNotFinite},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculator.Calculate(tt.op, tt.a, tt.b)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Calculate(%q, %v, %v) error = %v, want %v", tt.op, tt.a, tt.b, err, tt.wantErr)
			}
			if tt.wantErr == nil && got != tt.want {
				t.Errorf("Calculate(%q, %v, %v) = %v, want %v", tt.op, tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestSupported(t *testing.T) {
	known := []string{"add", "subtract", "multiply", "divide", "power", "sqrt", "percentage"}
	for _, op := range known {
		if !calculator.Supported(op) {
			t.Errorf("Supported(%q) = false, want true", op)
		}
	}
	for _, op := range []string{"", "cube", "ADD", "modulo"} {
		if calculator.Supported(op) {
			t.Errorf("Supported(%q) = true, want false", op)
		}
	}
}

func TestRequiresB(t *testing.T) {
	if calculator.RequiresB("sqrt") {
		t.Error(`RequiresB("sqrt") = true, want false`)
	}
	for _, op := range []string{"add", "subtract", "multiply", "divide", "power", "percentage"} {
		if !calculator.RequiresB(op) {
			t.Errorf("RequiresB(%q) = false, want true", op)
		}
	}
}
