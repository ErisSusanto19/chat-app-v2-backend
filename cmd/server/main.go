package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/config"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/handler"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/repository"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {

	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("could not ping database: %v", err)
	}

	fmt.Println("Successfully connected to the database!")
	fmt.Printf("Starting server on port %s\n", cfg.ServerPort)

	userRepo := repository.NewPostgresUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecretKey)
	authHandler := handler.NewAuthHandler(authService)

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Route("/api/v1", func(r chi.Router) {
		authHandler.RegisterRoutes(r.(*chi.Mux))
	})

	log.Printf("Starting server on port %s", cfg.ServerPort)
	if err := http.ListenAndServe(cfg.ServerPort, router); err != nil {
		log.Fatalf("could not start server: %v", err)
	}

}
