package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/pathvar"
)

var upgrader = websocket.Upgrader{
	// M3 简化：放行所有 origin；生产应收紧到白名单或同源。
	CheckOrigin: func(r *http.Request) bool { return true },
}

// SessionWSHandler 处理 /ws/session/:session_uuid 升级请求。
// 浏览器原生 WebSocket 不能加自定义 Header，因此鉴权先走 POST /ws-ticket 拿一次性 ticket。
func SessionWSHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tk := r.URL.Query().Get("ticket")
		if tk == "" {
			http.Error(w, "missing ticket", http.StatusUnauthorized)
			return
		}

		// go-zero 路由的路径参数不通过 r.PathValue 暴露，需要从 pathvar 上下文取。
		sessionUUID := pathvar.Vars(r)["session_uuid"]
		if sessionUUID == "" {
			http.Error(w, "missing session_uuid", http.StatusBadRequest)
			return
		}

		if _, ok := svcCtx.WSTickets.Consume(tk, sessionUUID); !ok {
			http.Error(w, "invalid or expired ticket", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Errorf("[ws] upgrade: %v", err)
			return
		}
		svcCtx.WSHub.Add(sessionUUID, conn)
		defer svcCtx.WSHub.Remove(sessionUUID, conn)
		defer conn.Close()

		// 客户端通常不发消息，但需要持续 ReadMessage 才能感知 Close 帧。
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}
}
