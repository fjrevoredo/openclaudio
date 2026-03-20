# openclaudio Implementation Plan

## Summary

`openclaudio` is a standalone open-source project at `/home/francisco/openclaudio`.

Purpose:

- provide a lightweight web panel for daily OpenClaw management
- focus on browsing and editing the OpenClaw workspace, especially markdown files
- expose key OpenClaw install and runtime information
- support a small set of safe gateway management actions

This plan is the current source of truth for implementation handoff.

## Current State

The repository exists locally and is connected to the public GitHub remote:

- local path: `/home/francisco/openclaudio`
- remote: `https://github.com/fjrevoredo/openclaudio.git`

Current repo state:

- git repository initialized
- `origin` remote configured
- no application code implemented yet
- no dependency manifests committed yet

Important implementation note:

- the agreed implementation stack is **Go**
- the previous attempt deviated to Node because the current machine does **not** have a Go toolchain installed
- that deviation has been reverted
- before implementation resumes, install or otherwise provide `go` on this machine

Observed environment fact:

- `go version` currently fails with `go: command not found`

## Product Decision

No single off-the-shelf product covers the full scope well enough.

Rejected as primary solution:

- OpenClaw Control UI: useful complementary operator surface, but not workspace-first file management
- File Browser: close on lightweight browsing, but not enough for markdown-first editing plus OpenClaw-specific operator data
- SilverBullet: strong markdown product, but too note-centric and weak on OpenClaw operations
- code-server: capable, but too heavy and too broad for a daily operations buddy

Decision:

- build a purpose-built standalone app called `openclaudio`
- keep OpenClaw Control UI as a secondary reference tool, not the main UI

## Implementation Stack

Required stack:

- backend: Go
- rendering: server-side HTML templates
- interactivity: HTMX
- raw editor: CodeMirror 6
- markdown rendering: server-side GitHub-flavored markdown with HTML sanitization

Deployment model:

- separate local service
- single trusted operator
- intranet or localhost only
- Linux-first
- user-level systemd service

## Runtime Defaults

Default runtime paths:

- workspace root: `/home/francisco/.openclaw/workspace`
- OpenClaw root: `/home/francisco/.openclaw`
- gateway unit: `openclaw-gateway.service`
- default panel port: `18890`

These should be configurable through environment variables or a small config file, but the defaults above should work on the current host.

## Functional Scope

### 1. Workspace Browser

Main focus:

- browse the full OpenClaw workspace
- prioritize markdown files in UX without hiding other files
- make file-path copying fast and obvious

Required behaviors:

- left-side file tree
- lazy or incremental loading if needed for responsiveness
- filename/path filter
- markdown-first affordances
- hide `.git` internals by default in v1

### 2. File Viewer And Editor

Required view modes:

- rendered markdown
- raw text
- split rendered/raw view

Required actions:

- save file
- reload from disk
- copy relative path
- copy relative path wrapped in backticks
- copy absolute path

Editing rules:

- text-based editing only in v1
- UTF-8 assumed
- invalid UTF-8 files may fall back to raw read-only handling in v1
- optimistic concurrency using file mtime plus content hash
- reject stale writes with clear conflict message
- warn on unsaved changes before navigation

### 3. OpenClaw Summary Dashboard

Expose these cards or panels:

- installed OpenClaw version
- configured primary model
- configured fallback models
- configured gateway port and bind mode
- active session count
- recent session list
- recent cron run counts or status
- gateway service state
- gateway process CPU/RSS/elapsed time
- recent gateway log tail

### 4. Gateway Actions

Allowed actions in v1:

- start gateway
- stop gateway
- restart gateway

Execution rules:

- no arbitrary command execution
- backend must use a fixed allowlist only
- results must be shown inline with success/failure and timestamps

## Data Sources

Use these runtime sources:

- installed version from OpenClaw `package.json`
- config from `~/.openclaw/openclaw.json`
- sessions from `~/.openclaw/agents/main/sessions/sessions.json`
- cron data from `~/.openclaw/cron/jobs.json` and `~/.openclaw/cron/runs/*.jsonl`
- logs from `/tmp/openclaw/openclaw-YYYY-MM-DD.log`
- service state from `systemctl --user show`
- process metrics from `ps` or `/proc`

Important implementation rule:

- do **not** depend on live OpenClaw CLI subcommands for core dashboard data
- in constrained environments, CLI subcommands may fail due to network-interface probing or sandbox behavior
- filesystem, `systemctl`, and process inspection are the reliable base

## Security Model

v1 security assumptions:

- single trusted operator
- not exposed to the public internet

Required controls:

- mandatory login
- one admin account for v1
- password hash stored outside repo-tracked files
- secure cookie sessions
- CSRF protection on all mutating routes
- path sandboxing to configured workspace root
- reject traversal, absolute path writes, and symlink escape

