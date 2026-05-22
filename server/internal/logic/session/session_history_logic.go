package session

import (
	"context"

	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SessionHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSessionHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SessionHistoryLogic {
	return &SessionHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SessionHistoryLogic) SessionHistory(req *types.SessionHistoryReq) (*types.SessionHistoryResp, error) {
	s, err := l.svcCtx.Sessions.FindByUUID(l.ctx, req.SessionUUID)
	if err != nil {
		return nil, err
	}
	// LEFT JOIN tasks p：拿父任务的 task_uuid（orchestration / relay 模式才有）。
	rows, err := l.svcCtx.DB.Query(l.ctx, `
		SELECT t.task_uuid, t.status, a.agent_uuid, t.id,
		       p.task_uuid AS parent_task_uuid,
		       COALESCE(t.node_id, '') AS node_id
		FROM tasks t
		JOIN messages m ON m.id = t.message_id
		JOIN agents a   ON a.id = t.agent_id
		LEFT JOIN tasks p ON p.id = t.parent_task_id
		WHERE m.session_id = $1
		ORDER BY t.id ASC`, s.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resp types.SessionHistoryResp
	for rows.Next() {
		var to types.TaskOutput
		var taskID int64
		// parent_task_uuid 在没有父任务时为 NULL，用 *string 兜底。
		var parentUUID *string
		if err := rows.Scan(&to.TaskUUID, &to.Status, &to.AgentUUID, &taskID, &parentUUID, &to.NodeID); err != nil {
			return nil, err
		}
		if parentUUID != nil {
			to.ParentTaskUUID = *parentUUID
		}
		chunks, err := l.svcCtx.TaskChunks.ListByTask(l.ctx, taskID)
		if err != nil {
			return nil, err
		}
		for _, c := range chunks {
			to.Chunks = append(to.Chunks, types.TaskChunk{
				Seq: c.Seq, Type: c.ChunkType, Chunk: c.Chunk, TS: c.TS,
			})
		}
		resp.Tasks = append(resp.Tasks, to)
	}
	return &resp, nil
}
