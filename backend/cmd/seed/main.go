package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"

	"shoe-store/internal/database"
)

// sheetName returns the first sheet name in the workbook.
func sheetName(f *excelize.File) string {
	list := f.GetSheetList()
	if len(list) == 0 {
		return "Sheet1"
	}
	return list[0]
}

func main() {
	importDir := flag.String("import", "../import", "path to import directory with Excel files and images")
	dbPath := flag.String("db", "data.db", "path to SQLite database file")
	flag.Parse()

	// Open and migrate DB
	db, err := database.Open(*dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	log.Println("Starting data import...")

	// 1. Import roles
	roleMap, err := importRoles(db, *importDir)
	if err != nil {
		log.Fatalf("import roles: %v", err)
	}
	log.Printf("Imported %d roles", len(roleMap))

	// 2. Import users
	userCount, err := importUsers(db, *importDir, roleMap)
	if err != nil {
		log.Fatalf("import users: %v", err)
	}
	log.Printf("Imported %d users", userCount)

	// 3. Import categories
	categoryMap, err := importRef(db, *importDir, "Tovar.xlsx", 6, "categories")
	if err != nil {
		log.Fatalf("import categories: %v", err)
	}
	log.Printf("Imported %d categories", len(categoryMap))

	// 4. Import manufacturers
	manufacturerMap, err := importRef(db, *importDir, "Tovar.xlsx", 5, "manufacturers")
	if err != nil {
		log.Fatalf("import manufacturers: %v", err)
	}
	log.Printf("Imported %d manufacturers", len(manufacturerMap))

	// 5. Import suppliers
	supplierMap, err := importRef(db, *importDir, "Tovar.xlsx", 4, "suppliers")
	if err != nil {
		log.Fatalf("import suppliers: %v", err)
	}
	log.Printf("Imported %d suppliers", len(supplierMap))

	// 6. Import units
	unitMap, err := importRef(db, *importDir, "Tovar.xlsx", 2, "units")
	if err != nil {
		log.Fatalf("import units: %v", err)
	}
	log.Printf("Imported %d units", len(unitMap))

	// 7. Import products
	productCount, err := importProducts(db, *importDir, categoryMap, manufacturerMap, supplierMap, unitMap)
	if err != nil {
		log.Fatalf("import products: %v", err)
	}
	log.Printf("Imported %d products", productCount)

	// 8. Import pickup points
	ppCount, err := importPickupPoints(db, *importDir)
	if err != nil {
		log.Fatalf("import pickup points: %v", err)
	}
	log.Printf("Imported %d pickup points", ppCount)

	// 9. Import order statuses
	statusMap, err := importRef(db, *importDir, "\u0417\u0430\u043a\u0430\u0437_import.xlsx", 7, "order_statuses")
	if err != nil {
		log.Fatalf("import order statuses: %v", err)
	}
	log.Printf("Imported %d order statuses", len(statusMap))

	// 10. Import orders
	orderCount, err := importOrders(db, *importDir, statusMap)
	if err != nil {
		log.Fatalf("import orders: %v", err)
	}
	log.Printf("Imported %d orders", orderCount)

	// 11. Import order items
	itemCount, err := importOrderItems(db, *importDir)
	if err != nil {
		log.Fatalf("import order items: %v", err)
	}
	log.Printf("Imported %d order items", itemCount)

	// 12. Copy images
	imgCount, err := copyImages(*importDir, "uploads")
	if err != nil {
		log.Fatalf("copy images: %v", err)
	}
	log.Printf("Copied %d images", imgCount)

	log.Println("Data import completed successfully!")
}

// roleNameMap maps Russian role names to internal names.
var roleNameMap = map[string]string{
	"\u0410\u0434\u043c\u0438\u043d\u0438\u0441\u0442\u0440\u0430\u0442\u043e\u0440":                "admin",
	"\u041c\u0435\u043d\u0435\u0434\u0436\u0435\u0440":                                              "manager",
	"\u0410\u0432\u0442\u043e\u0440\u0438\u0437\u0438\u0440\u043e\u0432\u0430\u043d\u043d\u044b\u0439 \u043a\u043b\u0438\u0435\u043d\u0442": "client",
}

// importRoles extracts unique roles from user_import.xlsx and inserts them.
// Returns map[russianName] -> role_id
func importRoles(db *sql.DB, importDir string) (map[string]int64, error) {
	f, err := excelize.OpenFile(filepath.Join(importDir, "user_import.xlsx"))
	if err != nil {
		return nil, fmt.Errorf("open user_import.xlsx: %w", err)
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName(f))
	if err != nil {
		return nil, fmt.Errorf("get rows: %w", err)
	}

	seen := map[string]bool{}
	roleMap := map[string]int64{}

	for i, row := range rows {
		if i == 0 { // skip header
			continue
		}
		if len(row) < 1 {
			continue
		}
		ruName := strings.TrimSpace(row[0])
		if ruName == "" || seen[ruName] {
			continue
		}
		seen[ruName] = true

		internalName, ok := roleNameMap[ruName]
		if !ok {
			log.Printf("WARNING: unknown role %q, using as-is", ruName)
			internalName = ruName
		}

		res, err := db.Exec("INSERT INTO roles (name) VALUES (?) ON CONFLICT(name) DO NOTHING", internalName)
		if err != nil {
			return nil, fmt.Errorf("insert role %q: %w", internalName, err)
		}

		id, _ := res.LastInsertId()
		if id == 0 {
			// Already existed, look it up
			if err := db.QueryRow("SELECT id FROM roles WHERE name = ?", internalName).Scan(&id); err != nil {
				return nil, fmt.Errorf("lookup role %q: %w", internalName, err)
			}
		}
		roleMap[ruName] = id
	}

	return roleMap, nil
}

// importUsers reads user_import.xlsx and inserts users.
func importUsers(db *sql.DB, importDir string, roleMap map[string]int64) (int, error) {
	f, err := excelize.OpenFile(filepath.Join(importDir, "user_import.xlsx"))
	if err != nil {
		return 0, fmt.Errorf("open user_import.xlsx: %w", err)
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName(f))
	if err != nil {
		return 0, fmt.Errorf("get rows: %w", err)
	}

	count := 0
	for i, row := range rows {
		if i == 0 { // skip header
			continue
		}
		if len(row) < 4 {
			log.Printf("WARNING: user row %d has fewer than 4 columns, skipping", i+1)
			continue
		}

		ruRole := strings.TrimSpace(row[0])
		fio := strings.TrimSpace(row[1])
		login := strings.TrimSpace(row[2])
		password := strings.TrimSpace(row[3])

		roleID, ok := roleMap[ruRole]
		if !ok {
			log.Printf("WARNING: role %q not found for user row %d, skipping", ruRole, i+1)
			continue
		}

		parts := strings.Fields(fio)
		var lastName, firstName, patronymic string
		if len(parts) >= 1 {
			lastName = parts[0]
		}
		if len(parts) >= 2 {
			firstName = parts[1]
		}
		if len(parts) >= 3 {
			patronymic = parts[2]
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			return 0, fmt.Errorf("bcrypt hash for user %q: %w", login, err)
		}

		_, err = db.Exec(
			`INSERT INTO users (login, password, last_name, first_name, patronymic, role_id)
			 VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT(login) DO NOTHING`,
			login, string(hash), lastName, firstName, patronymic, roleID,
		)
		if err != nil {
			return 0, fmt.Errorf("insert user %q: %w", login, err)
		}
		count++
	}
	return count, nil
}

// importRef extracts unique values from a specific column of an Excel file
// and inserts them into a reference table (categories, manufacturers, suppliers, units, order_statuses).
// colIdx is 0-based. Returns map[name] -> id.
func importRef(db *sql.DB, importDir, filename string, colIdx int, tableName string) (map[string]int64, error) {
	f, err := excelize.OpenFile(filepath.Join(importDir, filename))
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", filename, err)
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName(f))
	if err != nil {
		return nil, fmt.Errorf("get rows: %w", err)
	}

	seen := map[string]bool{}
	result := map[string]int64{}

	for i, row := range rows {
		if i == 0 { // skip header
			continue
		}
		if len(row) <= colIdx {
			continue
		}
		name := strings.TrimSpace(row[colIdx])
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true

		query := fmt.Sprintf("INSERT INTO %s (name) VALUES (?) ON CONFLICT(name) DO NOTHING", tableName)
		res, err := db.Exec(query, name)
		if err != nil {
			return nil, fmt.Errorf("insert %s %q: %w", tableName, name, err)
		}
		id, _ := res.LastInsertId()
		if id == 0 {
			lookupQ := fmt.Sprintf("SELECT id FROM %s WHERE name = ?", tableName)
			if err := db.QueryRow(lookupQ, name).Scan(&id); err != nil {
				return nil, fmt.Errorf("lookup %s %q: %w", tableName, name, err)
			}
		}
		result[name] = id
	}

	return result, nil
}

