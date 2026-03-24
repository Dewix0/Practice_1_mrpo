import type { Metadata } from "next";
import { AntdRegistry } from "@ant-design/nextjs-registry";
import { AuthProvider } from "@/lib/auth";
import AppHeader from "@/components/AppHeader";
import "./globals.css";

export const metadata: Metadata = {
  title: "ООО Обувь",
  description: "Информационная система магазина обуви",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ru">
      <body>
        <AntdRegistry>
          <AuthProvider>
            <AppHeader />
            <main style={{ padding: "24px", maxWidth: 1200, margin: "0 auto" }}>
              {children}
            </main>
          </AuthProvider>
        </AntdRegistry>
      </body>
    </html>
  );
}
