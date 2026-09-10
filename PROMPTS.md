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

---

## Session 5 — 2026-09-08 — Vite dev proxy fix

### Prompt 1

> @TECHNICAL_DEBTS.md Let's fix the vite proxy issue. Plan.

Read `vite.config.ts`, `DECISIONS.md`, and the README proxy references. Produced a plan to add a
`server.proxy` entry for `/calculate` → `http://localhost:8080` (string shorthand, no rewrite / no
`changeOrigin`), verify with a proxied `curl`, and move D1 to a resolved section. No files changed.

### Prompt 2

> yes, proceed

Added the `server.proxy` block to `frontend/vite.config.ts`. Moved D1 in `TECHNICAL_DEBTS.md` from
"Open gaps" to a new "Resolved" section. Verified: ran the Go backend + `npm run dev`, then
`curl -X POST http://localhost:5173/calculate -d '{"operation":"multiply","a":6,"b":7}'` returned
`{"result":42}`; `tsc` on `tsconfig.node.json` clean. Changes later stashed, then popped onto a new
`fix/vite-dev-proxy` branch off the updated `master` (which now has the merged backend).

### Prompt 3

> Log prompts for this session. After that you may commit and push

Ran this skill to append Session 5, then committed and pushed `fix/vite-dev-proxy`.

---

## Session 6 — 2026-09-08 — Frontend implementation

### Prompt 1

> Let's start with front-end development. Is there still any decision to be made before we start implementing? I know there are still 3 technical debts, but these will be resolved as we develop the front-end. Correct me if I'm wrong, but there is no decision to be made regarding these open gaps, right?

Read `DECISIONS.md`, `TECHNICAL_DEBTS.md`, `README.md`, `package.json` and `vite.config.ts`.
Confirmed D2/D3/D4 needed no decisions, and surfaced eight frontend questions not yet recorded
anywhere — interaction model, `sqrt`'s second operand, input handling, state, `api.ts` shape, error
display, styling, loading state — with a recommendation for each. No files changed.

### Prompt 2

> /plan Accept your recommendations. Plan implementation roadmap before implementing

Wrote a nine-step roadmap covering `api.ts`, `operations.ts`, `Calculator.tsx`, `Calculator.css`,
scaffold deletion, tests, README coverage, `DECISIONS.md` entries and this log. One follow-up
question asked whether to trim `operations.ts` and `api.test.ts`; the answer kept both. No project
files changed.

### Prompt 3

> Yes. Change model to Opus 5 before starting implementation. Create a new branch from master and implement plan as described, chunk by chunk, using auto mode. I'll review and commit each chunk before moving on to the next one.

Created branch `feat/frontend` off `master`. Flagged that the model cannot be switched from inside
the session — `/model opus` is the user's action.

### Prompt 4

> start with chunk 1

Added `frontend/src/api.ts`: the `Operation` string-literal union, `CalcRequest` / `CalcResponse`
interfaces mirroring the Go structs, an `ApiError` class carrying the backend's message, and
`calculate()`. `tsc --noEmit` and `eslint` clean.

### Prompt 5

> commit and push

Committed `80c3036`, pushing `feat/frontend` to the remote for the first time. Then chunk 2:
`frontend/src/operations.ts`. Dropped the plan's per-entry `unary` flag, keeping only `isUnary()`,
so the arity of `sqrt` has one source of truth rather than two that can disagree.

### Prompt 6

> commit and push. Move to chunk 3

Committed `3b3492a`. Chunk 3: `frontend/src/Calculator.tsx` — a `Status` discriminated union for
idle/loading/result/error, module-level `parseOperand` and `validate`, and a `role="status"` live
region. The empty-string check runs before `Number()`, since `Number('')` is `0`, not `NaN`.

### Prompt 7

> I agree with these plan deviations. Commit and push. Next chunk.

Committed `21c66bc`. Chunk 4: added `Calculator.css`, imported it, switched `main.tsx` to render
`Calculator`, moved the orphaned `#root` rule into `index.css`, and retitled `index.html` to
"Calculator". Verified with `npm run build` plus live `curl`s through the Vite proxy on `:5173`
covering add, divide-by-zero and `sqrt` without `b`.

### Prompt 8

> Add the favicon update to @TECHNICAL_DEBTS.md. commit and push. Move to next chunk

Added D5 (Vite favicon) and refreshed the status line; committed `707006d`. Chunks 5 and 6 were
then merged after a follow-up question, because deleting `App.test.tsx` on its own would leave a
commit where `npm test` exits 1 with "No test files found". Writing the tests exposed that
`type="number"` silently blanks any non-numeric value, making two validation branches unreachable;
a second question chose `type="text"` + `inputMode="decimal"`, which took frontend coverage to
100%. Committed `8565957`.

### Prompt 9

