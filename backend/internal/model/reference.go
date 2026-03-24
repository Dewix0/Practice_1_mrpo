package model

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Manufacturer struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Supplier struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Unit struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type OrderStatus struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type PickupPoint struct {
	ID      int64  `json:"id"`
	Address string `json:"address"`
}
