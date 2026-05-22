package daemon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/stellaris/stellaris/agent/internal/adapter"
	"github.com/stellaris/stellaris/agent/internal/config"
)

// registrar 把每个本地 declared agent 调中心 /api/planet/agent/register
// 拿持久 agent_uuid，替代 M1 时 Planet 本地拼接的方式。
type registrar struct {
	cfg    *config.Config
	client *http.Client
}

func newRegistrar(cfg *config.Config) *registrar {
	return &registrar{cfg: cfg, client: &http.Client{Timeout: 10 * time.Second}}
}

func (r *registrar) Register(name, typ string, caps adapter.Capabilities) (string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"type": typ, "name": name,
		"models": []string{"default"},
		"capabilities": map[string]interface{}{
			"streaming":      caps.Streaming,
			"model_switch":   caps.ModelSwitch,
			"context_window": caps.ContextWindow,
		},
	})
	req, _ := http.NewRequest("POST", r.cfg.CoreAddr+"/api/planet/agent/register",
		bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+r.cfg.PlanetJwt)
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("register http %d: %s", resp.StatusCode, data)
	}
	var x struct {
		AgentUUID string `json:"agent_uuid"`
	}
	if err := json.Unmarshal(data, &x); err != nil {
		return "", err
	}
	return x.AgentUUID, nil
}
