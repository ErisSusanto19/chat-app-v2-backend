// File: internal/service/chat_service.go
package service

import (
	"context"
	"errors"
	"slices"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/domain"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/repository"
	"github.com/google/uuid"
)

type ChatService interface {
	ProcessAndSaveMessage(ctx context.Context, senderID, conversationID uuid.UUID, content string) (*domain.Message, []uuid.UUID, error)
	StartOrGetPrivateConversation(ctx context.Context, creatorID, partnerID uuid.UUID) (*domain.Conversation, error)
}

type chatService struct {
	msgRepo  repository.MessageRepository
	convRepo repository.ConversationRepository
}

func NewChatService(msgRepo repository.MessageRepository, convRepo repository.ConversationRepository) ChatService {
	return &chatService{msgRepo: msgRepo, convRepo: convRepo}
}

func (s *chatService) ProcessAndSaveMessage(ctx context.Context, senderID, conversationID uuid.UUID, content string) (*domain.Message, []uuid.UUID, error) {
	participants, err := s.convRepo.GetParticipantIDs(ctx, conversationID)
	if err != nil {
		return nil, nil, err
	}
	if len(participants) == 0 {
		return nil, nil, errors.New("conversation not found or has no participants")
	}

	if !slices.Contains(participants, senderID) {
		return nil, nil, errors.New("user is not a participant of this conversation")
	}

	newMessage := &domain.Message{
		SenderID:       uuid.NullUUID{UUID: senderID, Valid: true},
		ConversationID: conversationID,
		Content: domain.MessageContent{
			Type:    "text",
			Message: &content,
		},
	}

	if err := s.msgRepo.CreateMessage(ctx, newMessage); err != nil {
		return nil, nil, err
	}

	return newMessage, participants, nil
}

func (s *chatService) StartOrGetPrivateConversation(ctx context.Context, creatorID, partnerID uuid.UUID) (*domain.Conversation, error) {
	if creatorID == partnerID {
		return nil, errors.New("cannot start a conversation with yourself")
	}

	existingConvID, err := s.convRepo.FindPrivateConversation(ctx, creatorID, partnerID)
	if err != nil {
		return nil, err
	}

	if existingConvID != nil {
		return &domain.Conversation{ID: *existingConvID}, nil
	}

	return s.convRepo.CreatePrivateConversation(ctx, creatorID, partnerID)
}
