package session

import (
	"context"

	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SessionTranscriptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSessionTranscriptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SessionTranscriptLogic {
	return &SessionTranscriptLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// SessionTranscript 返回会话的完整对话流：每条 user 指令一组，组内挂各 agent 的任务输出（已聚合 stdout）。
func (l *SessionTranscriptLogic) SessionTranscript(req *types.SessionTranscriptReq) (*types.SessionTranscriptResp, error) {
	s, err := l.svcCtx.Sessions.FindByUUID(l.ctx, req.SessionUUID)
	if err != nil {
		return nil, err
	}

	// 先按 message 顺序建组。
	mrows, err := l.svcCtx.DB.Query(l.ctx, `
		SELECT id, role, content, COALESCE(EXTRACT(EPOCH FROM created_at)::bigint, 0)
		FROM messages WHERE session_id = $1 ORDER BY id ASC`, s.ID)
	if err != nil {
		return nil, err
	}
	defer mrows.Close()

	messages := make([]types.TranscriptMessage, 0)
	idxByMessageID := map[int64]int{}
	for mrows.Next() {
		var m types.TranscriptMessage
		if err := mrows.Scan(&m.MessageID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.Tasks = make([]types.TranscriptTask, 0)
		idxByMessageID[m.MessageID] = len(messages)
		messages = append(messages, m)
	}
	if err := mrows.Err(); err != nil {
		return nil, err
	}

	// 再把 task 挂到对应 message；output 用 string_agg 聚合 stdout chunk。
	trows, err := l.svcCtx.DB.Query(l.ctx, `
		SELECT t.message_id, t.task_uuid, t.status, a.agent_uuid,
		       p.task_uuid AS parent_task_uuid, COALESCE(t.node_id, ''),
		       COALESCE((SELECT string_agg(tc.chunk, '' ORDER BY tc.seq)
		                 FROM task_chunks tc
		                 WHERE tc.task_id = t.id AND tc.chunk_type = 'stdout'), '')
		FROM tasks t
		JOIN agents a ON a.id = t.agent_id
		LEFT JOIN tasks p ON p.id = t.parent_task_id
		WHERE t.message_id IN (SELECT id FROM messages WHERE session_id = $1)
		ORDER BY t.id ASC`, s.ID)
	if err != nil {
		return nil, err
	}
	defer trows.Close()
	for trows.Next() {
		var messageID int64
		var t types.TranscriptTask
		var parentUUID *string
		if err := trows.Scan(&messageID, &t.TaskUUID, &t.Status, &t.AgentUUID, &parentUUID, &t.NodeID, &t.Output); err != nil {
			return nil, err
		}
		if parentUUID != nil {
			t.ParentTaskUUID = *parentUUID
		}
		if i, ok := idxByMessageID[messageID]; ok {
			messages[i].Tasks = append(messages[i].Tasks, t)
		}
	}
	if err := trows.Err(); err != nil {
		return nil, err
	}

	return &types.SessionTranscriptResp{
		SessionUUID: s.SessionUUID,
		Mode:        s.Mode,
		AgentUUIDs:  s.AgentUUIDs,
		Messages:    messages,
	}, nil
}
