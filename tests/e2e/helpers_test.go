//go:build e2e

// e2e 测试间共享的辅助函数。
package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// startThreeMockAgents 写 oc/hm/wb 三个 mock agent 的 yaml，
// 跑 `cli orbit + start` 直到 /api/galaxy/:gid/agents 返回 >=3 条记录，
// 然后把它们的 agent_uuid 按数据库返回顺序返回。
//
// 调用方必须负责 daemon 进程的回收——已经在 t.Cleanup 注册了 Kill。
func startThreeMockAgents(t *testing.T, cli, gid, nt, tok string) []string {
	t.Helper()

	cwd, _ := os.Getwd()
	agentsDir := t.TempDir()
	write := func(name, typ, mock string) {
		yaml := fmt.Sprintf("name: %s\ntype: %s\nbinary: %s\nargs: []\n",
			name, typ, filepath.Join(cwd, mock))
		if err := os.WriteFile(filepath.Join(agentsDir, name+".yaml"), []byte(yaml), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("oc", "openclaw", "mock-openclaw.sh")
	write("hm", "hermes", "mock-hermes.sh")
	write("wb", "workbuddy", "mock-workbuddy.sh")

	// 隔离 config 目录，避免与宿主机或其它 test 互相串
	cfgDir := t.TempDir()
	env := append(os.Environ(), "STELLARIS_CONFIG_DIR="+cfgDir)

	if out, err := runEnvM2(cli, env, "orbit", "127.0.0.1:4228", gid, "--token", nt); err != nil {
		t.Fatalf("orbit: %v\n%s", err, out)
	}
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	cfg, _ := os.ReadFile(cfgPath)
	cfg = append(cfg, []byte("\nagents_dir: "+agentsDir+"\n")...)
	if err := os.WriteFile(cfgPath, cfg, 0o600); err != nil {
		t.Fatal(err)
	}

	daemon := exec.Command(cli, "start")
	daemon.Env = env
	logBuf, _ := os.CreateTemp("", "e2e-daemon-*.log")
	daemon.Stdout, daemon.Stderr = logBuf, logBuf
	if err := daemon.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = daemon.Process.Kill()
		if t.Failed() {
			_, _ = logBuf.Seek(0, 0)
			b, _ := io.ReadAll(logBuf)
			t.Logf("--- daemon log ---\n%s", b)
		}
		_ = logBuf.Close()
		_ = os.Remove(logBuf.Name())
	})

	// 轮询直到 3 个 agent 全部注册
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		raw := mustGetM2(t, "/api/galaxy/"+gid+"/agents", tok)
		var r struct {
			Agents []struct {
				AgentUUID string `json:"agent_uuid"`
				Type      string `json:"type"`
			} `json:"agents"`
		}
		_ = json.Unmarshal(raw, &r)
		if len(r.Agents) >= 3 {
			out := make([]string, 0, len(r.Agents))
			for _, a := range r.Agents {
				out = append(out, a.AgentUUID)
			}
			return out
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatal("等不到 3 个 agent 全部注册（heartbeat 没生效？）")
	return nil
}
