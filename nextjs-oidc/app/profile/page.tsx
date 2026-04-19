import { auth } from "@/auth";
import { redirect } from "next/navigation";

export default async function ProfilePage() {
  const session = await auth();

  if (!session) {
    redirect("/api/auth/signin");
  }

  return (
    <main>
      <h1>Profile</h1>
      <p><a href="/">← Back</a></p>
      <table style={{ borderCollapse: "collapse", width: "100%" }}>
        <tbody>
          <tr>
            <th style={th}>Field</th>
            <th style={th}>Value</th>
          </tr>
          <tr>
            <td style={td}>Subject (sub)</td>
            <td style={td}><code>{session.user?.id}</code></td>
          </tr>
        </tbody>
      </table>
      <h2>Raw session</h2>
      <pre style={{ background: "#f4f4f4", padding: 16, borderRadius: 6, overflowX: "auto" }}>
        {JSON.stringify(session, null, 2)}
      </pre>
    </main>
  );
}

const th: React.CSSProperties = {
  textAlign: "left",
  padding: "8px 12px",
  borderBottom: "2px solid #ddd",
};

const td: React.CSSProperties = {
  padding: "8px 12px",
  borderBottom: "1px solid #eee",
};
