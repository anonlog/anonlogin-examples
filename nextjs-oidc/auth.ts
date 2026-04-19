import NextAuth from "next-auth";

export const { handlers, signIn, signOut, auth } = NextAuth({
  providers: [
    {
      id: "anonlogin",
      name: "anonlog.in",
      type: "oidc",
      issuer: process.env.AUTH_ANONLOGIN_ISSUER ?? "https://anonlog.in",
      clientId: process.env.AUTH_ANONLOGIN_ID,
      clientSecret: process.env.AUTH_ANONLOGIN_SECRET,
      authorization: { params: { scope: "openid" } },
      checks: ["pkce", "state"],
      // anonlogin preserves user anonymity — the only stable identifier is sub.
      // name, email, and image are intentionally unset.
      profile(profile) {
        return {
          id: profile.sub,
          name: undefined,
          email: null,
          image: null,
        };
      },
    },
  ],
  callbacks: {
    async session({ session, token }) {
      if (token.sub) {
        session.user.id = token.sub;
      }
      return session;
    },
  },
});
