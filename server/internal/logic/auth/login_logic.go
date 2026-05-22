package auth

import (
	"context"
	"errors"

	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"
	"github.com/stellaris/stellaris/shared/token"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (*types.LoginResp, error) {
	u, err := l.svcCtx.Users.FindByEmail(l.ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	t, err := token.Issue(l.svcCtx.Config.UserJwt.AccessSecret,
		l.svcCtx.Config.UserJwt.AccessExpire,
		token.Claims{Sub: token.SubjectUser, ID: u.ID})
	if err != nil {
		return nil, err
	}
	return &types.LoginResp{Token: t}, nil
}
