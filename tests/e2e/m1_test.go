//go:build e2e

// M1 端到端：register → create galaxy → orbit → start daemon → message → 拿到 OpenClaw mock 输出。
//
// 运行前置条件（由 `make e2e` 或人工准备）：
//  1. docker compose 启的 postgres/redis/emqx 都在跑
//  2. bin/stellaris-core 已构建并运行在 127.0.0.1:4228
//  3. bin/stellaris-cli 已构建
//
// 用 build tag `e2e` 隔离，平时 `go test ./...` 不会触发。
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

const coreAddr = "http://127.0.0.1:4228"

func postJSON(t *testing.T, path, token string, body any) map[string]any {
	t.Helper()
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", coreAddr+path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		t.Fatalf("POST %s → %d: %s", path, resp.StatusCode, raw)
	}
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	return m
}

func getJSON(t *testing.T, path, token string) []byte {
	t.Helper()
	req, _ := http.NewRequest("GET", coreAddr+path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		t.Fatalf("GET %s → %d: %s", path, resp.StatusCode, raw)
	}
	return raw
}

func TestM1EndToEnd(t *testing.T) {
	cwd, _ := os.Getwd()
	cli := filepath.Join(cwd, "..", "..", "bin", "stellaris-cli")
	if _, err := os.Stat(cli); err != nil {
		t.Skipf("bin/stellaris-cli 不存在，先执行 `make build`: %v", err)
	}
	if _, err := http.Get(coreAddr + "/api/auth/register"); err != nil {
		t.Skipf("Core 不在 %s 跑（先执行 make build && bin/stellaris-core &）: %v", coreAddr, err)
	}

	// 1. 注册用户
	uniqueEmail := fmt.Sprintf("e2e+%d@x.com", time.Now().UnixNano())
	reg := postJSON(t, "/api/auth/register", "", map[string]string{
		"email": uniqueEmail, "password": "pwd-12345",
	})
	userToken := reg["token"].(string)

	// 2. 创建星系
	g := postJSON(t, "/api/galaxy/create", userToken, map[string]string{"name": "E2E-Galaxy"})
	gid := g["gid"].(string)
	nodeToken := g["node_token"].(string)
	t.Logf("gid=%s", gid)

	// 3. 准备 agents.d 配置
	agentsDir := t.TempDir()
	mockBin := filepath.Join(cwd, "mock-openclaw.sh")
	yamlContent := fmt.Sprintf(`name: mock-claw
type: openclaw
binary: %s
args: []
`, mockBin)
	if err := os.WriteFile(filepath.Join(agentsDir, "claw.yaml"), []byte(yamlContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// 4. orbit + start
	cfgDir := t.TempDir()
	cliEnv := append(os.Environ(), "STELLARIS_CONFIG_DIR="+cfgDir)

	orbit := exec.Command(cli, "orbit", "127.0.0.1:4228", gid, "--token", nodeToken)
	orbit.Env = cliEnv
	if out, err := orbit.CombinedOutput(); err != nil {
		t.Fatalf("orbit failed: %v\n%s", err, out)
	}

	// 把 agents_dir 追加到 config（orbit 默认是 /etc/stellaris/agents.d）
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	cfg, _ := os.ReadFile(cfgPath)
	cfg = append(cfg, []byte(fmt.Sprintf("\nagents_dir: %s\n", agentsDir))...)
	_ = os.WriteFile(cfgPath, cfg, 0o600)

	daemon := exec.Command(cli, "start")
	daemon.Env = cliEnv
	daemonLog, err := os.CreateTemp("", "daemon-*.log")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(daemonLog.Name()) })
	daemon.Stdout = daemonLog
	daemon.Stderr = daemonLog
	if err := daemon.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = daemon.Process.Kill()
		_, _ = daemonLog.Seek(0, 0)
		buf, _ := io.ReadAll(daemonLog)
		if t.Failed() {
			t.Logf("--- daemon log ---\n%s\n--- end ---", buf)
		}
		_ = daemonLog.Close()
	})

	// 5. 等心跳让 agent 落库
	var agentUUID string
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		raw := getJSON(t, "/api/galaxy/"+gid+"/agents", userToken)
		var r struct {
			Agents []struct {
				AgentUUID string `json:"agent_uuid"`
				Type      string `json:"type"`
			} `json:"agents"`
		}
		_ = json.Unmarshal(raw, &r)
		if len(r.Agents) > 0 {
			agentUUID = r.Agents[0].AgentUUID
			t.Logf("agent_uuid=%s type=%s", agentUUID, r.Agents[0].Type)
			break
		}
		time.Sleep(1 * time.Second)
	}
	if agentUUID == "" {
		t.Fatal("等不到 agent 注册（heartbeat 没生效？）")
	}

	// 6. 建会话 + 发消息
	sess := postJSON(t, "/api/session/create", userToken, map[string]any{
		"gid": gid, "mode": "single", "agent_uuids": []string{agentUUID},
	})
	sessionUUID := sess["session_uuid"].(string)

	_ = postJSON(t, "/api/session/"+sessionUUID+"/message", userToken,
		map[string]string{"content": "ping"})

	// 7. 轮询 history，等待 task 进入 succeeded 且 chunks 里出现 OpenClaw 关键字
	deadline = time.Now().Add(30 * time.Second)
	var lastRaw []byte
	for time.Now().Before(deadline) {
		lastRaw = getJSON(t, "/api/session/"+sessionUUID+"/history", userToken)
		var hr struct {
			Tasks []struct {
				Status string `json:"status"`
				Chunks []struct {
					Chunk string `json:"chunk"`
				} `json:"chunks"`
			} `json:"tasks"`
		}
		_ = json.Unmarshal(lastRaw, &hr)
		if len(hr.Tasks) == 1 && hr.Tasks[0].Status == "succeeded" {
			var combined string
			for _, c := range hr.Tasks[0].Chunks {
				combined += c.Chunk
			}
			if bytes.Contains([]byte(combined), []byte("OpenClaw")) {
				t.Logf("✓ task succeeded，输出聚合后:\n%s", combined)
				return
			}
			t.Fatalf("task succeeded 但 chunks 不含 OpenClaw 关键字: %q", combined)
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("超时未等到 task succeeded，最后一次 history:\n%s", lastRaw)
}