// importProducts reads Tovar.xlsx and inserts products.
func importProducts(db *sql.DB, importDir string, categoryMap, manufacturerMap, supplierMap, unitMap map[string]int64) (int, error) {
	f, err := excelize.OpenFile(filepath.Join(importDir, "Tovar.xlsx"))
	if err != nil {
		return 0, fmt.Errorf("open Tovar.xlsx: %w", err)
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName(f))
	if err != nil {
		return 0, fmt.Errorf("get rows: %w", err)
	}

	count := 0
	for i, row := range rows {
		if i == 0 { // skip header
			continue
		}
		if len(row) < 10 {
			log.Printf("WARNING: product row %d has fewer than 10 columns, skipping", i+1)
			continue
		}

		article := strings.TrimSpace(row[0])
		name := strings.TrimSpace(row[1])
		unit := strings.TrimSpace(row[2])
		priceStr := strings.TrimSpace(row[3])
		supplier := strings.TrimSpace(row[4])
		manufacturer := strings.TrimSpace(row[5])
		category := strings.TrimSpace(row[6])
		discountStr := strings.TrimSpace(row[7])
		qtyStr := strings.TrimSpace(row[8])
		description := strings.TrimSpace(row[9])
		image := ""
		if len(row) >= 11 {
			image = strings.TrimSpace(row[10])
		}

		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			log.Printf("WARNING: invalid price %q in product row %d, setting to 0", priceStr, i+1)
			price = 0
		}

		discount, err := strconv.ParseFloat(discountStr, 64)
		if err != nil {
			log.Printf("WARNING: invalid discount %q in product row %d, setting to 0", discountStr, i+1)
			discount = 0
		}

		qty, err := strconv.Atoi(qtyStr)
		if err != nil {
			log.Printf("WARNING: invalid quantity %q in product row %d, setting to 0", qtyStr, i+1)
			qty = 0
		}

		catID, ok := categoryMap[category]
		if !ok {
			log.Printf("WARNING: category %q not found for product row %d, skipping", category, i+1)
			continue
		}
		mfgID, ok := manufacturerMap[manufacturer]
		if !ok {
			log.Printf("WARNING: manufacturer %q not found for product row %d, skipping", manufacturer, i+1)
			continue
		}
		supID, ok := supplierMap[supplier]
		if !ok {
			log.Printf("WARNING: supplier %q not found for product row %d, skipping", supplier, i+1)
			continue
		}
		unitID, ok := unitMap[unit]
		if !ok {
			log.Printf("WARNING: unit %q not found for product row %d, skipping", unit, i+1)
			continue
		}

		_, execErr := db.Exec(
			`INSERT INTO products (article, name, description, price, discount, quantity, image,
				category_id, manufacturer_id, supplier_id, unit_id)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(article) DO NOTHING`,
			article, name, description, price, discount, qty, image,
			catID, mfgID, supID, unitID,
		)
		if execErr != nil {
			log.Printf("WARNING: insert product %q (row %d): %v, skipping duplicate", article, i+1, execErr)
			continue
		}
		count++
	}
	return count, nil
}

