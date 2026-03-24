"use client";

import React, { useEffect, useState } from "react";
import { useRouter, useParams } from "next/navigation";
import { Spin, Button, Modal, notification } from "antd";
import { apiFetch } from "@/lib/api";
import { Order } from "@/types";
import OrderForm from "@/components/OrderForm";

export default function OrderEditPage() {
  const router = useRouter();
  const params = useParams();
  const id = params?.id as string;

  const [order, setOrder] = useState<Order | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    document.title = "Редактирование заказа — ООО Обувь";
  }, []);

  useEffect(() => {
    if (!id) return;
    apiFetch<Order>(`/api/orders/${id}`)
      .then(setOrder)
      .catch(() => setOrder(null))
      .finally(() => setLoading(false));
  }, [id]);

  const handleDelete = () => {
    Modal.confirm({
      title: "Удаление заказа",
      content: "Вы уверены, что хотите удалить этот заказ?",
      okText: "Удалить",
      okType: "danger",
      cancelText: "Отмена",
      onOk: async () => {
        try {
          await apiFetch(`/api/orders/${id}`, { method: "DELETE" });
          router.push("/orders");
        } catch (err: unknown) {
          const message = err instanceof Error ? err.message : "Неизвестная ошибка";
          notification.error({ message: "Ошибка", description: message });
        }
      },
    });
  };

  if (loading) {
    return (
      <div style={{ display: "flex", justifyContent: "center", padding: 48 }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!order) {
    return <div style={{ textAlign: "center", padding: 48 }}>Заказ не найден</div>;
  }

  return (
    <div>
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          marginBottom: 24,
        }}
      >
        <h1 style={{ fontSize: 24, fontWeight: 700, margin: 0 }}>
          Редактирование заказа
        </h1>
        <Button danger onClick={handleDelete}>
          Удалить
        </Button>
      </div>
      <OrderForm order={order} onSuccess={() => router.push("/orders")} />
    </div>
  );
}
