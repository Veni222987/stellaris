package galaxy

import (
	"context"

	"github.com/stellaris/stellaris/server/internal/jwtctx"
	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListGalaxyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListGalaxyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListGalaxyLogic {
	return &ListGalaxyLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListGalaxyLogic) ListGalaxy() (*types.GalaxyListResp, error) {
	userID, err := jwtctx.UserID(l.ctx)
	if err != nil {
		return nil, err
	}
	rows, err := l.svcCtx.Galaxies.ListByUser(l.ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]types.GalaxyItem, 0, len(rows))
	for _, g := range rows {
		items = append(items, types.GalaxyItem{GID: g.GID, Name: g.Name})
	}
	return &types.GalaxyListResp{Galaxies: items}, nil
}
