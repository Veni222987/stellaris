package galaxy

import (
	"net/http"

	"github.com/stellaris/stellaris/server/internal/logic/galaxy"
	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListGalaxyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := galaxy.NewListGalaxyLogic(r.Context(), svcCtx)
		resp, err := l.ListGalaxy()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
