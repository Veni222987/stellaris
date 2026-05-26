package daemon

import (
	"os/exec"

	"github.com/stellaris/stellaris/agent/internal/adapter"
	"github.com/stellaris/stellaris/agent/internal/discovery"
)

// AgentSource 标记一个 agent 声明的来源。
type AgentSource string

const (
	SourceAuto     AgentSource = "auto"     // PATH 自动发现
	SourceDeclared AgentSource = "declared" // agents.d 手动声明
)

// ResolvedAgent 是合并后的一个 agent 声明，附带来源标记。
type ResolvedAgent struct {
	Decl   discovery.AgentDecl
	Source AgentSource
}

// ResolveAgents 合并「PATH 自动发现的已知类型」与「agents.d 手动声明」。
// 规则：先按已知类型在 PATH 里探测默认 binary（命中即自动注册，name 取类型名）；
// 再读 agents.d，按 name 覆盖自动发现的同名项——用户可借此改 binary/args/env
// 或加同类型的多个实例。手动声明保持其首次出现的位置。
func ResolveAgents(agentsDir string) ([]ResolvedAgent, error) {
	byName := map[string]ResolvedAgent{}
	var order []string

	for _, kt := range adapter.KnownTypes() {
		path, err := exec.LookPath(kt.DefaultBinary)
		if err != nil {
			continue
		}
		name := kt.Type
		byName[name] = ResolvedAgent{
			Decl:   discovery.AgentDecl{Name: name, Type: kt.Type, Binary: path},
			Source: SourceAuto,
		}
		order = append(order, name)
	}

	decls, err := discovery.Scan(agentsDir)
	if err != nil {
		return nil, err
	}
	for _, d := range decls {
		if _, ok := byName[d.Name]; !ok {
			order = append(order, d.Name)
		}
		byName[d.Name] = ResolvedAgent{Decl: d, Source: SourceDeclared}
	}

	out := make([]ResolvedAgent, 0, len(order))
	for _, n := range order {
		out = append(out, byName[n])
	}
	return out, nil
}
