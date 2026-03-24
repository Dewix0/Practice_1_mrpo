package repository

import (
	"database/sql"
	"fmt"

	"shoe-store/internal/model"
)

type ReferenceRepo struct {
	DB *sql.DB
}

func NewReferenceRepo(db *sql.DB) *ReferenceRepo {
	return &ReferenceRepo{DB: db}
}

func (r *ReferenceRepo) ListCategories() ([]model.Category, error) {
	rows, err := r.DB.Query(`SELECT id, name FROM categories ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()
	var result []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		result = append(result, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = []model.Category{}
	}
	return result, nil
}

func (r *ReferenceRepo) ListManufacturers() ([]model.Manufacturer, error) {
	rows, err := r.DB.Query(`SELECT id, name FROM manufacturers ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list manufacturers: %w", err)
	}
	defer rows.Close()
	var result []model.Manufacturer
	for rows.Next() {
		var m model.Manufacturer
		if err := rows.Scan(&m.ID, &m.Name); err != nil {
			return nil, fmt.Errorf("scan manufacturer: %w", err)
		}
		result = append(result, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = []model.Manufacturer{}
	}
	return result, nil
}

func (r *ReferenceRepo) ListSuppliers() ([]model.Supplier, error) {
	rows, err := r.DB.Query(`SELECT id, name FROM suppliers ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list suppliers: %w", err)
	}
	defer rows.Close()
	var result []model.Supplier
	for rows.Next() {
		var s model.Supplier
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, fmt.Errorf("scan supplier: %w", err)
		}
		result = append(result, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = []model.Supplier{}
	}
	return result, nil
}

func (r *ReferenceRepo) ListUnits() ([]model.Unit, error) {
	rows, err := r.DB.Query(`SELECT id, name FROM units ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list units: %w", err)
	}
	defer rows.Close()
	var result []model.Unit
	for rows.Next() {
		var u model.Unit
		if err := rows.Scan(&u.ID, &u.Name); err != nil {
			return nil, fmt.Errorf("scan unit: %w", err)
		}
		result = append(result, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = []model.Unit{}
	}
	return result, nil
}

func (r *ReferenceRepo) ListOrderStatuses() ([]model.OrderStatus, error) {
	rows, err := r.DB.Query(`SELECT id, name FROM order_statuses ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list order statuses: %w", err)
	}
	defer rows.Close()
	var result []model.OrderStatus
	for rows.Next() {
		var os model.OrderStatus
		if err := rows.Scan(&os.ID, &os.Name); err != nil {
			return nil, fmt.Errorf("scan order status: %w", err)
		}
		result = append(result, os)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = []model.OrderStatus{}
	}
	return result, nil
}

func (r *ReferenceRepo) ListPickupPoints() ([]model.PickupPoint, error) {
	rows, err := r.DB.Query(`SELECT id, address FROM pickup_points ORDER BY address`)
	if err != nil {
		return nil, fmt.Errorf("list pickup points: %w", err)
	}
	defer rows.Close()
	var result []model.PickupPoint
	for rows.Next() {
		var pp model.PickupPoint
		if err := rows.Scan(&pp.ID, &pp.Address); err != nil {
			return nil, fmt.Errorf("scan pickup point: %w", err)
		}
		result = append(result, pp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = []model.PickupPoint{}
	}
	return result, nil
}
