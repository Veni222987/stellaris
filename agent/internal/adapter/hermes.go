package adapter

// HermesConfig 描述如何启动 HermesAgent CLI。
// 默认假设 hermes 二进制接受 prompt 在 stdin、流式输出到 stdout（与 OpenClaw 同一形态）；
// 用户可以通过 yaml 覆盖 binary/args/env。
type HermesConfig struct {
	Name   string
	Binary string
	Args   []string
	Env    map[string]string
}

// NewHermesAdapter 根据 HermesConfig 构造一个 StdioAdapter，
// 默认 binary=hermes、args=[--stream]。
func NewHermesAdapter(c HermesConfig) *StdioAdapter {
	if c.Binary == "" {
		c.Binary = "hermes"
	}
	if len(c.Args) == 0 {
		c.Args = []string{"--stream"}
	}
	return NewStdioAdapter(StdioConfig{
		Name:   c.Name,
		Type:   "hermes",
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
