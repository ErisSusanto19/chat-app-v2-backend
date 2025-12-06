// File: internal/service/user_service.go
package service

import (
	"context"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/repository"
	"github.com/google/uuid"
)

type UserService interface {
	SearchUsers(ctx context.Context, query string, selfID uuid.UUID) ([]*repository.UserSearchResult, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) SearchUsers(ctx context.Context, query string, selfID uuid.UUID) ([]*repository.UserSearchResult, error) {
	if len(query) < 2 {
		return []*repository.UserSearchResult{}, nil
	}
	return s.userRepo.SearchUsers(ctx, query, selfID)
}
