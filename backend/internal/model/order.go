package model

type Order struct {
	ID            int64       `json:"id"`
	OrderDate     string      `json:"orderDate"`
	DeliveryDate  *string     `json:"deliveryDate"`
	PickupCode    string      `json:"pickupCode"`
	StatusID      int64       `json:"statusId"`
	StatusName    string      `json:"statusName"`
	PickupPointID int64       `json:"pickupPointId"`
	PickupAddress string      `json:"pickupAddress"`
	UserID        *int64      `json:"userId"`
	Items         []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID             int64  `json:"id"`
	OrderID        int64  `json:"orderId"`
	ProductID      int64  `json:"productId"`
	ProductArticle string `json:"productArticle"`
	Quantity       int    `json:"quantity"`
}

type OrderInput struct {
	OrderDate     string           `json:"orderDate"`
	DeliveryDate  *string          `json:"deliveryDate"`
	PickupCode    string           `json:"pickupCode"`
	StatusID      int64            `json:"statusId"`
	PickupPointID int64            `json:"pickupPointId"`
	UserID        *int64           `json:"userId"`
	Items         []OrderItemInput `json:"items"`
}

type OrderItemInput struct {
	ProductID int64 `json:"productId"`
	Quantity  int   `json:"quantity"`
}
