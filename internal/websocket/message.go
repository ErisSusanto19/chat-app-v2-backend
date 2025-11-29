package websocket

import "github.com/google/uuid"

type Message struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

type SendMessagePayload struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	Content        string    `json:"content"`
}
