# ООО "Обувь" Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a shoe store information system with Go REST API backend, Next.js + Ant Design frontend, and SQLite database.

**Architecture:** Monorepo with `backend/` (Go chi REST API, layered: handler→service→repository) and `frontend/` (Next.js App Router + Ant Design). SQLite via modernc.org/sqlite (pure Go). JWT auth with bcrypt passwords. Static file serving for product images.

**Tech Stack:** Go 1.22+, chi, modernc.org/sqlite, golang-jwt/jwt/v5, rs/cors, bcrypt | Next.js 14+, Ant Design 5, TypeScript | SQLite

**Spec:** `docs/superpowers/specs/2026-03-24-shoe-store-design.md`

---

## File Map

### Backend

| File | Responsibility |
|------|----------------|
| `backend/cmd/server/main.go` | Entry point: wire up router, DB, start server |
| `backend/cmd/seed/main.go` | Import Excel data into DB |
| `backend/internal/database/db.go` | Open SQLite connection, run migrations |
| `backend/internal/database/migrations.go` | SQL CREATE TABLE statements |
| `backend/internal/model/user.go` | User, Role structs |
| `backend/internal/model/product.go` | Product struct |
| `backend/internal/model/order.go` | Order, OrderItem structs |
| `backend/internal/model/reference.go` | Category, Manufacturer, Supplier, Unit, OrderStatus, PickupPoint structs |
| `backend/internal/repository/user_repo.go` | SQL queries for users |
| `backend/internal/repository/product_repo.go` | SQL queries for products (CRUD, search, filter, sort) |
| `backend/internal/repository/order_repo.go` | SQL queries for orders + order_items |
| `backend/internal/repository/reference_repo.go` | SQL queries for all reference tables |
| `backend/internal/service/auth_service.go` | Login, JWT generation, password verification |
| `backend/internal/service/product_service.go` | Product business logic, image handling |
| `backend/internal/service/order_service.go` | Order business logic, deletion checks |
| `backend/internal/handler/auth_handler.go` | POST /api/auth/login, GET /api/auth/me, POST /api/auth/logout |
| `backend/internal/handler/product_handler.go` | Products CRUD + image upload endpoints |
| `backend/internal/handler/order_handler.go` | Orders CRUD endpoints |
| `backend/internal/handler/reference_handler.go` | GET endpoints for all reference data |
| `backend/internal/middleware/auth.go` | JWT extraction, RequireAuth, RequireRole middleware |
| `backend/internal/middleware/cors.go` | CORS configuration |
| `backend/go.mod` | Go module definition |

### Frontend

| File | Responsibility |
|------|----------------|
| `frontend/src/app/layout.tsx` | Root layout: Ant Design provider, global styles, auth context |
| `frontend/src/app/page.tsx` | Redirect to /products |
| `frontend/src/app/login/page.tsx` | Login form page |
| `frontend/src/app/products/page.tsx` | Product catalog with conditional styling |
| `frontend/src/app/products/new/page.tsx` | Add product form |
| `frontend/src/app/products/[id]/edit/page.tsx` | Edit product form |
| `frontend/src/app/orders/page.tsx` | Orders list table |
| `frontend/src/app/orders/new/page.tsx` | Add order form |
| `frontend/src/app/orders/[id]/edit/page.tsx` | Edit order form |
| `frontend/src/components/AppHeader.tsx` | Header: logo, nav, ФИО, logout button |
| `frontend/src/components/ProductCard.tsx` | Single product card with conditional styling |
| `frontend/src/components/ProductToolbar.tsx` | Search, supplier filter, sort buttons |
| `frontend/src/components/ProductForm.tsx` | Shared form for add/edit product |
| `frontend/src/components/OrderForm.tsx` | Shared form for add/edit order |
| `frontend/src/lib/api.ts` | Fetch wrapper, base URL, auth headers |
| `frontend/src/lib/auth.tsx` | AuthContext provider, useAuth hook, JWT cookie handling |
| `frontend/src/types/index.ts` | TypeScript interfaces for all entities |
| `frontend/next.config.js` | Next.js config (API proxy or CORS) |
| `frontend/package.json` | Dependencies |

---

## Task 1: Backend scaffold

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go`

- [ ] **Step 1: Initialize Go module**

```bash
cd backend && go mod init shoe-store && cd ..
```

- [ ] **Step 2: Install dependencies**

```bash
cd backend && go get github.com/go-chi/chi/v5 modernc.org/sqlite github.com/golang-jwt/jwt/v5 github.com/rs/cors golang.org/x/crypto/bcrypt && cd ..
```

- [ ] **Step 3: Create main.go with health check**

Create `backend/cmd/server/main.go`:

```go
package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
```

- [ ] **Step 4: Verify it compiles and runs**

```bash
cd backend && go build ./cmd/server && cd ..
```

Expected: no errors, binary created.

- [ ] **Step 5: Commit**

```bash
git add backend/
git commit -m "feat: backend scaffold with chi router and health check"
```

---

## Task 2: Database schema and connection

**Files:**
- Create: `backend/internal/database/db.go`
- Create: `backend/internal/database/migrations.go`

- [ ] **Step 1: Create database connection helper**

Create `backend/internal/database/db.go`:

```go
package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	// Enable WAL mode and foreign keys
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("set WAL: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return nil, fmt.Errorf("enable FK: %w", err)
	}
	return db, nil
}

