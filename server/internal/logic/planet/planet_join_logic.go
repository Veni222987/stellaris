package planet

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"
	"github.com/stellaris/stellaris/shared/token"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type PlanetJoinLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPlanetJoinLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlanetJoinLogic {
	return &PlanetJoinLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PlanetJoinLogic) PlanetJoin(req *types.PlanetJoinReq) (*types.PlanetJoinResp, error) {
	g, err := l.svcCtx.Galaxies.FindByGID(l.ctx, req.GID)
	if err != nil {
		return nil, errors.New("galaxy not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(g.NodeTokenHash), []byte(req.NodeToken)); err != nil {
		return nil, errors.New("invalid node token")
	}

	// 32 位十六进制 planet uuid
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return nil, err
	}
	planetUUID := hex.EncodeToString(b[:])

	planetID, err := l.svcCtx.Planets.Upsert(l.ctx, planetUUID, req.GID, req.IP, req.Hostname, req.OS)
	if err != nil {
		return nil, err
	}

	t, err := token.Issue(l.svcCtx.Config.PlanetJwt.AccessSecret,
		l.svcCtx.Config.PlanetJwt.AccessExpire,
		token.Claims{Sub: token.SubjectPlanet, ID: planetID, GID: req.GID, PlanetID: planetID})
	if err != nil {
		return nil, err
	}
	return &types.PlanetJoinResp{PlanetUUID: planetUUID, PlanetJwt: t}, nil
}
