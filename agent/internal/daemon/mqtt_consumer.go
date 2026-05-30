package daemon

import (
	"context"
	"encoding/json"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stellaris/stellaris/agent/internal/adapter"
	"github.com/stellaris/stellaris/agent/internal/config"
	"github.com/stellaris/stellaris/shared/protocol"
)

type consumer struct {
	cfg      *config.Config
	client   mqtt.Client
	registry *adapter.Registry
}

func newConsumer(ctx context.Context, cfg *config.Config, reg *adapter.Registry) (*consumer, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTTBroker).
		SetClientID("planet-" + cfg.PlanetUUID).
		SetAutoReconnect(true).
		SetConnectRetry(true)
	c := mqtt.NewClient(opts)
	tok := c.Connect()
	// ConnectRetry 下 token 仅在连上时完成；用 ctx 让"broker 未就绪"期间仍能被 SIGTERM 打断。
	select {
	case <-ctx.Done():
		c.Disconnect(0)
		return nil, ctx.Err()
	case <-tok.Done():
		if err := tok.Error(); err != nil {
			return nil, err
		}
	}
	return &consumer{cfg: cfg, client: c, registry: reg}, nil
}

func (c *consumer) start(ctx context.Context) error {
	topic := protocol.TaskTopic(c.cfg.GID, c.cfg.PlanetUUID)
	t := c.client.Subscribe(topic, 1, func(_ mqtt.Client, m mqtt.Message) {
		var task protocol.TaskMessage
		if err := json.Unmarshal(m.Payload(), &task); err != nil {
			log.Printf("[mqtt] 解析任务失败: %v", err)
			return
		}
		go c.execute(ctx, task)
	})
	t.Wait()
	log.Printf("[mqtt] 订阅 %s", topic)
	return t.Error()
}

func (c *consumer) execute(ctx context.Context, task protocol.TaskMessage) {
	// 单个任务的 panic 不应拖垮整个守护进程：兜住后回传 error chunk。
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[task] panic 已恢复 task=%s: %v", task.TaskUUID, r)
			c.publishError(task.TaskUUID, "internal panic")
		}
	}()
	log.Printf("[task] 收到 %s → agent=%s", task.TaskUUID, task.AgentUUID)
	a, err := c.registry.Get(task.AgentUUID)
	if err != nil {
		log.Printf("[task] registry.Get err=%v", err)
		c.publishError(task.TaskUUID, err.Error())
		return
	}
	timeout := time.Duration(task.TimeoutSec) * time.Second
	if timeout == 0 {
		timeout = 120 * time.Second
	}
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	log.Printf("[task] 启动 Chat type=%s name=%s history=%d", a.Type(), a.Name(), len(task.History))
	chunks, err := a.Chat(tctx, adapter.ChatRequest{
		SessionUUID: task.SessionUUID,
		Prompt:      task.Prompt,
		History:     toAdapterHistory(task.History),
	})
	if err != nil {
		log.Printf("[task] Chat err=%v", err)
		c.publishError(task.TaskUUID, err.Error())
		return
	}
	seq := 0
	for ch := range chunks {
		log.Printf("[task] chunk seq=%d type=%s len=%d", seq, ch.Type, len(ch.Chunk))
		c.publishChunk(task.TaskUUID, seq, protocol.ChunkType(ch.Type), ch.Chunk)
		seq++
	}
	log.Printf("[task] 完成 %s (共 %d chunks)", task.TaskUUID, seq)
}

func toAdapterHistory(h []protocol.HistoryEntry) []adapter.HistoryEntry {
	if len(h) == 0 {
		return nil
	}
	out := make([]adapter.HistoryEntry, len(h))
	for i, e := range h {
		out[i] = adapter.HistoryEntry{Role: e.Role, Content: e.Content}
	}
	return out
}

func (c *consumer) publishChunk(taskUUID string, seq int, typ protocol.ChunkType, chunk string) {
	payload, _ := json.Marshal(protocol.ResultChunk{
		TaskUUID: taskUUID, Seq: seq, Type: typ, Chunk: chunk,
		TS: time.Now().UnixMilli(),
	})
	c.client.Publish(protocol.ResultTopic(c.cfg.GID, taskUUID), 1, false, payload)
}

func (c *consumer) publishError(taskUUID, msg string) {
	c.publishChunk(taskUUID, 0, protocol.ChunkError, msg)
}
