package adapter

// GeminiConfig 描述如何启动 Google Gemini CLI。
// 默认以 --prompt 标志运行：非交互式地从 stdin 读取 prompt 并流式输出到 stdout。
type GeminiConfig struct {
	Name   string
	Binary string
	Args   []string
	Env    map[string]string
}

// NewGeminiAdapter 根据 GeminiConfig 构造一个 StdioAdapter，
// 默认 binary=gemini、args=[--prompt]。
func NewGeminiAdapter(c GeminiConfig) *StdioAdapter {
	if c.Binary == "" {
		c.Binary = "gemini"
	}
	if len(c.Args) == 0 {
		c.Args = []string{"--prompt"}
	}
	return NewStdioAdapter(StdioConfig{
		Name:   c.Name,
		Type:   "gemini",
		Binary: c.Binary,
		Args:   c.Args,
		Env:    c.Env,
		Capabilities: Capabilities{
			Streaming:     true,
			ModelSwitch:   true,
			ContextWindow: 1000000,
		},
	})
}
