# Changelog

All notable changes to Mini Diarium are documented here. This project uses [Semantic Versioning](https://semver.org/).

## Formatting

Title: ## [x.x.x] - <release_date>/Unreleased
Subtitle: ### Added/Changed/Fixed
Entry: - `- **Change title** — concise changelog entry description with scope and constraints` (<todo-id-if-applicable>)

# Logs

## [0.0.1] - Unreleased

### Changed

- **migrate frontend package management to pnpm** — Switched repository frontend tooling, lockfile expectations, and contributor docs from npm to pnpm with a pinned pnpm version and Corepack-based setup guidance. (TODO-TOOLING-PNPM-MIGRATION)
- **add basic github ci** — Added a least-privilege GitHub Actions workflow for `master` pull requests and pushes, with superseded-run cancellation and build/test validation for the standard frontend and Go path. (TODO-CI-BASIC)
