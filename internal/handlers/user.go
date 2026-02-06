package repository

import (
	"context"
	"database/sql"

	"github.com/0xhop3/outils/internal/models"
)

const (
	GET_USER = `SELECT id, firebase_uid, email, display_name, created_at, updated_at
		FROM users
		WHERE firebase_uid = $1`
	CREATE_USER = `INSERT INTO users (firebase_uid, email, display_name) VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, GET_USER, firebaseUID).Scan(
		&user.ID,
		&user.FirebaseUID,
		&user.Email,
		&user.DisplayName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.QueryRowContext(ctx, CREATE_USER, user.FirebaseUID, user.Email, user.DisplayName).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}