// importPickupPoints reads Пункты выдачи_import.xlsx (no header row).
func importPickupPoints(db *sql.DB, importDir string) (int, error) {
	f, err := excelize.OpenFile(filepath.Join(importDir, "\u041f\u0443\u043d\u043a\u0442\u044b \u0432\u044b\u0434\u0430\u0447\u0438_import.xlsx"))
	if err != nil {
		return 0, fmt.Errorf("open pickup points file: %w", err)
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName(f))
	if err != nil {
		return 0, fmt.Errorf("get rows: %w", err)
	}

	count := 0
	for _, row := range rows {
		if len(row) < 1 {
			continue
		}
		address := strings.TrimSpace(row[0])
		if address == "" {
			continue
		}
		_, err := db.Exec("INSERT INTO pickup_points (address) VALUES (?)", address)
		if err != nil {
			return 0, fmt.Errorf("insert pickup point: %w", err)
		}
		count++
	}
	return count, nil
}

// parseDate attempts to parse a date value from Excel.
// Excelize GetRows returns strings, but GetCellValue can return time.
// We handle various string formats and validate the date.
func parseDate(raw string) (*string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, true
	}

	formats := []string{
		"2006-01-02",
		"02.01.2006",
		"01-02-06",
		"2006-01-02 15:04:05",
		"01-02-06 15:04",
		"1/2/06 15:04",
	}

	for _, layout := range formats {
		t, err := time.Parse(layout, raw)
		if err == nil {
			if !isValidDate(t) {
				log.Printf("WARNING: invalid date %q (parsed as %v)", raw, t)
				return nil, false
			}
			s := t.Format("2006-01-02")
			return &s, true
		}
	}

	// Try parsing as Excel serial date number
	if serial, err := strconv.ParseFloat(raw, 64); err == nil {
		t := excelSerialToTime(serial)
		if !isValidDate(t) {
			log.Printf("WARNING: invalid date from serial %q (parsed as %v)", raw, t)
			return nil, false
		}
		s := t.Format("2006-01-02")
		return &s, true
	}

	log.Printf("WARNING: cannot parse date %q", raw)
	return nil, false
}

