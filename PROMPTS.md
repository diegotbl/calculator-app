# AI Prompts Log

Per the assignment instructions, this file records the AI prompts used while building the project,
in chronological order, with a short note on what each one produced.

Tool: Claude Code (Sonnet 5; switched to Opus 5 partway through Session 4).

---

## Session 1 — 2026-09-06 — Environment setup

### Prompt 1

> I don't have React installed in this machine yet, and I'll need React to implement this project.
> Help me with the setup.

Checked the toolchain (Node 20.10.0 / npm 10.2.4 present; Go not on PATH). Scaffolded the frontend
with `npm create vite@5.5.5 frontend -- --template react-ts` (latest `create-vite` requires Node
20.19+, so the scaffolder was pinned to v5). Ran `npm install` and verified with `vite build`.
Flagged the test-runner choice (Jest per CLAUDE.md vs. Vitest) as a decision to make.

### Prompt 2

> Let's use Vitest. Update CLAUDE.md and README.md to reflect this decision.

Installed `vitest`, `@vitest/coverage-v8`, `jsdom`, and the React Testing Library packages (v2 /
jsdom 25 pinned for Node 20.10). Added a `test` block to `vite.config.ts` (jsdom env, globals,
`src/test/setup.ts`, v8 coverage), `src/test/setup.ts` for jest-dom matchers, `test` /
`test:watch` / `test:coverage` scripts, and `vitest/globals` to `tsconfig.app.json`. Added a
throwaway `src/App.test.tsx` smoke test to verify the wiring (passes, build clean). Updated
CLAUDE.md and README.md to say Vitest instead of Jest and to document the Node version pins.

Also created `.claude/settings.local.json` with `"worktree": { "bgIsolation": "none" }` so edits
land directly in the working copy (Claude Code background-session setting, gitignored).

### Prompt 3

> Update Node 22 using nvm windows

Installed nvm-windows via `winget install CoreyButler.NVMforWindows` (its installer removed the
old 32-bit standalone Node and put nvm on PATH), then `nvm install 22` (22.23.2) and
`nvm use 22.23.2`. Reinstalled `frontend/node_modules` from scratch for the 32-bit → 64-bit
binary switch. Re-verified tests + build on Node 22. Updated the README toolchain note.

### Prompt 4

> Write prompts into PROMPTS.md

Created this file.

### Prompt 5

>  Create a skill for appending session's prompts into PROMPTS.md

Created .claude/skills/log-prompts/SKILL.md.

---

## Session 2 — 2026-09-07 — Decisions and error spec

### Prompt 1

> What's the next step for this project? Do I need to decide something (e.g. technical
> specifications, business rules) before we start implementing?

Reviewed project state (frontend scaffold + Vitest done, no backend, Go 1.27.1 now on PATH).
Surfaced three open decisions via a multiple-choice prompt; user chose: `percentage = (a/100)*b`,
Vite dev proxy for cross-origin (no backend CORS), and reject non-finite results with `400`.

### Prompt 2

> Specify all backend errors (overflow, missing operands, invalid operation, division by zero,
> etc)

Enumerated the full backend error surface — routing (`404`/`405`), decoding, semantic validation
(with check ordering), operation preconditions, and the finite-result guard — with exact messages
and status codes. No files changed yet.

### Prompt 3

> Let's create a DECISIONS.md for business and technical decisions. Keep README.md with an
> introduction about the application, setup instructions, basic architecture information, and
> information important to the user. Write the error status decisions into DECISIONS.md and a
> brief summary in README.md.

Created `DECISIONS.md` (business rules B1–B5, technical decisions T1–T10, full error-response
tables). Rewrote `README.md` to intro / setup / architecture / API / 5-row error summary that
links into `DECISIONS.md`; removed the requirements walls.

### Prompt 4

> Add a one-line pointer in CLAUDE.md referencind DECISIONS.md and log prompts.

Added a two-line pointer to `DECISIONS.md` near the top of `CLAUDE.md`. Ran this skill to append
Session 2.

---

## Session 3 — 2026-09-07 — Initial commit

### Prompt 1

> Create initial commit on master branch and push to origin.

Added `.claude/settings.local.json` to `.gitignore`, staged all files, created root commit
`bc63863` ("Initial commit: project scaffolding and assignment docs", 25 files) on `master` and
pushed with `-u` to `origin`. Flagged that `user.email` was still a work address.

### Prompt 2

> Change my user.email to diegotblima@gmail.com. Amend the commit

