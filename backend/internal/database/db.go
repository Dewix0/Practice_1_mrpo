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
