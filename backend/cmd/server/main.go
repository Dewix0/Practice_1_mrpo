package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"shoe-store/internal/database"
	"shoe-store/internal/handler"
	"shoe-store/internal/middleware"
	"shoe-store/internal/repository"
	"shoe-store/internal/service"
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

	userRepo := repository.NewUserRepo(db)
	authService := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthHandler(authService)

	r := chi.NewRouter()

	// Global middleware
	corsMiddleware := middleware.NewCORS()
	r.Use(corsMiddleware.Handler)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(middleware.JWTAuth(authService.JwtSecret))

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Auth routes
	r.Post("/api/auth/login", authHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth())
		r.Get("/api/auth/me", authHandler.Me)
		r.Post("/api/auth/logout", authHandler.Logout)
	})

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
