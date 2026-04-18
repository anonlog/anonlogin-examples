# CLI Reference — `anonlogin`

`anonlogin` is the first-party command-line interface for anonlog.in. It handles
authentication, token management, and management of API keys, OAuth clients, and
signing keys.

## Installation

```bash
git clone git@github.com:anonlog/anonlogin-examples.git
cd anonlogin-examples/anonlogin-cli
go install .
```

## Configuration

Settings are stored in `~/.config/anonlogin/config.json`. Tokens are stored in the
OS keychain where available, with a file fallback at
`~/.config/anonlogin/tokens.json`.

| Config key | Default | Description |
|------------|---------|-------------|
| `issuer` | `https://anonlog.in` | Base URL of the anonlogind instance |
| `client_id` | `anonlogin-cli` | OAuth client ID for the CLI |

Override a value with `anonlogin config set <key> <value>` if you are running a
self-hosted instance.

---

## Commands

### `anonlogin login`

Sign in using the OAuth 2.0 Device Authorization Grant (RFC 8628). This is the
recommended login method for interactive terminal use.

```
anonlogin login [--browser]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--browser` | Use browser-based authorization code + PKCE instead of device flow (experimental). |

**Device flow (default):**

1. CLI requests a device code from the server.
2. Prints an 8-character user code and a URL (e.g. `https://anonlog.in/device/activate`).
3. You open the URL in a browser, sign in (username + password + TOTP), and approve.
4. CLI polls until the server issues tokens, then stores them in the keychain.

```
$ anonlogin login

Open this URL in your browser:
  https://anonlog.in/device/activate?user_code=ABCD-1234

Or visit https://anonlog.in/device/activate and enter: ABCD-1234

Waiting for approval...
Logged in as 01HXYZ... (expires in 10m)
```

---

### `anonlogin logout`

Clear stored tokens from the keychain (and file fallback).

```
anonlogin logout
```

Does not call a server-side revocation endpoint. To revoke the underlying OAuth
grant use `anonlogin` via the management API or the dashboard.

---

### `anonlogin whoami`

Print identity information from the stored access token without contacting the
server (the JWT payload is decoded locally, not verified).

```
anonlogin whoami
```

Output:

```
Issuer:  https://anonlog.in
Subject: 01HXYZ...
Scopes:  openid profile
Expires: 2024-06-01 12:10:00 UTC (in 9m)
```

---

### `anonlogin token`

Subcommands for managing the stored token.

#### `anonlogin token print`

Print the raw access token to stdout. Useful for scripting:

```bash
curl -H "Authorization: Bearer $(anonlogin token print)" https://anonlog.in/v1/sessions
```

#### `anonlogin token refresh`

Exchange the stored refresh token for a new access + refresh token pair and
update the keychain.

```
anonlogin token refresh
```

---

### `anonlogin doctor`

Diagnose the CLI's connection to the server.

```
anonlogin doctor
```

Checks performed:

1. Fetches the OIDC discovery document from `{issuer}/.well-known/openid-configuration`.
2. Fetches the JWKS from the `jwks_uri` in the discovery document.
3. Decodes the stored access token and checks expiry.
4. Estimates clock skew from the `Date` response header on the JWKS request.

```
[✓] Discovery  https://anonlog.in/.well-known/openid-configuration
[✓] JWKS       1 key(s) found
[✓] Token      expires in 9m (subject: 01HXYZ...)
[✓] Clock skew ~0s
```

---

### `anonlogin config`

#### `anonlogin config set <key> <value>`

Persist a configuration value.

```bash
# Override the issuer when using a self-hosted instance
anonlogin config set issuer https://my-instance.example.com
anonlogin config set client_id my-cli-client
```

---

### `anonlogin api-key`

Manage long-lived API keys.

#### `anonlogin api-key list`

List all active API keys for the authenticated account.

```
anonlogin api-key list
```

Output:

```
ID              PREFIX        NAME          SCOPES      CREATED      LAST USED
01HXYZ...       ak1abc12      CI pipeline   api:read    2024-01-01   2024-06-01
```

#### `anonlogin api-key create`

Create a new API key. The secret is displayed once and not stored server-side.

```
anonlogin api-key create --name <name> [--scope <scope>]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-n, --name` | (required) | Human-readable name for this key. |
| `-s, --scope` | `api:read` | Space-separated scope string. |

```
$ anonlogin api-key create --name "CI pipeline" --scope "api:read"

New API key created — save the secret, it will not be shown again:

  ak_live_anon_ak1abc12_<32-character-secret>

ID:     01HXYZ...
Name:   CI pipeline
Scopes: api:read
```

#### `anonlogin api-key revoke <id>`

Revoke an API key immediately.

```
anonlogin api-key revoke 01HXYZ...
```

---

### `anonlogin auth-log`

Print the authentication audit log for the authenticated account.

```
anonlogin auth-log [--limit <n>]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-n, --limit` | `50` | Number of events to fetch. |

Output:

```
TIME                  RESULT   METHOD         IP          CLIENT
2024-06-01 12:00:00   ok       password+totp  1.2.3.4     —
2024-06-01 11:55:00   ok       device_flow    1.2.3.4     anonlogin-cli
2024-05-31 09:00:00   fail     password+totp  5.6.7.8     —
```

---

### `anonlogin client`

Manage OAuth clients registered to the authenticated account.

#### `anonlogin client list`

List all registered OAuth clients.

```
anonlogin client list
```

#### `anonlogin client create`

Register a new OAuth client.

