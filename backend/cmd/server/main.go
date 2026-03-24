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

	productRepo := repository.NewProductRepo(db)
	productService := service.NewProductService(productRepo, "uploads")
	productHandler := handler.NewProductHandler(productService)

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

	// Product routes
	r.Route("/api/products", func(r chi.Router) {
		r.Get("/", productHandler.List)
		r.Get("/{id}", productHandler.GetByID)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole("admin"))
			r.Post("/", productHandler.Create)
			r.Put("/{id}", productHandler.Update)
			r.Delete("/{id}", productHandler.Delete)
			r.Post("/{id}/image", productHandler.UploadImage)
		})
	})

	// Static file server for uploads
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
