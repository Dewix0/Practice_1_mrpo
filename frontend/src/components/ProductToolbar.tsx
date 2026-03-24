"use client";

import React, { useState, useEffect, useRef } from "react";
import { Input, Select, Button, Space } from "antd";
import { apiFetch } from "@/lib/api";
import { RefItem } from "@/types";

interface Filters {
  search: string;
  supplierId: number;
  sort: string;
}

interface ProductToolbarProps {
  onFilterChange: (filters: Filters) => void;
}

export default function ProductToolbar({ onFilterChange }: ProductToolbarProps) {
  const [search, setSearch] = useState("");
  const [supplierId, setSupplierId] = useState<number>(0);
  const [sort, setSort] = useState("");
  const [suppliers, setSuppliers] = useState<RefItem[]>([]);

  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Fetch suppliers on mount
  useEffect(() => {
    apiFetch<RefItem[]>("/api/suppliers")
      .then(setSuppliers)
      .catch(() => setSuppliers([]));
  }, []);

  // Report changes to parent
  const report = (newSearch: string, newSupplierId: number, newSort: string) => {
    onFilterChange({ search: newSearch, supplierId: newSupplierId, sort: newSort });
  };

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setSearch(value);
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => {
      report(value, supplierId, sort);
    }, 300);
  };

  const handleSupplierChange = (value: number) => {
    setSupplierId(value);
    report(search, value, sort);
  };

  const handleSortClick = (value: string) => {
    const newSort = sort === value ? "" : value;
    setSort(newSort);
    report(search, supplierId, newSort);
  };

  const supplierOptions = [
    { value: 0, label: "Все поставщики" },
    ...suppliers.map((s) => ({ value: s.id, label: s.name })),
  ];

  return (
    <Space wrap style={{ marginBottom: 16 }}>
      <Input
        placeholder="Поиск..."
        value={search}
        onChange={handleSearchChange}
        style={{ width: 240 }}
        allowClear
        onClear={() => {
          setSearch("");
          if (debounceRef.current) clearTimeout(debounceRef.current);
          report("", supplierId, sort);
        }}
      />

      <Select
        value={supplierId}
        onChange={handleSupplierChange}
        options={supplierOptions}
        style={{ width: 200 }}
      />

      <Button
        type={sort === "quantity_asc" ? "primary" : "default"}
        onClick={() => handleSortClick("quantity_asc")}
      >
        По количеству ↑
      </Button>

      <Button
        type={sort === "quantity_desc" ? "primary" : "default"}
        onClick={() => handleSortClick("quantity_desc")}
      >
        По количеству ↓
      </Button>
    </Space>
  );
}
