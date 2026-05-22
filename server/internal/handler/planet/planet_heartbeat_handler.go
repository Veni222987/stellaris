// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package planet

import (
	"net/http"

	"github.com/stellaris/stellaris/server/internal/logic/planet"
	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func PlanetHeartbeatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.HeartbeatReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := planet.NewPlanetHeartbeatLogic(r.Context(), svcCtx)
		resp, err := l.PlanetHeartbeat(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
