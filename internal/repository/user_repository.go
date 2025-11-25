package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/domain"
	"github.com/ErisSusanto19/chat-app-v2-backend/pkg/util"
	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	// GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
}

type postgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	user.ID = uuid.New()

	hashedPassword, err := util.HashPassword(user.HashedPassword)
	if err != nil {
		return err
	}
	user.HashedPassword = hashedPassword

	query := `
		INSERT INTO users (id, name, email, hashed_password, phone_number, image, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	now := time.Now()

	_, err = r.db.ExecContext(ctx, query,
		user.ID,
		user.Name,
		user.Email,
		user.HashedPassword,
		user.PhoneNumber,
		user.Image,
		now,
		now,
	)

	return err
}
