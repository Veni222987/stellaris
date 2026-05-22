// Package scheduler 的 DSL 用于 Orchestration 模式：一个会话对应一张 DAG。
package scheduler

import (
	"encoding/json"
	"fmt"
)

// DSL 一个 Orchestration 会话的图；节点 ID 不可重复，depends_on 必须指向已存在的节点。
type DSL struct {
	Nodes []DSLNode `json:"nodes"`
}

// DSLNode 单个节点：绑定 agent，依赖其它节点输出，PromptTpl 支持 {{user}} 和 {{node:X}} 占位。
type DSLNode struct {
	ID        string   `json:"id"`
	AgentUUID string   `json:"agent_uuid"`
	DependsOn []string `json:"depends_on,omitempty"`
	PromptTpl string   `json:"prompt"`
}

// ParseDSL 解析 + 校验 DAG（无环、引用合法、ID 唯一非空）。
func ParseDSL(raw []byte) (*DSL, error) {
	var d DSL
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil, fmt.Errorf("dsl unmarshal: %w", err)
	}
	if len(d.Nodes) == 0 {
		return nil, fmt.Errorf("dsl has no nodes")
	}
	idSet := map[string]bool{}
	for _, n := range d.Nodes {
		if n.ID == "" {
			return nil, fmt.Errorf("dsl node has empty id")
		}
		if idSet[n.ID] {
			return nil, fmt.Errorf("dsl duplicate node id %s", n.ID)
		}
		idSet[n.ID] = true
	}
	for _, n := range d.Nodes {
		for _, dep := range n.DependsOn {
			if !idSet[dep] {
				return nil, fmt.Errorf("dsl node %s depends on unknown %s", n.ID, dep)
			}
		}
	}
	if _, err := topoSort(&d); err != nil {
		return nil, err
	}
	return &d, nil
}

// topoSort 返回拓扑顺序，存在环时返回 error。
func topoSort(d *DSL) ([]string, error) {
	inDeg := map[string]int{}
	graph := map[string][]string{}
	for _, n := range d.Nodes {
		inDeg[n.ID] = 0
	}
	for _, n := range d.Nodes {
		for _, dep := range n.DependsOn {
			graph[dep] = append(graph[dep], n.ID)
			inDeg[n.ID]++
		}
	}
	queue := []string{}
	// 用 nodes 顺序遍历保证输出稳定，避免 map 随机化。
	for _, n := range d.Nodes {
		if inDeg[n.ID] == 0 {
			queue = append(queue, n.ID)
		}
	}
	var out []string
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		out = append(out, id)
		for _, nxt := range graph[id] {
			inDeg[nxt]--
			if inDeg[nxt] == 0 {
				queue = append(queue, nxt)
			}
		}
	}
	if len(out) != len(d.Nodes) {
		return nil, fmt.Errorf("dsl has cycle")
	}
	return out, nil
}
