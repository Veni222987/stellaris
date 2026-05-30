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

// HistoryRow 携带一条消息及其 agent 聚合输出（stdout）。
type HistoryRow struct {
	Role    string
	Content string
	Output  string // 聚合 stdout；空字符串表示 assistant 尚无输出
}

// LoadHistory 返回 session 内 id < beforeMsgID 的所有消息及其 stdout 聚合。
// 结果按消息 id 升序排列，用于构造多轮对话上下文。
func (m *MessageModel) LoadHistory(ctx context.Context, sessionID, beforeMsgID int64) ([]HistoryRow, error) {
	rows, err := m.db.Query(ctx, `
		SELECT m.role, m.content,
		       COALESCE((SELECT string_agg(tc.chunk, '' ORDER BY tc.seq)
		                 FROM task_chunks tc
		                 JOIN tasks t2 ON t2.id = tc.task_id
		                 WHERE t2.message_id = m.id AND tc.chunk_type = 'stdout'), '')
		FROM messages m
		WHERE m.session_id = $1 AND m.id < $2
		ORDER BY m.id ASC`, sessionID, beforeMsgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []HistoryRow
	for rows.Next() {
		var r HistoryRow
		if err := rows.Scan(&r.Role, &r.Content, &r.Output); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

// FindByID 按主键查 message。
func (m *MessageModel) FindByID(ctx context.Context, id int64) (*Message, error) {
	var msg Message
	err := m.db.QueryRow(ctx,
		`SELECT id, session_id, role, content FROM messages WHERE id = $1`, id).
		Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content)
	return &msg, err
}
