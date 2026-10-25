# Miyano — instructions for coding agents (any tool)

Offline-first workout app. Go API (Chi + sqlc + pgx + PostgreSQL) in `backend/`,
Vue 3 + TS + Tailwind v4 + pnpm in `frontend/`. This file is the universal entry
point: it works the same for ZCode, OpenCode, Claude Code, Cursor or a human.

## Truth sources — read before working

1. `docs/constitution.md` — the 10 binding principles (they outrank any agent's defaults).
2. `README.md` — how to run everything, env vars, endpoints, quick start.
3. `frontend/DESIGN.md` + `frontend/PRODUCT.md` — the visual system (tokens + rules). Never break its named rules.
4. `openspec/` (untracked, local) — living specs (`openspec/specs/`) and the active change (`openspec/changes/`).
5. **Miyano-Docs** (private repo of the owner, `AzuelZero/Miyano-Docs`) — bitácora with ADRs D001+, the backend runbook (§5.5) and per-lesson encyclopedias: the WHY of every decision. If you have access, the snapshot in its §3 says what to do next.

## Working rules (binding)

- **Everything through PRs** to `main` (protected). Conventional Commits.
- **Code, identifiers and comments in English; UI copy in Spanish** (ADR D009).
- **Spec-driven**: every feature or fix with observable behavior is an OpenSpec change (workflow below). No change, no code.
- **Verify every step before claiming it** (ADR D011): run the command, show the expected output. "If you don't know how to verify it, you're not done."
- **PR template + simulated review** (ADR D010): fill the author part when opening, the reviewer semáforo part before merging; 🔴 = 0 required.
- **Frontend**: one accent system (primary fills, primary-bright signals), tabular digits for all domain numbers, loading/empty/error states everywhere, Lucide icons, no new fonts or shadows (see DESIGN.md's named rules).

## OpenSpec workflow (plain CLI — no special tooling needed)

```bash
openspec new change "<kebab-name>"            # scaffold in openspec/changes/<name>/
# create proposal.md, specs/<capability>/spec.md, design.md, tasks.md
# (follow `openspec instructions <artifact> --change <name> --json` for each)
openspec validate <name>                      # must pass before implementing
# implement on a feature/* branch; tick tasks.md checkboxes ONLY with their
# verification evidence (command + expected output) next to them
openspec validate --specs                     # after syncing delta specs manually at archive time
```

Archiving = move `openspec/changes/<name>` to `openspec/changes/archive/YYYY-MM-DD-<name>`
after merging its delta requirements into `openspec/specs/<capability>/spec.md`.

## Commands

```bash
# Backend (needs Postgres up + DATABASE_URL + JWT_SECRET to RUN; build/test don't)
cd backend && go vet ./... && go test ./... && go build ./...

# Frontend
cd frontend && pnpm install && pnpm test && pnpm build

# Full local stack
docker compose -f docker-compose.dev.yml up -d postgres   # then API + pnpm dev (README §quick start)
```

CI runs both worlds on every PR; `-race` runs there (not on Windows-local Go).

## Session continuity (anti-hallucination)

Work state lives in **files, never in chat memory**:

- At session start: check the date, `git log --oneline -5`, `git status`, and the
  Miyano-Docs bitácora snapshot (§3). The most recent timeline entry's "➡️ Próximo"
  is the exact next step.
- If an agent (this includes you) asserts a state, a date or a file from memory:
  **verify it with a command before trusting it** — the files and the repo are the
  only truth.
- At session close: bitácora entry + snapshot updated, both repos pushed, tasks.md
  checkboxes honest.