func Migrate(db *sql.DB) error {
	for i, stmt := range migrations {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migration %d: %w", i, err)
		}
	}
	log.Println("Database migrations applied successfully")
	return nil
}
```

- [ ] **Step 2: Create migrations**

Create `backend/internal/database/migrations.go` with all 10 CREATE TABLE statements per the spec. Key points:
- `roles`, `categories`, `manufacturers`, `suppliers`, `units`, `order_statuses`, `pickup_points` — simple id + name/address
- `users` — with `role_id FK REFERENCES roles(id)`
- `products` — with `article TEXT NOT NULL UNIQUE`, 4 FK references
- `orders` — with `user_id`, `pickup_code`, `status_id`, `pickup_point_id` FKs
- `order_items` — with `ON DELETE CASCADE` for order_id, `ON DELETE RESTRICT` for product_id

```go
package database

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS roles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	)`,
	`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		login TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		last_name TEXT,
		first_name TEXT,
		patronymic TEXT,
		role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE RESTRICT
	)`,
	`CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	)`,
	`CREATE TABLE IF NOT EXISTS manufacturers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	)`,
	`CREATE TABLE IF NOT EXISTS suppliers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	)`,
	`CREATE TABLE IF NOT EXISTS units (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	)`,
	`CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		article TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		description TEXT,
		price REAL NOT NULL CHECK(price >= 0),
		discount REAL NOT NULL DEFAULT 0 CHECK(discount >= 0),
		quantity INTEGER NOT NULL DEFAULT 0 CHECK(quantity >= 0),
		image TEXT,
		category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
		manufacturer_id INTEGER NOT NULL REFERENCES manufacturers(id) ON DELETE RESTRICT,
		supplier_id INTEGER NOT NULL REFERENCES suppliers(id) ON DELETE RESTRICT,
		unit_id INTEGER NOT NULL REFERENCES units(id) ON DELETE RESTRICT
	)`,
	`CREATE TABLE IF NOT EXISTS order_statuses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	)`,
	`CREATE TABLE IF NOT EXISTS pickup_points (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		address TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_date TEXT,
		delivery_date TEXT,
		pickup_code TEXT,
		status_id INTEGER NOT NULL REFERENCES order_statuses(id) ON DELETE RESTRICT,
		pickup_point_id INTEGER NOT NULL REFERENCES pickup_points(id) ON DELETE RESTRICT,
		user_id INTEGER REFERENCES users(id) ON DELETE RESTRICT
	)`,
	`CREATE TABLE IF NOT EXISTS order_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
		product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
		quantity INTEGER NOT NULL CHECK(quantity > 0)
	)`,
}
```

- [ ] **Step 3: Wire DB into main.go**

Update `backend/cmd/server/main.go` to open DB and run migrations on startup:

```go
db, err := database.Open("data.db")
if err != nil {
    log.Fatal(err)
}
defer db.Close()
if err := database.Migrate(db); err != nil {
    log.Fatal(err)
}
```

- [ ] **Step 4: Verify it compiles**

```bash
cd backend && go build ./cmd/server && cd ..
```

- [ ] **Step 5: Commit**

```bash
git add backend/
git commit -m "feat: database schema with all 10 tables and migrations"
```

---

## Task 3: Models

**Files:**
- Create: `backend/internal/model/user.go`
- Create: `backend/internal/model/product.go`
- Create: `backend/internal/model/order.go`
- Create: `backend/internal/model/reference.go`

- [ ] **Step 1: Create all model structs**

`backend/internal/model/reference.go` — shared reference structs:
```go
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
```

`backend/internal/model/user.go`:
```go
package model

type Role struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
type User struct {
	ID         int64  `json:"id"`
	Login      string `json:"login"`
	Password   string `json:"-"`
	LastName   string `json:"lastName"`
	FirstName  string `json:"firstName"`
	Patronymic string `json:"patronymic"`
	RoleID     int64  `json:"roleId"`
	RoleName   string `json:"role"`
}
// FullName returns "Фамилия И.О." for display in header
func (u User) FullName() string {
	// e.g. "Никифорова В.Н."
}
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
```

`backend/internal/model/product.go`:
```go
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
	Search     string `json:"search"`
	Sort       string `json:"sort"`
	SupplierID int64  `json:"supplierId"`
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
```

`backend/internal/model/order.go`:
```go
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
```

- [ ] **Step 2: Verify it compiles**

```bash
cd backend && go build ./... && cd ..
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/model/
git commit -m "feat: add all data model structs"
```

---

## Task 4: Auth middleware and login endpoint

**Files:**
- Create: `backend/internal/middleware/auth.go`
- Create: `backend/internal/middleware/cors.go`
- Create: `backend/internal/repository/user_repo.go`
- Create: `backend/internal/service/auth_service.go`
- Create: `backend/internal/handler/auth_handler.go`

- [ ] **Step 1: Create user repository**

