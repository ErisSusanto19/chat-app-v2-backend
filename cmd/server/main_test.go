package main

import (
	"database/sql"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

var testDB *sql.DB

func TestMain(m *testing.M) {

	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("Error loading .env file for tests")
	}

	// 1. Setup
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		log.Fatal("TEST_DATABASE_URL is not set")
	}
	// log.Printf("DEBUG: Connecting to test database with URL: %s", dbURL)

	// _, b, _, _ := runtime.Caller(0)
	// projectRoot := filepath.Join(filepath.Dir(b), "../..")
	// migrationsPath := filepath.Join(projectRoot, "migrations")

	// Jalankan migrasi
	// mig, err := migrate.New("file://"+migrationsPath, dbURL)
	mig, err := migrate.New("file://../../migrations", dbURL)
	if err != nil {
		log.Fatalf("could not create migrate instance: %v", err)
	}
	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("could not run up migrations: %v", err)
	}

	sqlOpenURL := strings.Replace(dbURL, "pgx5://", "postgres://", 1)
	// log.Printf("DEBUG: Using sql.Open URL: %s", sqlOpenURL)

	// Terhubung ke database tes
	testDB, err = sql.Open("pgx", sqlOpenURL)
	if err != nil {
		log.Fatalf("could not connect to test database: %v", err)
	}
	defer testDB.Close()

	if err := testDB.Ping(); err != nil {
		log.Fatalf("Could not ping test database: %v", err)
	}

	// 2. Jalankan semua tes
	exitCode := m.Run()

	os.Exit(exitCode)
}
