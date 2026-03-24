"use client";

import React, { useEffect, useState } from "react";
import { useRouter, useParams } from "next/navigation";
import { Spin, Button, Modal, notification } from "antd";
import { apiFetch } from "@/lib/api";
import { Product } from "@/types";
import ProductForm from "@/components/ProductForm";

export default function ProductEditPage() {
  const router = useRouter();
  const params = useParams();
  const id = params?.id as string;

  const [product, setProduct] = useState<Product | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    document.title = "Редактирование товара — ООО Обувь";
  }, []);

  useEffect(() => {
    if (!id) return;
    apiFetch<Product>(`/api/products/${id}`)
      .then(setProduct)
      .catch(() => setProduct(null))
      .finally(() => setLoading(false));
  }, [id]);

  const handleDelete = () => {
    Modal.confirm({
      title: "Удаление товара",
      content: "Вы уверены, что хотите удалить этот товар?",
      okText: "Удалить",
      okType: "danger",
      cancelText: "Отмена",
      onOk: async () => {
        try {
          await apiFetch(`/api/products/${id}`, { method: "DELETE" });
          router.push("/products");
        } catch (err: unknown) {
          const message = err instanceof Error ? err.message : "";
          if (message.includes("409") || message.toLowerCase().includes("order")) {
            Modal.warning({
              title: "Невозможно удалить",
              content: "Товар присутствует в заказе и не может быть удалён",
            });
          } else {
            notification.error({ message: "Ошибка", description: message });
          }
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

  if (!product) {
    return <div style={{ textAlign: "center", padding: 48 }}>Товар не найден</div>;
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
          Редактирование товара
        </h1>
        <Button danger onClick={handleDelete}>
          Удалить
        </Button>
      </div>
      <ProductForm product={product} onSuccess={() => router.push("/products")} />
    </div>
  );
}
