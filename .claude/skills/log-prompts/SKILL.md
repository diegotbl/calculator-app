---
name: log-prompts
description: Append the current session's user prompts to PROMPTS.md (the assignment's AI prompt log). Use when the user says "log the prompts", "update PROMPTS.md", "write prompts into PROMPTS.md", or at the end of a work session.
---

# Log session prompts to PROMPTS.md

`PROMPTS.md` at the repo root is the assignment-required log of AI prompts used to build the
project. This skill appends the prompts from the current session to it.

## Steps

1. **Read `PROMPTS.md`** to see the existing format and find the last logged session number and
   the last logged prompt. If the file does not exist, create it with this header:

   ```markdown
   # AI Prompts Log

   Per the assignment instructions, this file records the AI prompts used while building the
   project, in chronological order, with a short note on what each one produced.

   Tool: Claude Code (Sonnet 5).

   ---
   ```

2. **Collect every user prompt from the current session that is not already in the file**, in
   order. Include only messages the user actually typed — skip tool results, system reminders,
   and your own replies.

3. **Append a session block** using this structure (start a new `## Session N` heading only if
   the current session is not already open in the file; otherwise continue the existing one):

   ```markdown
   ## Session N — YYYY-MM-DD — <short theme>

   ### Prompt K

   > <verbatim user prompt, blockquoted; keep it exact, one `>` per line>

   <1-3 sentence note on what this prompt produced: files changed, commands run, decisions made>
   ```

   - Number prompts continuously within a session (Prompt 1, 2, 3, …).
   - Use today's date from the environment for new session blocks.
   - Quote prompts **verbatim** — do not paraphrase, summarize, or fix typos.
   - Keep each note factual and brief. Mention concrete artifacts (file paths, package names,
     command names), not adjectives.

4. **Write the file** and show the user the appended block.

## Notes

- This skill only edits `PROMPTS.md`. It does not commit — leave that to the user unless they ask.
- If unsure whether a prior session is "open" (last entry is from today with no newer work),
  continue that session's numbering rather than starting a new one.
