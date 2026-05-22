// Package adapter 定义 Stellaris Planet 端把异构 Agent 接入统一接口的抽象。
//
// 所有 Agent 实现都通过同一个 Agent 接口暴露，Daemon 根据 MQTT 任务的 agent_uuid
// 路由到具体 Adapter 实例并调 Chat 拿流式输出。
package adapter

import "context"

type ChatRequest struct {
	Prompt string
	Model  string
}

type Chunk struct {
	Chunk string
	Type  string // stdout / stderr / done / error
}

type Capabilities struct {
	Streaming     bool `json:"streaming"`
	ModelSwitch   bool `json:"model_switch"`
	ContextWindow int  `json:"context_window"`
}

type Agent interface {
	Name() string
	Type() string
	Capabilities() Capabilities
	Chat(ctx context.Context, req ChatRequest) (<-chan Chunk, error)
}
