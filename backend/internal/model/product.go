package model

type Product struct {
	ID               int64   `json:"id"`
	Article          string  `json:"article"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	Price            float64 `json:"price"`
	Discount         float64 `json:"discount"`
	Quantity         int     `json:"quantity"`
	Image            string  `json:"image"`
	CategoryID       int64   `json:"categoryId"`
	CategoryName     string  `json:"categoryName"`
	ManufacturerID   int64   `json:"manufacturerId"`
	ManufacturerName string  `json:"manufacturerName"`
	SupplierID       int64   `json:"supplierId"`
	SupplierName     string  `json:"supplierName"`
	UnitID           int64   `json:"unitId"`
	UnitName         string  `json:"unitName"`
}

type ProductFilter struct {
	Search     string
	Sort       string
	SupplierID int64
}

type ProductInput struct {
	Article        string  `json:"article"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	Price          float64 `json:"price"`
	Discount       float64 `json:"discount"`
	Quantity       int     `json:"quantity"`
	CategoryID     int64   `json:"categoryId"`
	ManufacturerID int64   `json:"manufacturerId"`
	SupplierID     int64   `json:"supplierId"`
	UnitID         int64   `json:"unitId"`
}
