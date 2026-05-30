package adapter

// CodexConfig 描述如何启动 OpenAI Codex CLI。
// 默认以 --quiet 标志运行：抑制交互式提示，从 stdin 读取 prompt 并流式输出到 stdout。
type CodexConfig struct {
	Name   string
	Binary string
	Args   []string
	Env    map[string]string
}

// NewCodexAdapter 根据 CodexConfig 构造一个 StdioAdapter，
// 默认 binary=codex、args=[--quiet]。
func NewCodexAdapter(c CodexConfig) *StdioAdapter {
	if c.Binary == "" {
		c.Binary = "codex"
	}
	if len(c.Args) == 0 {
		c.Args = []string{"--quiet"}
	}
	return NewStdioAdapter(StdioConfig{
		Name:   c.Name,
		Type:   "codex",
		Binary: c.Binary,
		Args:   c.Args,
		Env:    c.Env,
		Capabilities: Capabilities{
			Streaming:     true,
			ModelSwitch:   true,
			ContextWindow: 128000,
		},
	})
}