Command execution guardrails:

- fixed allowlist for gateway actions only
- no shell passthrough
- no arbitrary subprocess feature

## HTTP Surface

Pages:

- `GET /login`
- `GET /`
- `GET /files/*path`

Data endpoints:

- `GET /api/tree?path=...`
- `GET /api/file?path=...&view=raw|rendered|split`
- `PUT /api/file?path=...`
- `POST /api/file/copy-path`
- `GET /api/openclaw/summary`
- `GET /api/openclaw/sessions`
- `GET /api/openclaw/cron`
- `GET /api/openclaw/logs?date=YYYY-MM-DD&lines=N`
- `POST /api/openclaw/gateway/start`
- `POST /api/openclaw/gateway/stop`
- `POST /api/openclaw/gateway/restart`

Save contract:

- client sends raw text
- client sends last-known mtime and content hash
- server rejects with conflict status if file changed on disk

## Suggested Go Project Structure

- `cmd/openclaudio/` main entrypoint
- `internal/config/` runtime config loading and validation
- `internal/files/` workspace tree, reads, writes, path safety, copy-path helpers
- `internal/markdown/` markdown rendering and HTML sanitization
- `internal/openclaw/` version/config/session/cron/log/service/process readers
- `internal/auth/` login, sessions, CSRF
- `internal/web/` handlers, templates, HTMX partials
- `web/templates/` HTML templates
- `web/static/` CSS, HTMX, CodeMirror wiring
- `deploy/systemd/` example user service
- `docs/` architecture and configuration notes

## Suggested Implementation Order

### Step 1. Bootstrap Repo

- add `README.md`
- add `go.mod`
- add initial project layout
- add configuration loading
- add basic server bootstrap and health route

Verify:

- app starts locally
- config resolves current host defaults

### Step 2. Add Auth Shell

- implement login page
- implement session cookies
- implement CSRF support
- protect all non-login routes

Verify:

- unauthenticated requests redirect to login
- authenticated session persists
- mutating routes reject missing/invalid CSRF

### Step 3. Implement Safe File Access

- workspace root config
- safe path resolution
- tree listing
- file read API
- file save with mtime/hash conflict detection
- copy-path helpers

Verify:

- traversal is blocked
- file browsing works on current workspace
- stale writes are rejected

### Step 4. Implement Markdown And Editor UI

- server-side markdown rendering
- raw editor
- split/raw/rendered modes
- save/reload/copy actions

Verify:

- markdown renders safely
- raw editing works
- path copy outputs correct values

### Step 5. Implement OpenClaw Summary Readers

- parse installed version
- parse current model/fallbacks
- parse sessions
- parse cron state
- tail current logs
- read service state
- read process metrics

Verify:

- dashboard cards show real host data
- missing files degrade gracefully

### Step 6. Implement Gateway Controls

- allowlisted `systemctl --user` wrapper
- start/stop/restart handlers
- inline action result UI

Verify:

- only allowed actions execute
- failures surface clearly
- no arbitrary command path exists

### Step 7. Polish And Ship Basics

- add styling
- add systemd service example
- add documentation
- add tests

Verify:

- app is usable without manual code inspection
- systemd example is coherent
- docs reflect actual behavior

## Test Plan

### Backend Tests

- path normalization blocks `..` and root escape
- symlink escape is blocked
- markdown rendering sanitizes HTML/script input
- file save succeeds with matching preconditions
- file save fails on stale mtime/hash
- non-UTF8 file handling is safe
- config/session/cron/log parsers handle missing files gracefully
- service wrapper rejects anything outside allowlist

### Integration Tests

- login required for all protected routes
- CSRF required for save and gateway actions
- markdown file opens in raw/rendered/split modes
- copy-path outputs correct relative/backticked/absolute values
- dashboard renders host-derived values
- gateway actions return clear success/failure payloads

### Manual Acceptance

On the current host:

- browse roughly current workspace scale without obvious lag
- open and edit `MEMORY.md`
- copy a nested relative path suitable for OpenClaw prompts
- inspect model and recent sessions
- restart gateway and observe updated state
- inspect today’s gateway log tail from the panel

## Non-Goals For V1

Do not include these in the first implementation unless requirements change:

- multi-user roles
- public internet deployment
- embedded OpenClaw Control UI
- arbitrary command execution
- delete, rename, move, upload, or bulk file operations
- collaborative editing
- plugin system

## Handoff Notes

When implementation resumes in a new session:

1. confirm Go is available locally
2. keep the implementation in Go as originally planned
3. do not reintroduce the Node fallback unless explicitly approved
4. use this file as the current implementation contract

At the moment, the repo is intentionally almost empty so the next session starts from a clean state.
