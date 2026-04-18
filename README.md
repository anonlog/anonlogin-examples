# anonlogin examples

Example projects showing how to integrate [anonlogin](https://anonlog.in)
as an identity provider.

## Examples

| Directory | What it shows |
|-----------|--------------|
| [`anonlogin-cli/`](./anonlogin-cli/) | First-party CLI — device flow login, token management, management API |
| [`nextjs-oidc/`](./nextjs-oidc/) | Next.js app using Auth.js with anonlogin as an OIDC provider |

## Prerequisites

- An [anonlog.in](https://anonlog.in) account
- [Go](https://go.dev/) toolchain (to install the CLI)
- [Node.js](https://nodejs.org/) 18+ (for the Next.js example)

## Quick start: nextjs-oidc via the CLI

Everything below uses only the terminal — no dashboard required.

### 1. Install the CLI

```bash
git clone git@github.com:anonlog/anonlogin-examples.git
cd anonlogin-examples/anonlogin-cli
go install .
```

The binary is installed as `anonlogin-cli`. Verify with:

```bash
anonlogin-cli --help
```

### 2. Sign in

```bash
anonlogin-cli login
```

This starts the OAuth device flow. The CLI prints a URL and a one-time code:

```
Activate your device:

  URL:  https://anonlog.in/device/activate
  Code: ABCD-WXYZ

  Or go directly to:
  https://anonlog.in/device/activate?code=ABCD-WXYZ
```

Open the URL in your browser, sign in with your anonlogin account
(username + password + TOTP), and approve the device. The CLI stores
tokens in your OS keychain.

### 3. Register an OAuth client

```bash
anonlogin-cli client create \
  --name "Next.js OIDC example" \
  --redirect-uri "http://localhost:3000/api/auth/callback/anonlogin"
```

The output shows the `client_id` and `client_secret` — save them, the
secret is only displayed once:

```
OAuth client registered
  Client ID: cid_...
  Name:      Next.js OIDC example
  Secret:    cs_...
  Note:      This secret will not be shown again.
```

### 4. Configure the Next.js app

```bash
cd nextjs-oidc
cp .env.local.example .env.local
```

Edit `.env.local` with the values from step 3:

```bash
AUTH_ANONLOGIN_ID=cid_...        # Client ID from step 3
AUTH_ANONLOGIN_SECRET=cs_...     # Client secret from step 3
AUTH_SECRET=$(openssl rand -base64 32)
```

### 5. Run

```bash
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) and click
**Sign in with anonlog.in**.

## CLI command reference

See [`anonlogin-cli/COMMANDS.md`](./anonlogin-cli/COMMANDS.md) for the
full list of commands (client management, API keys, invites, grants,
key rotation, and more).
