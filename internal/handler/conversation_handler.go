// File: internal/handler/conversation_handler.go
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/domain"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ConversationHandler struct {
	chatService service.ChatService
}

type conversationPreviewResponse struct {
	ID                   uuid.UUID  `json:"id"`
	IsGroup              bool       `json:"is_group"`
	Name                 string     `json:"name"`
	Image                *string    `json:"image,omitempty"`
	LastMessage          *string    `json:"last_message,omitempty"`
	LastMessageTimestamp *time.Time `json:"last_message_timestamp,omitempty"`
}

func NewConversationHandler(chatService service.ChatService) *ConversationHandler {
	return &ConversationHandler{chatService: chatService}
}

func (h *ConversationHandler) GetConversations(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserContextKey).(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	previews, err := h.chatService.GetConversationsForUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	responses := make([]conversationPreviewResponse, 0, len(previews))
	for _, p := range previews {
		responses = append(responses, conversationPreviewResponse{
			ID:                   p.ID,
			IsGroup:              p.IsGroup,
			Name:                 p.Name,
			Image:                p.Image,
			LastMessage:          p.LastMessage,
			LastMessageTimestamp: p.LastMessageTimestamp,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responses)
}

type startConversationRequest struct {
	PartnerID string `json:"partner_id"`
}

func (h *ConversationHandler) StartPrivateConversation(w http.ResponseWriter, r *http.Request) {
	creatorID, ok := r.Context().Value(UserContextKey).(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req startConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	partnerID, err := uuid.Parse(req.PartnerID)
	if err != nil {
		http.Error(w, "Invalid partner ID format", http.StatusBadRequest)
		return
	}

	conversation, err := h.chatService.StartOrGetPrivateConversation(r.Context(), creatorID, partnerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(conversation)
}

func (h *ConversationHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserContextKey).(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conversationIDStr := chi.URLParam(r, "conversationID")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		http.Error(w, "Invalid conversation ID format", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	offsetStr := r.URL.Query().Get("offset")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	messages, err := h.chatService.GetMessageHistory(r.Context(), userID, conversationID, limit, offset)
	if err != nil {
		if err.Error() == "user is not a participant of this conversation" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if messages == nil {
		messages = []*domain.Message{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(messages)
}
