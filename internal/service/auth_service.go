// File: internal/service/auth_service.go

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/domain"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/repository"
)

type AuthService interface {
	Register(ctx context.Context, name, email, password string) (*domain.User, error)
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
	}
}

func (s *authService) Register(ctx context.Context, name, email, password string) (*domain.User, error) {

	email = strings.ToLower(strings.TrimSpace(email))
	if name == "" || email == "" {
		return nil, errors.New("name and email cannot be empty")
	}
	if len(password) < 8 {
		return nil, errors.New("password must be at least 8 characters long")
	}

	existingUser, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("error checking for existing user: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	newUser := &domain.User{
		Name:           name,
		Email:          email,
		HashedPassword: password,
	}

	err = s.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("could not create user: %w", err)
	}

	newUser.HashedPassword = ""

	return newUser, nil
}
