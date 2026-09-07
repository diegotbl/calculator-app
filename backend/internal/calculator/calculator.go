// Package calculator holds the pure arithmetic behind the API. It knows nothing
// about HTTP or JSON: every operation takes float64 operands and returns
// (float64, error).
//
// An error is returned only when the request is mathematically undefined
// (divide by zero, square root of a negative) or the result is not a finite
// number (overflow to ±Inf, or NaN). DECISIONS.md B4/B5 explains why those are
// treated as client errors (HTTP 400) rather than server failures — the handler
// package does that mapping.
package calculator

import (
	"errors"
	"math"
)

// Sentinel errors. Go convention is to expose comparable error values (checked
// with errors.Is) rather than matching on message strings.
var (
	ErrDivideByZero     = errors.New("division by zero")
	ErrSqrtNegative     = errors.New("square root of a negative number is undefined")
	ErrNotFinite        = errors.New("result is not a finite number")
	ErrUnknownOperation = errors.New("unknown operation")
)

// operation bundles the math with its arity. Keeping arity here lets the handler
// validate a request ("is this a real operation? does it need a second
// operand?") without keeping its own copy of the operation list — the map below
// is the single source of truth (DECISIONS.md T7).
type operation struct {
	apply  func(a, b float64) (float64, error)
	binary bool // needs operand b; false only for sqrt
}

var operations = map[string]operation{
	"add":        {apply: add, binary: true},
	"subtract":   {apply: subtract, binary: true},
	"multiply":   {apply: multiply, binary: true},
	"divide":     {apply: divide, binary: true},
	"power":      {apply: power, binary: true},
	"sqrt":       {apply: sqrt, binary: false},
	"percentage": {apply: percentage, binary: true},
}

func add(a, b float64) (float64, error)      { return a + b, nil }
func subtract(a, b float64) (float64, error) { return a - b, nil }
func multiply(a, b float64) (float64, error) { return a * b, nil }
func power(a, b float64) (float64, error)    { return math.Pow(a, b), nil }

// percentage is "a percent of b" → (a / 100) * b, e.g. percentage(15, 200) = 30
// (DECISIONS.md B2).
func percentage(a, b float64) (float64, error) { return (a / 100) * b, nil }

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivideByZero
	}
	return a / b, nil
}

// sqrt is unary; the second parameter is ignored so it fits the shared function
// signature (DECISIONS.md B3).
func sqrt(a, _ float64) (float64, error) {
	if a < 0 {
		return 0, ErrSqrtNegative
	}
	return math.Sqrt(a), nil
}

// Supported reports whether op is one of the seven known operations.
func Supported(op string) bool {
	_, ok := operations[op]
	return ok
}

// RequiresB reports whether op needs a second operand. It is false for sqrt and
// for any unknown operation (callers check Supported first).
func RequiresB(op string) bool {
	return operations[op].binary
}

// Calculate runs op on the operands. It returns an error for an unknown
// operation, a mathematically undefined request, or a non-finite result. The
// non-finite guard runs once here so every operation is covered by it.
func Calculate(op string, a, b float64) (float64, error) {
	o, ok := operations[op]
	if !ok {
		return 0, ErrUnknownOperation
	}
	result, err := o.apply(a, b)
	if err != nil {
		return 0, err
	}
	if math.IsInf(result, 0) || math.IsNaN(result) {
		return 0, ErrNotFinite
	}
	return result, nil
}
