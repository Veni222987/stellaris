package adapter

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
)

// StdioConfig 描述一个通过 fork 子进程 + stdin/stdout 交互的 Agent。
type StdioConfig struct {
	Name         string
	Type         string
	Binary       string
	Args         []string
	Env          map[string]string
	Capabilities Capabilities
}

type StdioAdapter struct{ cfg StdioConfig }

func NewStdioAdapter(cfg StdioConfig) *StdioAdapter { return &StdioAdapter{cfg: cfg} }

func (a *StdioAdapter) Name() string               { return a.cfg.Name }
func (a *StdioAdapter) Type() string               { return a.cfg.Type }
func (a *StdioAdapter) Capabilities() Capabilities { return a.cfg.Capabilities }

// Chat fork 配置中的二进制，把 prompt 写入 stdin，把 stdout 按行流式输出。
// stderr 也会以 "stderr" 类型分片转发。子进程退出无错时发 "done"，有错时发 "error"。
func (a *StdioAdapter) Chat(ctx context.Context, req ChatRequest) (<-chan Chunk, error) {
	cmd := exec.CommandContext(ctx, a.cfg.Binary, a.cfg.Args...)
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

	// 写 prompt 后关闭 stdin
	go func() {
		_, _ = io.WriteString(stdin, req.Prompt+"\n")
		_ = stdin.Close()
	}()

	out := make(chan Chunk, 16)
	go func() {
		defer close(out)
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			out <- Chunk{Chunk: scanner.Text() + "\n", Type: "stdout"}
		}
		errScanner := bufio.NewScanner(stderr)
		for errScanner.Scan() {
			out <- Chunk{Chunk: errScanner.Text() + "\n", Type: "stderr"}
		}
		if err := cmd.Wait(); err != nil {
			out <- Chunk{Chunk: err.Error(), Type: "error"}
			return
		}
		out <- Chunk{Type: "done"}
	}()
	return out, nil
}
