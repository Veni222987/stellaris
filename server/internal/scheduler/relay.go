package scheduler

import (
	"context"
	"fmt"

	"github.com/stellaris/stellaris/server/internal/model"
)

// Relay 在 task 完成时检查 session.mode=relay，若有未派发的下游 agent，
// 用本 task 的聚合输出作为下游 prompt 再 Dispatch 一次。
type Relay struct {
	sessions  *model.SessionModel
	messages  *model.MessageModel
	tasks     *model.TaskModel
	scheduler *Scheduler
}

func NewRelay(sessions *model.SessionModel, messages *model.MessageModel,
	tasks *model.TaskModel, sch *Scheduler) *Relay {
	return &Relay{sessions: sessions, messages: messages, tasks: tasks, scheduler: sch}
}

// Advance 由 Subscriber 在 chunk.type=done 后调用。返回新派发的 task_uuid（如果没有下一个则 ""）。
func (r *Relay) Advance(ctx context.Context, completedTaskUUID string) (string, error) {
	t, err := r.tasks.FindByUUID(ctx, completedTaskUUID)
	if err != nil {
		return "", err
	}
	msgID, err := r.tasks.QueryMessageID(ctx, t.ID)
	if err != nil {
		return "", err
	}
	msg, err := r.messages.FindByID(ctx, msgID)
	if err != nil {
		return "", err
	}
	sess, err := r.sessions.FindByID(ctx, msg.SessionID)
	if err != nil {
		return "", err
	}
	if sess.Mode != "relay" {
		return "", nil
	}

	// 找当前 agent 在链中的位置；t.AgentID 是 DB 主键，需要换回 agent_uuid。
	curUUID, err := r.scheduler.agents.UUIDByID(ctx, t.AgentID)
	if err != nil {
		return "", err
	}
	idx := -1
	for i, u := range sess.AgentUUIDs {
		if u == curUUID {
			idx = i
			break
		}
	}
	if idx < 0 || idx >= len(sess.AgentUUIDs)-1 {
		return "", nil // 已是链尾
	}

	output, err := r.tasks.AggregateOutput(ctx, completedTaskUUID)
	if err != nil {
		return "", err
	}
	if output == "" {
		return "", fmt.Errorf("relay: upstream task %s produced empty output", completedTaskUUID)
	}
	nextAgent := sess.AgentUUIDs[idx+1]
	parentID := t.ID
	return r.scheduler.Dispatch(ctx, msg.ID, sess.SessionUUID, sess.GID, nextAgent, output, nil, &parentID, "")
}
