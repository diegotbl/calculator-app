# Technical debts

Known gaps and shortcuts, recorded so they are visible in review rather than discovered in it.

[DECISIONS.md](DECISIONS.md) records what was decided and why. This file records what is
**not done**, split into two kinds: gaps that should be closed, and omissions that were considered
and deliberately left out. The second list matters as much as the first — it is the difference
between "missed it" and "weighed it".

Status as of the end of the backend implementation: the Go backend is complete and tested; the
frontend is still Vite scaffolding.

---

## Open gaps

### D1 — The Vite dev proxy is documented but not configured

**Severity: high — the documentation currently describes behavior that does not exist.**

`frontend/vite.config.ts` has no `server.proxy` block, but two documents say it does:

- DECISIONS.md **T4** gives the dev proxy as the reason the backend carries no CORS code.
- README setup says the Vite dev server "proxies `/calculate` to :8080".

So today the backend is reachable by `curl` but not from a browser on `:5173` — a request to
`/calculate` would be served by Vite itself and 404. Nothing is broken in the backend; the
integration piece the backend's design depends on was simply never added.

Fix: a `server.proxy` entry in `vite.config.ts` mapping `/calculate` to `http://localhost:8080`.
Small, but it must land before any frontend work can talk to the API.

### D2 — Vite scaffold placeholders are still in the repo

`frontend/src/` still holds the generated demo app — `App.tsx`, `App.css`, `assets/react.svg` —
and `App.test.tsx`, which was written in session 1 only to prove the Vitest wiring worked. None of
it is part of the calculator.

Fix: delete them as `Calculator.tsx` and `Calculator.test.tsx` land. Leaving a throwaway smoke test
in a submission invites the reviewer to read it as a real test.

### D3 — Frontend dependencies are pinned to the Node 20 era

The frontend was scaffolded under Node 20.10, so `vite` 5, `vitest` 2, `jsdom` 25, and React 18 are
pinned to that generation. The project now runs on Node 22, which lifts the constraint that forced
those pins. Already noted in the README's toolchain note.

Fix: bump to current majors and re-run the suite. Deferred because it is churn with no functional
gain for a take-home, and a failed upgrade costs more time than it saves.

### D4 — Frontend coverage is an unfilled placeholder

README § Coverage lists real backend numbers and `_to be filled in_` for the frontend. Resolves
itself when the frontend tests exist; listed so it is not forgotten at submission time.

---

## Considered and deliberately not done

These are not oversights. Each was weighed against CLAUDE.md's instruction not to introduce
machinery this project's scope does not need.

### N1 — No graceful shutdown

`main()` does not trap `SIGINT`/`SIGTERM` to drain in-flight requests before exiting.

For a stateless service whose longest request is a floating-point multiply, there is nothing to
drain — no open transactions, no queued work, no external calls to finish. On a real service with a
database this would be required; here it would be ceremony.

### N2 — No request body size limit

The handler reads the whole body with `io.ReadAll` and no `http.MaxBytesReader` cap, so a
deliberately huge payload is read into memory before being rejected as invalid JSON.

The server's `ReadTimeout` bounds how long that can go on, and nothing in the spec calls for a
limit. On anything internet-facing, `http.MaxBytesReader` with a few KB would be the standard
guard, and it is a one-line addition if wanted.

### N3 — `main()` is not covered by tests

`cmd/server` sits at 48% coverage. The uncovered statements are all inside `main()`, which
constructs the `http.Server` and blocks on `ListenAndServe`.

`router()` and `port()` — the parts with actual branches — were split out precisely so they could
be tested, and both are covered. Restructuring further (injecting a listener, running the server in
a goroutine) to chase the number would add indirection to make a coverage report look better, which
is the wrong trade. Also explained in README § Coverage.
