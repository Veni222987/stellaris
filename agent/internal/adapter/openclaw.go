package adapter

// OpenClawConfig 描述如何启动一个 OpenClaw CLI 实例。
// M1 把它当成通用 stdio 的特化封装：默认 binary=openclaw、args=[chat --stream]。
// 后续 HermesAgent / Workbuddy 可以同样模式扩展独立 adapter。
type OpenClawConfig struct {
	Name   string
	Binary string
	Args   []string
	Env    map[string]string
}

func NewOpenClawAdapter(c OpenClawConfig) *StdioAdapter {
	if c.Binary == "" {
		c.Binary = "openclaw"
	}
	if len(c.Args) == 0 {
		c.Args = []string{"chat", "--stream"}
	}
	return NewStdioAdapter(StdioConfig{
		Name:   c.Name,
		Type:   "openclaw",
		Binary: c.Binary,
		Args:   c.Args,
		Env:    c.Env,
		Capabilities: Capabilities{
			Streaming:     true,
			ModelSwitch:   false,
			ContextWindow: 200000,
		},
	})
}
