// Typed client for the backend's single endpoint, POST /calculate.
//
// The types below mirror the Go structs in backend/internal/handler/handler.go,
// so the wire format is described once on each side and a change to one is an
// obvious prompt to change the other.

/**
 * The seven operations the backend accepts (DECISIONS.md B1).
 *
 * TS idiom: a union of string literals is TypeScript's usual stand-in for a Java
 * enum. It exists only at compile time — there is no runtime object — so what
 * actually goes on the wire is a plain string. The payoff is that a typo like
 * `'ad'` fails to compile, and editors autocomplete the seven valid values.
 */
export type Operation =
  | 'add'
  | 'subtract'
  | 'multiply'
  | 'divide'
  | 'power'
  | 'sqrt'
  | 'percentage'

/**
 * Request body. `b?` marks the field optional: `sqrt` is unary, and
 * JSON.stringify omits a property whose value is undefined, so a sqrt request
 * goes out as {"operation":"sqrt","a":9} (DECISIONS.md B3).
 */
export interface CalcRequest {
  operation: Operation
  a: number
  b?: number
}

/** Success body: {"result": <number>} */
export interface CalcResponse {
  result: number
}

/**
 * A failed call, carrying a message that is safe to render to the user as-is.
 *
 * The backend already writes its errors for humans ("division by zero"), so the
 * UI shows `error.message` directly instead of keeping its own table of status
 * codes to copy. Network failures get a message written here, since there is no
 * server response to quote.
 *
 * TS idiom: subclassing the built-in Error is reliable here because we compile
 * to ES2020. (Targeting ES5 breaks `instanceof` on subclassed built-ins — a
 * well-known TypeScript gotcha, and the reason this note exists.)
 */
export class ApiError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'ApiError'
  }
}

const NETWORK_ERROR_MESSAGE = 'Could not reach the server. Is the backend running?'

/**
 * Narrows an unknown JSON value to "some object with string keys", after which
 * the compiler lets us probe individual fields with typeof.
 *
 * TS idiom: the `value is X` return type makes this a type predicate — inside an
 * `if (isObject(v))` branch the compiler treats `v` as X. It is how TypeScript
 * expresses a shape check for a type that has no runtime class to instanceof.
 */
function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

/**
 * Sends one calculation to the backend.
 *
 * Resolves with the result, or rejects with an ApiError whose message is ready
 * to display. The URL is relative so the browser makes a same-origin request:
 * in development Vite proxies /calculate to :8080 (DECISIONS.md T4), and it
 * keeps working unchanged if both are ever served from one origin.
 */
export async function calculate(req: CalcRequest): Promise<CalcResponse> {
  let response: Response
  try {
    response = await fetch('/calculate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    })
  } catch {
    // fetch rejects only when the request never completed — connection refused,
    // DNS failure, offline. An HTTP 400 is a *successful* fetch carrying a
    // non-ok status, and is handled below. Conflating the two is the classic
    // fetch mistake, hence the separate try block.
    throw new ApiError(NETWORK_ERROR_MESSAGE)
  }

  // Every documented response, success or error, has a JSON body. A parse
  // failure (an empty or non-JSON body from something in between) is not worth
  // crashing on — it falls through to the generic messages below.
  //
  // Typing this `unknown` rather than letting it stay `any` (what response.json()
  // returns) is what forces the shape checks that follow.
  const body: unknown = await response.json().catch(() => undefined)

  if (!response.ok) {
    throw new ApiError(
      isObject(body) && typeof body.error === 'string'
        ? body.error
        : `Request failed (${response.status})`,
    )
  }

  if (!isObject(body) || typeof body.result !== 'number') {
    throw new ApiError('Unexpected response from the server.')
  }
  return { result: body.result }
}
