package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
)

// GeminiConfig 描述如何启动 Google Gemini CLI。
type GeminiConfig struct {
	Name   string
	Binary string
	Args   []string // 额外透传的 flag（如 --model 等）
	Env    map[string]string
}

// GeminiAdapter 使用 --prompt --output-format json 实现无头调用，
// 并通过 --resume <session_id> 实现原生多轮会话。
// Gemini JSON 输出中 session_id 为 UUID v4，可跨调用稳定恢复会话。
type GeminiAdapter struct {
	cfg   GeminiConfig
	store *SessionStore
}

func NewGeminiAdapter(c GeminiConfig, store *SessionStore) *GeminiAdapter {
	if c.Binary == "" {
		c.Binary = "gemini"
	}
	return &GeminiAdapter{cfg: c, store: store}
}

func (a *GeminiAdapter) Name() string { return a.cfg.Name }
func (a *GeminiAdapter) Type() string { return "gemini" }
func (a *GeminiAdapter) Capabilities() Capabilities {
	return Capabilities{Streaming: true, ModelSwitch: true, ContextWindow: 1000000}
}

type geminiResult struct {
	SessionID string `json:"session_id"`
	Response  string `json:"response"`
}

func (a *GeminiAdapter) Chat(ctx context.Context, req ChatRequest) (<-chan Chunk, error) {
	args := []string{"--prompt", "--output-format", "json"}
	if req.SessionUUID != "" {
		if sid, ok := a.store.Get(req.SessionUUID, a.cfg.Name); ok {
			args = append(args, "--resume", sid)
		}
	}
	skip := map[string]bool{"--prompt": true, "-p": true, "--output-format": true, "json": true, "text": true, "stream-json": true}
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
		var res geminiResult
		if err := json.Unmarshal(data, &res); err != nil {
			out <- Chunk{Chunk: string(data), Type: "stdout"}
			out <- Chunk{Type: "done"}
			return
		}
		if req.SessionUUID != "" && res.SessionID != "" {
			a.store.Set(req.SessionUUID, a.cfg.Name, res.SessionID)
		}
		if res.Response != "" {
			out <- Chunk{Chunk: res.Response, Type: "stdout"}
		}
		out <- Chunk{Type: "done"}
	}()
	return out, nil
}
