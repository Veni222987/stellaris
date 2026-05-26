package galaxy

import (
	"context"

	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GalaxyDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGalaxyDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GalaxyDetailLogic {
	return &GalaxyDetailLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// GalaxyDetail 返回星系名 + planets（含状态/心跳）+ 每个 planet 下的 agents（含状态）。
func (l *GalaxyDetailLogic) GalaxyDetail(req *types.GalaxyDetailReq) (*types.GalaxyDetailResp, error) {
	g, err := l.svcCtx.Galaxies.FindByGID(l.ctx, req.GID)
	if err != nil {
		return nil, err
	}

	// planets 按 id 升序，planet_uuid → 列表索引便于挂 agents。
	rows, err := l.svcCtx.DB.Query(l.ctx, `
		SELECT id, planet_uuid, hostname, ip, COALESCE(os, ''), status,
		       COALESCE(EXTRACT(EPOCH FROM last_heartbeat)::bigint, 0)
		FROM planets WHERE gid = $1 ORDER BY id ASC`, req.GID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	planets := make([]types.GalaxyPlanet, 0)
	idxByPlanetID := map[int64]int{}
	for rows.Next() {
		var pid int64
		var p types.GalaxyPlanet
		if err := rows.Scan(&pid, &p.PlanetUUID, &p.Hostname, &p.IP, &p.OS, &p.Status, &p.LastHeartbeat); err != nil {
			return nil, err
		}
		p.Agents = make([]types.GalaxyDetailAgent, 0)
		idxByPlanetID[pid] = len(planets)
		planets = append(planets, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	arows, err := l.svcCtx.DB.Query(l.ctx, `
		SELECT a.planet_id, a.agent_uuid, a.type, a.name, a.status
		FROM agents a JOIN planets p ON p.id = a.planet_id
		WHERE p.gid = $1 ORDER BY a.id ASC`, req.GID)
	if err != nil {
		return nil, err
	}
	defer arows.Close()
	for arows.Next() {
		var planetID int64
		var ag types.GalaxyDetailAgent
		if err := arows.Scan(&planetID, &ag.AgentUUID, &ag.Type, &ag.Name, &ag.Status); err != nil {
			return nil, err
		}
		if i, ok := idxByPlanetID[planetID]; ok {
			planets[i].Agents = append(planets[i].Agents, ag)
		}
	}
	if err := arows.Err(); err != nil {
		return nil, err
	}

	return &types.GalaxyDetailResp{GID: g.GID, Name: g.Name, Planets: planets}, nil
}
