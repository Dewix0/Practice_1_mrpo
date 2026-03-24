"use client";

import React, { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Button, Spin } from "antd";
import { useAuth } from "@/lib/auth";
import { apiFetch } from "@/lib/api";
import { Product } from "@/types";
import ProductCard from "@/components/ProductCard";
import ProductToolbar from "@/components/ProductToolbar";

interface Filters {
  search: string;
  supplierId: number;
  sort: string;
}

export default function ProductsPage() {
  const { isAdmin, isManager } = useAuth();
  const router = useRouter();

  const [filters, setFilters] = useState<Filters>({
    search: "",
    supplierId: 0,
    sort: "",
  });
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    document.title = "Товары — ООО Обувь";
  }, []);

  useEffect(() => {
    setLoading(true);
    const params = new URLSearchParams();
    if (filters.search) params.set("search", filters.search);
    if (filters.supplierId) params.set("supplier_id", String(filters.supplierId));
    if (filters.sort) params.set("sort", filters.sort);

    const query = params.toString();
    apiFetch<Product[]>(`/api/products${query ? `?${query}` : ""}`)
      .then(setProducts)
      .catch(() => setProducts([]))
      .finally(() => setLoading(false));
  }, [filters]);

  const handleFilterChange = (newFilters: Filters) => {
    setFilters(newFilters);
  };

  const handleCardClick = (id: number) => {
    if (isAdmin) {
      router.push(`/products/${id}/edit`);
    }
  };

  return (
    <div>
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          marginBottom: 16,
          flexWrap: "wrap",
          gap: 12,
        }}
      >
        <h1 style={{ margin: 0, fontSize: 24, fontWeight: 700 }}>Товары</h1>
        {isAdmin && (
          <Button
            onClick={() => router.push("/products/new")}
            style={{
              background: "#00FA9A",
              borderColor: "#00FA9A",
              color: "#000",
              fontWeight: 600,
            }}
          >
            Добавить товар
          </Button>
        )}
      </div>

      {(isAdmin || isManager) && (
        <ProductToolbar onFilterChange={handleFilterChange} />
      )}

      {loading ? (
        <div style={{ display: "flex", justifyContent: "center", padding: 48 }}>
          <Spin size="large" />
        </div>
      ) : products.length === 0 ? (
        <div style={{ textAlign: "center", padding: 48, color: "#999" }}>
          Товары не найдены
        </div>
      ) : (
        <div
          style={{
            border: "1px solid #e8e8e8",
            borderRadius: 8,
            overflow: "hidden",
          }}
        >
          {products.map((product) => (
            <ProductCard
              key={product.id}
              product={product}
              onClick={isAdmin ? () => handleCardClick(product.id) : undefined}
            />
          ))}
        </div>
      )}
    </div>
  );
}
