package repository

import (
	"context"
	"database/sql"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/domain"
	"github.com/google/uuid"
)

type ConversationRepository interface {
	GetParticipantIDs(ctx context.Context, conversationID uuid.UUID) ([]uuid.UUID, error)
	FindPrivateConversation(ctx context.Context, userID1, userID2 uuid.UUID) (*uuid.UUID, error)
	CreatePrivateConversation(ctx context.Context, creatorID, partnerID uuid.UUID) (*domain.Conversation, error)
}

type postgresConversationRepository struct {
	db *sql.DB
}

func NewPostgresConversationRepository(db *sql.DB) ConversationRepository {
	return &postgresConversationRepository{db: db}
}

func (r *postgresConversationRepository) GetParticipantIDs(ctx context.Context, conversationID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT user_id FROM user_conversations WHERE conversation_id = $1`
	rows, err := r.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participantIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		participantIDs = append(participantIDs, id)
	}
	return participantIDs, nil
}

func (r *postgresConversationRepository) FindPrivateConversation(ctx context.Context, userID1, userID2 uuid.UUID) (*uuid.UUID, error) {
	query := `
		SELECT uc1.conversation_id
		FROM user_conversations uc1
		JOIN user_conversations uc2 ON uc1.conversation_id = uc2.conversation_id
		JOIN conversations c ON uc1.conversation_id = c.id
		WHERE uc1.user_id = $1 AND uc2.user_id = $2 AND c.is_group = FALSE
	`
	var conversationID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, userID1, userID2).Scan(&conversationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &conversationID, nil
}

func (r *postgresConversationRepository) CreatePrivateConversation(ctx context.Context, creatorID, partnerID uuid.UUID) (*domain.Conversation, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	conv := &domain.Conversation{
		ID:      uuid.New(),
		IsGroup: false,
		CreatedBy: uuid.NullUUID{
			UUID:  creatorID,
			Valid: true,
		},
	}
	convQuery := `INSERT INTO conversations (id, is_group, created_by, created_at, updated_at)
	              VALUES ($1, $2, $3, NOW(), NOW())`
	_, err = tx.ExecContext(ctx, convQuery, conv.ID, conv.IsGroup, conv.CreatedBy)
	if err != nil {
		return nil, err
	}

	userConvQuery := `INSERT INTO user_conversations (id, user_id, conversation_id, created_at, updated_at)
	                  VALUES ($1, $2, $3, NOW(), NOW())`

	_, err = tx.ExecContext(ctx, userConvQuery, uuid.New(), creatorID, conv.ID)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, userConvQuery, uuid.New(), partnerID, conv.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return conv, nil
}
