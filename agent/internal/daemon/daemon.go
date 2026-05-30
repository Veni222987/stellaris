// Package daemon 是 stellaris-cli 守护进程实现：
// 启动时扫描 agents.d、为每个 Agent 实例化 adapter，然后并发跑心跳上报与 MQTT 任务消费。
package daemon

import (
	"context"
	"log"

	"github.com/stellaris/stellaris/agent/internal/adapter"
	"github.com/stellaris/stellaris/agent/internal/config"
)

type Daemon struct {
	cfg      *config.Config
	registry *adapter.Registry
}

// New 加载配置 + 解析 agent（PATH 自动发现 + agents.d 覆盖）并注册 adapter。
func New(cfg *config.Config) (*Daemon, error) {
	reg := adapter.NewRegistry()
	store := adapter.NewSessionStore(config.Dir())
	if cfg.AgentsDir == "" {
		cfg.AgentsDir = "/etc/stellaris/agents.d"
	}
	resolved, err := ResolveAgents(cfg.AgentsDir)
	if err != nil {
		return nil, err
	}
	r := newRegistrar(cfg)
	for _, ra := range resolved {
		d := ra.Decl
		a := adapter.Build(d.Name, d.Type, d.Binary, d.Args, d.Env, store)
		uuid, err := r.Register(d.Name, d.Type, a.Capabilities())
		if err != nil {
			log.Printf("[daemon] 注册 agent %s 失败: %v", d.Name, err)
			continue
		}
		reg.Set(uuid, a)
		log.Printf("[daemon] 已注册 agent name=%s type=%s source=%s uuid=%s", d.Name, d.Type, ra.Source, uuid)
	}
	return &Daemon{cfg: cfg, registry: reg}, nil
}

// Run 启动心跳与 MQTT 消费，阻塞直到 ctx 取消。
func (d *Daemon) Run(ctx context.Context) error {
	go newHeartbeat(d.cfg, d.registry).run(ctx)

	c, err := newConsumer(ctx, d.cfg, d.registry)
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
