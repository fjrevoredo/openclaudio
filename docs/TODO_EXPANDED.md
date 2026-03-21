# TODO

Full details about Open TODOs and planned improvements. For the list of TODOs, see [TODO.md](TODO.md).

Extended TODO Format

- `- [ ] **Task title** (TODO-ID) — concise requirement-style description with scope and constraints`
- Write items as requirements/acceptance criteria (what must be true), not implementation plans (how to build it)
- Add extended information about the task relevant for the planning of the implementation such as file references, conceptual references or contraints
- Put items under the appropriate priority section
- Use indented checkbox items only for true sub-tasks or explicit dependencies

---

## High Priority

- [ ] **Restrict split view to renderable files** (TODO-UI-SPLIT-RENDERABLE) — The split view must only be shown for files with a supported rendered preview, and unsupported file types must fall back to the single-pane editor/view experience without broken or empty preview states.
  - Relevant areas: `internal/files/`, `internal/web/`, `web/templates/file.html`, `web/static/src/app.js`
  - Current constraint: rendered preview support is currently markdown-focused, so availability should follow actual render support rather than file size or editor state alone.

- [ ] **Add basic GitHub CI** (TODO-CI-BASIC) — The repository must have a basic GitHub Actions CI workflow that runs on `push` and `pull_request` for the default branch, uses least-privilege permissions, enables concurrency cancellation for superseded runs, and validates the standard build and test path in line with `docs/CI_BEST_PRACTICES.md`.
  - Relevant references: `docs/CI_BEST_PRACTICES.md`, `Makefile`, `package.json`
  - Initial scope should cover the standard verification path for this repo, including frontend asset build and Go tests.

---

## Medium Priority

- [ ] **Add color scheme for raw markdown display** (TODO-UI-MARKDOWN-RAW-COLOR) — Raw markdown content must use a deliberate color scheme that improves readability and visual hierarchy while remaining consistent with the existing UI and preserving legibility in the editor/view experience.
  - Relevant areas: `web/templates/file.html`, `web/static/src/app.css`, `web/static/src/app.js`
  - Scope note: this is about the raw markdown presentation, not changing the rendered markdown theme independently unless the design requires shared tokens.

---

## Low Priority / Future

- [ ] **Migrate frontend package management to pnpm** (TODO-TOOLING-PNPM-MIGRATION) — The project must use `pnpm` as the primary JavaScript package manager, with repository scripts, lockfiles, CI expectations, and contributor documentation updated so installs and builds are reproducible without relying on `npm`.
  - Relevant areas: `package.json`, lockfiles, `README.md`, `AGENTS.md`, CI workflow files
  - Migration is lower priority than current product and CI behavior work because it is mostly tooling churn unless paired with a broader dependency-management change.
