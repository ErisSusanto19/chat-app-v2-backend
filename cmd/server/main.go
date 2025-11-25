package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/config"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/domain"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/repository"
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

	log.Println("Attempting to create a test user...")
	testUser := &domain.User{
		Name:           "Test User",
		Email:          "test@example.com",
		HashedPassword: "password123",
	}

	err = userRepo.CreateUser(ctx, testUser)
	if err != nil {
		log.Printf("Failed to create test user: %v", err)
	} else {
		log.Printf("Successfully created test user with ID: %s", testUser.ID)
	}

	fmt.Printf("Starting server on port %s\n", cfg.ServerPort)
}
