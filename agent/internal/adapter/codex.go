package adapter

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

// CodexConfig 描述如何启动 OpenAI Codex CLI。
type CodexConfig struct {
	Name   string
	Binary string
	Args   []string // 额外透传的 flag
	Env    map[string]string
}

// CodexAdapter 使用 codex exec <prompt> --json 实现无头调用，
// 并通过 codex exec resume <session_id> 实现原生多轮会话。
// 注意：Codex prompt 以 CLI 参数传入，不走 stdin。
type CodexAdapter struct {
	cfg   CodexConfig
	store *SessionStore
}

func NewCodexAdapter(c CodexConfig, store *SessionStore) *CodexAdapter {
	if c.Binary == "" {
		c.Binary = "codex"
	}
	return &CodexAdapter{cfg: c, store: store}
}

func (a *CodexAdapter) Name() string { return a.cfg.Name }
func (a *CodexAdapter) Type() string { return "codex" }
func (a *CodexAdapter) Capabilities() Capabilities {
	return Capabilities{Streaming: true, ModelSwitch: true, ContextWindow: 128000}
}

type codexEvent struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type codexSessionMeta struct {
	ID string `json:"id"`
}

type codexResponsePayload struct {
	Text string `json:"text"`
}

func (a *CodexAdapter) Chat(ctx context.Context, req ChatRequest) (<-chan Chunk, error) {
	var args []string
	if req.SessionUUID != "" {
		if sid, ok := a.store.Get(req.SessionUUID, a.cfg.Name); ok {
			args = []string{"exec", "resume", sid, req.Prompt, "--json"}
		} else {
			args = []string{"exec", req.Prompt, "--json"}
		}
	} else {
		args = []string{"exec", req.Prompt, "--json"}
	}
	args = append(args, a.cfg.Args...)

	cmd := exec.CommandContext(ctx, a.cfg.Binary, args...)
	for k, v := range a.cfg.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
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

	out := make(chan Chunk, 16)
	go func() {
		defer close(out)
		// 读取并丢弃 stderr（避免阻塞）
		go func() { buf := make([]byte, 4096); _, _ = stderr.Read(buf) }()

		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			var ev codexEvent
			if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
				continue
			}
			switch ev.Type {
			case "session_meta":
				if req.SessionUUID != "" {
					var meta codexSessionMeta
					if err := json.Unmarshal(ev.Payload, &meta); err == nil && meta.ID != "" {
						a.store.Set(req.SessionUUID, a.cfg.Name, meta.ID)
					}
				}
			case "response":
				var resp codexResponsePayload
				if err := json.Unmarshal(ev.Payload, &resp); err == nil && resp.Text != "" {
					out <- Chunk{Chunk: resp.Text, Type: "stdout"}
				}
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
