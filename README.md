# openclaudio

Lightweight Go web panel for daily OpenClaw management.

## Features

- browse the OpenClaw workspace with a markdown-first file tree
- edit UTF-8 text files with CodeMirror 6
- render markdown server-side with sanitization
- show OpenClaw version, models, sessions, cron status, logs, service state, and process metrics
- allow only `start`, `stop`, and `restart` for the configured gateway systemd unit

## Requirements

- Go
- npm
- a local OpenClaw install rooted at `~/.openclaw`

## Configuration

Copy `.env.example` and set at least:

- `OPENCLAUDIO_SESSION_SECRET`
- `OPENCLAUDIO_ADMIN_USER`
- `OPENCLAUDIO_ADMIN_PASSWORD_HASH`

Defaults:

- bind address: `127.0.0.1`
- workspace root: `~/.openclaw/workspace`
- OpenClaw root: `~/.openclaw`
- gateway unit: `openclaw-gateway.service`
- port: `18890`

The password hash must be bcrypt.

Generate one with a single command:

```bash
go run ./cmd/openclaudio hash-password 'your-password-here'
```

Or after building:

```bash
./openclaudio hash-password 'your-password-here'
```

Use the printed hash as the value of `OPENCLAUDIO_ADMIN_PASSWORD_HASH`.

`openclaudio` automatically loads `.env` from the current working directory. You do not need to `source` it before `make run`.

## Development

Install frontend dependencies:

```bash
npm install
```

Build assets and the Go binary:

```bash
make build
```

Run tests:

```bash
make test
```

Run locally:

```bash
OPENCLAUDIO_SESSION_SECRET=change-me \
OPENCLAUDIO_ADMIN_USER=admin \
OPENCLAUDIO_ADMIN_PASSWORD_HASH='$2a$10$replace.with.real.bcrypt.hash' \
make run
```

Then open `http://127.0.0.1:18890`.

To expose the app on your local network, set:

```bash
OPENCLAUDIO_BIND_ADDRESS=0.0.0.0
```

Then access it from another machine using `http://YOUR-HOST-IP:18890`.

## Deployment

An example user service is in `deploy/systemd/openclaudio.service`.

The app is intended for localhost or trusted intranet use. Secure cookies are enabled when the request is HTTPS, including when HTTPS is terminated by a reverse proxy that forwards the protocol correctly.

When the app is exposed directly over plain HTTP on a LAN, browser cookies cannot be marked `Secure`. If you need strict secure-cookie behavior, put the app behind HTTPS.

## Verification

```bash
npm run build
make test
```

## Notes

- invalid UTF-8 files are shown read-only in v1
- gateway control depends on `systemctl --user` access for the runtime user
- more detailed notes are in `docs/architecture.md` and `docs/configuration.md`
