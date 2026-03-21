# Repository Guidelines

## Project Structure & Module Organization

- `cmd/openclaudio/`: main entrypoint and CLI subcommands such as `hash-password`.
- `internal/config`, `internal/auth`, `internal/files`, `internal/markdown`, `internal/openclaw`, `internal/web`: core application packages.
- `web/templates/`: server-rendered HTML templates and HTMX partials.
- `web/static/src/`: source JS/CSS for HTMX and CodeMirror behavior.
- `web/static/dist/`: built assets embedded into the Go binary.
- `tools/build.mjs`: frontend build script used by `npm run build`.
- `docs/`: architecture, configuration, TODO, and planning notes.
- `deploy/systemd/`: example user service unit.
- `.agents/skills/` and `.claude/skills/`: checked-in skill definitions for agent workflows.
- `CHANGELOG.md` and `skills-lock.json`: changelog and skills metadata tracked at the repo root.

Keep new code inside `internal/` unless it is a public entrypoint or embedded asset.

## Build, Test, and Development Commands

- `npm install`: install frontend dependencies.
- `npm run build`: bundle `web/static/src/*` into `web/static/dist/*`.
- `make frontend-build`: rebuild frontend assets only.
- `make build`: build assets and compile the `openclaudio` binary.
- `make test`: run the Go test suite.
- `make run`: build assets and run the app locally.
- `go run ./cmd/openclaudio hash-password 'secret'`: generate a bcrypt hash for `OPENCLAUDIO_ADMIN_PASSWORD_HASH`.

The app auto-loads `.env` from the repo root during local development.

## Coding Style & Naming Conventions

- Go code must be `gofmt`-formatted. Use tabs/standard Go formatting; do not hand-align.
- Package names stay short and lowercase. Exported symbols use `CamelCase`; unexported symbols use `camelCase`.
- Keep handlers thin and push logic into `internal/*` services.
- Frontend code should stay small, dependency-light, and live in `web/static/src/app.js` and `app.css` unless a split is clearly justified.

## Testing Guidelines

- Use the standard Go `testing` package.
- Place tests next to the code they cover, using `*_test.go`.
- Prefer focused unit tests for config loading, auth/session behavior, file/path safety, OpenClaw service behavior, and HTTP handlers.
- Run `make test` before opening a PR. If you touch frontend assets, also run `npm run build`.

## Commit & Pull Request Guidelines

- Current history uses short, direct commit subjects, for example: `restore readme`, `fix error on .env parsing`.
- Follow that style: imperative, lowercase, concise, and specific.
- PRs should include a short summary, user-visible impact, config/deployment changes, and screenshots for UI changes.
- Link related issues when applicable and mention the verification commands you ran.

## Planning & Changelog Workflow

- Add newly identified work to `docs/TODO.md` in the documented format, under the correct priority section.
- Write TODO entries as requirement-style outcomes, not implementation plans; add deeper notes to `docs/TODO_EXPANDED.md` when useful.
- When a feature, task, or fix is completed, add an appropriate entry to `CHANGELOG.md` under the version currently in progress, following the file's documented format.

## Security & Configuration Tips

- Never commit `.env` or real secrets; use `.env.example`.
- Keep examples and docs aligned with the current config surface, including `OPENCLAUDIO_LOG_DIR` and gateway unit settings.
- Treat `OPENCLAUDIO_BIND_ADDRESS=0.0.0.0` as LAN exposure and review host firewall rules separately.
- Keep gateway control restricted to the configured systemd unit only.
