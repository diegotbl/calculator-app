# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

Business rules, API semantics, and the full error-response spec are recorded in
[DECISIONS.md](DECISIONS.md) — consult it before changing behavior.

# Calculator App — Take-Home Assignment (Sezzle)

## Context

This is a take-home coding assessment. It will be code-reviewed and I will be asked follow-up
interview questions about implementation details and design decisions — so favor code I can fully
explain over code that's merely clever or feature-rich. Prioritize correctness, clarity, and
idiomatic style over extra features. Budget: ~2-4 hours total.

I'm an experienced backend engineer (Java/Spring, 5+ years) but new to Go, React, and TypeScript,
so explain unfamiliar idioms briefly as you introduce them rather than assuming I already know
Go/TS conventions.

## Stack

- **Backend:** Go, standard library only (`net/http`, `encoding/json`) unless a dependency is
  clearly justified — no web framework needed for this scope.
- **Frontend:** React + TypeScript, built with Vite.
- **Tests:** Go's built-in `testing` package, table-driven style. Vitest + React Testing Library
  for the frontend (Vitest is Vite-native and Jest-API-compatible — same `describe`/`it`/`expect`,
  no separate transform/config to maintain).
- **No database** — this is a stateless calculator, computation only.

## Project structure

```
calculator-app/
  backend/
    cmd/server/main.go        # entrypoint, HTTP server setup, routing
    internal/calculator/      # pure operation logic (add, subtract, etc.) + unit tests
    internal/handler/         # HTTP handler, request/response structs, validation
    go.mod
    Dockerfile                # multi-stage: golang build -> static binary on scratch
  frontend/
    src/
      api.ts                  # typed client for the backend API
      Calculator.tsx          # main UI component
      Calculator.test.tsx
      test/setup.ts           # registers jest-dom matchers for Vitest
    vite.config.ts            # Vite config + Vitest `test` block (jsdom, coverage)
    package.json
    Dockerfile                # multi-stage: node build -> dist served by nginx
    nginx.conf                # static files + reverse proxy for /calculate
  compose.yaml                # runs both together; app on :8080, backend internal only
  README.md
  PROMPTS.md                  # log of AI prompts used, per assignment instructions
```

## API design

- Single endpoint: `POST /calculate`
- Request: `{"operation": "add" | "subtract" | "multiply" | "divide" | "power" | "sqrt" | "percentage", "a": number, "b": number}`
  (`b` is omitted/ignored for `sqrt`)
- Success response: `{"result": number}`
- Error response: `{"error": "string message"}` with appropriate 4xx status (e.g. 400 for
  malformed/missing input, 400 for division by zero — not 500; a bad request from the client is
  not a server fault).
- Keep operation dispatch simple and explicit (e.g. a switch or map of operation name → function)
  — don't over-engineer with plugin registries or reflection for four-to-seven operations.

## Coding conventions

- **Go:** idiomatic error handling (return `error`, don't panic on bad input), table-driven
  tests, small focused functions, exported names only where needed across packages. Run
  `gofmt`/`go vet` before considering something done.
- **TypeScript:** explicit types for API request/response shapes (mirror the Go structs), avoid
  `any`, functional components with hooks.
- Validate input on both the frontend (fast feedback) and backend (source of truth — never trust
  the client).
- Prefer clarity over abstraction. Don't introduce patterns (DI containers, generic op
  registries, middleware chains) that this app's scope doesn't need.

## Testing expectations

- Backend: unit tests for every operation including edge cases (division by zero, sqrt of
  negative number, overflow) and a handler-level test using `httptest` for at least one success
  and one error path.
- Frontend: component test(s) covering input validation and a rendered result/error.
- Run `go test ./... -cover` and the frontend test script with coverage before finishing; put the
  coverage summary in the README.

## Commands

- Backend: `cd backend && go run ./cmd/server` / `go test ./... -cover`
- Run a single Go test: `cd backend && go test ./internal/calculator -run TestAdd -v`
- Frontend: `cd frontend && npm install && npm run dev` / `npm test` (`npm run test:coverage`
  for the coverage summary, `npm run test:watch` while developing)
- Run a single frontend test: `cd frontend && npm test -- src/Calculator.test.tsx -t "validation"`
- Whole app in Docker: `docker compose up --build` (http://localhost:8080) / `docker compose down`

## Working style

- Explain the "why" behind non-obvious choices as you go (I'll reuse this reasoning in interviews)
  — a couple of sentences is enough, not an essay.
- After any meaningful chunk of generated code, pause for me to review before continuing to the
  next piece, rather than generating the whole app in one pass.
- Don't add CI config or extra endpoints unless I ask — those are explicitly optional in the
  assignment. (Docker was asked for and is done: see `compose.yaml` and DECISIONS.md T20. Keep it
  to those files — no Kubernetes manifests, no registry pushes, no extra services.)
- When you finish a feature, remind me to log the prompt(s) used in PROMPTS.md.
- Don't commit or push when you're done implementing. Changes will be reviewed and user will explicitely asked for commits and pushes.
