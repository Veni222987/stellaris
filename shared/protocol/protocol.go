package protocol

import "fmt"

// MQTT topic 构造

// TaskTopic 返回任务下发的 MQTT topic
func TaskTopic(gid, planetUUID string) string {
	return fmt.Sprintf("stellaris/%s/task/%s", gid, planetUUID)
}

// ResultTopic 返回任务结果分片的 MQTT topic
func ResultTopic(gid, taskUUID string) string {
	return fmt.Sprintf("stellaris/%s/result/%s", gid, taskUUID)
}

// CancelTopic 返回任务取消的 MQTT topic
func CancelTopic(gid, taskUUID string) string {
	return fmt.Sprintf("stellaris/%s/cancel/%s", gid, taskUUID)
}

// HistoryEntry 会话历史中的一条记录
type HistoryEntry struct {
	Role    string `json:"role"`    // "user" 或 "assistant"
	Content string `json:"content"` // 消息内容
}

// TaskMessage 任务下发消息结构
type TaskMessage struct {
	TaskUUID    string         `json:"task_uuid"`
	SessionUUID string         `json:"session_uuid"`
	AgentUUID   string         `json:"agent_uuid"`
	AgentType   string         `json:"agent_type"`
	AgentName   string         `json:"agent_name"`
	Prompt      string         `json:"prompt"`
	Stream      bool           `json:"stream"`
	TimeoutSec  int            `json:"timeout_sec"`
	History     []HistoryEntry `json:"history,omitempty"` // 本次消息之前的会话历史
}

// ChunkType 结果分片类型
type ChunkType string

const (
	ChunkStdout ChunkType = "stdout" // 标准输出
	ChunkStderr ChunkType = "stderr" // 标准错误
	ChunkDone   ChunkType = "done"   // 任务完成
	ChunkError  ChunkType = "error"  // 任务出错
)

// ResultChunk 结果分片消息结构
type ResultChunk struct {
	TaskUUID string    `json:"task_uuid"`
	Seq      int       `json:"seq"`
	Type     ChunkType `json:"type"`
	Chunk    string    `json:"chunk"`
	TS       int64     `json:"ts"`
}
