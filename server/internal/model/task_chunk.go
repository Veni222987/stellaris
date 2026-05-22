package model

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskChunkRow struct {
	Seq       int
	ChunkType string
	Chunk     string
	TS        int64 // 毫秒
}

type TaskChunkModel struct{ db *pgxpool.Pool }

func NewTaskChunkModel(db *pgxpool.Pool) *TaskChunkModel { return &TaskChunkModel{db: db} }

// Insert 写入单个分片。task_id+seq 是唯一键，重复插入会被忽略（保证 at-least-once 投递时幂等）。
func (m *TaskChunkModel) Insert(ctx context.Context, taskID int64, seq int,
	chunkType, chunk string) error {
	_, err := m.db.Exec(ctx,
		`INSERT INTO task_chunks (task_id, seq, chunk, chunk_type)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (task_id, seq) DO NOTHING`,
		taskID, seq, chunk, chunkType)
	return err
}

func (m *TaskChunkModel) ListByTask(ctx context.Context, taskID int64) ([]TaskChunkRow, error) {
	rows, err := m.db.Query(ctx,
		`SELECT seq, chunk_type, chunk, EXTRACT(EPOCH FROM ts) * 1000
		 FROM task_chunks WHERE task_id = $1 ORDER BY seq ASC`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TaskChunkRow
	for rows.Next() {
		var r TaskChunkRow
		var tsFloat float64
		if err := rows.Scan(&r.Seq, &r.ChunkType, &r.Chunk, &tsFloat); err != nil {
			return nil, err
		}
		r.TS = int64(tsFloat)
		out = append(out, r)
	}
	return out, nil
}
