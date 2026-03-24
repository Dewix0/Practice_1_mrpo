"use client";

import React, { useEffect } from "react";
import { useRouter } from "next/navigation";
import ProductForm from "@/components/ProductForm";

export default function ProductNewPage() {
  const router = useRouter();

  useEffect(() => {
    document.title = "Добавление товара — ООО Обувь";
  }, []);

  return (
    <div>
      <h1 style={{ fontSize: 24, fontWeight: 700, marginBottom: 24 }}>
        Добавление товара
      </h1>
      <ProductForm onSuccess={() => router.push("/products")} />
    </div>
  );
}
