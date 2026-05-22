// Package ws 维护 session_uuid → 一组 WebSocket 连接，作为 MQTT chunk 与浏览器的
// 中转：subscriber 收到 chunk 后调 Hub.Publish，所有订阅该 session 的连接都会收到。
package ws

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

// Chunk 是推送给浏览器的 chunk 结构，与 protocol.ResultChunk 字段对齐。
type Chunk struct {
	TaskUUID string `json:"task_uuid"`
	Seq      int    `json:"seq"`
	Type     string `json:"type"`
	Chunk    string `json:"chunk"`
	TS       int64  `json:"ts"`
}

// Hub 管理 session_uuid → WebSocket 连接集合的映射，所有操作均线程安全。
type Hub struct {
	mu    sync.RWMutex
	conns map[string]map[*websocket.Conn]struct{}
}

// New 创建一个空 Hub。
func New() *Hub { return &Hub{conns: map[string]map[*websocket.Conn]struct{}{}} }

// Add 把连接 c 注册到 sessionUUID 对应的连接集合。
func (h *Hub) Add(sessionUUID string, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conns[sessionUUID] == nil {
		h.conns[sessionUUID] = map[*websocket.Conn]struct{}{}
	}
	h.conns[sessionUUID][c] = struct{}{}
}

// Remove 从 sessionUUID 对应的连接集合中移除 c；集合为空时删除整个 key。
func (h *Hub) Remove(sessionUUID string, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.conns[sessionUUID]; ok {
		delete(set, c)
		if len(set) == 0 {
			delete(h.conns, sessionUUID)
		}
	}
}

// Publish 把 chunk 投递到指定 session 的所有连接。无连接则 no-op，不报错。
func (h *Hub) Publish(sessionUUID string, ch Chunk) {
	h.mu.RLock()
	conns := make([]*websocket.Conn, 0, len(h.conns[sessionUUID]))
	for c := range h.conns[sessionUUID] {
		conns = append(conns, c)
	}
	h.mu.RUnlock()

	payload, _ := json.Marshal(ch)
	for _, c := range conns {
		_ = c.WriteMessage(websocket.TextMessage, payload)
	}
}
