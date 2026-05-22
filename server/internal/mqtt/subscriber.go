package mqtt

import (
	"context"
	"encoding/json"

	mqttc "github.com/eclipse/paho.mqtt.golang"
	"github.com/stellaris/stellaris/server/internal/model"
	cws "github.com/stellaris/stellaris/server/internal/ws"
	"github.com/stellaris/stellaris/shared/protocol"
	"github.com/zeromicro/go-zero/core/logx"
)

// RelayAdvancer 在 task 完成时推进 relay 链；scheduler.Relay 实现该接口。
// 用接口反转避免 mqtt → scheduler → mqtt 的循环 import。
type RelayAdvancer interface {
	Advance(ctx context.Context, completedTaskUUID string) (string, error)
}

// OrchestratorAdvancer 在 task 完成时推进 DAG；scheduler.Orchestrator 实现该接口。
type OrchestratorAdvancer interface {
	Advance(ctx context.Context, completedTaskUUID string) error
}

// Subscriber 订阅所有星系的 result topic，把分片持久化并在 done/error 时更新任务状态，
// 同时通过 Hub 将 chunk fan-out 到所有订阅了该 session 的浏览器 WebSocket 连接。
type Subscriber struct {
	client       mqttc.Client
	tasks        *model.TaskModel
	chunks       *model.TaskChunkModel
	hub          *cws.Hub
	relay        RelayAdvancer
	orchestrator OrchestratorAdvancer
}

func NewSubscriber(client mqttc.Client, tasks *model.TaskModel, chunks *model.TaskChunkModel, hub *cws.Hub,
	relay RelayAdvancer, orch OrchestratorAdvancer) *Subscriber {
	return &Subscriber{client: client, tasks: tasks, chunks: chunks, hub: hub, relay: relay, orchestrator: orch}
}

// Start 注册全局订阅。需在 server.Start() 之前调用。
func (s *Subscriber) Start(ctx context.Context) error {
	topic := "stellaris/+/result/+"
	t := s.client.Subscribe(topic, 1, func(_ mqttc.Client, m mqttc.Message) {
		s.handleChunk(ctx, m.Payload())
	})
	t.Wait()
	return t.Error()
}

func (s *Subscriber) handleChunk(ctx context.Context, payload []byte) {
	var chunk protocol.ResultChunk
	if err := json.Unmarshal(payload, &chunk); err != nil {
		logx.Errorf("[mqtt-sub] 解析 chunk 失败: %v", err)
		return
	}
	task, err := s.tasks.FindByUUID(ctx, chunk.TaskUUID)
	if err != nil {
		logx.Errorf("[mqtt-sub] 找不到 task %s: %v", chunk.TaskUUID, err)
		return
	}
	if err := s.chunks.Insert(ctx, task.ID, chunk.Seq, string(chunk.Type), chunk.Chunk); err != nil {
		logx.Errorf("[mqtt-sub] 写入 chunk 失败: %v", err)
		return
	}
	logx.Infof("[mqtt-sub] inserted chunk task=%s seq=%d type=%q", chunk.TaskUUID, chunk.Seq, chunk.Type)
	switch chunk.Type {
	case protocol.ChunkDone:
		if err := s.tasks.UpdateStatus(ctx, chunk.TaskUUID, "succeeded"); err != nil {
			logx.Errorf("[mqtt-sub] UpdateStatus succeeded err=%v", err)
		} else {
			logx.Infof("[mqtt-sub] marked task %s succeeded", chunk.TaskUUID)
		}
		if s.relay != nil {
			if next, err := s.relay.Advance(ctx, chunk.TaskUUID); err != nil {
				logx.Errorf("[mqtt-sub] relay advance: %v", err)
			} else if next != "" {
				logx.Infof("[mqtt-sub] relay -> next task %s", next)
			}
		}
		if s.orchestrator != nil {
			if err := s.orchestrator.Advance(ctx, chunk.TaskUUID); err != nil {
				logx.Errorf("[mqtt-sub] orchestrator advance: %v", err)
			}
		}
	case protocol.ChunkError:
		if err := s.tasks.UpdateStatus(ctx, chunk.TaskUUID, "failed"); err != nil {
			logx.Errorf("[mqtt-sub] UpdateStatus failed err=%v", err)
		}
	}

	// chunk 落库 + 状态更新完成后，fan-out 到所有订阅该 session 的 WebSocket 连接。
	sessionUUID, err := s.tasks.SessionUUIDByTask(ctx, chunk.TaskUUID)
	if err == nil && sessionUUID != "" {
		s.hub.Publish(sessionUUID, cws.Chunk{
			TaskUUID: chunk.TaskUUID,
			Seq:      chunk.Seq,
			Type:     string(chunk.Type),
			Chunk:    chunk.Chunk,
			TS:       chunk.TS,
		})
	}
}
