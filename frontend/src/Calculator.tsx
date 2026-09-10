import { useState } from 'react'
import type { ChangeEvent, FormEvent } from 'react'
import { ApiError, calculate } from './api'
import type { CalcRequest, Operation } from './api'
import { OPERATIONS, isUnary } from './operations'
import './Calculator.css'

// What the status region under the form is showing. One value rather than
// separate booleans, so combinations like "loading while showing an error"
// cannot be represented.
type Status =
  | { kind: 'idle' }
  | { kind: 'loading' }
  | { kind: 'result'; value: number }
  | { kind: 'error'; message: string }

type ParsedOperand = { ok: true; value: number } | { ok: false; message: string }

// Client-side operand validation. Messages copy the backend's wording so the
// user cannot tell which layer rejected the input; the backend still re-checks
// everything.
function parseOperand(raw: string, name: 'a' | 'b'): ParsedOperand {
  const trimmed = raw.trim()
  // Must come first: Number('') and Number('  ') are 0, not NaN, so a blank
  // field would otherwise pass as a valid zero.
  if (trimmed === '') {
    return { ok: false, message: `operand "${name}" is required` }
  }
  const value = Number(trimmed)
  if (Number.isNaN(value)) {
    return { ok: false, message: `operand "${name}" must be a number` }
  }
  // Number('1e400') is Infinity, which the backend rejects too.
  if (!Number.isFinite(value)) {
    return { ok: false, message: `operand "${name}" must be a finite number` }
  }
  return { ok: true, value }
}

// Turns the three form fields into a request, or the first thing wrong with
// them. b is left off entirely for sqrt; JSON.stringify then drops it.
function validate(
  operation: Operation,
  a: string,
  b: string,
): { ok: true; request: CalcRequest } | { ok: false; message: string } {
  const parsedA = parseOperand(a, 'a')
  if (!parsedA.ok) {
    return parsedA
  }

  if (isUnary(operation)) {
    return { ok: true, request: { operation, a: parsedA.value } }
  }

  if (b.trim() === '') {
    return { ok: false, message: `operand "b" is required for operation "${operation}"` }
  }
  const parsedB = parseOperand(b, 'b')
  if (!parsedB.ok) {
    return parsedB
  }

  return { ok: true, request: { operation, a: parsedA.value, b: parsedB.value } }
}

export default function Calculator() {
  const [operation, setOperation] = useState<Operation>('add')
  // Operands are held as strings — that is what an input gives us, and it keeps
  // "empty" distinct from "zero" while the user types. Parsed once, on submit.
  const [a, setA] = useState('')
  const [b, setB] = useState('')
  const [status, setStatus] = useState<Status>({ kind: 'idle' })

  const unary = isUnary(operation)
  const loading = status.kind === 'loading'

  // Any edit invalidates the answer on screen, so the status region is cleared.
  function handleOperationChange(event: ChangeEvent<HTMLSelectElement>) {
    // Safe cast: every <option> value comes from OPERATIONS.
    setOperation(event.target.value as Operation)
    setStatus({ kind: 'idle' })
  }

  function handleAChange(event: ChangeEvent<HTMLInputElement>) {
    setA(event.target.value)
    setStatus({ kind: 'idle' })
  }

  function handleBChange(event: ChangeEvent<HTMLInputElement>) {
    setB(event.target.value)
    setStatus({ kind: 'idle' })
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const validated = validate(operation, a, b)
    if (!validated.ok) {
      setStatus({ kind: 'error', message: validated.message })
      return
    }

    setStatus({ kind: 'loading' })
    try {
      const { result } = await calculate(validated.request)
      setStatus({ kind: 'result', value: result })
    } catch (error) {
      // api.ts turns every expected failure into an ApiError. Anything else is a
      // bug on this side and gets a generic message rather than a leaked stack.
      setStatus({
        kind: 'error',
        message: error instanceof ApiError ? error.message : 'Something went wrong.',
      })
    }
  }

  return (
    <main className="calculator">
      <h1>Calculator</h1>

      <form className="calculator__form" onSubmit={handleSubmit}>
        <label className="calculator__field" htmlFor="operation">
          Operation
          <select id="operation" value={operation} onChange={handleOperationChange}>
            {OPERATIONS.map((op) => (
              <option key={op.value} value={op.value}>
                {op.label}
              </option>
            ))}
          </select>
        </label>

        {/*
          type="text", not type="number": a number input silently blanks any
          value that is not a valid float, so "abc" and "1e400" would never reach
          the validation that explains what is wrong. inputMode="decimal" still
          gets the numeric keypad on mobile.
        */}
        <label className="calculator__field" htmlFor="operand-a">
          Operand a
          <input
            id="operand-a"
            type="text"
            inputMode="decimal"
            autoComplete="off"
            value={a}
            onChange={handleAChange}
          />
        </label>

        <label className="calculator__field" htmlFor="operand-b">
          Operand b
          <input
            id="operand-b"
            type="text"
            inputMode="decimal"
            autoComplete="off"
            // b is kept in state while disabled so switching back to a binary
            // operation restores it, but blanked on screen — a value next to a
            // disabled field would read as "this number is being used".
            value={unary ? '' : b}
            onChange={handleBChange}
            // sqrt is unary. Disabled rather than hidden so the layout does not
            // jump when the operation changes.
            disabled={unary}
          />
        </label>

        <button type="submit" disabled={loading}>
          {loading ? 'Calculating…' : 'Calculate'}
        </button>
      </form>

      {/*
        One region for all three outcomes. role="status" (implies
        aria-live="polite") announces whatever appears here to a screen reader;
        it must be in the DOM before its content changes, so it is always
        rendered. The result is printed as-is — rounding for looks would hide
        real floating-point behaviour.
      */}
      <div className="calculator__status" role="status">
        {status.kind === 'result' && (
          <p className="calculator__result">Result: {status.value}</p>
        )}
        {status.kind === 'error' && (
          <p className="calculator__error">{status.message}</p>
        )}
      </div>
    </main>
  )
}
