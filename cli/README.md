# anonlogin CLI examples

The `anonlogin` CLI authenticates via the device flow (RFC 8628) — it opens a
URL in your browser, you approve a one-time code, and the CLI stores a refresh
token in your OS keychain. After that, `anonlogin token` prints a fresh access
token on demand.

## Install

Clone the main repo and build from source:

```bash
git clone git@github.com:anonlog/anonlogin.git
cd anonlogin
go install ./cmd/anonlogin
```

## First-time setup

```bash
# Point the CLI at your instance
anonlogin config set issuer https://anonlog.in

# Sign in — opens https://anonlog.in/device/activate?user_code=… in a browser
anonlogin login

# Confirm everything is working
anonlogin whoami
# => winton
```

## Common commands

```bash
# Print a fresh access token (JWT, ~10 min lifetime)
anonlogin token

# List your API keys
anonlogin apikey list

# Create an API key
anonlogin apikey create --name ci-deploy --scope api:read,api:write

# Revoke an API key
anonlogin apikey revoke <id>
```

## Scripting with the management API

Use `anonlogin token` to get a bearer token for one-off API calls:

```bash
TOKEN=$(anonlogin token)

# List recent auth events
curl "https://anonlog.in/v1/auth-events?limit=20" \
  -H "Authorization: Bearer $TOKEN"

# List active sessions
curl "https://anonlog.in/v1/sessions" \
  -H "Authorization: Bearer $TOKEN"

# Revoke a session
curl -X DELETE "https://anonlog.in/v1/sessions/<id>" \
  -H "Authorization: Bearer $TOKEN"
```

For long-running scripts or CI, use an API key instead of a short-lived token
(create one at `/dashboard/api-keys`):

```bash
# Store the key in an env var or secret manager, then:
curl "https://anonlog.in/v1/auth-events?limit=20" \
  -H "Authorization: Bearer ak_live_anon_..."
```

## Driving the device flow manually

You can replicate what the CLI does from any language:

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
