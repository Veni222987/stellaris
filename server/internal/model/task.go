package model

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Task struct {
	ID       int64
	TaskUUID string
	Status   string
	AgentID  int64
	PlanetID int64
	NodeID   string // Orchestration DSL 节点 ID；非 Orchestration 任务为空。
}

type TaskModel struct{ db *pgxpool.Pool }

func NewTaskModel(db *pgxpool.Pool) *TaskModel { return &TaskModel{db: db} }

// Create 写一条新 task。parentTaskID = nil 表示根任务；nodeID = "" 表示非 Orchestration。
func (m *TaskModel) Create(ctx context.Context, taskUUID string, messageID, agentID, planetID int64,
	parentTaskID *int64, nodeID string) (int64, error) {
	var id int64
	err := m.db.QueryRow(ctx,
		`INSERT INTO tasks (task_uuid, message_id, agent_id, planet_id, status, parent_task_id, node_id)
		 VALUES ($1, $2, $3, $4, 'pending', $5, NULLIF($6, '')) RETURNING id`,
		taskUUID, messageID, agentID, planetID, parentTaskID, nodeID).Scan(&id)
	return id, err
}

// FindByID 根据 PK 查 task。node_id 可能为 NULL，用 COALESCE 收口为空串。
func (m *TaskModel) FindByID(ctx context.Context, id int64) (*Task, error) {
	var t Task
	err := m.db.QueryRow(ctx,
		`SELECT id, task_uuid, status, agent_id, planet_id, COALESCE(node_id, '') FROM tasks WHERE id = $1`, id).
		Scan(&t.ID, &t.TaskUUID, &t.Status, &t.AgentID, &t.PlanetID, &t.NodeID)
	return &t, err
}

// AggregateOutput 把 task 所有 chunk_type='stdout' 的 chunk 按 seq 拼成字符串。
func (m *TaskModel) AggregateOutput(ctx context.Context, taskUUID string) (string, error) {
	rows, err := m.db.Query(ctx, `
		SELECT chunk FROM task_chunks tc
		JOIN tasks t ON t.id = tc.task_id
		WHERE t.task_uuid = $1 AND tc.chunk_type = 'stdout'
		ORDER BY tc.seq`, taskUUID)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var out []byte
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return "", err
		}
		out = append(out, s...)
	}
	return string(out), rows.Err()
}

func (m *TaskModel) FindByUUID(ctx context.Context, taskUUID string) (*Task, error) {
	var t Task
	err := m.db.QueryRow(ctx,
		`SELECT id, task_uuid, status, agent_id, planet_id, COALESCE(node_id, '') FROM tasks WHERE task_uuid = $1`,
		taskUUID).Scan(&t.ID, &t.TaskUUID, &t.Status, &t.AgentID, &t.PlanetID, &t.NodeID)
	return &t, err
}

// QueryByMessage 按 message_id 拉所有 task，给 Orchestrator 计算 done / dispatched 集合用。
func (m *TaskModel) QueryByMessage(ctx context.Context, messageID int64) ([]Task, error) {
	rows, err := m.db.Query(ctx,
		`SELECT id, task_uuid, status, agent_id, planet_id, COALESCE(node_id, '')
		 FROM tasks WHERE message_id = $1`, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.TaskUUID, &t.Status, &t.AgentID, &t.PlanetID, &t.NodeID); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// UpdateStatus 更新任务状态；终态会自动写入 finished_at。
// 显式 $1::text 让 pgx/PostgreSQL 在 CASE WHEN 中也能推导出参数类型。
func (m *TaskModel) UpdateStatus(ctx context.Context, taskUUID, status string) error {
	_, err := m.db.Exec(ctx,
		`UPDATE tasks SET status = $1::text,
			started_at = COALESCE(started_at, NOW()),
			finished_at = CASE WHEN $1::text IN ('succeeded','failed') THEN NOW() ELSE finished_at END
		 WHERE task_uuid = $2`, status, taskUUID)
	return err
}

// QueryMessageID 查 task 对应的 message_id，relay 链路里要用 message → session 反查。
func (m *TaskModel) QueryMessageID(ctx context.Context, taskID int64) (int64, error) {
	var mid int64
	err := m.db.QueryRow(ctx, `SELECT message_id FROM tasks WHERE id = $1`, taskID).Scan(&mid)
	return mid, err
}

// SessionUUIDByTask 通过 task → message → session 联表反查 session_uuid。
func (m *TaskModel) SessionUUIDByTask(ctx context.Context, taskUUID string) (string, error) {
	var sessionUUID string
	err := m.db.QueryRow(ctx, `
		SELECT s.session_uuid
		FROM tasks t
		JOIN messages m ON m.id = t.message_id
		JOIN sessions s ON s.id = m.session_id
		WHERE t.task_uuid = $1`, taskUUID).Scan(&sessionUUID)
	return sessionUUID, err
}