`backend/internal/repository/user_repo.go` — `FindByLogin(login string) (*model.User, error)`:
- Query joins users with roles to get role name
- Returns nil,nil if not found

- [ ] **Step 2: Create auth service**

`backend/internal/service/auth_service.go`:
- `Login(login, password string) (*model.LoginResponse, error)`
- Find user by login, bcrypt.CompareHashAndPassword, generate JWT with claims {userId, role, exp}
- JWT secret from env var `JWT_SECRET` (default "dev-secret" for dev)
- Token expiry: 24 hours

- [ ] **Step 3: Create auth middleware**

`backend/internal/middleware/auth.go`:
- `JWTAuth(secret string) func(http.Handler) http.Handler` — extracts JWT from httpOnly cookie `token`, parses claims, sets userId and role in context. Does not block — just enriches context.
- `RequireAuth() func(http.Handler) http.Handler` — rejects request with 401 if no valid JWT in context. Use for "auth"-level endpoints.
- `RequireRole(roles ...string) func(http.Handler) http.Handler` — checks role from context against allowed roles, rejects with 403.
- Helper functions: `GetUserID(ctx)`, `GetRole(ctx)`, `IsAuthenticated(ctx)`

- [ ] **Step 4: Create CORS middleware**

`backend/internal/middleware/cors.go`:
- Configure rs/cors: AllowedOrigins `["http://localhost:3000"]`, AllowCredentials true, AllowedMethods `["GET","POST","PUT","DELETE"]`, AllowedHeaders `["Content-Type"]`

- [ ] **Step 5: Create auth handler**

`backend/internal/handler/auth_handler.go`:
- `POST /api/auth/login` — decode JSON body, call service.Login, set httpOnly cookie with token, return LoginResponse JSON
- `GET /api/auth/me` — read JWT from httpOnly cookie, validate, return current User JSON. Returns 401 if no valid token. This is how the frontend restores session after page refresh (httpOnly cookies cannot be read by JS).
- `POST /api/auth/logout` — set `token` cookie to expired value (MaxAge=-1), return 200. Needed because httpOnly cookies cannot be cleared by frontend JS.

- [ ] **Step 6: Wire auth routes into main.go**

