# anonlogin examples

Example projects showing how to integrate [anonlogin](https://github.com/anonlog/anonlogin)
as an identity provider.

## Examples

| Directory | What it shows |
|-----------|--------------|
| [`anonlogin-cli/`](./anonlogin-cli/) | First-party CLI — device flow login, token management, management API |
| [`nextjs-oidc/`](./nextjs-oidc/) | Next.js app using Auth.js with anonlogin as an OIDC provider |
| [`cli/`](./cli/) | Raw shell examples — device flow and API calls without the CLI |

## Prerequisites

You need a running anonlogin instance. You can run one locally:

```bash
git clone git@github.com:anonlog/anonlogin.git
cd anonlogin
cp .env.example .env   # fill in BASE_URL and generated secrets
make up                # starts Postgres + anonlogind on :8080
```

Or point the examples at any existing instance (e.g. `https://anonlog.in`).
