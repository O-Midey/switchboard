import type { Metadata } from "next";
import "./styles.css";

export const metadata: Metadata = { title: "Switchboard Control Room", description: "Operational view of the Switchboard adaptive reverse proxy." };

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return <html lang="en"><body>{children}</body></html>;
}

