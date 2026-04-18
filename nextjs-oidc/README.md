# Next.js + anonlogin OIDC example

A minimal Next.js 15 app that signs users in via [anonlogin](https://anonlog.in)
using Auth.js (NextAuth v5) and the OpenID Connect authorization code flow with PKCE.

## What it demonstrates

- Configuring Auth.js with a custom OIDC provider pointed at your anonlogin instance
- Server-side session access with `auth()` in React Server Components
- A protected page that redirects unauthenticated users to sign-in
- Reading `preferred_username` and `sub` from the OIDC token

## Setup

### 1. Register an OAuth client

In your anonlogin dashboard at `/dashboard/clients`, click **Register a new client**:

| Field | Value |
|-------|-------|
| Application name | `Next.js example` (or anything) |
| Redirect URI | `http://localhost:3000/api/auth/callback/anonlogin` |
| Scopes | `openid profile` |
| Public client | Leave unchecked (confidential) |

Copy the `client_id` and `client_secret` shown after registration.

### 2. Configure environment variables

```bash
cp .env.local.example .env.local
```

Edit `.env.local`:

```bash
AUTH_ANONLOGIN_ID=cid_your_client_id
AUTH_ANONLOGIN_SECRET=cs_your_client_secret
AUTH_SECRET=$(openssl rand -base64 32)

# Only needed when using a self-hosted instance (defaults to https://anonlog.in)
# AUTH_ANONLOGIN_ISSUER=https://your-instance.example.com
```

### 3. Install and run

```bash
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) and click **Sign in with anonlog.in**.

## How it works

```
Browser → /api/auth/signin/anonlogin
        → GET /authorize (anonlogin — user logs in with password + TOTP)
        → consent screen (first use only)
        → GET /api/auth/callback/anonlogin?code=…
        → POST /token (Auth.js exchanges code for tokens)
        → session cookie set, user redirected to /
```

Auth.js handles the PKCE code challenge/verifier, the token exchange, and the
encrypted session cookie automatically. The app never sees the raw tokens.

## Files

| File | Purpose |
|------|---------|
| `auth.ts` | Auth.js configuration: provider, profile mapping, session/JWT callbacks |
| `app/api/auth/[...nextauth]/route.ts` | Mounts the Auth.js request handlers |
| `app/page.tsx` | Home page — shows sign-in/out button and session info |
| `app/profile/page.tsx` | Protected page — redirects to sign-in if no session |

## Using the access token

If your app needs to call APIs on behalf of the signed-in user, request the
`offline_access` scope and store the access token in the JWT callback:

```ts
// auth.ts
callbacks: {
  async jwt({ token, account }) {
    if (account?.access_token) {
      token.accessToken = account.access_token;
    }
    return token;
  },
  async session({ session, token }) {
    session.accessToken = token.accessToken as string;
    return session;
  },
},
```

Then use it in a Server Component or Route Handler:

```ts
const session = await auth();
const res = await fetch(`${process.env.AUTH_ANONLOGIN_ISSUER}/v1/auth-events`, {
  headers: { Authorization: `Bearer ${session?.accessToken}` },
});
```
