"use client";

import React from "react";
import { Tag } from "antd";
import { Product } from "@/types";
import { API_BASE } from "@/lib/api";

interface ProductCardProps {
  product: Product;
  onClick?: () => void;
}

export default function ProductCard({ product, onClick }: ProductCardProps) {
  let background = "#fff";
  let color = "#000";

  if (product.discount > 15) {
    background = "#2E8B57";
    color = "#fff";
  } else if (product.quantity === 0) {
    background = "#e0f2fe";
    color = "#000";
  }

  const imgSrc =
    product.image && product.image.trim() !== ""
      ? `${API_BASE}/uploads/${product.image}`
      : "/placeholder.png";

  const discountedPrice =
    product.discount > 0
      ? product.price * (1 - product.discount / 100)
      : null;

  return (
    <div
      onClick={onClick}
      style={{
        display: "flex",
        alignItems: "center",
        gap: 16,
        padding: "12px 16px",
        borderBottom: "1px solid #e8e8e8",
        background,
        color,
        cursor: onClick ? "pointer" : "default",
        transition: "opacity 0.15s",
      }}
      onMouseEnter={(e) => {
        if (onClick) (e.currentTarget as HTMLDivElement).style.opacity = "0.85";
      }}
      onMouseLeave={(e) => {
        if (onClick) (e.currentTarget as HTMLDivElement).style.opacity = "1";
      }}
    >
      {/* Image */}
      <img
        src={imgSrc}
        alt={product.name}
        style={{
          width: 60,
          height: 60,
          objectFit: "cover",
          borderRadius: 8,
          flexShrink: 0,
        }}
      />

      {/* Middle: name, meta, description */}
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ fontWeight: 700, fontSize: 15, marginBottom: 2 }}>
          {product.name}
        </div>
        <div style={{ fontSize: 13, opacity: 0.75, marginBottom: 2 }}>
          {[product.categoryName, product.manufacturerName, product.supplierName]
            .filter(Boolean)
            .join(" · ")}
        </div>
        {product.description && (
          <div
            style={{
              fontSize: 12,
              opacity: 0.6,
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            {product.description}
          </div>
        )}
      </div>

      {/* Right: price, quantity, discount badge */}
      <div style={{ textAlign: "right", flexShrink: 0, minWidth: 120 }}>
        {/* Price */}
        <div style={{ marginBottom: 4 }}>
          {product.discount > 0 && discountedPrice !== null ? (
            <>
              <span
                style={{
                  textDecoration: "line-through",
                  color: "red",
                  marginRight: 6,
                  fontSize: 13,
                }}
              >
                {product.price.toFixed(2)}₽
              </span>
              <span style={{ fontWeight: 700, color: "#000", fontSize: 15 }}>
                {discountedPrice.toFixed(2)}₽
              </span>
            </>
          ) : (
            <span style={{ fontWeight: 700, fontSize: 15, color }}>
              {product.price.toFixed(2)}₽
            </span>
          )}
        </div>

        {/* Quantity */}
        <div style={{ fontSize: 13, marginBottom: product.discount > 0 ? 4 : 0 }}>
          {product.quantity === 0 ? (
            <span style={{ color: "red" }}>Нет в наличии</span>
          ) : (
            <span>
              В наличии: {product.quantity} {product.unitName}
            </span>
          )}
        </div>

        {/* Discount badge */}
        {product.discount > 0 && (
          <Tag color="volcano" style={{ marginTop: 2 }}>
            -{product.discount}%
          </Tag>
        )}
      </div>
    </div>
  );
}
