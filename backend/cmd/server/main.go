package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"shoe-store/internal/database"
)

func main() {
	db, err := database.Open("data.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := database.Migrate(db); err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
