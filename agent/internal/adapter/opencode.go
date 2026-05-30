package adapter

// OpenCodeConfig 描述如何启动 OpenCode CLI（sst/opencode 开源编码 agent）。
// 默认以 run 子命令运行：从 stdin 读取 prompt 并流式输出结果到 stdout。
type OpenCodeConfig struct {
	Name   string
	Binary string
	Args   []string
	Env    map[string]string
}

// NewOpenCodeAdapter 根据 OpenCodeConfig 构造一个 StdioAdapter，
// 默认 binary=opencode、args=[run]。
func NewOpenCodeAdapter(c OpenCodeConfig) *StdioAdapter {
	if c.Binary == "" {
		c.Binary = "opencode"
	}
	if len(c.Args) == 0 {
		c.Args = []string{"run"}
	}
	return NewStdioAdapter(StdioConfig{
		Name:   c.Name,
		Type:   "opencode",
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
