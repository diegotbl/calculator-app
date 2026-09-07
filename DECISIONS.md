# Decisions

Business and technical decisions for the calculator app, with the reasoning behind each so it can
be explained in code review / interview follow-ups. README.md carries only a short summary and
points here for detail.

---

## Business rules

### B1 — Supported operations

`add`, `subtract`, `multiply`, `divide`, `power` (`a^b`), `sqrt` (√a), `percentage`.

The four basic arithmetic operations are the core requirement; `power`, `sqrt`, and `percentage`
are the "extended operations" from the assignment.

### B2 — `percentage` means "a percent of b" → `(a / 100) * b`

`percentage(15, 200) = 30` ("what is 15% of 200?").

Alternative considered: "a as a percent of b" → `(a / b) * 100`. Rejected because it is less
common as a calculator button and it introduces a division-by-zero case. The chosen formula has
no undefined inputs of its own (only the generic non-finite-result guard applies).

### B3 — `b` is optional for every operation, ignored for `sqrt`

`sqrt` is unary. Sending `b` with a `sqrt` request is not an error — it is silently ignored — so a
client that always sends both fields does not have to special-case the payload. `b` is required
for all six binary operations (see T7).

### B4 — Mathematically undefined results are client errors, not server errors

Division by zero and square root of a negative number return `400`, not `500`. The client asked
for something with no defined answer; the server did not fail. Same principle for a result that
overflows the range of a 64-bit float (see B5).

### B5 — Results must be finite real numbers

If an operation produces `±Inf` or `NaN` (overflow, `power(-2, 0.5)`, `0^-1`, …) the API rejects
it with `400` rather than returning a non-numeric value. The API contract is "numbers in, a
number out"; `Infinity`/`NaN` are not valid JSON numbers anyway (Go's `encoding/json` refuses to
encode them).

---

## Technical decisions

### T1 — Backend: Go standard library only

`net/http` for routing, `encoding/json` for the wire format. The service is one endpoint doing
pure computation; a web framework would add dependency and concept overhead with nothing to show
for it.

### T2 — Frontend: React + TypeScript via Vite; Vitest for tests

Vite is the current standard React toolchain. Vitest is Vite-native and uses the same
`describe`/`it`/`expect` API as Jest, so there is no second build/transform config to maintain.
(Originally CLAUDE.md said Jest; changed to Vitest early — see PROMPTS.md session 1.)

### T3 — No database, no state

Every request is self-contained: operation + operands in, result out. Nothing to persist.

### T4 — Dev cross-origin handled by a Vite dev proxy, not backend CORS

In development the frontend runs on `localhost:5173` and the backend on `localhost:8080`. Rather
than adding `Access-Control-Allow-Origin` handling and an `OPTIONS` preflight branch to the Go
handler, `vite.config.ts` proxies `/calculate` to the backend. The browser then sees a
same-origin request, and the backend stays free of CORS concerns. The frontend calls a relative
URL (`/calculate`), which also works unchanged if both are served from one origin in production.

### T5 — Server port is `8080`, overridable via `PORT`

A fixed sensible default with an environment-variable escape hatch. No config file for one value.

### T6 — Request operands are `*float64` (nullable), not `float64`

A plain `float64` struct field defaults to `0`, so the handler cannot distinguish "client omitted
`a`" from "client sent `a: 0`". A nil pointer means "absent" and is reported as a missing-operand
error; a non-nil pointer to `0` is a valid operand.

### T7 — Operation dispatch is an explicit `map[string]func`

Operation name → function, looked up once. No plugin registry, no reflection — for seven
operations that machinery would be harder to read than the thing it replaces.

### T8 — Layering: `calculator` (pure) vs `handler` (HTTP)

`internal/calculator` has no knowledge of HTTP — it takes `float64`s and returns
`(float64, error)`. `internal/handler` owns request parsing, validation, status codes, and JSON
shaping. This keeps the operation logic trivially unit-testable and the HTTP concerns in one
place.

### T9 — Validation lives on both sides, backend is authoritative

The frontend blocks obviously invalid input (empty fields, non-numeric text) for immediate
feedback. The backend re-validates everything because it must never trust the client. Where they
disagree, the backend wins.

### T10 — Unknown JSON fields are ignored

The decoder does not use `DisallowUnknownFields`. Extra keys in the payload are tolerated rather
than rejected — lenient in what we accept. (Easy to tighten later if strictness is wanted.)

---

## Error responses

Wire format: every error is `{"error": "<message>"}` as JSON. Success is `{"result": <number>}`.

Only routing errors deviate from HTTP `400`. A malformed or unanswerable request from the client
is a `400`, never a `500` (see B4). `500` is reachable only via an unexpected panic, which a
`recover` in the handler converts to `{"error": "internal server error"}` — no documented input
reaches it.

### Routing layer (`cmd/server/main.go`)

| Condition                       | Status | Message              |
| ------------------------------- | ------ | -------------------- |
| Path is not `/calculate`        | `404`  | `not found`          |
| Method is not `POST`            | `405`  | `method not allowed` (response also sets `Allow: POST`) |

### Request decoding (`internal/handler`)

| Condition                                             | Status | Message                          |
| ---------------------------------------------------- | ------ | -------------------------------- |
| Empty body                                           | `400`  | `request body is empty`          |
| Malformed JSON (syntax error)                        | `400`  | `invalid JSON in request body`   |
| Wrong JSON type for an operand (`"a": "x"`, `true`)  | `400`  | `operand "a" must be a number`   |
| `operation` is not a JSON string                     | `400`  | `operation must be a string`     |

### Semantic validation (`internal/handler`, checked in this order)

| Condition                                                                    | Status | Message                                         |
| --------------------------------------------------------------------------- | ------ | ---------------------------------------------- |
| `operation` missing or empty                                                 | `400`  | `operation is required`                        |
| `operation` not one of the seven                                             | `400`  | `unknown operation "cube"`                     |
| `a` missing (nil)                                                            | `400`  | `operand "a" is required`                      |
| `b` missing for a binary operation                                           | `400`  | `operand "b" is required for operation "add"`  |
| `a` or `b` is non-finite (`1e400` decodes to `+Inf`)                         | `400`  | `operand "a" must be a finite number`          |

### Operation preconditions (`internal/calculator`, checked before computing)

| Operation  | Condition | Status | Message                                            |
| ---------- | --------- | ------ | ------------------------------------------------- |
| `divide`   | `b == 0`  | `400`  | `division by zero`                                |
| `sqrt`     | `a < 0`   | `400`  | `square root of a negative number is undefined`   |

### Result check (`internal/calculator`, all operations)

| Condition                                  | Status | Message                          |
| ------------------------------------------ | ------ | -------------------------------- |
| Result is `±Inf` or `NaN` after computing  | `400`  | `result is not a finite number`  |

Covers overflow of `add` / `subtract` / `multiply` / `power` / `percentage` to `±Inf`, `power`
producing `NaN` (e.g. `power(-2, 0.5)`), and `0^-1` → `+Inf`.
