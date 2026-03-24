"use client";

import React, { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Table, Tag, Spin, Button } from "antd";
import type { TableColumnsType } from "antd";
import { apiFetch } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { Order } from "@/types";

export default function OrdersPage() {
  const router = useRouter();
  const { isAdmin } = useAuth();
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    document.title = "Заказы — ООО Обувь";
  }, []);

  useEffect(() => {
    apiFetch<Order[]>("/api/orders")
      .then(setOrders)
      .catch(() => setOrders([]))
      .finally(() => setLoading(false));
  }, []);

  const columns: TableColumnsType<Order> = [
    {
      title: "Артикулы",
      key: "articles",
      render: (_: unknown, record: Order) =>
        record.items.map((i) => i.productArticle).join(", ") || "—",
    },
    {
      title: "Статус",
      key: "status",
      render: (_: unknown, record: Order) => (
        <Tag color={record.statusName === "Завершен" ? "green" : "orange"}>
          {record.statusName}
        </Tag>
      ),
    },
    {
      title: "Пункт выдачи",
      dataIndex: "pickupAddress",
      key: "pickupAddress",
    },
    {
      title: "Дата заказа",
      dataIndex: "orderDate",
      key: "orderDate",
    },
    {
      title: "Дата доставки",
      key: "deliveryDate",
      render: (_: unknown, record: Order) => record.deliveryDate || "—",
    },
  ];

  if (loading) {
    return (
      <div style={{ display: "flex", justifyContent: "center", padding: 48 }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div>
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          marginBottom: 16,
        }}
      >
        <h1 style={{ margin: 0, fontSize: 24, fontWeight: 700 }}>Заказы</h1>
        {isAdmin && (
          <Button
            onClick={() => router.push("/orders/new")}
            style={{
              background: "#00FA9A",
              borderColor: "#00FA9A",
              color: "#000",
              fontWeight: 600,
            }}
          >
            Добавить заказ
          </Button>
        )}
      </div>

      <Table
        dataSource={orders}
        columns={columns}
        rowKey="id"
        onRow={
          isAdmin
            ? (record) => ({
                onClick: () => router.push(`/orders/${record.id}/edit`),
                style: { cursor: "pointer" },
              })
            : undefined
        }
      />
    </div>
  );
}
