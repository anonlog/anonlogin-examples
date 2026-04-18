import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "anonlogin Next.js example",
  description: "Example Next.js app using anonlogin as an OIDC provider.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body style={{ fontFamily: "system-ui, sans-serif", maxWidth: 640, margin: "64px auto", padding: "0 16px" }}>
        {children}
      </body>
    </html>
  );
}
