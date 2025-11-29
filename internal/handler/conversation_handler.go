// File: internal/handler/conversation_handler.go
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/service"
	"github.com/google/uuid"
)

type ConversationHandler struct {
	chatService service.ChatService
}

func NewConversationHandler(chatService service.ChatService) *ConversationHandler {
	return &ConversationHandler{chatService: chatService}
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
