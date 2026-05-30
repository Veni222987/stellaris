package adapter

// KnownType 描述一个内置 agent 类型及其默认可执行文件名。
// 自动发现（PATH 探测）与 adapter 构造共用此表，避免默认值漂移。
type KnownType struct {
	Type          string
	DefaultBinary string
}

// KnownTypes 返回所有内置 agent 类型。新增内置类型时在此登记，
// 并保证 DefaultBinary 与对应 NewXxxAdapter 的默认 binary 一致。
func KnownTypes() []KnownType {
	return []KnownType{
		{Type: "openclaw", DefaultBinary: "openclaw"},
		{Type: "hermes", DefaultBinary: "hermes"},
		{Type: "workbuddy", DefaultBinary: "workbuddy"},
		{Type: "claudecode", DefaultBinary: "claude"},
		{Type: "codex", DefaultBinary: "codex"},
		{Type: "opencode", DefaultBinary: "opencode"},
		{Type: "gemini", DefaultBinary: "gemini"},
	}
}

// Build 按类型构造对应 adapter；未知类型走通用 StdioAdapter。
// binary/args/env 为空时由各 adapter 落到自己的默认值。
// store 供原生会话 adapter（claudecode/codex/gemini/opencode）持久化 CLI session ID；
// openclaw/hermes/workbuddy 等 StdioAdapter 系不使用 store。
func Build(name, typ, binary string, args []string, env map[string]string, store *SessionStore) Agent {
	switch typ {
	case "openclaw":
		return NewOpenClawAdapter(OpenClawConfig{Name: name, Binary: binary, Args: args, Env: env})
	case "hermes":
		return NewHermesAdapter(HermesConfig{Name: name, Binary: binary, Args: args, Env: env})
	case "workbuddy":
		return NewWorkbuddyAdapter(WorkbuddyConfig{Name: name, Binary: binary, Args: args, Env: env})
	case "claudecode":
		return NewClaudeCodeAdapter(ClaudeCodeConfig{Name: name, Binary: binary, Args: args, Env: env}, store)
	case "codex":
		return NewCodexAdapter(CodexConfig{Name: name, Binary: binary, Args: args, Env: env}, store)
	case "opencode":
		return NewOpenCodeAdapter(OpenCodeConfig{Name: name, Binary: binary, Args: args, Env: env}, store)
	case "gemini":
		return NewGeminiAdapter(GeminiConfig{Name: name, Binary: binary, Args: args, Env: env}, store)
	default:
		return NewStdioAdapter(StdioConfig{Name: name, Type: typ, Binary: binary, Args: args, Env: env})
	}
}
