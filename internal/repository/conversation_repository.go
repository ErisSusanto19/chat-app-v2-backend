package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/domain"
	"github.com/google/uuid"
)

type ConversationPreview struct {
	ID                   uuid.UUID
	IsGroup              bool
	Name                 string
	Image                *string
	LastMessage          *string
	LastMessageTimestamp *time.Time
}

type ConversationRepository interface {
	GetParticipantIDs(ctx context.Context, conversationID uuid.UUID) ([]uuid.UUID, error)
	FindPrivateConversation(ctx context.Context, userID1, userID2 uuid.UUID) (*uuid.UUID, error)
	CreatePrivateConversation(ctx context.Context, creatorID, partnerID uuid.UUID) (*domain.Conversation, error)
	GetConversationPreviews(ctx context.Context, userID uuid.UUID) ([]*ConversationPreview, error)
	CreateGroupConversation(ctx context.Context, creatorID uuid.UUID, name string, participantIDs []uuid.UUID) (*domain.Conversation, error)
	AddParticipantsToGroup(ctx context.Context, conversationID uuid.UUID, userIDs []uuid.UUID) error
	GetUserRole(ctx context.Context, userID, conversationID uuid.UUID) (string, error)
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

func (r *postgresConversationRepository) GetConversationPreviews(ctx context.Context, userID uuid.UUID) ([]*ConversationPreview, error) {
	// query yang kompleks:
	// 1. mulai dari user_conversations (uc1) untuk menemukan semua percakapan milik userID.
	// 2. JOIN dengan conversations (c) untuk mendapatkan detail dasar.
	// 3. LEFT JOIN dengan messages (m) pada last_message_id untuk mendapatkan pesan terakhir. LEFT JOIN penting karena percakapan baru mungkin belum punya pesan.
	// 4. menemukan partner dengan LEFT JOIN lagi ke user_conversations (uc2) dengan kondisi `uc2.user_id != $1` untuk menemukan baris milik partner.
	// 5. LEFT JOIN ke users (p) untuk mendapatkan detail partner.
	// 6. COALESCE(c.name, p.name) adalah trik SQL: jika c.name (nama grup) ada, gunakan itu. Jika tidak, gunakan p.name (nama partner).
	// 7. ORDER BY m.created_at DESC NULLS LAST: Urutkan berdasarkan pesan terbaru. Percakapan tanpa pesan diletakkan di akhir.
	query := `
		SELECT
			c.id,
			c.is_group,
			COALESCE(c.name, p.name) AS conversation_name,
			COALESCE(c.image, p.image) AS conversation_image,
			m.content ->> 'message' AS last_message,
			m.created_at AS last_message_timestamp
		FROM user_conversations uc1
		JOIN conversations c ON uc1.conversation_id = c.id
		LEFT JOIN messages m ON c.last_message_id = m.id
		LEFT JOIN user_conversations uc2 ON c.id = uc2.conversation_id AND uc2.user_id != $1
		LEFT JOIN users p ON uc2.user_id = p.id AND c.is_group = FALSE
		WHERE uc1.user_id = $1
		ORDER BY m.created_at DESC NULLS LAST
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var previews []*ConversationPreview
	for rows.Next() {
		p := &ConversationPreview{}
		var lastMessage sql.NullString
		var lastMessageTimestamp sql.NullTime

		if err := rows.Scan(&p.ID, &p.IsGroup, &p.Name, &p.Image, &lastMessage, &lastMessageTimestamp); err != nil {
			return nil, err
		}

		if lastMessage.Valid {
			p.LastMessage = &lastMessage.String
		}
		if lastMessageTimestamp.Valid {
			p.LastMessageTimestamp = &lastMessageTimestamp.Time
		}

		previews = append(previews, p)
	}

	return previews, nil
}

func (r *postgresConversationRepository) CreateGroupConversation(ctx context.Context, creatorID uuid.UUID, name string, participantIDs []uuid.UUID) (*domain.Conversation, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	conv := &domain.Conversation{
		ID:        uuid.New(),
		IsGroup:   true,
		Name:      &name,
		CreatedBy: uuid.NullUUID{UUID: creatorID, Valid: true},
	}
	convQuery := `INSERT INTO conversations (id, is_group, name, created_by, created_at, updated_at)
	              VALUES ($1, $2, $3, $4, NOW(), NOW())`
	_, err = tx.ExecContext(ctx, convQuery, conv.ID, conv.IsGroup, conv.Name, conv.CreatedBy)
	if err != nil {
		return nil, err
	}

	userConvQuery := `INSERT INTO user_conversations (id, user_id, conversation_id, role, created_at, updated_at)
	                  VALUES ($1, $2, $3, $4, NOW(), NOW())`

	stmt, err := tx.PrepareContext(ctx, userConvQuery)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for _, participantID := range participantIDs {
		var role string
		if participantID == creatorID {
			role = "admin"
		} else {
			role = "member"
		}

		_, err := stmt.ExecContext(ctx, uuid.New(), participantID, conv.ID, role)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return conv, nil
}

func (r *postgresConversationRepository) AddParticipantsToGroup(ctx context.Context, conversationID uuid.UUID, userIDs []uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO user_conversations (id, user_id, conversation_id, role, created_at, updated_at)
	          VALUES ($1, $2, $3, 'member', NOW(), NOW())
			  ON CONFLICT (user_id, conversation_id) DO NOTHING`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, userID := range userIDs {
		_, err := stmt.ExecContext(ctx, uuid.New(), userID, conversationID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *postgresConversationRepository) GetUserRole(ctx context.Context, userID, conversationID uuid.UUID) (string, error) {
	var role sql.NullString
	query := `SELECT role FROM user_conversations WHERE user_id = $1 AND conversation_id = $2`
	err := r.db.QueryRowContext(ctx, query, userID, conversationID).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("user is not a member of this conversation")
		}
		return "", err
	}
	if !role.Valid {
		return "member", nil
	}
	return role.String, nil
}
