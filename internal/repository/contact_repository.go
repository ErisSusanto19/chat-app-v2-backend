package repository

import (
	"context"
	"database/sql"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/domain"
	"github.com/google/uuid"
)

type ContactRepository interface {
	CreateContact(ctx context.Context, contact *domain.Contact) error
	GetContactsByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*domain.Contact, error)
}

type postgresContactRepository struct {
	db *sql.DB
}

func NewPostgresContactRepository(db *sql.DB) ContactRepository {
	return &postgresContactRepository{db: db}
}

func (r *postgresContactRepository) CreateContact(ctx context.Context, contact *domain.Contact) error {
	contact.ID = uuid.New()
	query := `
		INSERT INTO contacts (id, owner_user_id, contact_user_id, alias_name, email, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`
	_, err := r.db.ExecContext(ctx, query,
		contact.ID,
		contact.OwnerUserID,
		contact.ContactUserID,
		contact.AliasName,
		contact.Email,
		contact.Status,
	)
	return err
}

func (r *postgresContactRepository) GetContactsByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*domain.Contact, error) {
	query := `
		SELECT id, owner_user_id, contact_user_id, alias_name, email, status, created_at, updated_at
		FROM contacts
		WHERE owner_user_id = $1
		ORDER BY alias_name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []*domain.Contact
	for rows.Next() {
		var c domain.Contact
		if err := rows.Scan(&c.ID, &c.OwnerUserID, &c.ContactUserID, &c.AliasName, &c.Email, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		contacts = append(contacts, &c)
	}
	return contacts, nil
}
