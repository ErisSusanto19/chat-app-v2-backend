package service

import (
	"context"
	"errors"
	"log"
	"slices"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/domain"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/repository"
	"github.com/google/uuid"
)

type ChatService interface {
	ProcessAndSaveMessage(ctx context.Context, senderID, conversationID uuid.UUID, content string) (*domain.Message, []uuid.UUID, error)
	StartOrGetPrivateConversation(ctx context.Context, creatorID, partnerID uuid.UUID) (*domain.Conversation, error)
	GetConversationsForUser(ctx context.Context, userID uuid.UUID) ([]*repository.ConversationPreview, error)
	GetMessageHistory(ctx context.Context, userID, conversationID uuid.UUID, limit, offset int) ([]*domain.Message, error)
	ProcessStatusUpdate(ctx context.Context, updaterID uuid.UUID, conversationID uuid.UUID, messageIDs []uuid.UUID, status string) ([]uuid.UUID, error)
	CreateGroup(ctx context.Context, creatorID uuid.UUID, name string, participantIDs []uuid.UUID) (*domain.Conversation, error)
	GetParticipantIDs(ctx context.Context, conversationID uuid.UUID) ([]uuid.UUID, error)
	AddGroupMembers(ctx context.Context, requesterID, conversationID uuid.UUID, newUserIDs []uuid.UUID) error
	RemoveGroupMember(ctx context.Context, requesterID, userIDToRemove, conversationID uuid.UUID) error
	LeaveGroup(ctx context.Context, userID, conversationID uuid.UUID) error
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

func (s *chatService) GetConversationsForUser(ctx context.Context, userID uuid.UUID) ([]*repository.ConversationPreview, error) {
	return s.convRepo.GetConversationPreviews(ctx, userID)
}

func (s *chatService) GetMessageHistory(ctx context.Context, userID, conversationID uuid.UUID, limit, offset int) ([]*domain.Message, error) {
	participants, err := s.convRepo.GetParticipantIDs(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if !slices.Contains(participants, userID) {
		return nil, errors.New("user is not a participant of this conversation")
	}

	return s.msgRepo.GetMessagesByConversationID(ctx, conversationID, limit, offset)
}

func (s *chatService) ProcessStatusUpdate(ctx context.Context, updaterID uuid.UUID, conversationID uuid.UUID, messageIDs []uuid.UUID, status string) ([]uuid.UUID, error) {
	participants, err := s.convRepo.GetParticipantIDs(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if !slices.Contains(participants, updaterID) {
		return nil, errors.New("user is not a participant of this conversation")
	}

	if err := s.msgRepo.UpdateMessagesStatus(ctx, messageIDs, status); err != nil {
		return nil, err
	}

	return participants, nil
}

func (s *chatService) CreateGroup(ctx context.Context, creatorID uuid.UUID, name string, participantIDs []uuid.UUID) (*domain.Conversation, error) {
	if name == "" {
		return nil, errors.New("group name cannot be empty")
	}
	if len(participantIDs) < 1 {
		return nil, errors.New("a group must have at least one participant besides the creator")
	}

	fullParticipantList := append(participantIDs, creatorID)

	uniqueParticipants := make(map[uuid.UUID]bool)
	finalParticipants := []uuid.UUID{}
	for _, id := range fullParticipantList {
		if !uniqueParticipants[id] {
			uniqueParticipants[id] = true
			finalParticipants = append(finalParticipants, id)
		}
	}

	return s.convRepo.CreateGroupConversation(ctx, creatorID, name, finalParticipants)
}

func (s *chatService) GetParticipantIDs(ctx context.Context, conversationID uuid.UUID) ([]uuid.UUID, error) {
	return s.convRepo.GetParticipantIDs(ctx, conversationID)
}

func (s *chatService) AddGroupMembers(ctx context.Context, requesterID, conversationID uuid.UUID, newUserIDs []uuid.UUID) error {

	role, err := s.convRepo.GetUserRole(ctx, requesterID, conversationID)

	if err != nil {
		return err
	}

	if role != "admin" {
		return errors.New("only admin can add members to the group")
	}

	if err := s.convRepo.AddParticipantsToGroup(ctx, conversationID, newUserIDs); err != nil {
		return err
	}

	return nil
}

func (s *chatService) RemoveGroupMember(ctx context.Context, requesterID, userIDToRemove, conversationID uuid.UUID) error {
	requesterRole, err := s.convRepo.GetUserRole(ctx, requesterID, conversationID)
	if err != nil {
		return err
	}
	if requesterRole != "admin" {
		return errors.New("only admin can remove members")
	}

	if requesterID == userIDToRemove {
		return errors.New("admin cannot remove themselves, use leave group endpoint")
	}

	targetRole, err := s.convRepo.GetUserRole(ctx, userIDToRemove, conversationID)
	if err != nil {
		return err
	}
	_ = targetRole

	return s.convRepo.RemoveParticipant(ctx, conversationID, userIDToRemove)
}

func (s *chatService) LeaveGroup(ctx context.Context, userID, conversationID uuid.UUID) error {
	role, err := s.convRepo.GetUserRole(ctx, userID, conversationID)
	if err != nil {
		return err
	}

	if err := s.convRepo.RemoveParticipant(ctx, conversationID, userID); err != nil {
		return err
	}

	if role == "admin" {
		remaining, err := s.convRepo.GetParticipantIDs(ctx, conversationID)
		if err != nil {
			log.Printf("Failed to get remaining participants after admin left: %v", err)
			return nil
		}

		if len(remaining) > 0 {
			if err := s.convRepo.PromoteNewAdmin(ctx, conversationID); err != nil {
				log.Printf("Failed to promote new admin: %v", err)
			}
		}

	}

	return nil
}
