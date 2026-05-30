package session

import (
	"context"
	"fmt"

	"github.com/stellaris/stellaris/server/internal/model"
	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"
	"github.com/stellaris/stellaris/shared/protocol"

	"github.com/zeromicro/go-zero/core/logx"
)

type SessionMessageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSessionMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SessionMessageLogic {
	return &SessionMessageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SessionMessageLogic) SessionMessage(req *types.SessionMessageReq) (*types.SessionMessageResp, error) {
	s, err := l.svcCtx.Sessions.FindByUUID(l.ctx, req.SessionUUID)
	if err != nil {
		return nil, err
	}
	msgID, err := l.svcCtx.Messages.Create(l.ctx, s.ID, "user", req.Content)
	if err != nil {
		return nil, err
	}

	// 加载当前消息之前的会话历史，供 Agent 建立多轮对话上下文。
	histRows, err := l.svcCtx.Messages.LoadHistory(l.ctx, s.ID, msgID)
	if err != nil {
		return nil, err
	}
	history := buildHistory(histRows)

	// 按 session.mode 分支：parallel 全 fan-out；relay 只派首个 agent；orchestration 走 DAG。
	var taskUUIDs []string
	switch s.Mode {
	case "relay":
		if len(s.AgentUUIDs) == 0 {
			return nil, fmt.Errorf("relay session has no agents")
		}
		tu, err := l.svcCtx.Scheduler.Dispatch(l.ctx, msgID, s.SessionUUID, s.GID, s.AgentUUIDs[0], req.Content, history, nil, "")
		if err != nil {
			return nil, err
		}
		taskUUIDs = append(taskUUIDs, tu)
	case "orchestration":
		tus, err := l.svcCtx.Orchestrator.StartRun(l.ctx, s, msgID, req.Content, history)
		if err != nil {
			return nil, err
		}
		taskUUIDs = tus
	default: // parallel / single
		for _, agentUUID := range s.AgentUUIDs {
			tu, err := l.svcCtx.Scheduler.Dispatch(l.ctx, msgID, s.SessionUUID, s.GID, agentUUID, req.Content, history, nil, "")
			if err != nil {
				return nil, err
			}
			taskUUIDs = append(taskUUIDs, tu)
		}
	}
	return &types.SessionMessageResp{MessageID: msgID, TaskUUIDs: taskUUIDs}, nil
}

// buildHistory 将 LoadHistory 返回的行转换为 protocol.HistoryEntry 列表。
// 每条 user 消息生成一条 user 条目；若该消息有 agent 输出，紧随一条 assistant 条目。
func buildHistory(rows []model.HistoryRow) []protocol.HistoryEntry {
	var out []protocol.HistoryEntry
	for _, r := range rows {
		out = append(out, protocol.HistoryEntry{Role: r.Role, Content: r.Content})
		if r.Role == "user" && r.Output != "" {
			out = append(out, protocol.HistoryEntry{Role: "assistant", Content: r.Output})
		}
	}
	return out
}
