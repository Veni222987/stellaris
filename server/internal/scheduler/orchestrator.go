package scheduler

import (
	"context"
	"fmt"
	"strings"

	"github.com/stellaris/stellaris/server/internal/model"
	"github.com/stellaris/stellaris/shared/protocol"
)

// Orchestrator 负责 DAG 模式调度：根节点由用户消息触发；后续节点由 task done 事件推进。
type Orchestrator struct {
	sessions  *model.SessionModel
	messages  *model.MessageModel
	tasks     *model.TaskModel
	agents    *model.AgentModel
	scheduler *Scheduler
}

func NewOrchestrator(sessions *model.SessionModel, messages *model.MessageModel,
	tasks *model.TaskModel, agents *model.AgentModel, sch *Scheduler) *Orchestrator {
	return &Orchestrator{sessions: sessions, messages: messages, tasks: tasks, agents: agents, scheduler: sch}
}

// StartRun 用户消息进入：派发所有 depends_on 为空的根节点。
func (o *Orchestrator) StartRun(ctx context.Context, s *model.Session, msgID int64, userPrompt string,
	history []protocol.HistoryEntry) ([]string, error) {
	dsl, err := dslOfSession(s)
	if err != nil {
		return nil, err
	}
	var dispatched []string
	for _, n := range dsl.Nodes {
		if len(n.DependsOn) == 0 {
			prompt := renderPrompt(n.PromptTpl, userPrompt, nil)
			tu, err := o.scheduler.Dispatch(ctx, msgID, s.SessionUUID, s.GID, n.AgentUUID, prompt, history, nil, n.ID)
			if err != nil {
				return nil, err
			}
			dispatched = append(dispatched, tu)
		}
	}
	if len(dispatched) == 0 {
		return nil, fmt.Errorf("orchestration: dsl has no root node")
	}
	return dispatched, nil
}

// Advance 由 Subscriber 在 task done 时调用：找所有依赖全 done 且自身未派发的节点，全部派发。
// 注意：签名 (ctx, taskUUID string) error 由 mqtt.OrchestratorAdvancer 接口约束，不可更动。
func (o *Orchestrator) Advance(ctx context.Context, completedTaskUUID string) error {
	t, err := o.tasks.FindByUUID(ctx, completedTaskUUID)
	if err != nil {
		return err
	}
	if t.NodeID == "" {
		return nil // 非 Orchestration 任务
	}
	msgID, err := o.tasks.QueryMessageID(ctx, t.ID)
	if err != nil {
		return err
	}
	msg, err := o.messages.FindByID(ctx, msgID)
	if err != nil {
		return err
	}
	sess, err := o.sessions.FindByID(ctx, msg.SessionID)
	if err != nil {
		return err
	}
	if sess.Mode != "orchestration" {
		return nil
	}
	dsl, err := dslOfSession(sess)
	if err != nil {
		return err
	}

	done, outputs, err := o.completedNodes(ctx, msgID)
	if err != nil {
		return err
	}
	dispatched, err := o.dispatchedNodes(ctx, msgID)
	if err != nil {
		return err
	}
	for _, n := range dsl.Nodes {
		if dispatched[n.ID] {
			continue
		}
		allDepsDone := true
		for _, dep := range n.DependsOn {
			if !done[dep] {
				allDepsDone = false
				break
			}
		}
		if !allDepsDone {
			continue
		}
		prompt := renderPrompt(n.PromptTpl, msg.Content, outputs)
		// parent 选第一个 depends_on 对应的 task_id（仅用于审计/UI）；行为以 done 集合为准。
		var parentID *int64
		if len(n.DependsOn) > 0 {
			if pid, ok := outputs[n.DependsOn[0]+":__taskid"]; ok {
				var id int64
				if _, e := fmt.Sscanf(pid, "%d", &id); e == nil {
					parentID = &id
				}
			}
		}
		if _, err := o.scheduler.Dispatch(ctx, msgID, sess.SessionUUID, sess.GID,
			n.AgentUUID, prompt, nil, parentID, n.ID); err != nil {
			return err
		}
	}
	return nil
}

// completedNodes 收集所有 succeeded 节点的输出；AggregateOutput 返空字符串时容忍并继续。
func (o *Orchestrator) completedNodes(ctx context.Context, msgID int64) (map[string]bool, map[string]string, error) {
	rows, err := o.tasks.QueryByMessage(ctx, msgID)
	if err != nil {
		return nil, nil, err
	}
	done := map[string]bool{}
	outputs := map[string]string{}
	for _, t := range rows {
		if t.Status != "succeeded" || t.NodeID == "" {
			continue
		}
		done[t.NodeID] = true
		out, _ := o.tasks.AggregateOutput(ctx, t.TaskUUID)
		outputs[t.NodeID] = out
		outputs[t.NodeID+":__taskid"] = fmt.Sprintf("%d", t.ID)
	}
	return done, outputs, nil
}

// dispatchedNodes 所有已派发（任意状态）的 node_id 集合，用于幂等。
func (o *Orchestrator) dispatchedNodes(ctx context.Context, msgID int64) (map[string]bool, error) {
	rows, err := o.tasks.QueryByMessage(ctx, msgID)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, t := range rows {
		if t.NodeID != "" {
			out[t.NodeID] = true
		}
	}
	return out, nil
}

// renderPrompt 把 {{user}} 与 {{node:X}} 替换为实际内容；不做完整模板引擎。
func renderPrompt(tpl, userPrompt string, outputs map[string]string) string {
	s := strings.ReplaceAll(tpl, "{{user}}", userPrompt)
	for nid, val := range outputs {
		if strings.HasSuffix(nid, ":__taskid") {
			continue
		}
		s = strings.ReplaceAll(s, "{{node:"+nid+"}}", val)
	}
	return s
}

func dslOfSession(s *model.Session) (*DSL, error) {
	if len(s.DSL) == 0 {
		return nil, fmt.Errorf("session has no dsl")
	}
	return ParseDSL(s.DSL)
}
