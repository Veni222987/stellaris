package config

import (
	"os"
	"path/filepath"
)

// AgentsDirPath 返回当前生效的 agents 配置目录。
// 优先级：环境变量 STELLARIS_AGENTS_DIR > 配置文件中的 AgentsDir > /etc/stellaris/agents.d。
func AgentsDirPath() string {
	if d := os.Getenv("STELLARIS_AGENTS_DIR"); d != "" {
		return d
	}
	if c, err := Load(); err == nil && c.AgentsDir != "" {
		return c.AgentsDir
	}
	return "/etc/stellaris/agents.d"
}

// EnsureAgentsDir 确保目录存在并返回路径。
func EnsureAgentsDir() (string, error) {
	d := AgentsDirPath()
	return d, os.MkdirAll(d, 0o755)
}

// AgentFilePath 返回给定 agent name 对应的 yaml 文件路径。
func AgentFilePath(name string) string {
	return filepath.Join(AgentsDirPath(), name+".yaml")
}
