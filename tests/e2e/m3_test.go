//go:build e2e

// M3 端到端：分别跑 parallel / relay / orchestration 三种调度模式，
// 都走 POST /ws-ticket 拿一次性 ticket 再升级 WS，确认能收到对应数量的 done chunk。
//
// 三个子测试共享同一个 Core 实例（由调用方先 bin/stellaris-core 起起来），
// 但各自隔离 STELLARIS_CONFIG_DIR + galaxy，避免互相干扰。
package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestM3_Parallel(t *testing.T) {
	runM3(t, "parallel", 3, "")
}

func TestM3_Relay(t *testing.T) {
	// relay 链长 2：第一个 task 完成后由 subscriber 触发下一个，
	// 最终至少 2 个 done（链尾 task succeeded）。
	runM3(t, "relay", 2, "")
}

func TestM3_Orchestration(t *testing.T) {
	// DSL：a 节点接受 user prompt，b 节点 depends_on a，prompt 模板引用 a 的输出。
	dslTpl := `{"nodes":[` +
		`{"id":"a","agent_uuid":"%s","prompt":"{{user}}"},` +
		`{"id":"b","agent_uuid":"%s","depends_on":["a"],"prompt":"polish: {{node:a}}"}` +
		`]}`
	runM3(t, "orchestration", 2, dslTpl)
}

// runM3 跑一次完整流程：注册用户 → 建星系 → 起 daemon → 等 3 agent 注册 →
// 按 mode 构造 session create 请求 → 拿 ticket → connect WS → 发消息 → 等够 expectedDones。
//
// dslTpl 仅 orchestration 模式用，格式化时按顺序填 agent_uuid。
func runM3(t *testing.T, mode string, expectedDones int, dslTpl string) {
	t.Helper()

	cwd, _ := os.Getwd()
	cli := filepath.Join(cwd, "..", "..", "bin", "stellaris-cli")
	if _, err := os.Stat(cli); err != nil {
		t.Skipf("先 make build: %v", err)
	}
	if _, err := http.Get(coreAddrM2 + "/api/auth/register"); err != nil {
		t.Skipf("Core 没起: %v", err)
	}

	// 每次测试用全新用户 + galaxy，相互隔离
	email := fmt.Sprintf("m3-%s+%d@x.com", mode, time.Now().UnixNano())
	reg := mustPostM2(t, "/api/auth/register", "", map[string]string{"email": email, "password": "pwd"})
	tok := reg["token"].(string)
	g := mustPostM2(t, "/api/galaxy/create", tok, map[string]string{"name": "M3-" + mode})
	gid := g["gid"].(string)
	nt := g["node_token"].(string)

	uuids := startThreeMockAgents(t, cli, gid, nt, tok)

	// 按 mode 构造请求
	body := map[string]any{
		"gid":  gid,
		"mode": mode,
	}
	switch mode {
	case "parallel":
		body["agent_uuids"] = uuids
	case "relay":
		// 取前 2 个 agent 组链
		body["agent_uuids"] = uuids[:2]
	case "orchestration":
		body["agent_uuids"] = uuids[:2]
		body["dsl"] = fmt.Sprintf(dslTpl, uuids[0], uuids[1])
	default:
		t.Fatalf("unknown mode: %s", mode)
	}

	sess := mustPostM2(t, "/api/session/create", tok, body)
	su := sess["session_uuid"].(string)

	// 先 POST 拿 ticket，再用 `?ticket=` 升级 WS（M3 起的新协议）
	tkResp := mustPostM2(t, "/api/session/"+su+"/ws-ticket", tok, map[string]any{})
	wsTicket := tkResp["ticket"].(string)
	wsURL := "ws://127.0.0.1:4228/ws/session/" + su + "?ticket=" + url.QueryEscape(wsTicket)
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("ws dial: %v", err)
	}
	defer wsConn.Close()

	_ = mustPostM2(t, "/api/session/"+su+"/message", tok, map[string]string{"content": "hello"})

	wsConn.SetReadDeadline(time.Now().Add(60 * time.Second))
	dones := 0
	for dones < expectedDones {
		_, msg, err := wsConn.ReadMessage()
		if err != nil {
			t.Fatalf("ws read: %v (got %d/%d done)", err, dones, expectedDones)
		}
		var p map[string]any
		_ = json.Unmarshal(msg, &p)
		if p["type"] == "done" {
			dones++
		}
	}
	t.Logf("✓ mode=%s 收到 %d 个 done chunk", mode, dones)
}
