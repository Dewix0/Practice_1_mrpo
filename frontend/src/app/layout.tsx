import type { Metadata } from "next";
import { AntdRegistry } from "@ant-design/nextjs-registry";
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
        <AntdRegistry>{children}</AntdRegistry>
      </body>
    </html>
  );
}