> 1. After testing manually I found an issue. Consider the following test: Given operand b was set to some number; When operation sqrt is selected; Then operand b must be cleaned and set to inactive. Current behaviour is setting to inactive without cleaning the previous value. Add this to @TECHNICAL_DEBTS.md; 2. Since this chunk was already commited, move to next one.

Recorded D6, noting the stale `b` is never actually sent because `validate()` omits it for unary
operations. Chunk 7: filled in README § Coverage (31 tests, 100% across all four metrics),
disclosed the `src/main.tsx` exclusion, and corrected the architecture diagram.

### Prompt 10

> Commit, push. Implement chunks 8 and 9 together

Committed `d1ee621`. Added T11–T17 to `DECISIONS.md` — submit-per-calculation, disabled `b`,
`type="text"`, message parity with the backend, the discriminated union, `ApiError` carrying a
message, and the mirrored operation list — and moved D2 and D4 to Resolved in `TECHNICAL_DEBTS.md`.
Ran this skill to append Session 6. Final verification also turned up D7: `gofmt -l` flags all six
Go files on Windows because git checks them out CRLF, which `gofmt` reads as wholly misformatted.
Confirmed a false alarm by running `gofmt -l` against the files extracted from the git object store.

### Prompt 11

> Commit and push

Committed `6f269da` and pushed, completing all nine planned chunks across seven commits on
`feat/frontend`. Reported the branch summary and flagged what remains: D6 needs a decision rather
than just a patch, D7's `.gitattributes` fix belongs on `master` since it touches every Go file, and
the `type="text"` change is worth confirming in a real browser.

### Prompt 12

> log prompts

Ran this skill to append prompts 11 and 12 and to add the D7 discovery to prompt 10's note, which
had omitted it. Session 6 spans 2026-09-08 into 2026-09-09; the heading keeps its start date.

### Prompt 13

> Let's deal with the remaining technical debts.

Asked which of D3/D5/D6/D7 to act on, since D3 and D5 were deliberate deferrals with no functional
gain. For D6, offered clearing `b` on switch vs. keeping it in state and rendering the disabled
field empty; user chose the latter. For D7, offered fixing now vs. leaving it for `master`; user
chose to leave it. Implemented D6 in `Calculator.tsx` (`value={unary ? '' : b}`), added two
regression tests to `Calculator.test.tsx` (33 tests total, coverage still 100%), moved D6 to
Resolved in `TECHNICAL_DEBTS.md`, and updated the test count in README § Coverage.

### Prompt 14

> Ok, I tested manually. log prompts, then commit and push

Ran this skill to append prompts 13 and 14.

---

## Session 7 — 2026-09-09 — D7: .gitattributes for Go line endings

### Prompt 1

> Let's deal with D7 now. Pull master, create a new branch and write .gitattributes

Pulled `master` (now `969dc31`), branched `fix/gitattributes-go-eol`. Added a repo-root
`.gitattributes` with `*.go text eol=lf`, re-checked-out the six `.go` files so the working copy is
LF, and confirmed `git add --renormalize .` had nothing to change (blobs were already LF — the
issue was only the Windows checkout). `gofmt -l .`, `go vet ./...` and `go test ./...` all clean.
Moved D7 to Resolved in `TECHNICAL_DEBTS.md`; open gaps now D3 and D5.

### Prompt 2

> Ok, I retested manually to guarantee correctness. Log prompts, commit and push

Ran this skill to append Session 7, then committed and pushed the branch.

---

## Session 8 — 2026-09-09 — D5: favicon

### Prompt 1

> Let's change the favicon. Walk me through

Read `frontend/index.html` and `frontend/public/`. Explained that `public/` files are copied to
the build root as-is, so the change is swap the file + repoint `<link rel="icon">`. Offered three
options (emoji-as-SVG, hand-drawn SVG, `.ico`/PNG) and recommended the emoji SVG. No files changed.

### Prompt 2

> Okay. Since there is no calculator icon, let's use the Abacus one. Checkout to master, pull, and create a new branch. Then you can implement the favicon update.

Checked out `master` and pulled (fast-forwarded to `2294aec`, picking up the merged D7
`.gitattributes` work), branched `feat/favicon-abacus`. Added `frontend/public/abacus.svg` (🧮 in
an SVG `<text>`, 100×100 viewBox), pointed `index.html` at `/abacus.svg`, and `git rm`'d
`public/vite.svg` after confirming nothing else referenced it. `vite build` clean; `dist/` is
gitignored.

### Prompt 3

> Add to gitattributes

Added `*.svg text eol=lf` to `.gitattributes` under the `*.go` rule with a one-line comment.
`git check-attr eol` then reports `lf` for `abacus.svg` and the CRLF warning is gone;
`git add --renormalize .` found nothing else.

### Prompt 4

> Ok, tested and it's working. Log prompts, commit and push.

