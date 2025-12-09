package mocks

import (
	"context"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/domain"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockConversationRepository struct {
	mock.Mock
}

func (m *MockConversationRepository) GetParticipantIDs(ctx context.Context, conversationID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, conversationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func (m *MockConversationRepository) FindPrivateConversation(ctx context.Context, userID1, userID2 uuid.UUID) (*uuid.UUID, error) {
	args := m.Called(ctx, userID1, userID2)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*uuid.UUID), args.Error(1)
}

func (m *MockConversationRepository) CreatePrivateConversation(ctx context.Context, creatorID, partnerID uuid.UUID) (*domain.Conversation, error) {
	args := m.Called(ctx, creatorID, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Conversation), args.Error(1)
}

func (m *MockConversationRepository) CreateGroupConversation(ctx context.Context, creatorID uuid.UUID, name string, participantIDs []uuid.UUID) (*domain.Conversation, error) {
	args := m.Called(ctx, creatorID, name, participantIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Conversation), args.Error(1)
}

func (m *MockConversationRepository) GetConversationPreviews(ctx context.Context, userID uuid.UUID) ([]*repository.ConversationPreview, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repository.ConversationPreview), args.Error(1)
}

func (m *MockConversationRepository) AddParticipantsToGroup(ctx context.Context, conversationID uuid.UUID, userIDs []uuid.UUID) error {
	args := m.Called(ctx, conversationID, userIDs)
	return args.Error(0)
}

func (m *MockConversationRepository) GetUserRole(ctx context.Context, userID, conversationID uuid.UUID) (string, error) {
	args := m.Called(ctx, userID, conversationID)
	return args.String(0), args.Error(1)
}

func (m *MockConversationRepository) RemoveParticipant(ctx context.Context, conversationID, userID uuid.UUID) error {
	args := m.Called(ctx, conversationID, userID)
	return args.Error(0)
}

func (m *MockConversationRepository) PromoteNewAdmin(ctx context.Context, conversationID uuid.UUID) error {
	args := m.Called(ctx, conversationID)
	return args.Error(0)
}