```
anonlogin client create --name <name> --redirect-uri <uri> [flags]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-n, --name` | (required) | Display name shown on the consent screen. |
| `-r, --redirect-uri` | (required, repeatable) | Allowed redirect URI. Can be repeated. |
| `-s, --scope` | `openid profile` | Scope (can repeat; e.g. `-s openid -s profile`). |
| `--public` | `false` | Public client (no secret, PKCE required). |
| `--subject-type` | `public` | `public` or `pairwise`. Pairwise derives an opaque per-client `sub`. |
| `--sector-identifier` | _(redirect URI host)_ | Hostname used as the pairwise sector identifier. |
| `--description` | _(empty)_ | Short description shown on the consent screen. |
| `--homepage-url` | _(empty)_ | Client homepage link on the consent screen. |
| `--logo-url` | _(empty)_ | Absolute URL of the client logo for the consent screen. |

The client secret (for confidential clients) is shown once after creation.

#### `anonlogin client rotate-secret <client-id>`

Generate a new client secret. The old secret is invalidated immediately.

```
anonlogin client rotate-secret my-app-a1b2c3d4
```

#### `anonlogin client delete <id>`

Disable an OAuth client. Existing tokens remain valid until expiry.

```
anonlogin client delete 01HXYZ...
```

---

### `anonlogin keys`

#### `anonlogin keys rotate`

Rotate the server's RSA JWT signing key. The old key remains in JWKS for a grace
period so existing valid tokens can still be verified.

```
anonlogin keys rotate
```

Output:

```
New signing key: 01HNEW...
Old key 01HOLD... retained in JWKS for grace period.
```

---

### `anonlogin invite`

Manage registration invite codes. Requires invite-admin privileges (see
`INVITE_ADMINS` — see your instance's configuration for details.

#### `anonlogin invite create`

Create a new invite code.

```
anonlogin invite create [--note <text>] [--expires-in-days <n>]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-m, --note` | _(empty)_ | Optional note describing who this code is for. |
| `-e, --expires-in-days` | `0` | Expire after N days. 0 means no expiry. |

```
$ anonlogin invite create --note "For Alice" --expires-in-days 7

Invite code created
  Code: a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
  URL:  https://anonlog.in/register?invite=a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
  Note: For Alice
```

#### `anonlogin invite list`

List all invite codes (used and unused).

```
anonlogin invite list
```

#### `anonlogin invite delete <code>`

Hard-delete an unused invite code. Fails if the code has already been consumed.

```
anonlogin invite delete a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
```

---

### `anonlogin apps`

Manage connected apps — the per-user consent grants that control which OAuth
clients can act on your behalf.

#### `anonlogin apps list`

List all apps you have approved.

```
anonlogin apps list
```

Output:

```
CLIENT_ID                   NAME                    SCOPES                           GRANTED_AT
────────────────────────────────────────────────────────────────────────────────────────────────────
my-app-a1b2c3d4             My App                  openid profile                   2024-01-01T00:00:00Z
```

#### `anonlogin apps revoke <client_id>`

Revoke a consent grant for a specific client. All active access tokens and
refresh tokens issued to that client on your behalf are invalidated immediately.

```
anonlogin apps revoke my-app-a1b2c3d4
```

---

### `anonlogin grants`

Manage active OAuth refresh-token grants (one per CLI or app login session).

#### `anonlogin grants list`

List all active grants.

```
anonlogin grants list
```

Output:

```
REQUEST_ID                  CLIENT_ID                 SCOPES                           CREATED_AT
────────────────────────────────────────────────────────────────────────────────────────────────────
01HXYZ...                   anonlogin-cli               openid profile                   2024-06-01T12:00:00Z
```

#### `anonlogin grants revoke <request_id>`

Revoke a grant and its entire refresh-token family. The next refresh attempt
using any token from this grant will fail.

```
anonlogin grants revoke 01HXYZ...
```

---

## Scripting

The CLI is designed to be scriptable. Use `anonlogin token print` to get a bearer
token for `curl` or other HTTP clients:

```bash
# List sessions via API
curl -s \
  -H "Authorization: Bearer $(anonlogin token print)" \
  https://anonlog.in/v1/sessions | jq .

# Revoke a session
curl -s -X DELETE \
  -H "Authorization: Bearer $(anonlogin token print)" \
  https://anonlog.in/v1/sessions/01HXYZ...

# Use an API key instead (no CLI required)
curl -s \
  -H "Authorization: Bearer ak_live_anon_ak1abc12_<secret>" \
  https://anonlog.in/v1/api-keys
```

## Driving the device flow without the CLI

You can replicate what the CLI does from any language or shell script:

```bash
ISSUER=https://anonlog.in

# 1. Request a device code
RESPONSE=$(curl -s -X POST "$ISSUER/device/code" \
  -d "client_id=anonlogin-cli&scope=openid+profile+offline_access")

DEVICE_CODE=$(echo "$RESPONSE" | jq -r .device_code)
USER_CODE=$(echo "$RESPONSE"   | jq -r .user_code)
INTERVAL=$(echo "$RESPONSE"    | jq -r .interval)

echo "Go to: $ISSUER/device/activate?user_code=$USER_CODE"

# 2. Poll until approved (respect the interval)
while true; do
  sleep "$INTERVAL"
  TOKEN_RESPONSE=$(curl -s -X POST "$ISSUER/device/token" \
    -d "grant_type=urn:ietf:params:oauth:grant-type:device_code" \
    -d "device_code=$DEVICE_CODE" \
    -d "client_id=anonlogin-cli")

  ERROR=$(echo "$TOKEN_RESPONSE" | jq -r '.error // empty')
  if [ -z "$ERROR" ]; then
    echo "Access token:"
    echo "$TOKEN_RESPONSE" | jq -r .access_token
    break
  elif [ "$ERROR" != "authorization_pending" ]; then
    echo "Error: $ERROR"
    break
  fi
done
```
