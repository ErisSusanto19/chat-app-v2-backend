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