Update `backend/cmd/server/main.go`:
- Add CORS middleware
- Add JWTAuth middleware (optional extraction — doesn't block, just sets context)
- Mount `/api/auth/login`, `/api/auth/me`, `/api/auth/logout` routes
- Pass `db` to repository → service → handler chain

- [ ] **Step 7: Verify it compiles**

```bash
cd backend && go build ./cmd/server && cd ..
```

- [ ] **Step 8: Commit**

```bash
git add backend/
git commit -m "feat: auth system with JWT, bcrypt, login endpoint, role middleware"
```

---

## Task 5: Products repository and service

**Files:**
- Create: `backend/internal/repository/product_repo.go`
- Create: `backend/internal/service/product_service.go`

- [ ] **Step 1: Create product repository**

`backend/internal/repository/product_repo.go`:

Methods:
- `List(filter model.ProductFilter) ([]model.Product, error)` — SELECT with JOINs on categories, manufacturers, suppliers, units. Dynamic WHERE for search (LIKE across name, description, article, category, manufacturer, supplier), supplier_id filter. ORDER BY for quantity sort.
- `GetByID(id int64) (*model.Product, error)` — single product with all joins
- `Create(input model.ProductInput) (int64, error)` — INSERT, return new ID
- `Update(id int64, input model.ProductInput) error` — UPDATE
- `Delete(id int64) error` — DELETE (FK constraint will block if in order_items)
- `UpdateImage(id int64, path string) error` — UPDATE image column
- `GetImage(id int64) (string, error)` — SELECT image path for a product

Key SQL for List with search:
```sql
SELECT p.id, p.article, p.name, p.description, p.price, p.discount, p.quantity, p.image,
       p.category_id, c.name, p.manufacturer_id, m.name, p.supplier_id, s.name, p.unit_id, u.name
FROM products p
JOIN categories c ON p.category_id = c.id
JOIN manufacturers m ON p.manufacturer_id = m.id
JOIN suppliers s ON p.supplier_id = s.id
JOIN units u ON p.unit_id = u.id
WHERE 1=1
-- if search: AND (p.name LIKE ? OR p.description LIKE ? OR p.article LIKE ? OR c.name LIKE ? OR m.name LIKE ? OR s.name LIKE ?)
-- if supplier_id: AND p.supplier_id = ?
-- if sort=quantity_asc: ORDER BY p.quantity ASC
-- if sort=quantity_desc: ORDER BY p.quantity DESC
```

- [ ] **Step 2: Create product service**

`backend/internal/service/product_service.go`:

Methods:
- `List(filter model.ProductFilter) ([]model.Product, error)` — pass through to repo
- `GetByID(id int64) (*model.Product, error)` — pass through, return 404-style error if nil
- `Create(input model.ProductInput) (int64, error)` — validate (price ≥ 0, quantity ≥ 0, required fields), then repo.Create
- `Update(id int64, input model.ProductInput) error` — validate, then repo.Update
- `Delete(id int64) error` — call repo.Delete, catch FK violation → return specific "product is in an order" error
- `UploadImage(id int64, file multipart.File, header *multipart.FileHeader) error` — validate MIME (jpeg/png), resize to fit within 300×200 preserving aspect ratio using `golang.org/x/image/draw` (BiLinear), decode any format → save as JPEG (quality 85) to `uploads/`, delete old file if exists, update image path in DB

- [ ] **Step 3: Install image resize dependency**

```bash
cd backend && go get golang.org/x/image && cd ..
```

- [ ] **Step 4: Verify it compiles**

```bash
cd backend && go build ./... && cd ..
```

- [ ] **Step 5: Commit**

```bash
git add backend/
git commit -m "feat: products repository and service with search, filter, sort, image upload"
```

---

## Task 6: Products HTTP handlers

**Files:**
- Create: `backend/internal/handler/product_handler.go`
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: Create product handler**

`backend/internal/handler/product_handler.go`:

```go
type ProductHandler struct {
	service *service.ProductService
}
```

Methods:
- `List(w, r)` — parse query params (search, sort, supplier_id), call service.List, write JSON
- `GetByID(w, r)` — parse `{id}` from chi.URLParam, call service.GetByID, write JSON
- `Create(w, r)` — decode JSON body into ProductInput, call service.Create, write JSON `{id: ...}`
- `Update(w, r)` — parse id, decode body, call service.Update
- `Delete(w, r)` — parse id, call service.Delete, handle "in order" error → 409
- `UploadImage(w, r)` — parse id, `r.FormFile("image")`, call service.UploadImage

- [ ] **Step 2: Mount product routes in main.go**

```go
r.Route("/api/products", func(r chi.Router) {
    r.Get("/", productHandler.List)           // public
    r.Get("/{id}", productHandler.GetByID)    // public

    r.Group(func(r chi.Router) {
        r.Use(middleware.RequireRole("admin"))
        r.Post("/", productHandler.Create)
        r.Put("/{id}", productHandler.Update)
        r.Delete("/{id}", productHandler.Delete)
        r.Post("/{id}/image", productHandler.UploadImage)
    })
})

// Serve uploaded images
r.Handle("/uploads/*", http.StripPrefix("/uploads/",
    http.FileServer(http.Dir("uploads"))))
```

- [ ] **Step 3: Create `backend/uploads/` directory with .gitkeep**

```bash
mkdir -p backend/uploads && touch backend/uploads/.gitkeep
```

- [ ] **Step 4: Verify it compiles**

```bash
cd backend && go build ./cmd/server && cd ..
```

- [ ] **Step 5: Runtime verification (after seed is done in Task 9)**

Start the server and test with curl:
```bash
curl -s http://localhost:8080/api/products | head -c 200
# Expected: JSON array of products
```

- [ ] **Step 6: Commit**

```bash
git add backend/
git commit -m "feat: product CRUD HTTP handlers with image upload"
```

---

## Task 7: Orders repository, service, and handlers

**Files:**
- Create: `backend/internal/repository/order_repo.go`
- Create: `backend/internal/service/order_service.go`
- Create: `backend/internal/handler/order_handler.go`
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: Create order repository**

`backend/internal/repository/order_repo.go`:

Methods:
- `List() ([]model.Order, error)` — SELECT with JOINs on order_statuses, pickup_points. For each order, also load its items (separate query or subquery).
- `GetByID(id int64) (*model.Order, error)` — single order + items
- `Create(input model.OrderInput) (int64, error)` — INSERT order, then INSERT each order_item in a transaction
- `Update(id int64, input model.OrderInput) error` — in transaction: UPDATE order, DELETE old items, INSERT new items
- `Delete(id int64) error` — DELETE (CASCADE removes items)

For List, load order_items in a second query grouped by order_id, then merge in Go. Or use a LEFT JOIN and group in Go.

- [ ] **Step 2: Create order service**

`backend/internal/service/order_service.go`:
- Validate required fields (order_date, status_id, pickup_point_id)
- Validate items not empty
- Pass through to repository

- [ ] **Step 3: Create order handler**

`backend/internal/handler/order_handler.go`:
- Same pattern as products: List, GetByID, Create, Update, Delete
- JSON encode/decode

- [ ] **Step 4: Mount order routes in main.go**

```go
r.Route("/api/orders", func(r chi.Router) {
    r.Use(middleware.RequireRole("manager", "admin"))
    r.Get("/", orderHandler.List)
    r.Get("/{id}", orderHandler.GetByID)

    r.Group(func(r chi.Router) {
        r.Use(middleware.RequireRole("admin"))
        r.Post("/", orderHandler.Create)
        r.Put("/{id}", orderHandler.Update)
        r.Delete("/{id}", orderHandler.Delete)
    })
})
```

- [ ] **Step 5: Verify it compiles**

```bash
cd backend && go build ./cmd/server && cd ..
```

- [ ] **Step 6: Runtime verification (after seed is done in Task 9)**

```bash
# Login as admin first to get cookie, then test orders
curl -s -c cookies.txt -X POST http://localhost:8080/api/auth/login -H "Content-Type: application/json" -d '{"login":"94d5ous@gmail.com","password":"uzWC67"}'
curl -s -b cookies.txt http://localhost:8080/api/orders | head -c 200
# Expected: JSON array of orders
```

- [ ] **Step 7: Commit**

```bash
git add backend/
git commit -m "feat: orders CRUD with order_items support"
```

---

## Task 8: Reference data handlers

**Files:**
- Create: `backend/internal/repository/reference_repo.go`
- Create: `backend/internal/handler/reference_handler.go`
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: Create reference repository**

`backend/internal/repository/reference_repo.go`:

Generic pattern — one method per table, all just `SELECT id, name FROM table ORDER BY name`:
- `ListCategories() ([]model.Category, error)`
- `ListManufacturers() ([]model.Manufacturer, error)`
- `ListSuppliers() ([]model.Supplier, error)`
- `ListUnits() ([]model.Unit, error)`
- `ListOrderStatuses() ([]model.OrderStatus, error)`
- `ListPickupPoints() ([]model.PickupPoint, error)` — SELECT id, address

- [ ] **Step 2: Create reference handler**

`backend/internal/handler/reference_handler.go`:
- One handler method per endpoint, each calls the corresponding repo method and writes JSON

- [ ] **Step 3: Mount reference routes in main.go**

```go
// Auth required (any authenticated user)
r.Group(func(r chi.Router) {
    r.Use(middleware.RequireAuth())
    r.Get("/api/categories", refHandler.Categories)
    r.Get("/api/manufacturers", refHandler.Manufacturers)
    r.Get("/api/units", refHandler.Units)
    r.Get("/api/pickup-points", refHandler.PickupPoints)
    r.Get("/api/order-statuses", refHandler.OrderStatuses)
})
// Manager+ for suppliers (used in filter toolbar)
r.Group(func(r chi.Router) {
    r.Use(middleware.RequireRole("manager", "admin"))
    r.Get("/api/suppliers", refHandler.Suppliers)
})
```

- [ ] **Step 4: Verify it compiles**

```bash
cd backend && go build ./cmd/server && cd ..
```

- [ ] **Step 5: Runtime verification (after seed is done in Task 9)**

```bash
curl -s -b cookies.txt http://localhost:8080/api/categories
# Expected: JSON array of categories like [{"id":1,"name":"Женская обувь"},...]
```

- [ ] **Step 6: Commit**

```bash
git add backend/
git commit -m "feat: reference data endpoints for dropdowns and filters"
```

---

## Task 9: Data import (seed command)

**Files:**
- Create: `backend/cmd/seed/main.go`

- [ ] **Step 1: Install Excel library**

```bash
cd backend && go get github.com/xuri/excelize/v2 && cd ..
```

- [ ] **Step 2: Verify import files exist**

```bash
ls import/user_import.xlsx import/Tovar.xlsx "import/Заказ_import.xlsx" "import/Пункты выдачи_import.xlsx" import/picture.png import/1.jpg
```

Expected: all files listed without errors.

- [ ] **Step 3: Create seed command**

`backend/cmd/seed/main.go`:

This is a standalone CLI tool that reads Excel files from `../import/` (relative to project root) and populates the DB.

**Excel file details** (sheet name is `Sheet1` for all files):

`user_import.xlsx` columns (row 1 headers):
- A: "Роль сотрудника", B: "ФИО", C: "Логин", D: "Пароль"

`Tovar.xlsx` columns (row 1 headers):
- A: "Артикул", B: "Наименование товара", C: "Единица измерения", D: "Цена", E: "Поставщик", F: "Производитель", G: "Категория товара", H: "Действующая скидка", I: "Кол-во на складе", J: "Описание товара", K: "Фото"

`Заказ_import.xlsx` columns (row 1 headers):
- A: "Номер заказа", B: "Артикул заказа", C: "Дата заказа", D: "Дата доставки", E: "Адрес пункта выдачи", F: "ФИО авторизированного клиента", G: "Код для получения", H: "Статус заказа"

`Пункты выдачи_import.xlsx`: **no header row** — first row is already data (an address). Single column A.

Implementation order (matching spec import steps 1-12):
1. Open DB, run migrations
2. Read `user_import.xlsx`:
   - Extract unique roles → insert into `roles` with mapping (Администратор→admin, Менеджер→manager, Авторизированный клиент→client)
   - For each row: split ФИО into last/first/patronymic, bcrypt hash password, insert into `users`
3. Read `Tovar.xlsx`:
   - Extract unique categories, manufacturers, suppliers, units → insert into respective tables
   - For each row: lookup FK IDs by name, insert into `products`
4. Read `Пункты выдачи_import.xlsx`:
   - NOTE: no header row, first row is already data
   - Insert each address into `pickup_points` (ID assigned by row order)
5. Read `Заказ_import.xlsx`:
   - Extract unique statuses → insert into `order_statuses`
   - For each row:
     - Parse order_date and delivery_date (handle invalid dates like 30.02.2025 → NULL, log warning)
     - Lookup pickup_point_id (column value is numeric index = row number in pickup_points file)
     - Lookup user_id by matching ФИО against users table
     - Insert into `orders`
     - Parse "Артикул заказа" column: split by ", " → pairs of (article, quantity) → lookup product_id by article → insert into `order_items`
6. Copy images: `1.jpg`–`10.jpg` and `picture.png` from `../import/` to `uploads/`

Key date parsing logic:
```go
func parseDate(val interface{}) *string {
    // Excel may give float (serial date) or string
    // Try time.Parse with formats: "2006-01-02", "02.01.2006"
    // Validate the date is real (e.g. reject Feb 30)
    // Return nil for invalid dates, log warning
}
```

Key article parsing:
```go
func parseOrderArticles(raw string) []struct{ Article string; Qty int } {
    // Split by ", " → ["А112Т4", "2", "F635R4", "2"]
    // Take pairs: (parts[0],parts[1]), (parts[2],parts[3]), ...
    // Parse quantity as int
}
```

- [ ] **Step 4: Verify it compiles**

```bash
cd backend && go build ./cmd/seed && cd ..
```

- [ ] **Step 5: Run seed and verify data**

```bash
cd backend && go run ./cmd/seed && cd ..
```

Expected: log output showing imported counts for each table, any warnings for invalid data.

Verify with sqlite3:
```bash
sqlite3 backend/data.db "SELECT COUNT(*) FROM products; SELECT COUNT(*) FROM users; SELECT COUNT(*) FROM orders;"
```

Expected: 43 products, 15 users, 53 orders (approximately).

- [ ] **Step 6: Commit**

```bash
git add backend/cmd/seed/
git commit -m "feat: data import seed command for Excel files"
```

---

## Task 10: Frontend scaffold

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/next.config.js`
- Create: `frontend/tsconfig.json`
- Create: `frontend/src/app/layout.tsx`
- Create: `frontend/src/app/page.tsx`
- Create: `frontend/src/types/index.ts`
- Create: `frontend/src/lib/api.ts`

- [ ] **Step 1: Initialize Next.js project**

```bash
npx create-next-app@latest frontend --typescript --app --src-dir --no-tailwind --no-eslint --import-alias "@/*"
```

When prompted, accept defaults. This creates the project structure.

- [ ] **Step 2: Install Ant Design**

```bash
cd frontend && npm install antd @ant-design/icons && cd ..
```

- [ ] **Step 3: Create TypeScript types**

`frontend/src/types/index.ts` — interfaces matching all backend models:
```typescript
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
```

- [ ] **Step 4: Create API client**

`frontend/src/lib/api.ts`:
```typescript
const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    credentials: "include",
    headers: { "Content-Type": "application/json", ...options?.headers },
    ...options,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(body.error || res.statusText);
  }
  return res.json();
}
```

- [ ] **Step 5: Configure root layout with Ant Design**

`frontend/src/app/layout.tsx`:
- Import Ant Design CSS
- Set Times New Roman as default font via global CSS
- Wrap children in AntdRegistry (for SSR compatibility)
- Set `<title>` default: "ООО Обувь"

- [ ] **Step 6: Create root page redirect**

`frontend/src/app/page.tsx` — redirect to `/products`:
```typescript
import { redirect } from "next/navigation";
export default function Home() {
  redirect("/products");
}
```

- [ ] **Step 7: Configure next.config.js**

Set up image domains for backend if needed. No proxy — CORS is handled on Go side.

- [ ] **Step 8: Verify it builds**

```bash
cd frontend && npm run build && cd ..
```

- [ ] **Step 9: Commit**

```bash
git add frontend/
git commit -m "feat: frontend scaffold with Next.js, Ant Design, types, API client"
```

---

## Task 11: Auth context and login page

**Files:**
- Create: `frontend/src/lib/auth.tsx`
- Create: `frontend/src/app/login/page.tsx`
- Create: `frontend/src/components/AppHeader.tsx`
- Modify: `frontend/src/app/layout.tsx`

- [ ] **Step 1: Create auth context**

`frontend/src/lib/auth.tsx`:
- React context with `user: User | null`, `login(email, password)`, `logout()`, `isGuest`, `isAdmin`, `isManager`, `loading`
- On mount: fetch `GET /api/auth/me` (with credentials: "include") to restore session from httpOnly cookie. If 401 → user stays null (guest). This is the only way to restore session since httpOnly cookies cannot be read by JS.
- `login()` → POST `/api/auth/login` (credentials: "include"), store returned user in state
- `logout()` → POST `/api/auth/logout` (credentials: "include"), set user to null, redirect to /login

- [ ] **Step 2: Create login page**

`frontend/src/app/login/page.tsx`:
- Ant Design Form with Input (login), Input.Password (password)
- "Войти" button with accent color #00FA9A
- "Войти как гость" link → navigates to /products without auth
- On submit: call auth.login(), redirect to /products on success
- On error: show `message.error()` with server error text
- Page metadata title: "Вход — ООО Обувь"

- [ ] **Step 3: Create AppHeader component**

`frontend/src/components/AppHeader.tsx`:
- Background color #7FFF00
- Left: logo (import/Icon.png → copy to public/), nav links ("Товары", "Заказы" — show "Заказы" only for manager/admin)
- Right: ФИО (lastName + initials) if logged in, "Выход" button. If guest: "Войти" link.
- Uses useAuth() hook

- [ ] **Step 4: Wire auth provider into layout**

Update `frontend/src/app/layout.tsx`:
- Wrap children with `<AuthProvider>`
- Add `<AppHeader />` above children

- [ ] **Step 5: Copy assets to public/**

```bash
cp import/Icon.png frontend/public/logo.png
cp import/Icon.ico frontend/public/favicon.ico
cp import/picture.png frontend/public/placeholder.png
```

- [ ] **Step 6: Verify it builds**

```bash
cd frontend && npm run build && cd ..
```

- [ ] **Step 7: Commit**

```bash
git add frontend/ import/
git commit -m "feat: auth context, login page, app header with role-based nav"
```

---

## Task 12: Products catalog page

**Files:**
- Create: `frontend/src/app/products/page.tsx`
- Create: `frontend/src/components/ProductCard.tsx`
- Create: `frontend/src/components/ProductToolbar.tsx`

- [ ] **Step 1: Create ProductCard component**

`frontend/src/components/ProductCard.tsx`:

Renders a single product row/card with conditional styling per spec:
- **Discount > 15%**: background `#2E8B57`, old price struck through in red, new price in black
- **Out of stock (quantity === 0)**: background `#e0f2fe` (light blue)
- **Normal**: white background