Corrected local `user.email` to `diegotblima@gmail.com`, amended again with `--reset-author`
(commit `23d2533`), force-pushed with `--force-with-lease`.

### Prompt 3

> log prompts

Ran this skill to append Session 3.

---

## Session 4 — 2026-09-07 — Backend implementation

### Prompt 1

> Let's start with backend implementation. Plan

Entered plan mode. Confirmed `backend/` did not exist and Go 1.27.1 was on PATH; read
`DECISIONS.md` and treated it as the spec. Wrote a four-chunk plan (calculator → handler → server
→ docs) to the plan file. No project files changed.

### Prompt 2

> Instead of keeping all calculation logic inside calculator.go, what do you think about crating
> small files each dealing with only one operation?

Recommended against one file per operation: the seven operations are one-line functions, and
idiomatic Go groups small related functions in a single file per package. Offered three layouts;
user selected one `calculator.go`. No files changed.

### Prompt 3

> Change model to Opus 5. Implement. Create a new branch for that implementation.

Noted that the model cannot be switched from inside the session (user ran `/model`; the session
continued on Opus 5). Created branch `feat/backend`. Implemented chunk 1: `backend/go.mod` (module
`calculator-app/backend`) and `internal/calculator/` — seven operations, a name → operation map
carrying each operation's arity, sentinel errors, and one shared `IsInf`/`IsNaN` result guard.
`gofmt`/`go vet` clean, 100% coverage.

### Prompt 4

> Ok. Move to chunk 2

Added `internal/handler/`. Request fields decode as `json.RawMessage` and are unmarshalled
individually so "absent", "wrong type", and "out of range" produce different messages; operands go
through `json.Number` + `strconv.ParseFloat` after verifying that unmarshalling `1e400` straight
into a `float64` returns `+Inf` *and* an `UnmarshalTypeError`. Split `compute()` from the HTTP
plumbing; added a `recover` that turns a panic into a JSON `500`. 100% coverage.

### Prompt 5

> Next chunk

Chunk 3: `cmd/server/main.go` with hand-written path/method routing (JSON `404`/`405`, `Allow:
POST`), `PORT` override, and explicit `http.Server` timeouts, plus `main_test.go`. Exported
`handler.WriteError` so routing reuses the error wire format. Live smoke test on `PORT=8099`
matched all eight documented responses.

### Prompt 6

> Next chunk

Chunk 4: filled the README coverage placeholder (`internal/calculator` 100%, `internal/handler`
100%, `cmd/server` 48% — only `main()` uncovered) and rewrote `DECISIONS.md` T6, which still
described operands as `*float64`, correcting the claim that `1e400` "decodes to `+Inf`".

### Prompt 7

> Commit and push only chunk 1

Staged only `backend/go.mod` and `internal/calculator/`, verified the package tests standalone,
committed `ba56e45` and pushed with `-u` to `origin/feat/backend`.

### Prompt 8

> Now commit and push chunk 2

Verified the handler builds and tests against the already-committed calculator alone, then
committed `f248d76` and pushed. `WriteError`, which was written during chunk 3, went in here since
it lives in the handler files.

### Prompt 9

> Commit and push chunk 3

Ran `gofmt`, `go vet`, `go build`, and the full test suite, then committed `f771b95` and pushed.

### Prompt 10

> Commit and push chunk 4

Reviewed both doc diffs, confirmed the README numbers still matched a live coverage run, committed
`f1bdbd1` and pushed. Left `CLAUDE.md` and `PROMPTS.md` uncommitted as user-owned edits.

### Prompt 11

> log prompts and commit CLAUDE.md and PROMPTS.md

Ran this skill to append Session 4, then committed `CLAUDE.md` and `PROMPTS.md`.

### Prompt 12

> We finished for this session or is there still something missing for backend implementation?

Re-ran `gofmt`, `go vet`, and the full test suite, and cross-checked every row of the
`DECISIONS.md` error tables against a test. Backend confirmed complete. Found that
`frontend/vite.config.ts` has no `server.proxy` block even though DECISIONS.md T4 and the README
both describe one, so the API is reachable by `curl` but not from a browser on `:5173`.

### Prompt 13

> Document the gaps you found in a file. Might be TECHNICAL_DEBTS.md for example

Created `TECHNICAL_DEBTS.md`: four open gaps (missing Vite proxy, leftover Vite scaffold files,
Node 20-era dependency pins, unfilled frontend coverage) and three deliberate omissions with their
reasoning (no graceful shutdown, no request body size limit, `main()` left uncovered at 48%).

