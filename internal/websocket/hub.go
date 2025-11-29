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

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}
