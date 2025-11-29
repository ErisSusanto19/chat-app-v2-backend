package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type ConversationRepository interface {
	GetParticipantIDs(ctx context.Context, conversationID uuid.UUID) ([]uuid.UUID, error)
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
