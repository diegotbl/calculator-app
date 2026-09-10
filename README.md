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
- Responsive single-column layout with basic mobile support (fluid width, numeric keypad on
  touch devices) — see [DECISIONS.md T19](DECISIONS.md)
- Clear, human-readable error messages for invalid input and undefined results (division by
  zero, square root of a negative number, overflow)

## Setup

Prerequisites: **Go 1.27+** and **Node 22 LTS**. To run the app without either, see
[Run with Docker](#run-with-docker) below.

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

## Run with Docker

Requires only Docker (Desktop, or Engine with the Compose v2 plugin) — no Go and no Node on the
host. From the repository root:

```
docker compose up --build      # builds both images and starts them
```

Then open **http://localhost:8080**. Stop with `Ctrl+C`, or from another terminal:

```
docker compose down
```

Two services, one published port:

```
browser ──▶ frontend (nginx, :8080) ─┬─ /           → the built React app
                                     └─ /calculate  → backend (Go, :8080, internal only)
```

The frontend image builds `dist/` with Node and serves it from `nginx:alpine`; the backend image
compiles a static binary and ships it `FROM scratch` (~6.6 MB, no shell or package manager
inside). nginx reverse-proxies `/calculate` over the compose network, so the browser makes a
same-origin request exactly as it does behind the Vite dev proxy — the same reason the backend
needs no CORS code (see [DECISIONS.md T20](DECISIONS.md)).

The backend publishes no host port: it is reachable only through the proxy. The API examples
below therefore work unchanged against `localhost:8080` while the stack is up.

> Port `8080` is the *whole app* here, while `go run ./cmd/server` uses `8080` for the API alone.
> Stop a locally running backend first, or change the left side of the `ports` mapping in
> `compose.yaml`.

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

## Design decisions & assumptions

The summary below is enough to review the shape of the solution; every point is expanded, with
the alternatives that were weighed, in [DECISIONS.md](DECISIONS.md) (referenced by ID, e.g.
`T7`). That file is the single source of truth for the business rules and the full
error-response spec.

**Assumptions**

- `percentage` means "*a* percent of *b*" → `(a / 100) * b`, e.g. `percentage(15, 200) = 30`
  (`B2`).
- `sqrt` is the only unary operation; sending `b` with it is silently ignored, not an error
  (`B3`).
- A mathematically undefined or out-of-range request (÷0, √negative, overflow to `±Inf`/`NaN`)
  is the *client's* mistake → `400`, never `500` (`B4`, `B5`).
- No auth, rate limiting, persistence, or multi-tenancy is in scope — it is a stateless
  compute endpoint (`T3`).

**Backend**

- Go standard library only — `net/http` + `encoding/json`; one endpoint doing arithmetic does
  not need a framework (`T1`).
- Two packages: `internal/calculator` is pure `float64 → (float64, error)` with no HTTP
  knowledge; `internal/handler` owns decoding, validation, status codes, JSON (`T8`).
- Operation dispatch is an explicit `map[string]func`, not a registry or reflection (`T7`).
- Request fields decode as `json.RawMessage`, then operands via `json.Number`, so the handler
  can tell "absent" from "wrong type" from "not finite" and return the right message for each
  (`T6`).
- Request body capped at 1 MiB via `http.MaxBytesReader` → `413` (`T18`).

**Frontend**

- Validation runs on both sides; the backend re-checks everything and wins any disagreement.
  The client copies the backend's exact error wording so the user cannot tell which layer
  rejected the input (`T9`, `T14`).
- One explicit **Calculate** button — no recalculation per keystroke, which would need
  debouncing and out-of-order handling (`T11`).
- Operand fields are `type="text"` + `inputMode="decimal"`, not `type="number"`, whose value
  sanitisation would swallow invalid input before validation could explain it (`T13`).
- Status region is one discriminated union (`idle | loading | result | error`), so impossible
  combinations like "loading while showing an error" are unrepresentable (`T15`).
- `api.ts` surfaces the backend's message rather than mapping status codes to its own copy
  (`T16`).

**Integration**

- Dev cross-origin is handled by a Vite dev proxy, not backend CORS; the frontend calls the
  relative URL `/calculate`, which also works if both are served from one origin (`T4`).
- Docker runs the two halves as two images behind an nginx reverse proxy rather than baking both
  into one — the proxy reproduces the dev arrangement in production, so no application code
  changes shape for deployment (`T20`).

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

### Examples

```
# success
$ curl -s -X POST localhost:8080/calculate \
    -H 'Content-Type: application/json' \
    -d '{"operation":"multiply","a":6,"b":7}'
{"result":42}

# unary operation — b omitted
$ curl -s -X POST localhost:8080/calculate \
    -H 'Content-Type: application/json' \
    -d '{"operation":"sqrt","a":144}'
{"result":12}

# error — HTTP 400
$ curl -s -X POST localhost:8080/calculate \
    -H 'Content-Type: application/json' \
    -d '{"operation":"divide","a":1,"b":0}'
{"error":"division by zero"}
```

During development the same calls work against the Vite dev server on `localhost:5173`, which
proxies `/calculate` to the backend.

### Error summary

| Status | When | Examples |
| ------ | ---- | -------- |
| `404`  | Unknown path | anything other than `/calculate` |
| `405`  | Wrong method | `GET /calculate` (response sets `Allow: POST`) |
| `413`  | Body too large | request body over 1 MiB |
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

Backend — `cd backend && go test ./... -cover` for the per-package summary. For a detailed
report, write a coverage profile first and then render it:

```
cd backend
go test ./... -coverprofile=coverage.out   # write the profile
go tool cover -func=coverage.out           # per-function breakdown + total, in the terminal
go tool cover -html=coverage.out           # annotated source, opens in a browser
```

`-func` prints one line per function plus an overall total (88.7% of statements); `-html` colours
each line of source green (covered) or red (not covered), which is the quickest way to see *which*
lines in `main()` are missed. `coverage.out` is a build artifact — it is gitignored.

Per-package summary:

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

33 tests across `src/api.test.ts` (the client's response-to-result and
response-to-error mapping) and `src/Calculator.test.tsx` (rendering, input validation, and the
result / error / loading states).

`src/main.tsx` is excluded from the report — see `test.coverage.exclude` in `vite.config.ts`. It
only mounts the React root and has no branches, which is the same reason `main()` is uncovered on
the Go side.

### Toolchain note

Built and tested on Node 22 LTS (managed via [nvm-windows](https://github.com/coreybutler/nvm-windows)).
The frontend was initially scaffolded under Node 20.10, so `vite` 5 / `vitest` 2 / `jsdom` 25 are
still pinned to that era; they can be bumped to current majors now that Node 22 is in use.

`npm audit` flags advisories in this pinned `vite` / `vitest` toolchain; `npm audit --omit=dev`
reports none. They are all `devDependency`-only dev-server / test-runner issues with no effect on
the built output or the runtime dependencies. See
[TECHNICAL_DEBTS.md D3](TECHNICAL_DEBTS.md).
