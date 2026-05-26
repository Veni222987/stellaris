package planet

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/stellaris/stellaris/server/internal/jwtctx"
	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AgentRegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAgentRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AgentRegisterLogic {
	return &AgentRegisterLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AgentRegisterLogic) AgentRegister(req *types.AgentRegisterReq) (*types.AgentRegisterResp, error) {
	planetID, err := jwtctx.PlanetID(l.ctx)
	if err != nil {
		return nil, err
	}

	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return nil, err
	}
	agentUUID := hex.EncodeToString(b[:])

	actualUUID, err := l.svcCtx.Agents.Upsert(l.ctx, agentUUID, planetID,
		req.Type, req.Name, req.Models, req.CapabilitiesJSON, "ready")
	if err != nil {
		return nil, err
	}
	return &types.AgentRegisterResp{AgentUUID: actualUUID}, nil
}
