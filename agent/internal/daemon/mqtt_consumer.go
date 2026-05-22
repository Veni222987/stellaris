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

func newConsumer(cfg *config.Config, reg *adapter.Registry) (*consumer, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTTBroker).
		SetClientID("planet-" + cfg.PlanetUUID).
		SetAutoReconnect(true).
		SetConnectRetry(true)
	c := mqtt.NewClient(opts)
	if t := c.Connect(); t.Wait() && t.Error() != nil {
		return nil, t.Error()
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

	log.Printf("[task] 启动 Chat type=%s name=%s", a.Type(), a.Name())
	chunks, err := a.Chat(tctx, adapter.ChatRequest{Prompt: task.Prompt})
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
