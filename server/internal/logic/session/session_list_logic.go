package session

import (
	"context"
	"encoding/json"

	"github.com/stellaris/stellaris/server/internal/jwtctx"
	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SessionListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSessionListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SessionListLogic {
	return &SessionListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// SessionList 列出当前用户的会话，可按 gid 过滤；title 取首条 user 指令做预览。
func (l *SessionListLogic) SessionList(req *types.SessionListReq) (*types.SessionListResp, error) {
	userID, err := jwtctx.UserID(l.ctx)
	if err != nil {
		return nil, err
	}
	rows, err := l.svcCtx.DB.Query(l.ctx, `
		SELECT s.session_uuid, s.gid, s.mode, s.agent_uuids,
		       COALESCE(EXTRACT(EPOCH FROM s.created_at)::bigint, 0),
		       COALESCE((SELECT m.content FROM messages m
		                 WHERE m.session_id = s.id AND m.role = 'user'
		                 ORDER BY m.id ASC LIMIT 1), '')
		FROM sessions s
		WHERE s.user_id = $1 AND ($2 = '' OR s.gid = $2)
		ORDER BY s.id DESC`, userID, req.GID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resp := &types.SessionListResp{Sessions: make([]types.SessionListItem, 0)}
	for rows.Next() {
		var it types.SessionListItem
		var auJSON []byte
		if err := rows.Scan(&it.SessionUUID, &it.GID, &it.Mode, &auJSON, &it.CreatedAt, &it.Title); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(auJSON, &it.AgentUUIDs)
		if it.AgentUUIDs == nil {
			it.AgentUUIDs = []string{}
		}
		resp.Sessions = append(resp.Sessions, it)
	}
	return resp, rows.Err()
}
