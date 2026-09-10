// Operation metadata for the UI: the dropdown entries and which operation is
// unary. This mirrors the operations map in the backend, which stays
// authoritative — it re-checks the name and arity on every request, so drift
// here is a UI bug, never a correctness hole.

import type { Operation } from './api'

export interface OperationInfo {
  value: Operation
  label: string
}

// The four basic operations first, then the extended ones.
export const OPERATIONS: readonly OperationInfo[] = [
  { value: 'add', label: 'Add (a + b)' },
  { value: 'subtract', label: 'Subtract (a - b)' },
  { value: 'multiply', label: 'Multiply (a × b)' },
  { value: 'divide', label: 'Divide (a ÷ b)' },
  { value: 'power', label: 'Power (a ^ b)' },
  { value: 'sqrt', label: 'Square root (√a)' },
  { value: 'percentage', label: 'Percentage (a% of b)' },
]

// sqrt is the only unary operation. This decides whether the second input is
// disabled and whether b is validated and sent.
export function isUnary(operation: Operation): boolean {
  return operation === 'sqrt'
}
