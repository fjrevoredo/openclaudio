# TODO

Open tasks and planned improvements. For full context, see [TODO_EXPANDED.md](TODO_EXPANDED.md).

TODO entry format:

- `- [ ] **Task title** (TODO-ID) — concise requirement-style description with scope and constraints`
- Write items as requirements/acceptance criteria (what must be true), not implementation plans (how to build it)
- Keep implementation details minimal in TODO entries; move deep implementation notes to `TODO_EXPANDED.md` when needed
- Put items under the appropriate priority section
- Use indented checkbox items only for true sub-tasks or explicit dependencies


---

## High Priority

- [ ] **Stabilize editor lifecycle and dirty-state tracking** (TODO-UI-EDITOR-LIFECYCLE) — The file editor experience must clean up replaced editor instances correctly so HTMX swaps never leave stale unsaved-change warnings, stale listeners, or incorrect active-editor state behind.
- [ ] **Make file saves atomic and crash-safe** (TODO-FILES-ATOMIC-SAVE) — Workspace file saves must use an atomic write strategy that preserves permissions and avoids truncating or corrupting the target file if the process or write path fails mid-save.
- [ ] **Expose degraded dashboard data honestly** (TODO-DASHBOARD-PARTIAL-FAILURES) — Summary and dashboard views must distinguish missing or failed OpenClaw data sources from healthy empty values, and must surface partial-read failures explicitly in the UI and tests.
- [ ] **Cover critical HTTP and editor flows with automated tests** (TODO-TEST-HTTP-EDITOR-FLOWS) — The test suite must exercise login/logout, CSRF rejection, file read/save/conflict flows, dashboard partial behavior, and the main edit-save-navigate browser flow with deterministic automated coverage.
- [ ] **Restrict split view to renderable files** (TODO-UI-SPLIT-RENDERABLE) — The split view must only be shown for files with a supported rendered preview, and unsupported file types must fall back to the single-pane editor/view experience without broken or empty preview states.
- [x] **Add basic GitHub CI** (TODO-CI-BASIC) — The repository must have a basic GitHub Actions CI workflow that runs on `push` and `pull_request` for the default branch, uses least-privilege permissions, enables concurrency cancellation for superseded runs, and validates the standard build and test path in line with `docs/CI_BEST_PRACTICES.md`.

---

## Medium Priority

- [ ] **Harden accessibility semantics for async and tree interactions** (TODO-UI-A11Y-ASYNC-TREE) — Async status messages, workspace tree expansion, and keyboard focus states must be communicated semantically and visibly so the interface remains usable with keyboard and assistive technology.
- [ ] **Introduce a narrow host-command test seam** (TODO-OPENCLAW-COMMAND-SEAM) — OpenClaw service status, process inspection, and gateway actions must be testable through a small command-execution seam without introducing broad abstraction or diluting the single-host design.
- [x] **Add color scheme for raw markdown display** (TODO-UI-MARKDOWN-RAW-COLOR) — Raw markdown content must use a deliberate color scheme that improves readability and visual hierarchy while remaining consistent with the existing UI and preserving legibility in the editor/view experience.

---


## Low Priority / Future

- [x] **Migrate frontend package management to pnpm** (TODO-TOOLING-PNPM-MIGRATION) — The project must use `pnpm` as the primary JavaScript package manager, with repository scripts, lockfiles, CI expectations, and contributor documentation updated so installs and builds are reproducible without relying on `npm`.

--
