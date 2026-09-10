# Technical debts

Known gaps and shortcuts, recorded so they are visible in review rather than discovered in it.

[DECISIONS.md](DECISIONS.md) records what was decided and why. This file records what is
**not done** and why it was left that way — the difference between "missed it" and "weighed it".

Status: the Go backend and the React frontend are both implemented, tested and documented. No
gaps remain open; the items below were considered and deliberately left as they are, each weighed
against CLAUDE.md's instruction not to introduce machinery this project's scope does not need.

---

## Considered and deliberately not done

### D3 — Frontend toolchain stays on the versions Vite scaffolded

`vite` 5, `vitest` 2 and `jsdom` 25 are the versions `npm create vite` produced. They are left
as-is rather than bumped to current majors.

The upgrade is a breaking change (`vite` 8, `vitest` 5) with no functional gain for a calculator
this size, and a scaffold's pinned toolchain is not where a take-home's remaining hours are best
spent. All three versions run cleanly on the Node 22 LTS the project is built with.

`npm audit` reports 6 advisories (2 critical, 1 high, 3 moderate), all inside this toolchain —
`vite`, `vitest`, `esbuild`, `@vitest/*`. `npm audit --omit=dev` reports **0**: every advisory is
in a `devDependency`, and each one is a local dev-server or test-runner attack surface (a
malicious page reaching `localhost:5173`, the Vitest UI server, dev-server path traversal) with
nothing exposed in the built `dist/` output or the runtime dependencies (`react`, `react-dom`).
The `vite` 8 / `vitest` 5 bump clears all six if it is ever wanted.

### N1 — No graceful shutdown

`main()` does not trap `SIGINT`/`SIGTERM` to drain in-flight requests before exiting.

For a stateless service whose longest request is a floating-point multiply, there is nothing to
drain — no open transactions, no queued work, no external calls to finish. On a real service with a
database this would be required; here it would be ceremony.

### N4 — The compose stack has no healthchecks and nginx runs as root

`compose.yaml` orders startup with `depends_on` and nothing more: no `healthcheck` blocks, no
`condition: service_healthy`, no restart-until-ready loop.

`depends_on` guarantees the backend container exists before nginx starts, which is all nginx
actually needs — it resolves `backend` in `proxy_pass` at startup and would refuse to start if the
name did not resolve. It does not wait for the Go server to be *listening*, but that server binds a
port and nothing else; there is no migration, no connection pool, no warm-up. A healthcheck would
be polling for an event that has already happened. On a stack with a database it would be
required.

The backend image also could not run one as written: `FROM scratch` has no shell and no `wget`, so
a `CMD`-style healthcheck would need a purpose-built probe binary added to the image — real work to
buy nothing here.

Two related gaps, both deliberate: `nginx:alpine`'s master process runs as root (its workers drop
to the `nginx` user), where `nginxinc/nginx-unprivileged` would avoid it; and the images are built
locally by compose rather than tagged and pushed to a registry, since there is no deployment target
to push to. The backend image does run as `nobody` (`USER 65534`).

### N3 — `main()` is not covered by tests

`cmd/server` sits at 48% coverage. The uncovered statements are all inside `main()`, which
constructs the `http.Server` and blocks on `ListenAndServe`.

`router()` and `port()` — the parts with actual branches — were split out precisely so they could
be tested, and both are covered. Restructuring further (injecting a listener, running the server in
a goroutine) to chase the number would add indirection to make a coverage report look better, which
is the wrong trade. Also explained in README § Coverage.

---

## Resolved

### D8 — No request body size limit *(fixed)*

`internal/handler` now wraps `r.Body` in `http.MaxBytesReader` with a 1 MiB cap before reading
it. A body past the limit is refused with `413 request body too large` rather than being read
into memory and then failing as invalid JSON. One line in `Calculate`, one branch in `compute`
to turn the resulting `*http.MaxBytesError` into the `413`, and one handler test. Was N2 in the
"deliberately not done" list; promoted because the guard is idiomatic (`net/http` ships it),
costs almost nothing, and the question came up in review. See DECISIONS.md T18.

### D5 — The favicon was still Vite's default logo *(fixed)*

`frontend/index.html` now links `/abacus.svg` — a small SVG in `frontend/public/` that renders the
🧮 (abacus) emoji through an SVG `<text>` element, so there is no binary asset to maintain and it
stays crisp at any tab size. `public/vite.svg` is deleted (nothing else referenced it). The
repo-wide `.gitattributes` LF rule covers the new asset too, so it does not churn between Windows
and CI checkouts (see D7).

### D7 — `gofmt -l` reported every Go file on Windows *(fixed)*

A `.gitattributes` at the repo root now sets `* text=auto eol=lf`, so every text file checks out
with LF even on Windows and Go sources no longer read as CRLF-misformatted. The six committed
`.go` files were re-checked-out once under the new rule to normalise the working copy;
`git add --renormalize .` found nothing to change, confirming the blobs were already stored LF —
the problem was purely the Windows checkout. Verified: `gofmt -l .`, `go vet ./...` and
`go test ./...` are all clean.

The rule started as `*.go` (plus `*.svg` for D5) and was later widened to every text file. The
narrow version left Markdown ungoverned, so `PROMPTS.md` sat CRLF in the working tree while
generated files were LF; appending LF lines to it produced a mixed-ending file. One rule for the
whole repo removes the class of problem rather than the instances.

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

README § Coverage now carries real frontend numbers: 33 tests across two files, 100% of statements,
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
