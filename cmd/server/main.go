package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/config"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/repository"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/service"
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

	if err := db.Ping(); err != nil {
		log.Fatalf("could not ping database: %v", err)
	}

	fmt.Println("Successfully connected to the database!")

	userRepo := repository.NewPostgresUserRepository(db)

	authService := service.NewAuthService(userRepo)

	log.Println("Attempting to register a new user via auth service...")

	newUser, err := authService.Register(ctx, "Service User", "service.user@example.com", "strongpassword123")
	if err != nil {
		log.Printf("Failed to register user: %v", err)
	} else {
		log.Printf("Successfully registered user with ID: %s and Name: %s", newUser.ID, newUser.Name)
	}

	fmt.Printf("Starting server on port %s\n", cfg.ServerPort)
}
