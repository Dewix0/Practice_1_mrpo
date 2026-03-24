export interface User {
  id: number;
  login: string;
  lastName: string;
  firstName: string;
  patronymic: string;
  roleId: number;
  role: string;
}

export interface Product {
  id: number;
  article: string;
  name: string;
  description: string;
  price: number;
  discount: number;
  quantity: number;
  image: string;
  categoryId: number;
  categoryName: string;
  manufacturerId: number;
  manufacturerName: string;
  supplierId: number;
  supplierName: string;
  unitId: number;
  unitName: string;
}

export interface Order {
  id: number;
  orderDate: string;
  deliveryDate: string | null;
  pickupCode: string;
  statusId: number;
  statusName: string;
  pickupPointId: number;
  pickupAddress: string;
  userId: number | null;
  items: OrderItem[];
}

export interface OrderItem {
  id: number;
  orderId: number;
  productId: number;
  productArticle: string;
  quantity: number;
}

export interface RefItem {
  id: number;
  name: string;
}

export interface PickupPoint {
  id: number;
  address: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}
