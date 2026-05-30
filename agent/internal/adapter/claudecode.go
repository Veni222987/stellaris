package adapter

// ClaudeCodeConfig 描述如何启动 Claude Code CLI（Anthropic 官方 claude 命令行工具）。
// 默认以 --print 标志运行：从 stdin 读取 prompt，非交互式输出到 stdout。
type ClaudeCodeConfig struct {
	Name   string
	Binary string
	Args   []string
	Env    map[string]string
}

// NewClaudeCodeAdapter 根据 ClaudeCodeConfig 构造一个 StdioAdapter，
// 默认 binary=claude、args=[--print]。
func NewClaudeCodeAdapter(c ClaudeCodeConfig) *StdioAdapter {
	if c.Binary == "" {
		c.Binary = "claude"
	}
	if len(c.Args) == 0 {
		c.Args = []string{"--print"}
	}
	return NewStdioAdapter(StdioConfig{
		Name:   c.Name,
		Type:   "claudecode",
		Binary: c.Binary,
		Args:   c.Args,
		Env:    c.Env,
		Capabilities: Capabilities{
			Streaming:     true,
			ModelSwitch:   true,
			ContextWindow: 200000,
		},
	})
}
