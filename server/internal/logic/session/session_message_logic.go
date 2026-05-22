package session

import (
	"context"
	"fmt"

	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"

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

	// 按 session.mode 分支：parallel 全 fan-out；relay 只派首个 agent；orchestration 走 DAG。
	var taskUUIDs []string
	switch s.Mode {
	case "relay":
		if len(s.AgentUUIDs) == 0 {
			return nil, fmt.Errorf("relay session has no agents")
		}
		tu, err := l.svcCtx.Scheduler.Dispatch(l.ctx, msgID, s.SessionUUID, s.GID, s.AgentUUIDs[0], req.Content, nil, "")
		if err != nil {
			return nil, err
		}
		taskUUIDs = append(taskUUIDs, tu)
	case "orchestration":
		tus, err := l.svcCtx.Orchestrator.StartRun(l.ctx, s, msgID, req.Content)
		if err != nil {
			return nil, err
		}
		taskUUIDs = tus
	default: // parallel / single
		for _, agentUUID := range s.AgentUUIDs {
			tu, err := l.svcCtx.Scheduler.Dispatch(l.ctx, msgID, s.SessionUUID, s.GID, agentUUID, req.Content, nil, "")
			if err != nil {
				return nil, err
			}
			taskUUIDs = append(taskUUIDs, tu)
		}
	}
	return &types.SessionMessageResp{MessageID: msgID, TaskUUIDs: taskUUIDs}, nil
}
