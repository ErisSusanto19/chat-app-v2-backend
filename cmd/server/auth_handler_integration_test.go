package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/handler"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/repository"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer() (http.Handler, func()) {
	teardown := func() {
		_, err := testDB.Exec("TRUNCATE users, contacts, conversations, user_conversations, messages RESTART IDENTITY CASCADE")
		if err != nil {
			log.Fatalf("could not truncate tables: %v", err)
		}
	}

	userRepo := repository.NewPostgresUserRepository(testDB)
	authService := service.NewAuthService(userRepo, "test-secret", nil)
	authHandler := handler.NewAuthHandler(authService)

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Post("/register", authHandler.Register)

	return router, teardown
}

func TestRegister_Integration(t *testing.T) {
	router, teardown := setupTestServer()
	defer teardown()

	requestBody := map[string]string{
		"name":     "Integration Test User",
		"email":    "integration@test.com",
		"password": "password123",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code, "handler returned wrong status code")

	var userCount int
	err := testDB.QueryRow("SELECT COUNT(*) FROM users WHERE email = 'integration@test.com'").Scan(&userCount)
	require.NoError(t, err, "could not query database")
	assert.Equal(t, 1, userCount, "user was not created in the database")
}
