package adapter

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
)

// OpenCodeConfig 描述如何启动 OpenCode CLI（sst/opencode 开源编码 agent）。
type OpenCodeConfig struct {
	Name   string
	Binary string
	Args   []string // 额外透传的 flag
	Env    map[string]string
}

// OpenCodeAdapter 使用 opencode run --format json 实现无头调用，
// 并通过 --session <sessionID> 实现原生多轮会话。
type OpenCodeAdapter struct {
	cfg   OpenCodeConfig
	store *SessionStore
}

func NewOpenCodeAdapter(c OpenCodeConfig, store *SessionStore) *OpenCodeAdapter {
	if c.Binary == "" {
		c.Binary = "opencode"
	}
	return &OpenCodeAdapter{cfg: c, store: store}
}

func (a *OpenCodeAdapter) Name() string { return a.cfg.Name }
func (a *OpenCodeAdapter) Type() string { return "opencode" }
func (a *OpenCodeAdapter) Capabilities() Capabilities {
	return Capabilities{Streaming: true, ModelSwitch: true, ContextWindow: 200000}
}

// opencodeEvent 覆盖 opencode --format json 输出的事件结构（字段按需提取）。
type opencodeEvent struct {
	SessionID  string `json:"sessionID"`
	Type       string `json:"type"`
	Properties struct {
		Text    string `json:"text"`
		Content string `json:"content"`
	} `json:"properties"`
}

func (a *OpenCodeAdapter) Chat(ctx context.Context, req ChatRequest) (<-chan Chunk, error) {
	args := []string{"run", "--format", "json"}
	if req.SessionUUID != "" {
		if sid, ok := a.store.Get(req.SessionUUID, a.cfg.Name); ok {
			args = append(args, "--session", sid)
		}
	}
	skip := map[string]bool{"run": true, "--format": true, "json": true, "default": true}
	for _, arg := range a.cfg.Args {
		if !skip[arg] {
			args = append(args, arg)
		}
	}

	cmd := exec.CommandContext(ctx, a.cfg.Binary, args...)
	for k, v := range a.cfg.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	go func() {
		_, _ = io.WriteString(stdin, req.Prompt+"\n")
		_ = stdin.Close()
	}()

	out := make(chan Chunk, 16)
	go func() {
		defer close(out)
		go func() { buf := make([]byte, 4096); _, _ = stderr.Read(buf) }()

		var sessionStored bool
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			var ev opencodeEvent
			if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
				continue
			}
			if !sessionStored && req.SessionUUID != "" && ev.SessionID != "" {
				a.store.Set(req.SessionUUID, a.cfg.Name, ev.SessionID)
				sessionStored = true
			}
			text := ev.Properties.Text
			if text == "" {
				text = ev.Properties.Content
			}
			if text != "" {
				out <- Chunk{Chunk: text, Type: "stdout"}
			}
		}
		if err := cmd.Wait(); err != nil {
			out <- Chunk{Chunk: err.Error(), Type: "error"}
			return
		}
		out <- Chunk{Type: "done"}
	}()
	return out, nil
}
