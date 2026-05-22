package adapter

// WorkbuddyConfig 描述如何启动 Workbuddy CLI。
type WorkbuddyConfig struct {
	Name   string
	Binary string
	Args   []string
	Env    map[string]string
}

// NewWorkbuddyAdapter 根据 WorkbuddyConfig 构造一个 StdioAdapter，
// 默认 binary=workbuddy、args=[chat]。
func NewWorkbuddyAdapter(c WorkbuddyConfig) *StdioAdapter {
	if c.Binary == "" {
		c.Binary = "workbuddy"
	}
	if len(c.Args) == 0 {
		c.Args = []string{"chat"}
	}
	return NewStdioAdapter(StdioConfig{
		Name:   c.Name,
		Type:   "workbuddy",
		Binary: c.Binary,
		Args:   c.Args,
		Env:    c.Env,
		Capabilities: Capabilities{
			Streaming:     true,
			ModelSwitch:   false,
			ContextWindow: 32000,
		},
	})
}
