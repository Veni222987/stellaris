package model

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Message struct {
	ID        int64
	SessionID int64
	Role      string
	Content   string
}

type MessageModel struct{ db *pgxpool.Pool }

func NewMessageModel(db *pgxpool.Pool) *MessageModel { return &MessageModel{db: db} }

func (m *MessageModel) Create(ctx context.Context, sessionID int64, role, content string) (int64, error) {
	var id int64
	err := m.db.QueryRow(ctx,
		`INSERT INTO messages (session_id, role, content) VALUES ($1, $2, $3) RETURNING id`,
		sessionID, role, content).Scan(&id)
	return id, err
}

// FindByID 按主键查 message。
func (m *MessageModel) FindByID(ctx context.Context, id int64) (*Message, error) {
	var msg Message
	err := m.db.QueryRow(ctx,
		`SELECT id, session_id, role, content FROM messages WHERE id = $1`, id).
		Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content)
	return &msg, err
}
