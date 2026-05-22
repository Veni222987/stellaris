package session

import (
	"context"

	"github.com/stellaris/stellaris/server/internal/jwtctx"
	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// SessionWSTicketLogic 颁发 WS 连接票据。
type SessionWSTicketLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSessionWSTicketLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SessionWSTicketLogic {
	return &SessionWSTicketLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// Issue 校验当前用户对 session 可见后颁发一次性 ticket。
func (l *SessionWSTicketLogic) Issue(req *types.WSTicketReq) (*types.WSTicketResp, error) {
	uid, err := jwtctx.UserID(l.ctx)
	if err != nil {
		return nil, err
	}
	// 先确保 session 存在（避免给不存在的 session 发 ticket）。
	if _, err := l.svcCtx.Sessions.FindByUUID(l.ctx, req.SessionUUID); err != nil {
		return nil, err
	}
	tk, err := l.svcCtx.WSTickets.Issue(req.SessionUUID, uid)
	if err != nil {
		return nil, err
	}
	return &types.WSTicketResp{Ticket: tk}, nil
}
