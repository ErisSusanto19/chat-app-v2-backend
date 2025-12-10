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
	"github.com/ErisSusanto19/chat-app-v2-backend/pkg/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
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
	router.Post("/login", authHandler.Login)

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

func TestLogin_Integration(t *testing.T) {
	router, teardown := setupTestServer()
	teardown()
	defer teardown()

	password := "strongpassword123"
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	userID := uuid.New()
	userEmail := "login@test.com"

	_, err = testDB.Exec(`INSERT INTO users (id, name, email, hashed_password) VALUES ($1, $2, $3, $4)`,
		userID, "Login Test User", userEmail, hashedPassword)
	require.NoError(t, err, "failed to insert test user")

	t.Run("should login successfully with correct credentials", func(t *testing.T) {
		requestBody := map[string]string{
			"email":    userEmail,
			"password": password,
		}
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var responseBody map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
		require.NoError(t, err)

		assert.NotEmpty(t, responseBody["token"])
	})
	t.Run("should fail with incorrect password", func(t *testing.T) {
		requestBody := map[string]string{
			"email":    userEmail,
			"password": "wrongpassword",
		}
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
