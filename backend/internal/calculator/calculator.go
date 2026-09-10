// Package calculator holds the pure arithmetic behind the API: each operation
// takes float64 operands and returns (float64, error), with no knowledge of HTTP
// or JSON.
//
// An operation errors only when the request has no defined answer (divide by
// zero, square root of a negative) or the result is not finite (overflow to
// ±Inf, or NaN).
package calculator

import (
	"errors"
	"math"
)

// Sentinel errors, compared by the handler with errors.Is.
var (
	ErrDivideByZero     = errors.New("division by zero")
	ErrSqrtNegative     = errors.New("square root of a negative number is undefined")
	ErrNotFinite        = errors.New("result is not a finite number")
	ErrUnknownOperation = errors.New("unknown operation")
)

// operation pairs an operation's function with whether it needs operand b. This
// map is the single source of truth for the operation names and their arity.
type operation struct {
	apply  func(a, b float64) (float64, error)
	binary bool // false only for sqrt
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

// percentage returns a percent of b: (a / 100) * b, e.g. percentage(15, 200) = 30.
func percentage(a, b float64) (float64, error) { return (a / 100) * b, nil }

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivideByZero
	}
	return a / b, nil
}

// sqrt is unary; b is ignored so it fits the shared function signature.
func sqrt(a, _ float64) (float64, error) {
	if a < 0 {
		return 0, ErrSqrtNegative
	}
	return math.Sqrt(a), nil
}

// Supported reports whether op is a known operation.
func Supported(op string) bool {
	_, ok := operations[op]
	return ok
}

// RequiresB reports whether op needs a second operand. False for sqrt and for
// any unknown operation.
func RequiresB(op string) bool {
	return operations[op].binary
}

// Calculate runs op on the operands, returning an error for an unknown
// operation, a request with no defined answer, or a non-finite result. The
// non-finite guard runs here so every operation is covered by it.
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
