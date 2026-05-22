package model

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID           int64
	Email        string
	PasswordHash string
}

type UserModel struct{ db *pgxpool.Pool }

func NewUserModel(db *pgxpool.Pool) *UserModel { return &UserModel{db: db} }

func (m *UserModel) Create(ctx context.Context, email, passwordHash string) (int64, error) {
	var id int64
	err := m.db.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		email, passwordHash).Scan(&id)
	return id, err
}

func (m *UserModel) FindByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := m.db.QueryRow(ctx,
		`SELECT id, email, password_hash FROM users WHERE email = $1`, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return &u, err
}
