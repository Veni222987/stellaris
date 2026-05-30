package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
)

// ClaudeCodeConfig 描述如何启动 Claude Code CLI（Anthropic 官方 claude 命令行工具）。
type ClaudeCodeConfig struct {
	Name   string
	Binary string
	Args   []string // 额外透传的 flag（如 --model、--dangerously-skip-permissions 等）
	Env    map[string]string
}

// ClaudeCodeAdapter 使用 --print --output-format json 实现无头调用，
// 并通过 --resume <session_id> 实现原生多轮会话。
type ClaudeCodeAdapter struct {
	cfg   ClaudeCodeConfig
	store *SessionStore
}

func NewClaudeCodeAdapter(c ClaudeCodeConfig, store *SessionStore) *ClaudeCodeAdapter {
	if c.Binary == "" {
		c.Binary = "claude"
	}
	return &ClaudeCodeAdapter{cfg: c, store: store}
}

func (a *ClaudeCodeAdapter) Name() string { return a.cfg.Name }
func (a *ClaudeCodeAdapter) Type() string { return "claudecode" }
func (a *ClaudeCodeAdapter) Capabilities() Capabilities {
	return Capabilities{Streaming: true, ModelSwitch: true, ContextWindow: 200000}
}

type claudeResult struct {
	SessionID string `json:"session_id"`
	Result    string `json:"result"`
	IsError   bool   `json:"is_error"`
}

func (a *ClaudeCodeAdapter) Chat(ctx context.Context, req ChatRequest) (<-chan Chunk, error) {
	args := []string{"--print", "--output-format", "json"}
	if req.SessionUUID != "" {
		if sid, ok := a.store.Get(req.SessionUUID, a.cfg.Name); ok {
			args = append(args, "--resume", sid)
		}
	}
	// 透传用户自定义 flag（跳过与内置冲突的）
	skip := map[string]bool{"--print": true, "-p": true, "--output-format": true, "json": true, "text": true, "stream-json": true}
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

	out := make(chan Chunk, 4)
	go func() {
		defer close(out)
		// claude --output-format json 输出单个完整 JSON，读完再解析
		data, _ := io.ReadAll(stdout)
		errData, _ := io.ReadAll(stderr)
		if err := cmd.Wait(); err != nil {
			msg := err.Error()
			if len(errData) > 0 {
				msg = string(errData)
			}
			out <- Chunk{Chunk: msg, Type: "error"}
			return
		}
		var res claudeResult
		if err := json.Unmarshal(data, &res); err != nil {
			// JSON 解析失败时原样输出
			out <- Chunk{Chunk: string(data), Type: "stdout"}
			out <- Chunk{Type: "done"}
			return
		}
		if res.IsError {
			out <- Chunk{Chunk: res.Result, Type: "error"}
			return
		}
		if req.SessionUUID != "" && res.SessionID != "" {
			a.store.Set(req.SessionUUID, a.cfg.Name, res.SessionID)
		}
		if res.Result != "" {
			out <- Chunk{Chunk: res.Result, Type: "stdout"}
		}
		out <- Chunk{Type: "done"}
	}()
	return out, nil
}
