package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/domain"
	"github.com/google/uuid"
)

type ContactDetail struct {
	ID            uuid.UUID     `db:"id"`
	AliasName     string        `db:"alias_name"`
	Email         string        `db:"email"`
	Status        string        `db:"status"`
	ContactUserID uuid.NullUUID `db:"contact_user_id"`
	ContactName   *string       `db:"contact_name"`
	ContactImage  *string       `db:"contact_image"`
}

type ContactRepository interface {
	CreateContact(ctx context.Context, contact *domain.Contact) error
	GetContactsByOwnerID(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*domain.Contact, error)
	UpdateContactAlias(ctx context.Context, contactID, ownerID uuid.UUID, newAliasName string) error
	DeleteContact(ctx context.Context, contactID, ownerID uuid.UUID) error
	GetContactByID(ctx context.Context, contactID, ownerID uuid.UUID) (*ContactDetail, error)
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

func (r *postgresContactRepository) GetContactsByOwnerID(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*domain.Contact, error) {
	query := `
		SELECT id, owner_user_id, contact_user_id, alias_name, email, status, created_at, updated_at
		FROM contacts
		WHERE owner_user_id = $1
		ORDER BY alias_name ASC
		LIMIT $2 OFFSET $3
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

func (r *postgresContactRepository) UpdateContactAlias(ctx context.Context, contactID, ownerID uuid.UUID, newAliasName string) error {
	query := `
		UPDATE contacts
		SET alias_name = $1, updated_at = NOW()
		WHERE id = $2 AND owner_user_id = $3
	`
	res, err := r.db.ExecContext(ctx, query, newAliasName, contactID, ownerID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("contact not found or you don't have permission to edit it")
	}

	return nil
}

func (r *postgresContactRepository) DeleteContact(ctx context.Context, contactID, ownerID uuid.UUID) error {
	query := `DELETE FROM contacts WHERE id = $1 AND owner_user_id = $2`
	res, err := r.db.ExecContext(ctx, query, contactID, ownerID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("contact not found or you don't have permission to delete it")
	}

	return nil
}

func (r *postgresContactRepository) GetContactByID(ctx context.Context, contactID, ownerID uuid.UUID) (*ContactDetail, error) {
	query := `
		SELECT
			c.id, c.alias_name, c.email, c.status,
			u.id AS contact_user_id, u.name AS contact_name, u.image_public_id AS contact_image
		FROM contacts c
		LEFT JOIN users u ON c.contact_user_id = u.id
		WHERE c.id = $1 AND c.owner_user_id = $2
	`
	var detail ContactDetail
	err := r.db.QueryRowContext(ctx, query, contactID, ownerID).Scan(
		&detail.ID, &detail.AliasName, &detail.Email, &detail.Status,
		&detail.ContactUserID, &detail.ContactName, &detail.ContactImage,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("contact not found or you don't have permission to view it")
		}
		return nil, err
	}
	return &detail, nil
}
