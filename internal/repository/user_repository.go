package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/domain"
	"github.com/ErisSusanto19/chat-app-v2-backend/pkg/util"
	"github.com/google/uuid"
)

type UserSearchResult struct {
	ID    uuid.UUID `db:"id"`
	Name  string    `db:"name"`
	Image *string   `db:"image"`
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, newHashedPassword string) error
	SearchUsers(ctx context.Context, query string, selfID uuid.UUID) ([]*UserSearchResult, error)
}

type postgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, name, email, hashed_password, phone_number, image, created_at, updated_at
		FROM users WHERE email = $1
	`

	user := &domain.User{}

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.HashedPassword,
		&user.PhoneNumber,
		&user.Image,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return user, nil
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

func (r *postgresUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, name, email, hashed_password, phone_number, image, created_at, updated_at FROM users WHERE id = $1`
	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.HashedPassword, &user.PhoneNumber,
		&user.Image, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *postgresUserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET name = $1, phone_number = $2, image = $3, updated_at = NOW()
		WHERE id = $4
	`
	res, err := r.db.ExecContext(ctx, query, user.Name, user.PhoneNumber, user.Image, user.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *postgresUserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, newHashedPassword string) error {
	query := `UPDATE users SET hashed_password = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, newHashedPassword, userID)
	return err
}

func (r *postgresUserRepository) SearchUsers(ctx context.Context, query string, selfID uuid.UUID) ([]*UserSearchResult, error) {
	searchQuery := "%" + query + "%"

	sqlQuery := `
		SELECT id, name, image
		FROM users
		WHERE (name ILIKE $1 OR email ILIKE $1) AND id != $2
		LIMIT 20 -- Batasi hasil untuk mencegah penyalahgunaan
	`

	rows, err := r.db.QueryContext(ctx, sqlQuery, searchQuery, selfID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*UserSearchResult
	for rows.Next() {
		var u UserSearchResult
		if err := rows.Scan(&u.ID, &u.Name, &u.Image); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}
