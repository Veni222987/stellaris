package model

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Session struct {
	ID          int64
	SessionUUID string
	GID         string
	UserID      int64
	Mode        string
	AgentUUIDs  []string
	DSL         []byte // Orchestration 模式的原始 JSON；其它模式为 nil。
}

type SessionModel struct{ db *pgxpool.Pool }

func NewSessionModel(db *pgxpool.Pool) *SessionModel { return &SessionModel{db: db} }

// Create 写入会话；dsl 非 nil 时写 dsl 列（Orchestration 模式）。
func (m *SessionModel) Create(ctx context.Context, sessionUUID, gid string, userID int64,
	mode string, agentUUIDs []string, dsl []byte) (int64, error) {
	if agentUUIDs == nil {
		agentUUIDs = []string{}
	}
	auJSON, _ := json.Marshal(agentUUIDs)
	var id int64
	err := m.db.QueryRow(ctx,
		`INSERT INTO sessions (session_uuid, gid, user_id, mode, agent_uuids, dsl)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		sessionUUID, gid, userID, mode, auJSON, dsl).Scan(&id)
	return id, err
}

func (m *SessionModel) FindByUUID(ctx context.Context, sessionUUID string) (*Session, error) {
	var s Session
	var auJSON []byte
	var dsl *[]byte
	err := m.db.QueryRow(ctx,
		`SELECT id, session_uuid, gid, user_id, mode, agent_uuids, dsl
		 FROM sessions WHERE session_uuid = $1`, sessionUUID).
		Scan(&s.ID, &s.SessionUUID, &s.GID, &s.UserID, &s.Mode, &auJSON, &dsl)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(auJSON, &s.AgentUUIDs)
	if dsl != nil {
		s.DSL = *dsl
	}
	return &s, nil
}

// FindByID 按主键查 session，relay/orchestration 链路里 message → session 时使用。
func (m *SessionModel) FindByID(ctx context.Context, id int64) (*Session, error) {
	var s Session
	var auJSON []byte
	var dsl *[]byte
	err := m.db.QueryRow(ctx,
		`SELECT id, session_uuid, gid, user_id, mode, agent_uuids, dsl
		 FROM sessions WHERE id = $1`, id).
		Scan(&s.ID, &s.SessionUUID, &s.GID, &s.UserID, &s.Mode, &auJSON, &dsl)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(auJSON, &s.AgentUUIDs)
	if dsl != nil {
		s.DSL = *dsl
	}
	return &s, nil
}
