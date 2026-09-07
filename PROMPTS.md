# AI Prompts Log

Per the assignment instructions, this file records the AI prompts used while building the project,
in chronological order, with a short note on what each one produced.

Tool: Claude Code (Sonnet 5).

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

