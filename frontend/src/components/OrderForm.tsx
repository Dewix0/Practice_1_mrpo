"use client";

import React, { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import {
  Form,
  Input,
  InputNumber,
  Select,
  Button,
  DatePicker,
  Row,
  Col,
  notification,
  Space,
} from "antd";
import { MinusCircleOutlined, PlusOutlined } from "@ant-design/icons";
import dayjs from "dayjs";
import { apiFetch } from "@/lib/api";
import { Order, RefItem, PickupPoint, Product } from "@/types";

interface OrderFormProps {
  order?: Order;
  onSuccess: () => void;
}

interface FormValues {
  statusId: number;
  pickupPointId: number;
  orderDate: ReturnType<typeof dayjs>;
  deliveryDate?: ReturnType<typeof dayjs>;
  pickupCode?: string;
  items: Array<{ productId: number; quantity: number }>;
}

export default function OrderForm({ order, onSuccess }: OrderFormProps) {
  const router = useRouter();
  const [form] = Form.useForm<FormValues>();
  const [statuses, setStatuses] = useState<RefItem[]>([]);
  const [pickupPoints, setPickupPoints] = useState<PickupPoint[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    Promise.all([
      apiFetch<RefItem[]>("/api/order-statuses"),
      apiFetch<PickupPoint[]>("/api/pickup-points"),
      apiFetch<Product[]>("/api/products"),
    ])
      .then(([sts, pts, prods]) => {
        setStatuses(sts);
        setPickupPoints(pts);
        setProducts(prods);
      })
      .catch(() => {});
  }, []);

  useEffect(() => {
    if (order) {
      form.setFieldsValue({
        statusId: order.statusId,
        pickupPointId: order.pickupPointId,
        orderDate: dayjs(order.orderDate),
        deliveryDate: order.deliveryDate ? dayjs(order.deliveryDate) : undefined,
        pickupCode: order.pickupCode,
        items: order.items.map((item) => ({
          productId: item.productId,
          quantity: item.quantity,
        })),
      });
    }
  }, [order, form]);

  const handleFinish = async (values: FormValues) => {
    setSubmitting(true);
    try {
      const body = {
        statusId: values.statusId,
        pickupPointId: values.pickupPointId,
        orderDate: values.orderDate.format("YYYY-MM-DD"),
        deliveryDate: values.deliveryDate
          ? values.deliveryDate.format("YYYY-MM-DD")
          : null,
        pickupCode: values.pickupCode || "",
        items: (values.items || []).map((item) => ({
          productId: item.productId,
          quantity: item.quantity,
        })),
      };

      if (order) {
        await apiFetch(`/api/orders/${order.id}`, {
          method: "PUT",
          body: JSON.stringify(body),
        });
      } else {
        await apiFetch("/api/orders", {
          method: "POST",
          body: JSON.stringify(body),
        });
      }

      onSuccess();
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Неизвестная ошибка";
      notification.error({ message: "Ошибка", description: message });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Form
      form={form}
      layout="vertical"
      onFinish={handleFinish}
    >
      <Row gutter={16}>
        <Col xs={24} sm={12}>
          <Form.Item
            name="statusId"
            label="Статус"
            rules={[{ required: true, message: "Выберите статус" }]}
          >
            <Select
              options={statuses.map((s) => ({ value: s.id, label: s.name }))}
              placeholder="Выберите статус"
            />
          </Form.Item>
        </Col>
        <Col xs={24} sm={12}>
          <Form.Item
            name="pickupPointId"
            label="Пункт выдачи"
            rules={[{ required: true, message: "Выберите пункт выдачи" }]}
          >
            <Select
              options={pickupPoints.map((p) => ({ value: p.id, label: p.address }))}
              placeholder="Выберите пункт выдачи"
            />
          </Form.Item>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col xs={24} sm={8}>
          <Form.Item
            name="orderDate"
            label="Дата заказа"
            rules={[{ required: true, message: "Выберите дату заказа" }]}
          >
            <DatePicker style={{ width: "100%" }} format="YYYY-MM-DD" />
          </Form.Item>
        </Col>
        <Col xs={24} sm={8}>
          <Form.Item name="deliveryDate" label="Дата доставки">
            <DatePicker style={{ width: "100%" }} format="YYYY-MM-DD" />
          </Form.Item>
        </Col>
        <Col xs={24} sm={8}>
          <Form.Item name="pickupCode" label="Код для получения">
            <Input />
          </Form.Item>
        </Col>
      </Row>

      <Form.List name="items">
        {(fields, { add, remove }) => (
          <div>
            <div style={{ fontWeight: 600, marginBottom: 8 }}>Товары в заказе</div>
            {fields.map(({ key, name, ...restField }) => (
              <Space key={key} align="baseline" style={{ display: "flex", marginBottom: 8 }}>
                <Form.Item
                  {...restField}
                  name={[name, "productId"]}
                  rules={[{ required: true, message: "Выберите товар" }]}
                  style={{ marginBottom: 0 }}
                >
                  <Select
                    style={{ width: 320 }}
                    placeholder="Выберите товар"
                    options={products.map((p) => ({
                      value: p.id,
                      label: `${p.article} — ${p.name}`,
                    }))}
                    showSearch
                    filterOption={(input, option) =>
                      String(option?.label ?? "")
                        .toLowerCase()
                        .includes(input.toLowerCase())
                    }
                  />
                </Form.Item>
                <Form.Item
                  {...restField}
                  name={[name, "quantity"]}
                  rules={[{ required: true, message: "Укажите кол-во" }]}
                  style={{ marginBottom: 0 }}
                >
                  <InputNumber min={1} placeholder="Кол-во" style={{ width: 100 }} />
                </Form.Item>
                <MinusCircleOutlined onClick={() => remove(name)} style={{ color: "red" }} />
              </Space>
            ))}
            <Button
              type="dashed"
              onClick={() => add()}
              icon={<PlusOutlined />}
              style={{ marginTop: 4 }}
            >
              Добавить товар
            </Button>
          </div>
        )}
      </Form.List>

      <div style={{ display: "flex", gap: 12, marginTop: 24 }}>
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
