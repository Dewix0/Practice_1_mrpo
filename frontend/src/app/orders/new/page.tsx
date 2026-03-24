"use client";

import React, { useEffect } from "react";
import { useRouter } from "next/navigation";
import OrderForm from "@/components/OrderForm";

export default function OrderNewPage() {
  const router = useRouter();

  useEffect(() => {
    document.title = "Добавление заказа — ООО Обувь";
  }, []);

  return (
    <div>
      <h1 style={{ fontSize: 24, fontWeight: 700, marginBottom: 24 }}>
        Добавление заказа
      </h1>
      <OrderForm onSuccess={() => router.push("/orders")} />
    </div>
  );
}
