import { ApiError, calculate } from './api'

// These tests are about how api.ts maps a response onto a result or an
// ApiError, so fetch is replaced wholesale.
const mockFetch = vi.fn()
beforeEach(() => {
  mockFetch.mockReset()
  vi.stubGlobal('fetch', mockFetch)
})
afterEach(() => {
  vi.unstubAllGlobals()
})

/** A real Response, so .ok, .status and .json() behave as they do in a browser. */
function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), { status })
}

describe('calculate', () => {
  it('posts the operation and operands as JSON', async () => {
    mockFetch.mockResolvedValue(jsonResponse(200, { result: 5 }))

    await calculate({ operation: 'add', a: 2, b: 3 })

    expect(mockFetch).toHaveBeenCalledWith('/calculate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: '{"operation":"add","a":2,"b":3}',
    })
  })

  it('omits b for a unary operation', async () => {
    mockFetch.mockResolvedValue(jsonResponse(200, { result: 3 }))

    await calculate({ operation: 'sqrt', a: 9 })

    // JSON.stringify drops an undefined property, so b never reaches the wire.
    expect(mockFetch.mock.calls[0][1].body).toBe('{"operation":"sqrt","a":9}')
  })

  it('returns the parsed result on success', async () => {
    mockFetch.mockResolvedValue(jsonResponse(200, { result: 5 }))

    await expect(calculate({ operation: 'add', a: 2, b: 3 })).resolves.toEqual({
      result: 5,
    })
  })

  // The messages below are the backend's own, passed through untouched.
  it.each([
    { status: 400, error: 'division by zero' },
    { status: 400, error: 'square root of a negative number is undefined' },
    { status: 400, error: 'operand "a" is required' },
    { status: 404, error: 'not found' },
    { status: 405, error: 'method not allowed' },
  ])('surfaces the backend message from a $status response', async ({ status, error }) => {
    mockFetch.mockResolvedValue(jsonResponse(status, { error }))

    await expect(calculate({ operation: 'divide', a: 1, b: 0 })).rejects.toThrow(
      new ApiError(error),
    )
  })

  it('falls back to the status code when an error body is not JSON', async () => {
    mockFetch.mockResolvedValue(new Response('<html>gateway</html>', { status: 502 }))

    await expect(calculate({ operation: 'add', a: 2, b: 3 })).rejects.toThrow(
      'Request failed (502)',
    )
  })

  it('rejects a 200 whose body is not the documented shape', async () => {
    mockFetch.mockResolvedValue(jsonResponse(200, { unexpected: true }))

    await expect(calculate({ operation: 'add', a: 2, b: 3 })).rejects.toThrow(
      'Unexpected response from the server.',
    )
  })

  it('reports a failed connection as a network error, not a calculation error', async () => {
    // fetch rejects only when the request never completed — not to be confused
    // with an HTTP 400.
    mockFetch.mockRejectedValue(new TypeError('Failed to fetch'))

    await expect(calculate({ operation: 'add', a: 2, b: 3 })).rejects.toThrow(
      'Could not reach the server. Is the backend running?',
    )
  })

  it('throws ApiError specifically, so callers can tell expected failures apart', async () => {
    mockFetch.mockResolvedValue(jsonResponse(400, { error: 'division by zero' }))

    await expect(calculate({ operation: 'divide', a: 1, b: 0 })).rejects.toBeInstanceOf(
      ApiError,
    )
  })
})
