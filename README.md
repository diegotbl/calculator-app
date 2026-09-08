# Calculator App

A small full-stack calculator. A React + TypeScript frontend collects an operation and its
operands and sends them to a Go HTTP backend, which performs the computation and returns the
result. The backend is the source of truth; the app holds no state and uses no database.

For the reasoning behind the business rules, API shape, and error handling, see
[DECISIONS.md](DECISIONS.md).

## Features

- Arithmetic: **add**, **subtract**, **multiply**, **divide**
- Extended: **power** (`a^b`), **square root** (`sqrt` of `a`), **percentage** (`a` percent of
  `b`, i.e. `(a / 100) * b`)
- Input validation on both the client (immediate feedback) and the server (authoritative)
- Clear, human-readable error messages for invalid input and undefined results (division by
  zero, square root of a negative number, overflow)

## Setup

Prerequisites: **Go 1.27+** and **Node 22 LTS**.

Backend:

```
cd backend
go run ./cmd/server        # serves on :8080 (override with PORT)
```

Frontend (in a second terminal):

```
cd frontend
npm install
npm run dev                # Vite dev server on :5173, proxies /calculate to :8080
```

Open the URL Vite prints. The frontend must reach the backend, so start the backend first.

## Architecture

```
frontend (React + TS, Vite)  ──POST /calculate──▶  backend (Go, net/http)
  src/api.ts          typed client                   internal/handler     request parsing + validation
  src/operations.ts   operations shown in the UI     internal/calculator  pure operation logic
  src/Calculator.tsx  form, validation, rendering
```

- **`internal/calculator`** knows nothing about HTTP: it takes `float64`s and returns
  `(float64, error)`. Each operation is unit-tested in isolation.
- **`internal/handler`** owns JSON decoding, validation, status codes, and response shaping.
  Operations are dispatched through an explicit name → function map.
- **`src/api.ts`** is the only place that speaks HTTP on the frontend. It turns a non-2xx response
  into an error carrying the backend's own message, so the UI displays that message rather than
  keeping a second copy of the error spec.
- **`src/Calculator.tsx`** holds the form state and validates before sending. Its messages match
  the backend's wording exactly, so the user cannot tell which layer rejected the input.
- In development, `vite.config.ts` proxies `/calculate` to the backend so the browser sees a
  same-origin request and the backend needs no CORS code.

## API

Single endpoint: `POST /calculate`

Request:

```json
{ "operation": "add", "a": 2, "b": 3 }
```

`b` is optional for `sqrt` (ignored) and required for the other six operations.

Success (`200`):

```json
{ "result": 5 }
```

Error (`4xx`):

```json
{ "error": "division by zero" }
```

### Error summary

| Status | When | Examples |
| ------ | ---- | -------- |
| `404`  | Unknown path | anything other than `/calculate` |
| `405`  | Wrong method | `GET /calculate` (response sets `Allow: POST`) |
| `400`  | Malformed request | empty body, invalid JSON, operand that isn't a number |
| `400`  | Invalid input | missing `operation`, unknown operation, missing operand, non-finite operand |
| `400`  | Undefined / out-of-range result | division by zero, `sqrt` of a negative number, overflow to `±Inf`/`NaN` |

A bad request from the client is always a `4xx`, never a `500`. Full table with exact messages:
[DECISIONS.md § Error responses](DECISIONS.md#error-responses).

## Tests

Backend uses Go's `testing` package (table-driven). Frontend uses **Vitest** (Vite-native,
Jest-compatible API) with React Testing Library and `jsdom`.

```
cd backend  && go test ./... -cover
cd frontend && npm test              # one-off run
cd frontend && npm run test:coverage # run + coverage summary
```

### Coverage

Backend — `cd backend && go test ./... -cover`:

| Package | Coverage |
| ------- | -------- |
| `internal/calculator` | **100.0%** |
| `internal/handler` | **100.0%** |
| `cmd/server` | 48.0% |

`cmd/server` is lower because `main()` itself — building the `http.Server` and blocking on
`ListenAndServe` — can't be exercised from a unit test. The logic that has branches, `router()`
and `port()`, is fully covered by `cmd/server/main_test.go`.

Frontend — `cd frontend && npm run test:coverage`:

| File | Statements | Branches | Functions | Lines |
| ---- | ---------- | -------- | --------- | ----- |
| `src/api.ts` | **100%** | **100%** | **100%** | **100%** |
| `src/operations.ts` | **100%** | **100%** | **100%** | **100%** |
| `src/Calculator.tsx` | **100%** | **100%** | **100%** | **100%** |

31 tests across `src/api.test.ts` (the client's response-to-result and
response-to-error mapping) and `src/Calculator.test.tsx` (rendering, input validation, and the
result / error / loading states).

`src/main.tsx` is excluded from the report — see `test.coverage.exclude` in `vite.config.ts`. It
only mounts the React root and has no branches, which is the same reason `main()` is uncovered on
the Go side.

### Toolchain note

Built and tested on Node 22 LTS (managed via [nvm-windows](https://github.com/coreybutler/nvm-windows)).
The frontend was initially scaffolded under Node 20.10, so `vite` 5 / `vitest` 2 / `jsdom` 25 are
still pinned to that era; they can be bumped to current majors now that Node 22 is in use.
