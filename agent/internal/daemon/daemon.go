// Package daemon 是 stellaris-cli 守护进程实现：
// 启动时扫描 agents.d、为每个 Agent 实例化 adapter，然后并发跑心跳上报与 MQTT 任务消费。
package daemon

import (
	"context"
	"log"

	"github.com/stellaris/stellaris/agent/internal/adapter"
	"github.com/stellaris/stellaris/agent/internal/config"
	"github.com/stellaris/stellaris/agent/internal/discovery"
)

type Daemon struct {
	cfg      *config.Config
	registry *adapter.Registry
}

// New 加载配置 + 扫描 agent 声明并注册 adapter。
func New(cfg *config.Config) (*Daemon, error) {
	reg := adapter.NewRegistry()
	if cfg.AgentsDir == "" {
		cfg.AgentsDir = "/etc/stellaris/agents.d"
	}
	decls, err := discovery.Scan(cfg.AgentsDir)
	if err != nil {
		return nil, err
	}
	r := newRegistrar(cfg)
	for _, d := range decls {
		var a adapter.Agent
		switch d.Type {
		case "openclaw":
			a = adapter.NewOpenClawAdapter(adapter.OpenClawConfig{
				Name: d.Name, Binary: d.Binary, Args: d.Args, Env: d.Env,
			})
		case "hermes":
			a = adapter.NewHermesAdapter(adapter.HermesConfig{
				Name: d.Name, Binary: d.Binary, Args: d.Args, Env: d.Env,
			})
		case "workbuddy":
			a = adapter.NewWorkbuddyAdapter(adapter.WorkbuddyConfig{
				Name: d.Name, Binary: d.Binary, Args: d.Args, Env: d.Env,
			})
		default:
			a = adapter.NewStdioAdapter(adapter.StdioConfig{
				Name: d.Name, Type: d.Type, Binary: d.Binary, Args: d.Args, Env: d.Env,
			})
		}
		uuid, err := r.Register(d.Name, d.Type, a.Capabilities())
		if err != nil {
			log.Printf("[daemon] 注册 agent %s 失败: %v", d.Name, err)
			continue
		}
		reg.Set(uuid, a)
		log.Printf("[daemon] 已注册 agent name=%s type=%s uuid=%s", d.Name, d.Type, uuid)
	}
	return &Daemon{cfg: cfg, registry: reg}, nil
}

// Run 启动心跳与 MQTT 消费，阻塞直到 ctx 取消。
func (d *Daemon) Run(ctx context.Context) error {
	go newHeartbeat(d.cfg, d.registry).run(ctx)

	c, err := newConsumer(d.cfg, d.registry)
	if err != nil {
		return err
	}
	if err := c.start(ctx); err != nil {
		return err
	}
	log.Printf("[daemon] 已运行 gid=%s planet=%s", d.cfg.GID, d.cfg.PlanetUUID)
	<-ctx.Done()
	return nil
}
