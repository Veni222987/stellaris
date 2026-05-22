// Package scheduler 把用户消息分解为单个 Agent 的 Task 并通过 MQTT 投递。
//
// M1 实现最简策略：会话里每个 agent_uuid 各创建一个 Task，下发到对应 Planet。
// M3 引入 Parallel/Relay/Orchestration 模式时在此扩展。
package scheduler

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/stellaris/stellaris/server/internal/model"
	cmqtt "github.com/stellaris/stellaris/server/internal/mqtt"
	"github.com/stellaris/stellaris/shared/protocol"
)

type Scheduler struct {
	agents    *model.AgentModel
	planets   *model.PlanetModel
	tasks     *model.TaskModel
	publisher *cmqtt.Publisher
}

func New(agents *model.AgentModel, planets *model.PlanetModel, tasks *model.TaskModel,
	pub *cmqtt.Publisher) *Scheduler {
	return &Scheduler{agents: agents, planets: planets, tasks: tasks, publisher: pub}
}

// Dispatch 给指定 agent 创建一个 Task 并把 prompt 通过 MQTT 推到所在 Planet。
// parentTaskID = nil 表示用户消息直接派发（Parallel / Relay 头 / DAG 根）；nodeID = "" 表示非 Orchestration 模式。
func (s *Scheduler) Dispatch(ctx context.Context, messageID int64, sessionUUID, gid,
	agentUUID, prompt string, parentTaskID *int64, nodeID string) (string, error) {

	a, err := s.agents.FindByUUID(ctx, agentUUID)
	if err != nil {
		return "", err
	}
	p, err := s.planets.FindByID(ctx, a.PlanetID)
	if err != nil {
		return "", err
	}

	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	taskUUID := hex.EncodeToString(b[:])
	if _, err := s.tasks.Create(ctx, taskUUID, messageID, a.ID, p.ID, parentTaskID, nodeID); err != nil {
		return "", err
	}
	if err := s.publisher.PublishTask(gid, p.PlanetUUID, protocol.TaskMessage{
		TaskUUID:    taskUUID,
		SessionUUID: sessionUUID,
		AgentUUID:   agentUUID,
		AgentType:   a.Type,
		AgentName:   a.Name,
		Prompt:      prompt,
		Stream:      true,
		TimeoutSec:  120,
	}); err != nil {
		_ = s.tasks.UpdateStatus(ctx, taskUUID, "failed")
		return "", err
	}
	return taskUUID, nil
}
