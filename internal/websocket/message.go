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

type DeliveredMessagePayload struct {
	MessageID      uuid.UUID `json:"message_id"`
	ConversationID uuid.UUID `json:"conversation_id"`
}

type ReadMessagesPayload struct {
	ConversationID uuid.UUID `json:"conversation_id"`
}

type StatusUpdatePayload struct {
	MessageIDs     []uuid.UUID `json:"message_ids"`
	ConversationID uuid.UUID   `json:"conversation_id"`
	Status         string      `json:"status"`
}

type TypingPayload struct {
	ConversationID uuid.UUID `json:"conversation_id"`
}

type TypingNotificationPayload struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	UserID         uuid.UUID `json:"user_id"`
	UserName       string    `json:"user_name"`
}
