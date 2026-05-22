//go:build e2e

// M2 端到端：注册 → 建星系 → orbit → daemon → 三个 mock agent 全部走中心注册 →
// 同一会话挂三个 agent 发一次消息 → 期望从 WebSocket 收到 ≥3 个 done chunk。
//
// M3 起 WS 鉴权改为先 POST /ws-ticket 拿一次性 ticket，再用 `?ticket=...` 升级。
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

const coreAddrM2 = "http://127.0.0.1:4228"

func TestM2_MultiAgent_WS(t *testing.T) {
	cwd, _ := os.Getwd()
	cli := filepath.Join(cwd, "..", "..", "bin", "stellaris-cli")
	if _, err := os.Stat(cli); err != nil {
		t.Skipf("先 make build: %v", err)
	}
	if _, err := http.Get(coreAddrM2 + "/api/auth/register"); err != nil {
		t.Skipf("Core 没起: %v", err)
	}

	email := fmt.Sprintf("m2+%d@x.com", time.Now().UnixNano())
	reg := mustPostM2(t, "/api/auth/register", "", map[string]string{"email": email, "password": "pwd"})
	tok := reg["token"].(string)
	g := mustPostM2(t, "/api/galaxy/create", tok, map[string]string{"name": "M2"})
	gid := g["gid"].(string)
	nt := g["node_token"].(string)

	uuids := startThreeMockAgents(t, cli, gid, nt, tok)

	sess := mustPostM2(t, "/api/session/create", tok, map[string]any{
		"gid": gid, "mode": "single", "agent_uuids": uuids,
	})
	su := sess["session_uuid"].(string)

	// 先 POST 拿一次性 ticket，再用 `?ticket=` 升级 WS。
	tkResp := mustPostM2(t, "/api/session/"+su+"/ws-ticket", tok, map[string]any{})
	wsTicket := tkResp["ticket"].(string)
	wsURL := "ws://127.0.0.1:4228/ws/session/" + su + "?ticket=" + url.QueryEscape(wsTicket)
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("ws dial: %v", err)
	}
	defer wsConn.Close()

	_ = mustPostM2(t, "/api/session/"+su+"/message", tok, map[string]string{"content": "hi"})

	wsConn.SetReadDeadline(time.Now().Add(30 * time.Second))
	dones := 0
	for dones < 3 {
		_, msg, err := wsConn.ReadMessage()
		if err != nil {
			t.Fatalf("ws read: %v (got %d done)", err, dones)
		}
		if bytes.Contains(msg, []byte(`"type":"done"`)) {
			dones++
		}
	}
	t.Logf("✓ 收到 %d 个 done chunk", dones)
}

func mustPostM2(t *testing.T, path, token string, body any) map[string]any {
	t.Helper()
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", coreAddrM2+path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		t.Fatalf("POST %s %d: %s", path, resp.StatusCode, raw)
	}
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	return m
}

func mustGetM2(t *testing.T, path, token string) []byte {
	t.Helper()
	req, _ := http.NewRequest("GET", coreAddrM2+path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		t.Fatalf("GET %s %d: %s", path, resp.StatusCode, raw)
	}
	return raw
}

func runEnvM2(bin string, env []string, args ...string) ([]byte, error) {
	c := exec.Command(bin, args...)
	c.Env = env
	return c.CombinedOutput()
}
