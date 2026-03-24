"use client";

import React, { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import {
  Form,
  Input,
  InputNumber,
  Select,
  Button,
  Upload,
  Row,
  Col,
  notification,
} from "antd";
import { UploadOutlined } from "@ant-design/icons";
import type { UploadFile } from "antd";
import { apiFetch, apiUpload, API_BASE } from "@/lib/api";
import { Product, RefItem } from "@/types";

const { TextArea } = Input;

interface ProductFormProps {
  product?: Product;
  onSuccess: () => void;
}

interface FormValues {
  name: string;
  article?: string;
  categoryId: number;
  manufacturerId: number;
  supplierId: number;
  description?: string;
  price: number;
  discount?: number;
  quantity: number;
  unitId: number;
}

export default function ProductForm({ product, onSuccess }: ProductFormProps) {
  const router = useRouter();
  const [form] = Form.useForm<FormValues>();
  const [categories, setCategories] = useState<RefItem[]>([]);
  const [manufacturers, setManufacturers] = useState<RefItem[]>([]);
  const [suppliers, setSuppliers] = useState<RefItem[]>([]);
  const [units, setUnits] = useState<RefItem[]>([]);
  const [fileList, setFileList] = useState<UploadFile[]>([]);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    Promise.all([
      apiFetch<RefItem[]>("/api/categories"),
      apiFetch<RefItem[]>("/api/manufacturers"),
      apiFetch<RefItem[]>("/api/suppliers"),
      apiFetch<RefItem[]>("/api/units"),
    ])
      .then(([cats, mans, sups, uns]) => {
        setCategories(cats);
        setManufacturers(mans);
        setSuppliers(sups);
        setUnits(uns);
      })
      .catch(() => {});
  }, []);

  useEffect(() => {
    if (product) {
      form.setFieldsValue({
        name: product.name,
        article: product.article,
        categoryId: product.categoryId,
        manufacturerId: product.manufacturerId,
        supplierId: product.supplierId,
        description: product.description,
        price: product.price,
        discount: product.discount,
        quantity: product.quantity,
        unitId: product.unitId,
      });
    }
  }, [product, form]);

  const handleFinish = async (values: FormValues) => {
    setSubmitting(true);
    try {
      const body = {
        name: values.name,
        categoryId: values.categoryId,
        manufacturerId: values.manufacturerId,
        supplierId: values.supplierId,
        description: values.description || "",
        price: values.price,
        discount: values.discount ?? 0,
        quantity: values.quantity,
        unitId: values.unitId,
      };

      let productId: number;
      if (product) {
        await apiFetch(`/api/products/${product.id}`, {
          method: "PUT",
          body: JSON.stringify(body),
        });
        productId = product.id;
      } else {
        const created = await apiFetch<Product>("/api/products", {
          method: "POST",
          body: JSON.stringify(body),
        });
        productId = created.id;
      }

      if (fileList.length > 0 && fileList[0].originFileObj) {
        const formData = new FormData();
        formData.append("image", fileList[0].originFileObj);
        await apiUpload(`/api/products/${productId}/image`, formData);
      }

      onSuccess();
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Неизвестная ошибка";
      notification.error({ message: "Ошибка", description: message });
    } finally {
      setSubmitting(false);
    }
  };

  const currentImageUrl =
    product?.image && product.image.trim() !== ""
      ? `${API_BASE}/uploads/${product.image}`
      : null;

  return (
    <Form
      form={form}
      layout="vertical"
      onFinish={handleFinish}
      initialValues={{ discount: 0 }}
    >
      <Row gutter={24}>
        {/* Left: image */}
        <Col xs={24} md={8}>
          {currentImageUrl && (
            <div style={{ marginBottom: 16 }}>
              <img
                src={currentImageUrl}
                alt="Фото товара"
                style={{
                  width: "100%",
                  maxHeight: 240,
                  objectFit: "contain",
                  borderRadius: 8,
                  border: "1px solid #e8e8e8",
                }}
              />
            </div>
          )}
          <Form.Item label="Фото товара">
            <Upload
              fileList={fileList}
              beforeUpload={() => false}
              onChange={({ fileList: newList }) => setFileList(newList)}
              accept="image/*"
              maxCount={1}
              listType="picture"
            >
              <Button icon={<UploadOutlined />}>Выбрать файл</Button>
            </Upload>
          </Form.Item>
        </Col>

        {/* Right: fields */}
        <Col xs={24} md={16}>
          {product && (
            <Form.Item name="article" label="Артикул">
              <Input disabled />
            </Form.Item>
          )}

          <Form.Item
            name="name"
            label="Наименование"
            rules={[{ required: true, message: "Введите наименование" }]}
          >
            <Input />
          </Form.Item>

          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <Form.Item
                name="categoryId"
                label="Категория"
                rules={[{ required: true, message: "Выберите категорию" }]}
              >
                <Select
                  options={categories.map((c) => ({ value: c.id, label: c.name }))}
                  placeholder="Выберите категорию"
                />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item
                name="manufacturerId"
                label="Производитель"
                rules={[{ required: true, message: "Выберите производителя" }]}
              >
                <Select
                  options={manufacturers.map((m) => ({ value: m.id, label: m.name }))}
                  placeholder="Выберите производителя"
                />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <Form.Item
                name="supplierId"
                label="Поставщик"
                rules={[{ required: true, message: "Выберите поставщика" }]}
              >
                <Select
                  options={suppliers.map((s) => ({ value: s.id, label: s.name }))}
                  placeholder="Выберите поставщика"
                />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item
                name="unitId"
                label="Единица измерения"
                rules={[{ required: true, message: "Выберите единицу измерения" }]}
              >
                <Select
                  options={units.map((u) => ({ value: u.id, label: u.name }))}
                  placeholder="Выберите единицу"
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="description" label="Описание">
            <TextArea rows={3} />
          </Form.Item>

          <Row gutter={16}>
            <Col xs={24} sm={8}>
              <Form.Item
                name="price"
                label="Цена (₽)"
                rules={[{ required: true, message: "Введите цену" }]}
              >
                <InputNumber
                  min={0}
                  step={0.01}
                  precision={2}
                  style={{ width: "100%" }}
                />
              </Form.Item>
            </Col>
            <Col xs={24} sm={8}>
              <Form.Item name="discount" label="Скидка (%)">
                <InputNumber min={0} max={100} style={{ width: "100%" }} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={8}>
              <Form.Item
                name="quantity"
                label="Количество"
                rules={[{ required: true, message: "Введите количество" }]}
              >
                <InputNumber min={0} precision={0} style={{ width: "100%" }} />
              </Form.Item>
            </Col>
          </Row>
        </Col>
      </Row>

      <div style={{ display: "flex", gap: 12, marginTop: 8 }}>
        <Button
          type="primary"
          htmlType="submit"
          loading={submitting}
          style={{
            background: "#00FA9A",
            borderColor: "#00FA9A",
            color: "#000",
            fontWeight: 600,
          }}
        >
          Сохранить
        </Button>
        <Button onClick={() => router.back()}>Назад</Button>
      </div>
    </Form>
  );
}
