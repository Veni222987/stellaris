// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package session

import (
	"encoding/json"
	"net/http"

	"github.com/stellaris/stellaris/server/internal/logic/session"
	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func SessionCreateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 直接用 encoding/json 解码，跳过 go-zero httpx.Parse 对 DSL 字段必填校验。
		// 这样 types.go 可以保持 `json:"dsl,omitempty"` 而不引入 go-zero 专属 tag。
		var req types.SessionCreateReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := session.NewSessionCreateLogic(r.Context(), svcCtx)
		resp, err := l.SessionCreate(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
