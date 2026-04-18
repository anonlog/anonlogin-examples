# anonlogin CLI

Command-line client for [anonlog.in](https://anonlog.in). Authenticates via the
OAuth 2.0 Device Authorization Grant (RFC 8628), stores tokens in the OS keychain,
and wraps the management API with typed subcommands.

## Install

```bash
git clone git@github.com:anonlog/anonlogin-examples.git
cd anonlogin-examples/anonlogin-cli
go install .
```

## Quick start

```bash
# Sign in — opens a browser URL with a one-time code
anonlogin login

anonlogin whoami
anonlogin token print        # prints a fresh JWT to stdout

# API key management
anonlogin api-key list
anonlogin api-key create --name ci-deploy --scope api:read,api:write
anonlogin api-key revoke <id>
```

See [COMMANDS.md](./COMMANDS.md) for the full command reference.
