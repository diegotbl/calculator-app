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

### T6 — Request fields decode as `json.RawMessage`, operands then via `json.Number`

A plain `float64` struct field defaults to `0`, so the handler cannot distinguish "client omitted
`a`" from "client sent `a: 0`". The obvious fix is `*float64` — nil means absent. But the error
table below needs **three** outcomes per operand, and a `*float64` only gives two:

| Client sent | Required message |
| ----------- | ---------------- |
| nothing, or `null` | `operand "a" is required` |
| `"x"`, `true` | `operand "a" must be a number` |
| `1e400` | `operand "a" must be a finite number` |

So `request` holds each field as `json.RawMessage` (the undecoded bytes of that JSON value) and
decodes them one at a time. Absent stays nil; a wrong-typed value fails its individual unmarshal;
an out-of-range one is caught by the finite check.

The last case needs one more step. Unmarshalling `1e400` straight into a `float64` returns *both*
`+Inf` and an `UnmarshalTypeError`, so that path would report it as `must be a number` — the wrong
message. Decoding into a `json.Number` (a string alias holding the literal) and calling
`strconv.ParseFloat` ourselves gives `+Inf` with `ErrRange`, which the finite check turns into the
right one.

The same treatment on `operation` is what separates `operation must be a string` (sent `5`) from
`operation is required` (sent nothing, `null`, or `""`).

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

### T11 — One calculation per submit, not live recalculation

The UI is a form with an explicit **Calculate** button. It does not recompute as the user types.

Recalculating on every keystroke would mean one HTTP request per character, which needs debouncing
and out-of-order response handling to behave — machinery this app has no reason to carry. It also
maps cleanly onto the API: one `POST /calculate` per calculation, which is what the endpoint is.

### T12 — `sqrt`'s second operand is disabled, not hidden

Selecting `sqrt` disables the `b` input rather than removing it. Hiding it would reflow the form
every time the operation changed; disabling keeps the layout still while making the field plainly
inert. `b` is omitted from the request entirely (B3), not sent as `0`.

### T13 — Operand inputs are `type="text"` with `inputMode="decimal"`, not `type="number"`

The obvious choice is `<input type="number">`. It is the wrong one here.

A number input applies the HTML spec's *value sanitisation*: any value that is not a valid
floating-point number is silently replaced with the empty string. So `abc` and `1e400` never reach
our validation at all — the field simply goes blank and the user watches their typing vanish with
no explanation. Two of the three messages in T14 would have been unreachable prose, and T9's claim
that the frontend blocks non-numeric text would have been false.

Holding the raw string lets the validation actually run and say what is wrong. `inputMode="decimal"`
keeps the numeric keypad on mobile, which is the only thing worth having from `type="number"`; the
spinner arrows, and their habit of changing the value on a stray scroll, are no loss.

Found while writing the tests: the branches were provably dead, which is also why frontend coverage
reaches 100% only after this change.

### T14 — Client-side validation messages copy the backend's wording

`operand "a" is required`, `operand "b" is required for operation "add"`, `must be a number`,
`must be a finite number` — the frontend uses the backend's exact strings from § Error responses.

The user should not be able to tell which layer rejected the input; the frontend is simply faster
to say it. The backend still re-validates everything and wins any disagreement (T9). The cost is
that the two lists must be kept in step, which is why both point at the same table here.

### T15 — UI state is one discriminated union, not several booleans

The status region is driven by a single value: `idle`, `loading`, `result`, or `error`. Separate
`isLoading` / `error` / `result` fields would allow "loading while showing an error" and similar
states that should not exist, and would need coordinated resets to avoid them. One value makes
those combinations unrepresentable instead of merely avoided.

### T16 — `api.ts` raises an error carrying the backend's message, not its status code

`ApiError.message` holds the text the server sent, ready to display. The UI does not map status
codes to its own copy, which would make the error spec a second source of truth alongside
§ Error responses.

The one message written on the client is for a failed connection, because there is no response to
quote. `fetch` rejects only when the request never completed; an HTTP `400` is a *successful* fetch
with `ok === false`. Those are handled separately — conflating them would report `division by zero`
as a network outage.

### T17 — The UI's operation list mirrors the backend map

`src/operations.ts` holds the dropdown entries so `Calculator.tsx` does not hardcode the backend's
vocabulary. It duplicates the `operations` map in `internal/calculator`, deliberately: the backend
re-checks the operation name and arity on every request (T9), so drift between the two is a UI bug,
never a correctness hole. Generating one from the other would need a schema endpoint or a build
step, which is more machinery than seven fixed operations justify.

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
| `a` or `b` is outside the `float64` range (`1e400` parses to `±Inf`)         | `400`  | `operand "a" must be a finite number`          |

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
