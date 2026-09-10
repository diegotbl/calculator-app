// Typed client for the backend's single endpoint, POST /calculate. The types
// mirror the Go structs in backend/internal/handler.

export type Operation =
  | 'add'
  | 'subtract'
  | 'multiply'
  | 'divide'
  | 'power'
  | 'sqrt'
  | 'percentage'

// b is optional: sqrt is unary, and JSON.stringify omits an undefined property,
// so a sqrt request goes out as {"operation":"sqrt","a":9}.
export interface CalcRequest {
  operation: Operation
  a: number
  b?: number
}

export interface CalcResponse {
  result: number
}

// A failed call, carrying a message safe to render as-is. The backend writes its
// errors for humans ("division by zero"); a failed connection gets the message
// below, since there is no server response to quote.
export class ApiError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'ApiError'
  }
}

const NETWORK_ERROR_MESSAGE = 'Could not reach the server. Is the backend running?'

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

// Sends one calculation. Resolves with the result, or rejects with an ApiError
// whose message is ready to display. The URL is relative so the browser makes a
// same-origin request; Vite proxies it to the backend in development.
export async function calculate(req: CalcRequest): Promise<CalcResponse> {
  let response: Response
  try {
    response = await fetch('/calculate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    })
  } catch {
    // fetch rejects only when the request never completed. An HTTP 400 is a
    // successful fetch with ok === false, handled below.
    throw new ApiError(NETWORK_ERROR_MESSAGE)
  }

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
