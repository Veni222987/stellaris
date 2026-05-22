// Package discovery 扫描本机 agents.d 配置目录，把每个 yaml 解析成 AgentDecl 列表。
package discovery

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type AgentDecl struct {
	Name   string            `yaml:"name"`
	Type   string            `yaml:"type"`
	Binary string            `yaml:"binary"`
	Args   []string          `yaml:"args"`
	Env    map[string]string `yaml:"env"`
}

// Scan 列出 dir 下所有 *.yaml 文件并解析。目录不存在时返回 nil（不算错误）。
func Scan(dir string) ([]AgentDecl, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []AgentDecl
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".yaml" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var d AgentDecl
		if err := yaml.Unmarshal(data, &d); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}