Layout: image (60×60, fallback to `/placeholder.png`), name, category, manufacturer, supplier, description, price, quantity, unit, discount.

Image src: `${API_BASE}/uploads/${product.image}` (import `API_BASE` from `@/lib/api`) or `/placeholder.png` if image is empty.

Price display logic:
```typescript
if (product.discount > 0) {
  const discountedPrice = product.price * (1 - product.discount / 100);
  // Show: <span style={{textDecoration:'line-through', color:'red'}}>{price}</span>
  //        <span style={{color:'black', fontWeight:'bold'}}>{discountedPrice}</span>
}
```

- [ ] **Step 2: Create ProductToolbar component**

`frontend/src/components/ProductToolbar.tsx`:

Only rendered if role is manager or admin. Contains:
- `Input.Search` — value bound to search state, onChange fires immediately (real-time search, no button)
- `Select` — supplier filter. Options from `/api/suppliers`. First option: "Все поставщики" (value: 0). onChange fires immediately.
- Sort buttons — "По количеству ↑" / "По количеству ↓" toggle

All changes call parent callback `onFilterChange({ search, supplierId, sort })`.

- [ ] **Step 3: Create products page**

`frontend/src/app/products/page.tsx`:

- Fetches products from `GET /api/products` with query params from toolbar
- Renders ProductToolbar (if manager+) and list of ProductCard
- Admin: "Добавить товар" button (accent #00FA9A) → navigates to /products/new
- Admin: click on product card → navigates to /products/{id}/edit
- Uses `useEffect` to re-fetch when filters change (real-time)
- Page metadata title: "Товары — ООО Обувь"

- [ ] **Step 4: Verify it builds**

```bash
cd frontend && npm run build && cd ..
```

- [ ] **Step 5: Commit**

```bash
git add frontend/
git commit -m "feat: product catalog with conditional styling, search, filter, sort"
```

---

## Task 13: Product add/edit form

**Files:**
- Create: `frontend/src/components/ProductForm.tsx`
- Create: `frontend/src/app/products/new/page.tsx`
- Create: `frontend/src/app/products/[id]/edit/page.tsx`

- [ ] **Step 1: Create ProductForm component**

`frontend/src/components/ProductForm.tsx`:

Shared component used by both add and edit pages. Props: `product?: Product` (if editing), `onSuccess: () => void`.

Ant Design Form with fields (matching spec widget mapping):
- `article` → Input (hidden on add, read-only on edit — shows the product article code like "А112Т4")
- `name` → Input (required)
- `category` → Select, options from `/api/categories` (required)
- `manufacturer` → Select, options from `/api/manufacturers` (required)
- `supplier` → Select, options from `/api/suppliers` (required)
- `description` → TextArea
- `price` → InputNumber, min=0, step=0.01, precision=2 (required)
- `discount` → InputNumber, min=0
- `quantity` → InputNumber, min=0, precision=0 (required)
- `unit` → Select, options from `/api/units` (required)
- `image` → Upload (Ant Upload with beforeUpload returning false for manual control). Preview of current image.

Buttons: "Сохранить" (#00FA9A), "Назад" (navigates back)

On submit:
- Add mode: POST `/api/products`, then POST `/api/products/{id}/image` if file selected
- Edit mode: PUT `/api/products/{id}`, then POST `/api/products/{id}/image` if file changed

On error: `notification.error()` with title "Ошибка" and description from server.

- [ ] **Step 2: Create add product page**

`frontend/src/app/products/new/page.tsx`:
- Renders `<ProductForm />` without product prop
- Page metadata title: "Добавление товара — ООО Обувь"
- On success: redirect to /products

- [ ] **Step 3: Create edit product page**

`frontend/src/app/products/[id]/edit/page.tsx`:
- Fetches product by ID from `/api/products/{id}`
- Renders `<ProductForm product={data} />`
- Page metadata title: "Редактирование товара — ООО Обувь"
- On success: redirect to /products

- [ ] **Step 4: Add delete button to edit page (admin only)**

On the edit page, add a "Удалить" button:
- Calls `Modal.confirm()` with warning text "Вы уверены, что хотите удалить этот товар?"
- On confirm: DELETE `/api/products/{id}`
- Handle 409 error: show `Modal.warning({ title: "Невозможно удалить", content: "Товар присутствует в заказе" })`
- On success: redirect to /products

- [ ] **Step 5: Verify it builds**

```bash
cd frontend && npm run build && cd ..
```

- [ ] **Step 6: Commit**

```bash
git add frontend/
git commit -m "feat: product add/edit/delete forms with validation and image upload"
```

---

## Task 14: Orders list page

**Files:**
- Create: `frontend/src/app/orders/page.tsx`

- [ ] **Step 1: Create orders page**

`frontend/src/app/orders/page.tsx`:

- Fetches orders from `GET /api/orders`
- Ant Design Table with columns:
  - Артикулы (join order items articles with comma)
  - Статус (with colored Tag — Ant Design)
  - Пункт выдачи (address)
  - Дата заказа
  - Дата доставки (or "—" if null)
- Admin: "Добавить заказ" button (#00FA9A) → /orders/new
- Admin: click row → /orders/{id}/edit
- Page metadata title: "Заказы — ООО Обувь"

- [ ] **Step 2: Verify it builds**

```bash
cd frontend && npm run build && cd ..
```

- [ ] **Step 3: Commit**

```bash
git add frontend/
git commit -m "feat: orders list page with table display"
```

---

## Task 15: Order add/edit form

**Files:**
- Create: `frontend/src/components/OrderForm.tsx`
- Create: `frontend/src/app/orders/new/page.tsx`
- Create: `frontend/src/app/orders/[id]/edit/page.tsx`

- [ ] **Step 1: Create OrderForm component**

`frontend/src/components/OrderForm.tsx`:

Shared component. Props: `order?: Order`, `onSuccess: () => void`.

Fields:
- `status` → Select, options from `/api/order-statuses` (required)
- `pickup_point` → Select, options from `/api/pickup-points` (required)
- `order_date` → DatePicker (required)
- `delivery_date` → DatePicker (optional)
- `pickup_code` → Input
- **Позиции заказа** → Ant Form.List:
  - Each row: Select (product, options from `/api/products` — display article + name), InputNumber (quantity, min=1)
  - "Добавить товар" button to add row
  - "Удалить" button per row to remove

Buttons: "Сохранить" (#00FA9A), "Назад"

On submit:
- Build OrderInput object with items array
- Add: POST `/api/orders`
- Edit: PUT `/api/orders/{id}`

- [ ] **Step 2: Create add order page**

`frontend/src/app/orders/new/page.tsx`:
- Renders `<OrderForm />`
- Title: "Добавление заказа — ООО Обувь"
- On success: redirect to /orders

- [ ] **Step 3: Create edit order page**

`frontend/src/app/orders/[id]/edit/page.tsx`:
- Fetch order by ID (with items)
- Renders `<OrderForm order={data} />`
- Title: "Редактирование заказа — ООО Обувь"
- Add "Удалить" button (admin): confirm modal, DELETE `/api/orders/{id}`, redirect to /orders

- [ ] **Step 4: Verify it builds**

```bash
cd frontend && npm run build && cd ..
```

- [ ] **Step 5: Commit**

```bash
git add frontend/
git commit -m "feat: order add/edit/delete forms with dynamic order items"
```

---

## Task 16: End-to-end smoke test

**Files:** none (manual testing)

- [ ] **Step 1: Start backend**

```bash
cd backend && go run ./cmd/seed && go run ./cmd/server
```

- [ ] **Step 2: Start frontend**

```bash
cd frontend && npm run dev
```

- [ ] **Step 3: Test guest flow**

Open http://localhost:3000. Click "Войти как гость". Verify:
- Product list loads with data from seed
- Conditional styling works (discount > 15% green, out of stock blue)
- No toolbar visible (guest has no filter/sort/search)
- No "Добавить" button visible

- [ ] **Step 4: Test admin flow**

Login with admin credentials from seed data. Verify:
- ФИО shows in header
- Toolbar visible with search, filter, sort
- Search works in real-time
- Supplier filter works
- Quantity sort works
- Can add product (with image upload)
- Can edit product
- Cannot delete product that is in an order (409 error shown)
- Can navigate to Orders
- Can add/edit/delete orders
- Logout returns to login screen

- [ ] **Step 5: Test manager flow**

Login with manager credentials. Verify:
- Toolbar visible
- No "Добавить" / edit / delete buttons on products
- Orders page visible (view only, no add/edit/delete)

- [ ] **Step 6: Test client flow**

Login with client credentials. Verify:
- Product list visible without toolbar
- No orders navigation

- [ ] **Step 7: Final commit**

```bash
git add -A
git commit -m "chore: final verification and cleanup"
```
