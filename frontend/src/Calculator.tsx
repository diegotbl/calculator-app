import { useState } from 'react'
import type { ChangeEvent, FormEvent } from 'react'
import { ApiError, calculate } from './api'
import type { CalcRequest, Operation } from './api'
import { OPERATIONS, isUnary } from './operations'
import './Calculator.css'

/**
 * What the status region under the form is currently showing. The four states
 * are mutually exclusive, so they are one value rather than four booleans —
 * there is no way to represent "loading and also showing an error" by accident.
 *
 * TS idiom: this is a discriminated union. Every member has a `kind` field with
 * a distinct literal type, so `if (status.kind === 'result')` narrows the type,
 * and `status.value` is only reachable inside that branch. Java would need a
 * sealed interface with four records plus pattern matching to say the same thing.
 */
type Status =
  | { kind: 'idle' }
  | { kind: 'loading' }
  | { kind: 'result'; value: number }
  | { kind: 'error'; message: string }

type ParsedOperand = { ok: true; value: number } | { ok: false; message: string }

/**
 * Client-side operand validation. The messages deliberately copy the backend's
 * wording from DECISIONS.md § Error responses, so a user cannot tell which layer
 * rejected their input — the frontend just gets there faster. The backend still
 * re-checks everything and wins any disagreement (DECISIONS.md T9).
 *
 * The empty check has to come first: Number('') and Number('  ') are both 0, not
 * NaN, so a blank field would otherwise sail through as a valid zero.
 */
function parseOperand(raw: string, name: 'a' | 'b'): ParsedOperand {
  const trimmed = raw.trim()
  if (trimmed === '') {
    return { ok: false, message: `operand "${name}" is required` }
  }
  const value = Number(trimmed)
  if (Number.isNaN(value)) {
    return { ok: false, message: `operand "${name}" must be a number` }
  }
  // Number('1e400') is Infinity, which the backend rejects too (DECISIONS.md B5).
  if (!Number.isFinite(value)) {
    return { ok: false, message: `operand "${name}" must be a finite number` }
  }
  return { ok: true, value }
}

/**
 * Turns the three form fields into a request, or into the first thing wrong with
 * them. Kept outside the component as a plain function of its inputs: no state,
 * no hooks, nothing to re-create on every render.
 *
 * `b` is left off entirely for sqrt. JSON.stringify drops undefined properties,
 * so the request goes out as {"operation":"sqrt","a":9} (DECISIONS.md B3).
 */
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

  // Binary operations name the operation in the "required" message, matching the
  // backend's `operand "b" is required for operation "add"`.
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
  // Operands are held as strings, not numbers: that is what an input element
  // actually gives us, and it keeps "empty" distinguishable from "zero" while
  // the user is still typing. They are parsed once, on submit.
  const [a, setA] = useState('')
  const [b, setB] = useState('')
  const [status, setStatus] = useState<Status>({ kind: 'idle' })

  const unary = isUnary(operation)
  const loading = status.kind === 'loading'

  // Any edit invalidates the answer on screen, so the status region is cleared
  // rather than left showing a result computed from different inputs.
  function handleOperationChange(event: ChangeEvent<HTMLSelectElement>) {
    // The cast is safe because every <option> below is generated from OPERATIONS,
    // so the only values this element can produce are Operation values. The DOM
    // types have no way to know that — event.target.value is always string.
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
    // Without this the browser does a full-page form GET and the app reloads.
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
      // api.ts turns every expected failure into an ApiError with a message meant
      // for display. Anything else is a bug on this side, so it gets a generic
      // message instead of leaking a stack trace into the UI.
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
              // React needs a stable `key` per list item to tell entries apart
              // across renders. The operation name is naturally unique here.
              <option key={op.value} value={op.value}>
                {op.label}
              </option>
            ))}
          </select>
        </label>

        {/*
          type="text" rather than type="number", deliberately. A number input
          applies the HTML spec's value sanitisation: anything that is not a
          valid floating-point number is silently replaced with an empty string,
          so "abc" and "1e400" never reach our validation at all — the user just
          watches their typing disappear with no explanation. Holding the raw
          string means the messages below can actually say what is wrong.
          inputMode="decimal" keeps the numeric keypad on mobile, which is the
          only thing worth having from type="number" here.
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
            value={b}
            onChange={handleBChange}
            // sqrt is unary. The field stays visible so the layout does not jump
            // when the operation changes; disabling makes it plainly inert.
            disabled={unary}
          />
        </label>

        <button type="submit" disabled={loading}>
          {loading ? 'Calculating…' : 'Calculate'}
        </button>
      </form>

      {/*
        One region for all three outcomes, since they are mutually exclusive.
        role="status" implies aria-live="polite", so a screen reader announces
        whatever appears here without interrupting. The container is always
        rendered, empty or not — a live region has to be in the DOM before its
        content changes, or the change goes unannounced.

        The result is printed as-is. The backend's float64 round-trips exactly
        through JSON into a JS number, and rounding for looks would hide real
        floating-point behaviour that a calculator's user is entitled to see.
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
