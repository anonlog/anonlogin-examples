import { auth, signIn, signOut } from "@/auth";

export default async function Home() {
  const session = await auth();

  return (
    <main>
      <h1>anonlogin Next.js example</h1>

      {session ? (
        <>
          <p>
            Signed in as <strong>{session.user?.name}</strong>
            {" "}(sub: <code>{session.user?.id}</code>)
          </p>
          <p>
            <a href="/profile">View profile page (server-side protected)</a>
          </p>
          <form
            action={async () => {
              "use server";
              await signOut();
            }}
          >
            <button type="submit">Sign out</button>
          </form>
        </>
      ) : (
        <>
          <p>You are not signed in.</p>
          <form
            action={async () => {
              "use server";
              await signIn("anonlogin");
            }}
          >
            <button type="submit">Sign in with anonlog.in</button>
          </form>
        </>
      )}
    </main>
  );
}
