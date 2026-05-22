-- Relay / Orchestration 的链路追溯：parent_task_id 指向上游任务；node_id 在 Orchestration 模式下记录 DSL 节点 ID。
ALTER TABLE tasks ADD COLUMN parent_task_id BIGINT NULL REFERENCES tasks(id);
ALTER TABLE tasks ADD COLUMN node_id VARCHAR(64) NULL;
CREATE INDEX idx_tasks_parent ON tasks(parent_task_id);

-- Orchestration 模式的 DSL（JSONB）：节点列表 + 依赖。
ALTER TABLE sessions ADD COLUMN dsl JSONB NULL;