// excelSerialToTime converts an Excel serial date number to time.Time.
func excelSerialToTime(serial float64) time.Time {
	// Excel epoch is January 1, 1900 (but with a bug: treats 1900 as leap year)
	// Day 1 = Jan 1, 1900
	// Day 60 = Feb 29, 1900 (doesn't exist, Excel bug)
	// Day 61 = Mar 1, 1900
	epoch := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	days := int(serial)
	fraction := serial - float64(days)
	t := epoch.AddDate(0, 0, days)
	t = t.Add(time.Duration(fraction * 24 * float64(time.Hour)))
	return t
}

// isValidDate checks if a parsed date has valid month/day ranges.
func isValidDate(t time.Time) bool {
	y, m, d := t.Date()
	// Reconstruct from y/m/d and check it matches
	reconstructed := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return reconstructed.Year() == y && reconstructed.Month() == m && reconstructed.Day() == d
}

// importOrders reads Заказ_import.xlsx and inserts orders.
func importOrders(db *sql.DB, importDir string, statusMap map[string]int64) (int, error) {
	fpath := filepath.Join(importDir, "\u0417\u0430\u043a\u0430\u0437_import.xlsx")
	f, err := excelize.OpenFile(fpath)
	if err != nil {
		return 0, fmt.Errorf("open orders file: %w", err)
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName(f))
	if err != nil {
		return 0, fmt.Errorf("get rows: %w", err)
	}

	count := 0
	for i, row := range rows {
		if i == 0 { // skip header
			continue
		}
		if len(row) < 8 {
			log.Printf("WARNING: order row %d has fewer than 8 columns, skipping", i+1)
			continue
		}

		// A: order number (we use auto-increment, but read for reference)
		// B: articles (handled in importOrderItems)
		orderDateRaw := strings.TrimSpace(row[2])   // C
		deliveryDateRaw := strings.TrimSpace(row[3]) // D
		pickupPointRaw := strings.TrimSpace(row[4])  // E
		fio := strings.TrimSpace(row[5])             // F
		pickupCode := strings.TrimSpace(row[6])      // G
		statusName := strings.TrimSpace(row[7])      // H

		statusID, ok := statusMap[statusName]
		if !ok {
			log.Printf("WARNING: status %q not found for order row %d, skipping", statusName, i+1)
			continue
		}

		// Parse pickup point ID from column E (numeric index)
		ppID, err := parseIntFromExcel(pickupPointRaw)
		if err != nil {
			log.Printf("WARNING: invalid pickup point %q in order row %d: %v", pickupPointRaw, i+1, err)
			continue
		}

		// Parse dates
		var orderDate *string
		var deliveryDate *string

		orderDate, _ = parseDate(orderDateRaw)
		deliveryDate, _ = parseDate(deliveryDateRaw)

		// Lookup user by FIO
		var userID *int64
		if fio != "" {
			parts := strings.Fields(fio)
			var lastName, firstName, patronymic string
			if len(parts) >= 1 {
				lastName = parts[0]
			}
			if len(parts) >= 2 {
				firstName = parts[1]
			}
			if len(parts) >= 3 {
				patronymic = parts[2]
			}

			var uid int64
			err := db.QueryRow(
				"SELECT id FROM users WHERE last_name = ? AND first_name = ? AND patronymic = ?",
				lastName, firstName, patronymic,
			).Scan(&uid)
			if err != nil {
				log.Printf("WARNING: user %q not found for order row %d", fio, i+1)
			} else {
				userID = &uid
			}
		}

		_, execErr := db.Exec(
			`INSERT INTO orders (order_date, delivery_date, pickup_code, status_id, pickup_point_id, user_id)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			orderDate, deliveryDate, pickupCode, statusID, ppID, userID,
		)
		if execErr != nil {
			return 0, fmt.Errorf("insert order row %d: %w", i+1, execErr)
		}
		count++
	}
	return count, nil
}

// parseIntFromExcel parses an integer that may come as a float string from Excel.
func parseIntFromExcel(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	// Try direct int parse first
	if v, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return v, nil
	}
	// Try float parse (Excel may export "1.0")
	if v, err := strconv.ParseFloat(raw, 64); err == nil {
		return int64(math.Round(v)), nil
	}
	return 0, fmt.Errorf("cannot parse %q as integer", raw)
}

// importOrderItems reads Заказ_import.xlsx column B and inserts order items.
func importOrderItems(db *sql.DB, importDir string) (int, error) {
	fpath := filepath.Join(importDir, "\u0417\u0430\u043a\u0430\u0437_import.xlsx")
	f, err := excelize.OpenFile(fpath)
	if err != nil {
		return 0, fmt.Errorf("open orders file: %w", err)
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName(f))
	if err != nil {
		return 0, fmt.Errorf("get rows: %w", err)
	}

	// Build article -> product_id map
	articleMap := map[string]int64{}
	dbRows, err := db.Query("SELECT id, article FROM products")
	if err != nil {
		return 0, fmt.Errorf("query products: %w", err)
	}
	defer dbRows.Close()
	for dbRows.Next() {
		var id int64
		var article string
		if err := dbRows.Scan(&id, &article); err != nil {
			return 0, fmt.Errorf("scan product: %w", err)
		}
		articleMap[article] = id
	}
	if err := dbRows.Err(); err != nil {
		return 0, fmt.Errorf("iterate products: %w", err)
	}

	count := 0
	orderIdx := 0 // tracks order_id (1-based, matching auto-increment)
	for i, row := range rows {
		if i == 0 { // skip header
			continue
		}
		orderIdx++

		if len(row) < 2 {
			continue
		}
		raw := strings.TrimSpace(row[1]) // column B
		if raw == "" {
			continue
		}

		// Parse "А112Т4, 2, F635R4, 2" -> pairs
		parts := strings.Split(raw, ", ")
		for j := 0; j < len(parts)-1; j += 2 {
			article := strings.TrimSpace(parts[j])
			qtyStr := strings.TrimSpace(parts[j+1])
			qty, err := strconv.Atoi(qtyStr)
			if err != nil {
				log.Printf("WARNING: invalid quantity %q for article %q in order row %d", qtyStr, article, i+1)
				continue
			}

			productID, ok := articleMap[article]
			if !ok {
				log.Printf("WARNING: article %q not found in products for order row %d", article, i+1)
				continue
			}

			_, execErr := db.Exec(
				"INSERT INTO order_items (order_id, product_id, quantity) VALUES (?, ?, ?)",
				orderIdx, productID, qty,
			)
			if execErr != nil {
				return 0, fmt.Errorf("insert order item for order %d, article %q: %w", orderIdx, article, execErr)
			}
			count++
		}
	}
	return count, nil
}

// copyImages copies image files from import dir to uploads dir.
func copyImages(importDir, uploadsDir string) (int, error) {
	if err := os.MkdirAll(uploadsDir, 0o755); err != nil {
		return 0, fmt.Errorf("create uploads dir: %w", err)
	}

	files := []string{
		"1.jpg", "2.jpg", "3.jpg", "4.jpg", "5.jpg",
		"6.jpg", "7.jpg", "8.jpg", "9.jpg", "10.jpg",
		"picture.png",
	}

	count := 0
	for _, name := range files {
		src := filepath.Join(importDir, name)
		dst := filepath.Join(uploadsDir, name)

		if _, err := os.Stat(src); os.IsNotExist(err) {
			log.Printf("WARNING: image %s not found in import dir, skipping", name)
			continue
		}

		if err := copyFile(src, dst); err != nil {
			return 0, fmt.Errorf("copy %s: %w", name, err)
		}
		count++
	}
	return count, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
