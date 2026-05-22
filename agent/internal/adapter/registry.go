package adapter

import (
	"fmt"
	"sync"
)

// Registry 维护 agent_uuid → Agent 的映射，daemon 收到 MQTT task 时按 uuid 查找。
type Registry struct {
	mu     sync.RWMutex
	agents map[string]Agent
}

func NewRegistry() *Registry { return &Registry{agents: map[string]Agent{}} }

func (r *Registry) Set(uuid string, a Agent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[uuid] = a
}

func (r *Registry) Get(uuid string) (Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.agents[uuid]
	if !ok {
		return nil, fmt.Errorf("agent %s 未注册", uuid)
	}
	return a, nil
}

// Snapshot 返回当前所有已注册 agent 的拷贝，用于心跳上报。
func (r *Registry) Snapshot() map[string]Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]Agent, len(r.agents))
	for k, v := range r.agents {
		out[k] = v
	}
	return out
}
