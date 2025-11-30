// File: internal/repository/message_repository.go
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/domain"
	"github.com/google/uuid"
)

type MessageRepository interface {
	CreateMessage(ctx context.Context, message *domain.Message) error
	GetMessagesByConversationID(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]*domain.Message, error)
	UpdateMessagesStatus(ctx context.Context, messageIDs []uuid.UUID, status string) error
}

type postgresMessageRepository struct {
	db *sql.DB
}

func NewPostgresMessageRepository(db *sql.DB) MessageRepository {
	return &postgresMessageRepository{db: db}
}

func (r *postgresMessageRepository) CreateMessage(ctx context.Context, message *domain.Message) error {
	message.ID = uuid.New()
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()

	contentJSON, err := json.Marshal(message.Content)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	msgQuery := `INSERT INTO messages (id, sender_id, conversation_id, content, created_at, updated_at)
               VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = tx.ExecContext(ctx, msgQuery, message.ID, message.SenderID, message.ConversationID, contentJSON, message.CreatedAt, message.UpdatedAt)
	if err != nil {
		return err
	}

	convQuery := `UPDATE conversations SET last_message_id = $1, updated_at = $2 WHERE id = $3`
	_, err = tx.ExecContext(ctx, convQuery, message.ID, message.UpdatedAt, message.ConversationID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *postgresMessageRepository) GetMessagesByConversationID(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]*domain.Message, error) {
	query := `
		SELECT id, sender_id, conversation_id, status, status_changed_at, content,
		       disappear_for_all, is_edited, created_at, updated_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at DESC
		LIMIT $2
		OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, conversationID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		var msg domain.Message
		var contentJSON []byte
		var statusChangedAtJSON []byte

		err := rows.Scan(
			&msg.ID, &msg.SenderID, &msg.ConversationID, &msg.Status, &statusChangedAtJSON,
			&contentJSON, &msg.DisappearForAll, &msg.IsEdited, &msg.CreatedAt, &msg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if contentJSON != nil {
			if err := json.Unmarshal(contentJSON, &msg.Content); err != nil {
				return nil, err
			}
		}
		if statusChangedAtJSON != nil {
			if err := json.Unmarshal(statusChangedAtJSON, &msg.StatusChangedAt); err != nil {
				return nil, err
			}
		}

		messages = append(messages, &msg)
	}

	return messages, nil
}

func (r *postgresMessageRepository) UpdateMessagesStatus(ctx context.Context, messageIDs []uuid.UUID, status string) error {
	statusChangeField := "delivered_at"
	if status == "read" {
		statusChangeField = "read_at"
	}

	jsonbUpdateQuery := "jsonb_build_object($1, to_jsonb(NOW()))"

	query := `
		UPDATE messages
		SET 
			status = $2,
			status_changed_at = COALESCE(status_changed_at, '{}'::jsonb) || ` + jsonbUpdateQuery + `
		WHERE id = ANY($3) AND status != 'read' -- Hanya update jika belum dibaca
	`

	_, err := r.db.ExecContext(ctx, query, statusChangeField, status, messageIDs)
	return err
}
