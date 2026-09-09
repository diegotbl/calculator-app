# Technical debts

Known gaps and shortcuts, recorded so they are visible in review rather than discovered in it.

[DECISIONS.md](DECISIONS.md) records what was decided and why. This file records what is
**not done**, split into two kinds: gaps that should be closed, and omissions that were considered
and deliberately left out. The second list matters as much as the first — it is the difference
between "missed it" and "weighed it".

Status: the Go backend and the React frontend are both implemented, tested and documented. One
gap remains open — D3 — not blocking, and recorded below with what it would take to close.

---

## Open gaps

### D3 — Frontend dependencies are pinned to the Node 20 era

The frontend was scaffolded under Node 20.10, so `vite` 5, `vitest` 2, `jsdom` 25, and React 18 are
pinned to that generation. The project now runs on Node 22, which lifts the constraint that forced
those pins. Already noted in the README's toolchain note.

Fix: bump to current majors and re-run the suite. Deferred because it is churn with no functional
gain for a take-home, and a failed upgrade costs more time than it saves.

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

---

## Resolved

### D5 — The favicon was still Vite's default logo *(fixed)*

`frontend/index.html` now links `/abacus.svg` — a small SVG in `frontend/public/` that renders the
🧮 (abacus) emoji through an SVG `<text>` element, so there is no binary asset to maintain and it
stays crisp at any tab size. `public/vite.svg` is deleted (nothing else referenced it). The
`.gitattributes` rule was extended with `*.svg text eol=lf` so the new asset does not churn between
Windows and CI checkouts, the same reason `*.go` is pinned (see D7).

### D7 — `gofmt -l` reported every Go file on Windows *(fixed)*

A `.gitattributes` at the repo root now sets `*.go text eol=lf`, so Go sources check out with LF
even on Windows and no longer read as CRLF-misformatted. The six committed `.go` files were
re-checked-out once under the new rule to normalise the working copy; `git add --renormalize .`
found nothing to change, confirming the blobs were already stored LF — the problem was purely the
Windows checkout. Verified: `gofmt -l .`, `go vet ./...` and `go test ./...` are all clean.

The false alarm was worth verifying before acting: `gofmt -l` on the working copy still cried wolf,
but extracting the files from the object store showed the committed code was correctly formatted,
so the fix was a checkout-rule change rather than rewriting six files.

### D1 — The Vite dev proxy is documented but not configured *(fixed)*

`frontend/vite.config.ts` now has a `server.proxy` entry mapping `/calculate` to
`http://localhost:8080`, so a browser on `:5173` reaches the Go backend as a same-origin request —
the integration piece DECISIONS.md T4 and the README setup section already described.

### D2 — Vite scaffold placeholders are still in the repo *(fixed)*

`App.tsx`, `App.css`, `assets/react.svg` and `App.test.tsx` are deleted. They went out in the same
commit as `Calculator.test.tsx` and `api.test.ts` deliberately: `App.test.tsx` was the only test
file, so removing it alone would have left a commit where `npm test` exits 1 with "No test files
found".

`index.html`'s `<title>` was part of the same scaffold leftover and now reads "Calculator". The
favicon is not — see D5.

### D4 — Frontend coverage is an unfilled placeholder *(fixed)*

README § Coverage now carries real frontend numbers: 31 tests across two files, 100% of statements,
branches, functions and lines for `api.ts`, `operations.ts` and `Calculator.tsx`. The table also
states that `src/main.tsx` is excluded from the report, since a bare 100% without that disclosure
would be misleading.

### D6 — Selecting `sqrt` disables operand `b` but leaves its value on screen *(fixed)*

`b` stays in component state even while disabled, so a user who picks `sqrt` by mistake and
switches back to a binary operation gets their value back. Only the *displayed* value is blanked
while the operation is unary (`value={unary ? '' : b}` in `Calculator.tsx`) — the field never shows
an inactive number next to a "disabled" state, and nothing about the request changes: `validate()`
already omitted `b` for unary operations regardless of what the box showed (DECISIONS.md B3).

Chose this over clearing `b` on the way in, which was the two-line alternative — clearing would
have lost the typed value for good on switching back, and preserving it costs one ternary.
