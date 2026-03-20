# openclaudio configuration

Configuration is loaded in this order:

1. hardcoded defaults
2. `~/.config/openclaudio/config.json` or `OPENCLAUDIO_CONFIG`
3. environment variables

Required environment variables:

- `OPENCLAUDIO_SESSION_SECRET`
- `OPENCLAUDIO_ADMIN_USER`
- `OPENCLAUDIO_ADMIN_PASSWORD_HASH`

Optional overrides:

- `OPENCLAUDIO_PORT`
- `OPENCLAUDIO_WORKSPACE_ROOT`
- `OPENCLAUDIO_OPENCLAW_ROOT`
- `OPENCLAUDIO_GATEWAY_UNIT`
- `OPENCLAUDIO_LOG_DIR`
- `OPENCLAUDIO_OPENCLAW_PACKAGE_JSON`

`OPENCLAUDIO_ADMIN_PASSWORD_HASH` must be a bcrypt hash.
