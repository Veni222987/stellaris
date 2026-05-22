package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/stellaris/stellaris/agent/internal/adapter"
	"github.com/stellaris/stellaris/agent/internal/config"
)

type heartbeat struct {
	cfg      *config.Config
	registry *adapter.Registry
	client   *http.Client
	period   time.Duration
}

func newHeartbeat(cfg *config.Config, reg *adapter.Registry) *heartbeat {
	return &heartbeat{
		cfg:      cfg,
		registry: reg,
		client:   &http.Client{Timeout: 8 * time.Second},
		period:   5 * time.Second,
	}
}

func (h *heartbeat) run(ctx context.Context) {
	t := time.NewTicker(h.period)
	defer t.Stop()
	for {
		if err := h.send(ctx); err != nil {
			log.Printf("[heartbeat] %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

func (h *heartbeat) send(ctx context.Context) error {
	// M2: agent 已通过中心 register 落库，心跳不再上报 agents 列表。
	body, _ := json.Marshal(map[string]interface{}{
		"agents": []any{}, "cpu": 0.0, "mem": 0.0,
	})
	req, _ := http.NewRequestWithContext(ctx, "POST",
		h.cfg.CoreAddr+"/api/planet/heartbeat", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+h.cfg.PlanetJwt)
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("heartbeat http %d: %s", resp.StatusCode, string(data))
	}
	return nil
}
