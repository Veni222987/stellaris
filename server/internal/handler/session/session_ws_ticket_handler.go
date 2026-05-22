// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package session

import (
	"net/http"

	"github.com/stellaris/stellaris/server/internal/logic/session"
	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// SessionWSTicketHandler 处理 POST /api/session/:session_uuid/ws-ticket。
func SessionWSTicketHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.WSTicketReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := session.NewSessionWSTicketLogic(r.Context(), svcCtx)
		resp, err := l.Issue(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
