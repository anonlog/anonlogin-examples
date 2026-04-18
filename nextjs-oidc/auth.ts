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
      // anonlogin returns the username in preferred_username
      profile(profile) {
        return {
          id: profile.sub,
          name: profile.preferred_username ?? profile.sub,
          email: null,
          image: null,
        };
      },
    },
  ],
  callbacks: {
    // Expose the raw OIDC profile fields on the session so components can
    // read preferred_username without an extra /userinfo round-trip.
    async session({ session, token }) {
      if (token.preferred_username) {
        session.user.name = token.preferred_username as string;
      }
      if (token.sub) {
        session.user.id = token.sub;
      }
      return session;
    },
    async jwt({ token, profile }) {
      if (profile?.preferred_username) {
        token.preferred_username = profile.preferred_username;
      }
      return token;
    },
  },
});
