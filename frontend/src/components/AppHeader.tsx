"use client";

import React from "react";
import Link from "next/link";
import Image from "next/image";
import { Button, Space } from "antd";
import { useAuth } from "@/lib/auth";
import { User } from "@/types";

function getFullName(user: User): string {
  let name = user.lastName || "";
  if (user.firstName) name += " " + user.firstName.charAt(0) + ".";
  if (user.patronymic) name += user.patronymic.charAt(0) + ".";
  return name;
}

export default function AppHeader() {
  const { user, isGuest, isAdmin, isManager, logout } = useAuth();

  return (
    <header
      style={{
        backgroundColor: "#7FFF00",
        padding: "0 24px",
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        height: 56,
      }}
    >
      <Space size="large" align="center">
        <Link href="/products" style={{ display: "flex", alignItems: "center" }}>
          <Image src="/logo.png" alt="Логотип" height={40} width={0} style={{ width: "auto", height: 40 }} />
        </Link>
        <span style={{ fontWeight: 700, fontSize: 18 }}>ООО Обувь</span>
        <Link href="/products" style={{ color: "#000", textDecoration: "none", fontWeight: 500 }}>
          Товары
        </Link>
        {(isAdmin || isManager) && (
          <Link href="/orders" style={{ color: "#000", textDecoration: "none", fontWeight: 500 }}>
            Заказы
          </Link>
        )}
      </Space>

      <Space align="center">
        {!isGuest && user ? (
          <>
            <span style={{ fontWeight: 500 }}>{getFullName(user)}</span>
            <Button onClick={logout} size="small">
              Выход
            </Button>
          </>
        ) : (
          <Link href="/login" style={{ color: "#000", fontWeight: 500 }}>
            Войти
          </Link>
        )}
      </Space>
    </header>
  );
}
