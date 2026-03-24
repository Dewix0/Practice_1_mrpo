"use client";

import React, { useEffect } from "react";
import { useRouter } from "next/navigation";
import Image from "next/image";
import { Form, Input, Button, message } from "antd";
import { useAuth } from "@/lib/auth";

export default function LoginPage() {
  const { login } = useAuth();
  const router = useRouter();
  const [messageApi, contextHolder] = message.useMessage();

  useEffect(() => {
    document.title = "Вход — ООО Обувь";
  }, []);

  const onFinish = async (values: { login: string; password: string }) => {
    try {
      await login(values.login, values.password);
      router.push("/products");
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : "Ошибка входа";
      messageApi.error(errorMessage);
    }
  };

  return (
    <>
      {contextHolder}
      <div
        style={{
          minHeight: "80vh",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <div
          style={{
            maxWidth: 360,
            width: "100%",
            padding: 32,
            borderRadius: 12,
            border: "1px solid #e8e8e8",
            background: "#fff",
            boxShadow: "0 2px 12px rgba(0,0,0,0.08)",
          }}
        >
          <div style={{ textAlign: "center", marginBottom: 24 }}>
            <Image src="/logo.png" alt="Логотип" height={64} width={0} style={{ width: "auto", height: 64 }} />
            <div style={{ fontWeight: 700, fontSize: 20, marginTop: 8 }}>ООО Обувь</div>
          </div>

          <Form layout="vertical" onFinish={onFinish}>
            <Form.Item
              name="login"
              rules={[{ required: true, message: "Введите логин" }]}
            >
              <Input placeholder="Введите логин" />
            </Form.Item>

            <Form.Item
              name="password"
              rules={[{ required: true, message: "Введите пароль" }]}
            >
              <Input.Password placeholder="Введите пароль" />
            </Form.Item>

            <Form.Item>
              <Button
                type="primary"
                htmlType="submit"
                style={{ width: "100%", background: "#00FA9A", borderColor: "#00FA9A", color: "#000" }}
              >
                Войти
              </Button>
            </Form.Item>
          </Form>

          <div style={{ textAlign: "center" }}>
            <a
              onClick={() => router.push("/products")}
              style={{ cursor: "pointer", color: "#1677ff" }}
            >
              Войти как гость →
            </a>
          </div>
        </div>
      </div>
    </>
  );
}
