package galaxy

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/stellaris/stellaris/server/internal/jwtctx"
	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type CreateGalaxyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateGalaxyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateGalaxyLogic {
	return &CreateGalaxyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateGalaxyLogic) CreateGalaxy(req *types.GalaxyCreateReq) (*types.GalaxyCreateResp, error) {
	userID, err := jwtctx.UserID(l.ctx)
	if err != nil {
		return nil, err
	}

	// 16 位十六进制 gid（8 字节随机）
	var gb [8]byte
	if _, err := rand.Read(gb[:]); err != nil {
		return nil, err
	}
	gid := hex.EncodeToString(gb[:])

	// 64 位十六进制 node token（32 字节随机），仅本次返回明文
	var tb [32]byte
	if _, err := rand.Read(tb[:]); err != nil {
		return nil, err
	}
	nodeToken := hex.EncodeToString(tb[:])
	hash, err := bcrypt.GenerateFromPassword([]byte(nodeToken), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	if _, err := l.svcCtx.Galaxies.Create(l.ctx, gid, req.Name, string(hash), userID); err != nil {
		return nil, err
	}
	return &types.GalaxyCreateResp{GID: gid, NodeToken: nodeToken}, nil
}
