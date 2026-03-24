package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"shoe-store/internal/model"
)

type ProductRepo struct {
	DB *sql.DB
}

func NewProductRepo(db *sql.DB) *ProductRepo {
	return &ProductRepo{DB: db}
}

const productBaseQuery = `
SELECT p.id, p.article, p.name, p.description, p.price, p.discount, p.quantity, COALESCE(p.image,''),
       p.category_id, c.name, p.manufacturer_id, m.name, p.supplier_id, s.name, p.unit_id, u.name
FROM products p
JOIN categories c ON p.category_id = c.id
JOIN manufacturers m ON p.manufacturer_id = m.id
JOIN suppliers s ON p.supplier_id = s.id
JOIN units u ON p.unit_id = u.id`

func scanProduct(row interface {
	Scan(...interface{}) error
}) (*model.Product, error) {
	var p model.Product
	err := row.Scan(
		&p.ID, &p.Article, &p.Name, &p.Description,
		&p.Price, &p.Discount, &p.Quantity, &p.Image,
		&p.CategoryID, &p.CategoryName,
		&p.ManufacturerID, &p.ManufacturerName,
		&p.SupplierID, &p.SupplierName,
		&p.UnitID, &p.UnitName,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) List(filter model.ProductFilter) ([]model.Product, error) {
	query := productBaseQuery
	var args []interface{}
	var conditions []string

	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		conditions = append(conditions, `(p.name LIKE ? OR p.description LIKE ? OR p.article LIKE ? OR c.name LIKE ? OR m.name LIKE ? OR s.name LIKE ?)`)
		args = append(args, like, like, like, like, like, like)
	}

	if filter.SupplierID > 0 {
		conditions = append(conditions, `p.supplier_id = ?`)
		args = append(args, filter.SupplierID)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	switch filter.Sort {
	case "quantity_asc":
		query += " ORDER BY p.quantity ASC"
	case "quantity_desc":
		query += " ORDER BY p.quantity DESC"
	default:
		query += " ORDER BY p.id"
	}

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("product list query: %w", err)
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, fmt.Errorf("product scan: %w", err)
		}
		products = append(products, *p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if products == nil {
		products = []model.Product{}
	}
	return products, nil
}

func (r *ProductRepo) GetByID(id int64) (*model.Product, error) {
	query := productBaseQuery + " WHERE p.id = ?"
	row := r.DB.QueryRow(query, id)
	p, err := scanProduct(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("product get by id: %w", err)
	}
	return p, nil
}

func (r *ProductRepo) Create(input model.ProductInput) (int64, error) {
	query := `INSERT INTO products (article, name, description, price, discount, quantity, category_id, manufacturer_id, supplier_id, unit_id)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := r.DB.Exec(query,
		input.Article, input.Name, input.Description,
		input.Price, input.Discount, input.Quantity,
		input.CategoryID, input.ManufacturerID, input.SupplierID, input.UnitID,
	)
	if err != nil {
		return 0, fmt.Errorf("product create: %w", err)
	}
	return res.LastInsertId()
}

func (r *ProductRepo) Update(id int64, input model.ProductInput) error {
	query := `UPDATE products SET article=?, name=?, description=?, price=?, discount=?, quantity=?,
	          category_id=?, manufacturer_id=?, supplier_id=?, unit_id=? WHERE id=?`
	_, err := r.DB.Exec(query,
		input.Article, input.Name, input.Description,
		input.Price, input.Discount, input.Quantity,
		input.CategoryID, input.ManufacturerID, input.SupplierID, input.UnitID,
		id,
	)
	if err != nil {
		return fmt.Errorf("product update: %w", err)
	}
	return nil
}

func (r *ProductRepo) Delete(id int64) error {
	_, err := r.DB.Exec(`DELETE FROM products WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("product delete: %w", err)
	}
	return nil
}

func (r *ProductRepo) UpdateImage(id int64, path string) error {
	_, err := r.DB.Exec(`UPDATE products SET image=? WHERE id=?`, path, id)
	if err != nil {
		return fmt.Errorf("product update image: %w", err)
	}
	return nil
}

func (r *ProductRepo) GetImage(id int64) (string, error) {
	var img string
	err := r.DB.QueryRow(`SELECT COALESCE(image,'') FROM products WHERE id=?`, id).Scan(&img)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("product get image: %w", err)
	}
	return img, nil
}
