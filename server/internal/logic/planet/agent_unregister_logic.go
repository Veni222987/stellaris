package planet

import (
	"context"

	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AgentUnregisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAgentUnregisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AgentUnregisterLogic {
	return &AgentUnregisterLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AgentUnregisterLogic) AgentUnregister(req *types.AgentUnregisterReq) (*types.AgentUnregisterResp, error) {
	if err := l.svcCtx.Agents.DeleteByUUID(l.ctx, req.AgentUUID); err != nil {
		return nil, err
	}
	return &types.AgentUnregisterResp{OK: true}, nil
}
