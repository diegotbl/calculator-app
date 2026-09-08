// Operation metadata for the UI: what to put in the dropdown, and which
// operation is unary. Kept out of Calculator.tsx so the component holds layout
// and state rather than a hardcoded list of the backend's vocabulary.
//
// This mirrors the operations map in backend/internal/calculator/calculator.go.
// The backend stays authoritative — it re-checks the operation name and the
// arity on every request (DECISIONS.md T9) — so a drift here is a UI bug, never
// a correctness hole.

import type { Operation } from './api'

/**
 * One entry in the operation dropdown.
 *
 * TS idiom: `import type` above imports only the compile-time type and is erased
 * from the emitted JS, so this module adds no runtime dependency on api.ts.
 */
export interface OperationInfo {
  value: Operation
  /** Dropdown text. The formula is spelled out so the operand roles are obvious. */
  label: string
}

/**
 * The seven supported operations, in the order they appear in the dropdown:
 * the four basic ones first, then the extended ones (DECISIONS.md B1).
 *
 * Labels name the operands `a` and `b` to match the input fields below them, so
 * "what is b for percentage?" is answerable without reading the docs
 * (percentage is "a percent of b" — DECISIONS.md B2).
 */
export const OPERATIONS: readonly OperationInfo[] = [
  { value: 'add', label: 'Add (a + b)' },
  { value: 'subtract', label: 'Subtract (a - b)' },
  { value: 'multiply', label: 'Multiply (a × b)' },
  { value: 'divide', label: 'Divide (a ÷ b)' },
  { value: 'power', label: 'Power (a ^ b)' },
  { value: 'sqrt', label: 'Square root (√a)' },
  { value: 'percentage', label: 'Percentage (a% of b)' },
]

/**
 * Whether an operation takes only operand `a`, which decides whether the second
 * input is disabled and whether `b` is validated and sent.
 *
 * `sqrt` is the only unary operation (DECISIONS.md B3), so this is a comparison
 * rather than a flag on each entry above — a per-entry `unary: false` repeated
 * six times carries less meaning than the one name that matters.
 */
export function isUnary(operation: Operation): boolean {
  return operation === 'sqrt'
}
