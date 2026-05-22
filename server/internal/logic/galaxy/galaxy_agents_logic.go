package galaxy

import (
	"context"

	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GalaxyAgentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGalaxyAgentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GalaxyAgentsLogic {
	return &GalaxyAgentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GalaxyAgentsLogic) GalaxyAgents(req *types.GalaxyAgentsReq) (*types.GalaxyAgentsResp, error) {
	rows, err := l.svcCtx.DB.Query(l.ctx, `
		SELECT a.agent_uuid, a.type, a.name
		FROM agents a JOIN planets p ON p.id = a.planet_id
		WHERE p.gid = $1`, req.GID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var resp types.GalaxyAgentsResp
	for rows.Next() {
		var x types.GalaxyAgent
		if err := rows.Scan(&x.AgentUUID, &x.Type, &x.Name); err != nil {
			return nil, err
		}
		resp.Agents = append(resp.Agents, x)
	}
	return &resp, nil
}
