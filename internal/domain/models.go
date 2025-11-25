package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `db:"id" json:"id"`
	Name           string    `db:"name" json:"name"`
	Email          string    `db:"email" json:"email"`
	HashedPassword string    `db:"hashed_password" json:"-"`
	PhoneNumber    *string   `db:"phone_number" json:"phone_number,omitempty"`
	Image          *string   `db:"image" json:"image,omitempty"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

type Contact struct {
	ID            uuid.UUID     `db:"id" json:"id"`
	OwnerUserID   uuid.UUID     `db:"owner_user_id" json:"-"`
	ContactUserID uuid.NullUUID `db:"contact_user_id" json:"contact_user_id"`
	AliasName     string        `db:"alias_name" json:"alias_name"`
	Email         string        `db:"email" json:"email"`
	Status        string        `db:"status" json:"status"`
	CreatedAt     time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time     `db:"updated_at" json:"updated_at"`
}

type Conversation struct {
	ID            uuid.UUID     `db:"id" json:"id"`
	IsGroup       bool          `db:"is_group" json:"is_group"`
	Name          *string       `db:"name" json:"name,omitempty"`
	Image         *string       `db:"image" json:"image,omitempty"`
	Description   *string       `db:"description" json:"description,omitempty"`
	CreatedBy     uuid.NullUUID `db:"created_by" json:"created_by"`
	LastMessageID uuid.NullUUID `db:"last_message_id" json:"-"`
	CreatedAt     time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time     `db:"updated_at" json:"updated_at"`
}

type UserConversation struct {
	ID             uuid.UUID `db:"id" json:"id"`
	UserID         uuid.UUID `db:"user_id" json:"user_id"`
	ConversationID uuid.UUID `db:"conversation_id" json:"conversation_id"`
	Role           *string   `db:"role" json:"role,omitempty"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

type Message struct {
	ID              uuid.UUID       `db:"id" json:"id"`
	SenderID        uuid.NullUUID   `db:"sender_id" json:"sender_id"`
	ConversationID  uuid.UUID       `db:"conversation_id" json:"conversation_id"`
	Status          string          `db:"status" json:"status"`
	StatusChangedAt StatusChangedAt `db:"status_changed_at" json:"status_changed_at"`
	Content         MessageContent  `db:"content" json:"content"`
	DisappearForAll bool            `db:"disappear_for_all" json:"disappear_for_all"`
	IsEdited        bool            `db:"is_edited" json:"is_edited"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`
}

type StatusChangedAt struct {
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
}

type MessageContent struct {
	Type    string  `json:"type"`
	URL     *string `json:"url,omitempty"`
	Message *string `json:"message,omitempty"`
}
