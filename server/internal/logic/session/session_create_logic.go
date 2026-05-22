package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/stellaris/stellaris/server/internal/jwtctx"
	"github.com/stellaris/stellaris/server/internal/scheduler"
	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// 会话调度模式常量；single 是 M1/M2 旧名，等价于 parallel。
const (
	ModeParallel      = "parallel"
	ModeRelay         = "relay"
	ModeOrchestration = "orchestration"
)

var validModes = map[string]bool{
	ModeParallel: true, ModeRelay: true, ModeOrchestration: true,
	"single": true,
}

type SessionCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSessionCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SessionCreateLogic {
	return &SessionCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SessionCreateLogic) SessionCreate(req *types.SessionCreateReq) (*types.SessionCreateResp, error) {
	userID, err := jwtctx.UserID(l.ctx)
	if err != nil {
		return nil, err
	}
	if req.Mode == "" || req.Mode == "single" {
		req.Mode = ModeParallel
	}
	if !validModes[req.Mode] {
		return nil, fmt.Errorf("invalid mode: %s", req.Mode)
	}
	if len(req.AgentUUIDs) == 0 {
		return nil, errors.New("agent_uuids required")
	}

	// Orchestration 模式需要 DSL；其它模式 dslRaw 为 nil。
	var dslRaw []byte
	if req.Mode == ModeOrchestration {
		if req.DSL == "" {
			return nil, errors.New("orchestration mode requires dsl")
		}
		if _, err := scheduler.ParseDSL([]byte(req.DSL)); err != nil {
			return nil, fmt.Errorf("invalid dsl: %w", err)
		}
		dslRaw = []byte(req.DSL)
	}

	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return nil, err
	}
	suuid := hex.EncodeToString(b[:])
	if _, err := l.svcCtx.Sessions.Create(l.ctx, suuid, req.GID, userID, req.Mode, req.AgentUUIDs, dslRaw); err != nil {
		return nil, err
	}
	return &types.SessionCreateResp{SessionUUID: suuid}, nil
}
