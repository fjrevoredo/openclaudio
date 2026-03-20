# openclaudio architecture

`openclaudio` is a single-process Go web app with embedded templates and static assets.

## Subsystems

- `internal/config`: env and optional JSON config loading
- `internal/auth`: signed cookie auth and double-submit CSRF protection
- `internal/files`: workspace tree browsing, safe file access, optimistic save flow
- `internal/markdown`: GitHub-flavored markdown rendering plus sanitization
- `internal/openclaw`: OpenClaw config, sessions, cron, logs, service state, and gateway actions
- `internal/web`: routing, templates, HTMX partials, and JSON save/copy endpoints

## Runtime assumptions

- single trusted operator
- localhost or intranet deployment
- existing OpenClaw install rooted at `~/.openclaw`
- gateway management only through `systemctl --user`

## Frontend

- HTMX handles panel and partial refreshes
- CodeMirror 6 powers the raw editor
- the JS bundle adds unsaved-change warnings, save flow, copy-path actions, and CSRF headers
