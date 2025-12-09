package service

import (
	"context"
	"testing"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/domain"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/repository/mocks"
	"github.com/ErisSusanto19/chat-app-v2-backend/pkg/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthService_Register(t *testing.T) {
	t.Run("should register a new user successfully", func(t *testing.T) {
		// Setup
		mockUserRepo := new(mocks.MockUserRepository)

		authService := NewAuthService(mockUserRepo, "test-secret", nil)

		ctx := context.Background()
		testEmail := "test@example.com"

		// Ekspektasi
		mockUserRepo.On("GetUserByEmail", ctx, testEmail).Return(nil, nil)

		mockUserRepo.On("CreateUser", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

		// Eksekusi
		newUser, err := authService.Register(ctx, "Test User", testEmail, "password123")

		// Asersi
		assert.NoError(t, err)
		assert.NotNil(t, newUser)
		assert.Equal(t, testEmail, newUser.Email)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should return an error if email already exists", func(t *testing.T) {
		// Setup
		mockUserRepo := new(mocks.MockUserRepository)
		authService := NewAuthService(mockUserRepo, "test-secret", nil)

		ctx := context.Background()
		testEmail := "existing@example.com"
		existingUser := &domain.User{Email: testEmail}

		// Ekspektasi
		mockUserRepo.On("GetUserByEmail", ctx, testEmail).Return(existingUser, nil)

		// Eksekusi
		newUser, err := authService.Register(ctx, "Another User", testEmail, "password123")

		// Asersi
		assert.Error(t, err)
		assert.Nil(t, newUser)
		assert.Equal(t, "user with this email already exists", err.Error())

		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthService_Login(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.MockUserRepository)

	authService := NewAuthService(mockUserRepo, "test-secret-key", nil)
	ctx := context.Background()

	testEmail := "test@example.com"
	correctPassword := "password123"

	hashedPassword, _ := util.HashPassword(correctPassword)

	mockUser := &domain.User{
		ID:             uuid.New(),
		Email:          testEmail,
		HashedPassword: string(hashedPassword),
	}
	t.Run("should login successfully with correct credentials", func(t *testing.T) {
		// Ekspektasi
		mockUserRepo.On("GetUserByEmail", ctx, testEmail).Return(mockUser, nil).Once()

		// Eksekusi
		token, err := authService.Login(ctx, testEmail, correctPassword)

		// Asersi
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		mockUserRepo.AssertExpectations(t)
	})
	t.Run("should return an error for a non-existent email", func(t *testing.T) {
		nonExistentEmail := "notfound@example.com"

		// Ekspektasi
		mockUserRepo.On("GetUserByEmail", ctx, nonExistentEmail).Return(nil, nil).Once()

		// Eksekusi
		token, err := authService.Login(ctx, nonExistentEmail, "anypassword")

		// Asersi
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, "invalid email or password", err.Error())

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should return an error for an incorrect password", func(t *testing.T) {
		incorrectPassword := "wrongpassword"

		// Ekspektasi
		mockUserRepo.On("GetUserByEmail", ctx, testEmail).Return(mockUser, nil).Once()

		// Eksekusi
		token, err := authService.Login(ctx, testEmail, incorrectPassword)

		// Asersi
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, "invalid email or password", err.Error())

		mockUserRepo.AssertExpectations(t)
	})
}
