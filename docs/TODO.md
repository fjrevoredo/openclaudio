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

- [ ] **Restrict split view to renderable files** (TODO-UI-SPLIT-RENDERABLE) — The split view must only be shown for files with a supported rendered preview, and unsupported file types must fall back to the single-pane editor/view experience without broken or empty preview states.
- [x] **Add basic GitHub CI** (TODO-CI-BASIC) — The repository must have a basic GitHub Actions CI workflow that runs on `push` and `pull_request` for the default branch, uses least-privilege permissions, enables concurrency cancellation for superseded runs, and validates the standard build and test path in line with `docs/CI_BEST_PRACTICES.md`.

---

## Medium Priority

- [ ] **Add color scheme for raw markdown display** (TODO-UI-MARKDOWN-RAW-COLOR) — Raw markdown content must use a deliberate color scheme that improves readability and visual hierarchy while remaining consistent with the existing UI and preserving legibility in the editor/view experience.

---


## Low Priority / Future

- [x] **Migrate frontend package management to pnpm** (TODO-TOOLING-PNPM-MIGRATION) — The project must use `pnpm` as the primary JavaScript package manager, with repository scripts, lockfiles, CI expectations, and contributor documentation updated so installs and builds are reproducible without relying on `npm`.

--
