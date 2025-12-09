package service

import (
	"context"
	"testing"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/domain"
	repoMocks "github.com/ErisSusanto19/chat-app-v2-backend/internal/repository/mocks"
	serviceMocks "github.com/ErisSusanto19/chat-app-v2-backend/internal/service/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestChatService_CreateGroup(t *testing.T) {
	// Setup
	ctx := context.Background()
	creatorID := uuid.New()
	participant1ID := uuid.New()
	participant2ID := uuid.New()
	groupName := "Test Group"

	t.Run("should create a group successfully with valid participants", func(t *testing.T) {
		// 1. Setup
		mockUserRepo := new(repoMocks.MockUserRepository)
		mockConvRepo := new(repoMocks.MockConversationRepository)
		mockNotifier := new(serviceMocks.MockNotifier)
		chatService := NewChatService(nil, mockConvRepo, mockUserRepo, mockNotifier)

		participantIDs := []uuid.UUID{participant1ID, participant2ID}
		expectedParticipants := append(participantIDs, creatorID)

		// Ekspektasi
		mockUserRepo.On("CountExistingUsers", ctx, mock.MatchedBy(func(ids []uuid.UUID) bool {
			if len(ids) != len(expectedParticipants) {
				return false
			}

			found := make(map[uuid.UUID]bool)
			for _, id := range ids {
				found[id] = true
			}
			return len(found) == len(expectedParticipants)
		})).Return(len(expectedParticipants), nil).Once()

		mockConversation := &domain.Conversation{ID: uuid.New(), IsGroup: true, Name: &groupName}

		mockConvRepo.On(
			"CreateGroupConversation",
			ctx,
			creatorID,
			groupName,
			mock.MatchedBy(func(ids []uuid.UUID) bool { return len(ids) == len(expectedParticipants) }),
		).Return(mockConversation, nil).Once()

		// Eksekusi
		createdGroup, err := chatService.CreateGroup(ctx, creatorID, groupName, participantIDs)

		// Asersi
		assert.NoError(t, err)
		assert.NotNil(t, createdGroup)
		assert.Equal(t, groupName, *createdGroup.Name)
		assert.True(t, createdGroup.IsGroup)

		mockUserRepo.AssertExpectations(t)
		mockConvRepo.AssertExpectations(t)
	})

	t.Run("should return an error if any participant is invalid", func(t *testing.T) {
		// 1. Setup
		mockUserRepo := new(repoMocks.MockUserRepository)
		mockConvRepo := new(repoMocks.MockConversationRepository)
		mockNotifier := new(serviceMocks.MockNotifier)
		chatService := NewChatService(nil, mockConvRepo, mockUserRepo, mockNotifier)

		participantIDs := []uuid.UUID{participant1ID, uuid.New()}
		finalParticipants := append(participantIDs, creatorID)

		// Ekspektasi
		mockUserRepo.On("CountExistingUsers", ctx, mock.Anything).Return(len(finalParticipants)-1, nil).Once()

		// Eksekusi
		createdGroup, err := chatService.CreateGroup(ctx, creatorID, groupName, participantIDs)

		// Asersi
		assert.Error(t, err)
		assert.Nil(t, createdGroup)
		assert.Equal(t, "one or more participant IDs are invalid or do not exist", err.Error())

		mockUserRepo.AssertExpectations(t)
		mockConvRepo.AssertNotCalled(t, "CreateGroupConversation", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
}
