package websocket

import (
	"context"
	"encoding/json"
	"log"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/domain"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/service"
	"github.com/google/uuid"
)

type HubMessage struct {
	RawMessage []byte
	Sender     *Client
}

type Hub struct {
	Clients     map[uuid.UUID]map[*Client]bool
	Broadcast   chan *HubMessage
	Register    chan *Client
	Unregister  chan *Client
	ChatService service.ChatService
}

func NewHub(chatService service.ChatService) *Hub {
	return &Hub{
		Broadcast:   make(chan *HubMessage),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Clients:     make(map[uuid.UUID]map[*Client]bool),
		ChatService: chatService,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			if _, ok := h.Clients[client.UserID]; !ok {
				h.Clients[client.UserID] = make(map[*Client]bool)
			}
			h.Clients[client.UserID][client] = true
			log.Printf("Client connected: UserID %s, RemoteAddr %s", client.UserID, client.Conn.RemoteAddr())

		case client := <-h.Unregister:
			if userClients, ok := h.Clients[client.UserID]; ok {
				if _, ok := userClients[client]; ok {
					delete(userClients, client)
					close(client.Send)

					if len(userClients) == 0 {
						delete(h.Clients, client.UserID)
					}
				}
			}
			log.Printf("Client disconnected: UserID %s, RemoteAddr %s", client.UserID, client.Conn.RemoteAddr())

		case hubMsg := <-h.Broadcast:
			h.handleIncomingMessage(hubMsg)
		}
	}
}

func (h *Hub) handleIncomingMessage(hubMsg *HubMessage) {
	var msg Message
	if err := json.Unmarshal(hubMsg.RawMessage, &msg); err != nil {
		log.Printf("Error unmarshalling message: %v", err)
		return
	}

	switch msg.Type {
	case "send_message":
		payloadBytes, _ := json.Marshal(msg.Payload)
		var payload SendMessagePayload
		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			log.Printf("Error unmarshalling send_message payload: %v", err)
			return
		}

		savedMsg, participants, err := h.ChatService.ProcessAndSaveMessage(
			context.Background(),
			hubMsg.Sender.UserID,
			payload.ConversationID,
			payload.Content,
		)
		if err != nil {
			log.Printf("Error processing message: %v", err)
			return
		}

		replyPayload := domain.Message{
			ID:             savedMsg.ID,
			SenderID:       savedMsg.SenderID,
			ConversationID: savedMsg.ConversationID,
			Content:        savedMsg.Content,
			CreatedAt:      savedMsg.CreatedAt,
		}
		replyMsg := Message{Type: "new_message", Payload: replyPayload}
		replyBytes, _ := json.Marshal(replyMsg)

		for _, participantID := range participants {
			if userClients, ok := h.Clients[participantID]; ok {
				for client := range userClients {
					client.Send <- replyBytes
				}
			}
		}

	case "message_delivered_ack":
		payloadBytes, _ := json.Marshal(msg.Payload)
		var payload DeliveredMessagePayload
		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			log.Printf("Error unmarshalling delivered_ack payload: %v", err)
			return
		}

		participants, err := h.ChatService.ProcessStatusUpdate(
			context.Background(),
			hubMsg.Sender.UserID,
			payload.ConversationID,
			[]uuid.UUID{payload.MessageID},
			"delivered",
		)
		if err != nil {
			log.Printf("Error processing delivered ack: %v", err)
			return
		}

		h.broadcastStatusUpdate(participants, payload.ConversationID, []uuid.UUID{payload.MessageID}, "delivered")

	case "start_typing":
		h.handleTypingEvent(hubMsg, msg, "user_typing")

	case "stop_typing":
		h.handleTypingEvent(hubMsg, msg, "user_stopped_typing")

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

func (h *Hub) broadcastStatusUpdate(participants []uuid.UUID, conversationID uuid.UUID, messageIDs []uuid.UUID, status string) {
	replyPayload := StatusUpdatePayload{
		ConversationID: conversationID,
		MessageIDs:     messageIDs,
		Status:         status,
	}
	replyMsg := Message{Type: "message_status_update", Payload: replyPayload}
	replyBytes, _ := json.Marshal(replyMsg)

	for _, participantID := range participants {
		if userClients, ok := h.Clients[participantID]; ok {
			for client := range userClients {
				client.Send <- replyBytes
			}
		}
	}
}

func (h *Hub) handleTypingEvent(hubMsg *HubMessage, msg Message, notificationType string) {
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload TypingPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		log.Printf("Error unmarshalling typing payload: %v", err)
		return
	}

	participants, err := h.ChatService.GetParticipantIDs(context.Background(), payload.ConversationID)
	if err != nil {
		log.Printf("Could not get participants for typing event: %v", err)
		return
	}

	notificationPayload := TypingNotificationPayload{
		ConversationID: payload.ConversationID,
		UserID:         hubMsg.Sender.UserID,
		UserName:       "Someone",
	}
	replyMsg := Message{Type: notificationType, Payload: notificationPayload}
	replyBytes, _ := json.Marshal(replyMsg)

	for _, participantID := range participants {
		if participantID == hubMsg.Sender.UserID {
			continue
		}

		if userClients, ok := h.Clients[participantID]; ok {
			for client := range userClients {
				client.Send <- replyBytes
			}
		}
	}
}

func (h *Hub) NotifyUserAddedToGroup(addedByUserID uuid.UUID, newMemberIDs []uuid.UUID, conversationID uuid.UUID) {
	participants, err := h.ChatService.GetParticipantIDs(context.Background(), conversationID)
	if err != nil {
		log.Printf("Notifier could not get participants: %v", err)
		return
	}

	for _, newMemberID := range newMemberIDs {
		for _, participantID := range participants {
			var msg Message
			if participantID == newMemberID {
				payload := YouWereAddedToGroupPayload{
					ConversationID: conversationID,
					AddedByUserID:  addedByUserID,
				}
				msg = Message{Type: "you_were_added_to_group", Payload: payload}
			} else {
				payload := UserAddedToGroupPayload{
					ConversationID: conversationID,
					AddedByUserID:  addedByUserID,
					NewMemberID:    newMemberID,
				}
				msg = Message{Type: "user_added_to_group", Payload: payload}
			}

			if userClients, ok := h.Clients[participantID]; ok {
				replyBytes, _ := json.Marshal(msg)
				for client := range userClients {
					client.Send <- replyBytes
				}
			}
		}
	}
}

func (h *Hub) SetChatService(chatService service.ChatService) {
	h.ChatService = chatService
}
