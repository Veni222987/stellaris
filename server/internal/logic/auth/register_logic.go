package auth

import (
	"context"

	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"
	"github.com/stellaris/stellaris/shared/token"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (*types.RegisterResp, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	id, err := l.svcCtx.Users.Create(l.ctx, req.Email, string(hash))
	if err != nil {
		return nil, err
	}
	t, err := token.Issue(l.svcCtx.Config.UserJwt.AccessSecret,
		l.svcCtx.Config.UserJwt.AccessExpire,
		token.Claims{Sub: token.SubjectUser, ID: id})
	if err != nil {
		return nil, err
	}
	return &types.RegisterResp{Token: t}, nil
}
