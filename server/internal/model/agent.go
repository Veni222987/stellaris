package model

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AgentRow struct {
	ID        int64
	AgentUUID string
	PlanetID  int64
	Type      string
	Name      string
}

type AgentModel struct{ db *pgxpool.Pool }

func NewAgentModel(db *pgxpool.Pool) *AgentModel { return &AgentModel{db: db} }

// Upsert 按 (planet_id, name) 去重：同一 planet 重启时复用已有 agent_uuid。
// 返回 (actualAgentUUID, error)；若行已存在则返回原有 UUID。
func (m *AgentModel) Upsert(ctx context.Context, agentUUID string, planetID int64,
	typ, name string, models []string, capabilities map[string]any, status string) (string, error) {

	if models == nil {
		models = []string{}
	}
	if capabilities == nil {
		capabilities = map[string]any{}
	}
	modelsJSON, _ := json.Marshal(models)
	capsJSON, _ := json.Marshal(capabilities)
	var actualUUID string
	err := m.db.QueryRow(ctx, `
		INSERT INTO agents (agent_uuid, planet_id, type, name, models_json, capabilities_json, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (planet_id, name) DO UPDATE SET
			type = EXCLUDED.type,
			models_json = EXCLUDED.models_json,
			capabilities_json = EXCLUDED.capabilities_json,
			status = EXCLUDED.status
		RETURNING agent_uuid`,
		agentUUID, planetID, typ, name, modelsJSON, capsJSON, status).Scan(&actualUUID)
	return actualUUID, err
}

func (m *AgentModel) FindByUUID(ctx context.Context, agentUUID string) (*AgentRow, error) {
	var a AgentRow
	err := m.db.QueryRow(ctx,
		`SELECT id, agent_uuid, planet_id, type, name FROM agents WHERE agent_uuid = $1`,
		agentUUID).Scan(&a.ID, &a.AgentUUID, &a.PlanetID, &a.Type, &a.Name)
	return &a, err
}

func (m *AgentModel) DeleteByUUID(ctx context.Context, agentUUID string) error {
	_, err := m.db.Exec(ctx, `DELETE FROM agents WHERE agent_uuid = $1`, agentUUID)
	return err
}

// UUIDByID 把 agents.id 反查为 agent_uuid，用于 relay 链中定位下一跳。
func (m *AgentModel) UUIDByID(ctx context.Context, id int64) (string, error) {
	var u string
	err := m.db.QueryRow(ctx, `SELECT agent_uuid FROM agents WHERE id = $1`, id).Scan(&u)
	return u, err
}
