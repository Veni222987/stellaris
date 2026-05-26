package planet

import (
	"context"

	"github.com/stellaris/stellaris/server/internal/jwtctx"
	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PlanetHeartbeatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPlanetHeartbeatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlanetHeartbeatLogic {
	return &PlanetHeartbeatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PlanetHeartbeatLogic) PlanetHeartbeat(req *types.HeartbeatReq) (*types.HeartbeatResp, error) {
	planetID, err := jwtctx.PlanetID(l.ctx)
	if err != nil {
		return nil, err
	}
	if err := l.svcCtx.Planets.UpdateHeartbeat(l.ctx, planetID); err != nil {
		return nil, err
	}
	for _, a := range req.Agents {
		if _, err := l.svcCtx.Agents.Upsert(l.ctx,
			a.AgentUUID, planetID, a.Type, a.Name,
			a.Models, a.CapabilitiesJSON, a.Status); err != nil {
			return nil, err
		}
	}
	return &types.HeartbeatResp{OK: true}, nil
}