Moved D5 to Resolved in `TECHNICAL_DEBTS.md` and updated the status line (only D3 open now). Ran
this skill to append Session 8, then committed and pushed `feat/favicon-abacus`.

---

## Session 9 — 2026-09-10 — Pre-submission review, cleanup, comment trim

### Prompt 1

> Back end and front end have been implemented and the application is working well. analyze the project and try to find anything we should fix before I submit this assignment.

Reviewed the whole project (`go vet`, `gofmt -l`, `go test ./... -cover`, `npm test`,
`npm run lint`, `npm run build` — all clean). Reported five issues: the stale Vite-template
`frontend/README.md`, ESLint linting the generated `coverage/` directory, a "31 vs 33 tests"
drift in `TECHNICAL_DEBTS.md`, dead Vite-scaffold rules in `src/index.css`, and the unguarded
request body size (N2).

### Prompt 2

> Create a new branch for these issues

Checked out `master`, fetched, and created `chore/pre-submission-cleanup`, rebasing it onto the
latest `origin/master` (favicon PR #5 already merged there).

### Prompt 3

> /plan let's fix all the 5 issues you mentioned

Entered plan mode. Asked two questions — `413` vs `400` for an oversized body (chose `413`), and
how far to trim `index.css` (dead rules only) — wrote the plan to the plan file, approved.

### Prompt 4

> Go

Implemented the body-size chunk: `maxBodyBytes` (1 MiB) and `http.MaxBytesReader` in
`internal/handler`, oversized body → `413 request body too large`, plus a handler test.
Documented as `DECISIONS.md` T18, a new error-table row, a `README.md` error-summary row, and
N2 → D8 in `TECHNICAL_DEBTS.md`. Verified both suites and the `413` via `curl`.

### Prompt 5

> Commit and push

Split the five fixes into commits `44c0e6a` (frontend/README stub), `a09a112` (eslint ignore
`coverage`), `6568932` (D4 count 31→33), `76d0742` (drop dead `index.css` anchor rules),
`47d94cc` (body cap). Rewrote one messy intermediate commit before pushing the branch.

### Prompt 6

> Read @../misc/ASSIGNEMENT_INSTRUCTIONS.md and see if we missed anything

Checked the assignment brief against the repo. Found two documentation gaps — no runnable `curl`
API examples in the README, and responsive/mobile support implemented but never stated — plus a
note that this session still needed logging in `PROMPTS.md`.

### Prompt 7

> yes, do #1 and #2

Added a `### Examples` block to README § API (three `curl` calls, output verified against a
running server) and documented the responsive layout in README Features + `DECISIONS.md` T19.
Commit `aefc3cb`.

### Prompt 8

> According to the @misc/ASSIGNMENT_INSTRUCTIONS.md, "Design decisions or assumptions" should be included in README.md. Should we move what's in @DECISIONS.md to the readme file? At least a summary of the decisions, referencing @DECISIONS.md ?

Recommended not moving the full file. Added a "Design decisions & assumptions" section to
`README.md` — a one-line-per-point summary (assumptions plus backend/frontend/integration
decisions) that references `DECISIONS.md` by ID; updated the `DECISIONS.md` intro to match.
Commit `9d01e0a`.

### Prompt 9

> Now let's remove the unnecessary comments from the project. Comments should only be used when necessary to explain some business logic or important implementation decision. Don't reference codes from the docs, like "D1", "B1", "N1", etc. When needed state the business rule in the inline comment. Keep comments clean and simple.

Trimmed comments across all Go and TS/TSX/CSS source and tests: removed narration, TypeScript
idiom explanations, and every `DECISIONS.md` / `B#` / `T#` / `D#` cross-reference, keeping only
business rules and non-obvious implementation notes in plain language. Comments only, no code
change (confirmed with a filtered diff); ~150 net lines removed. Commit `e4d3949`.

### Prompt 10

> Perfect. Now run npm audit and let's see if we have any issues

Ran `npm audit`: 6 advisories (2 critical, 1 high, 3 moderate), all in the `vite` / `vitest` /
`esbuild` / `@vitest/*` dev toolchain; `npm audit --omit=dev` reports 0. Documented the status
under `TECHNICAL_DEBTS.md` D3 and the README toolchain note rather than forcing the breaking
`vite` 8 / `vitest` 5 upgrade. Commit `9a170c9`.

### Prompt 11

> D3 is still listed as an open gap, but we have deliberately not done. Update @TECHNICAL_DEBTS.md accordingly and ammend last commit

Moved D3 from "Open gaps" to "Considered and deliberately not done" in `TECHNICAL_DEBTS.md`,
reframed it as a deliberate choice, and updated the status line to "no gaps remain open".
Amended and force-pushed the previous commit → `b4d1c92`.

### Prompt 12

> log prompts and ammend last commit to include the prompts

Ran this skill to append Session 9, then amended `b4d1c92` to include `PROMPTS.md`.

