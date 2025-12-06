// File: internal/service/auth_service.go

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/domain"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/repository"
	"github.com/ErisSusanto19/chat-app-v2-backend/pkg/util"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, name, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, name string, phoneNumber, image *string) (*domain.User, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
}

type authService struct {
	userRepo     repository.UserRepository
	jwtSecretKey string
}

func NewAuthService(userRepo repository.UserRepository, jwtSecretKey string) AuthService {
	return &authService{
		userRepo:     userRepo,
		jwtSecretKey: jwtSecretKey,
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

func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password == "" {
		return "", errors.New("email and password cannot be empty")
	}

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("database error: %w", err)
	}
	if user == nil {
		return "", errors.New("invalid email or password")
	}

	if !util.CheckPasswordHash(password, user.HashedPassword) {
		return "", errors.New("invalid email or password")
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.jwtSecretKey))
	if err != nil {
		return "", fmt.Errorf("could not create token: %w", err)
	}

	return tokenString, nil
}

func (s *authService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *authService) UpdateProfile(ctx context.Context, userID uuid.UUID, name string, phoneNumber, image *string) (*domain.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	if name != "" {
		user.Name = name
	}
	user.PhoneNumber = phoneNumber
	user.Image = image

	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	user.HashedPassword = ""
	return user, nil
}

func (s *authService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("user not found")
	}

	if !util.CheckPasswordHash(oldPassword, user.HashedPassword) {
		return errors.New("incorrect old password")
	}

	if len(newPassword) < 8 {
		return errors.New("new password must be at least 8 characters long")
	}
	if newPassword == oldPassword {
		return errors.New("new password cannot be the same as the old password")
	}

	newHashedPassword, err := util.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(ctx, userID, newHashedPassword)
}
