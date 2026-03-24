package repository

import (
	"database/sql"
	"fmt"

	"shoe-store/internal/model"
)

type OrderRepo struct {
	DB *sql.DB
}

func NewOrderRepo(db *sql.DB) *OrderRepo {
	return &OrderRepo{DB: db}
}

const orderBaseQuery = `
SELECT o.id, o.order_date, o.delivery_date, COALESCE(o.pickup_code,''),
       o.status_id, os.name, o.pickup_point_id, pp.address, o.user_id
FROM orders o
JOIN order_statuses os ON o.status_id = os.id
JOIN pickup_points pp ON o.pickup_point_id = pp.id`

func scanOrder(row interface {
	Scan(...interface{}) error
}) (*model.Order, error) {
	var o model.Order
	var orderDate sql.NullString
	var deliveryDate sql.NullString
	var userID sql.NullInt64
	err := row.Scan(
		&o.ID, &orderDate, &deliveryDate, &o.PickupCode,
		&o.StatusID, &o.StatusName,
		&o.PickupPointID, &o.PickupAddress,
		&userID,
	)
	if err != nil {
		return nil, err
	}
	if orderDate.Valid {
		o.OrderDate = orderDate.String
	}
	if deliveryDate.Valid {
		o.DeliveryDate = &deliveryDate.String
	}
	if userID.Valid {
		o.UserID = &userID.Int64
	}
	return &o, nil
}

func (r *OrderRepo) loadItems(orderID int64) ([]model.OrderItem, error) {
	rows, err := r.DB.Query(`
		SELECT oi.id, oi.order_id, oi.product_id, p.article, oi.quantity
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = ?`, orderID)
	if err != nil {
		return nil, fmt.Errorf("order items query: %w", err)
	}
	defer rows.Close()

	var items []model.OrderItem
	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductArticle, &item.Quantity); err != nil {
			return nil, fmt.Errorf("order item scan: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if items == nil {
		items = []model.OrderItem{}
	}
	return items, nil
}

func (r *OrderRepo) List() ([]model.Order, error) {
	rows, err := r.DB.Query(orderBaseQuery + " ORDER BY o.id DESC")
	if err != nil {
		return nil, fmt.Errorf("order list query: %w", err)
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, fmt.Errorf("order scan: %w", err)
		}
		items, err := r.loadItems(o.ID)
		if err != nil {
			return nil, err
		}
		o.Items = items
		orders = append(orders, *o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if orders == nil {
		orders = []model.Order{}
	}
	return orders, nil
}

func (r *OrderRepo) GetByID(id int64) (*model.Order, error) {
	row := r.DB.QueryRow(orderBaseQuery+" WHERE o.id = ?", id)
	o, err := scanOrder(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("order get by id: %w", err)
	}
	items, err := r.loadItems(o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items
	return o, nil
}

func (r *OrderRepo) Create(input model.OrderInput) (int64, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return 0, fmt.Errorf("order create begin tx: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
		INSERT INTO orders (order_date, delivery_date, pickup_code, status_id, pickup_point_id, user_id)
		VALUES (?, ?, ?, ?, ?, ?)`,
		input.OrderDate, input.DeliveryDate, input.PickupCode,
		input.StatusID, input.PickupPointID, input.UserID,
	)
	if err != nil {
		return 0, fmt.Errorf("order insert: %w", err)
	}
	orderID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("order last insert id: %w", err)
	}

	for _, item := range input.Items {
		_, err := tx.Exec(`INSERT INTO order_items (order_id, product_id, quantity) VALUES (?, ?, ?)`,
			orderID, item.ProductID, item.Quantity)
		if err != nil {
			return 0, fmt.Errorf("order item insert: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("order create commit: %w", err)
	}
	return orderID, nil
}

func (r *OrderRepo) Update(id int64, input model.OrderInput) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("order update begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE orders SET order_date=?, delivery_date=?, pickup_code=?, status_id=?, pickup_point_id=?, user_id=?
		WHERE id=?`,
		input.OrderDate, input.DeliveryDate, input.PickupCode,
		input.StatusID, input.PickupPointID, input.UserID,
		id,
	)
	if err != nil {
		return fmt.Errorf("order update: %w", err)
	}

	_, err = tx.Exec(`DELETE FROM order_items WHERE order_id=?`, id)
	if err != nil {
		return fmt.Errorf("order items delete: %w", err)
	}

	for _, item := range input.Items {
		_, err := tx.Exec(`INSERT INTO order_items (order_id, product_id, quantity) VALUES (?, ?, ?)`,
			id, item.ProductID, item.Quantity)
		if err != nil {
			return fmt.Errorf("order item insert: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("order update commit: %w", err)
	}
	return nil
}

func (r *OrderRepo) Delete(id int64) error {
	_, err := r.DB.Exec(`DELETE FROM orders WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("order delete: %w", err)
	}
	return nil
}
