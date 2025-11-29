package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL  string
	ServerPort   string
	JWTSecretKey string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("DATABASE_URL is not set")
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = ":8080"
	}

	jwtKey := os.Getenv("JWT_SECRET_KEY")
	if jwtKey == "" {
		log.Fatal("JWT_SECRET_KEY is not set")
	}

	return &Config{
		DatabaseURL:  dbURL,
		ServerPort:   port,
		JWTSecretKey: jwtKey,
	}, nil
}
