// Package ticket 提供 WebSocket 一次性票据：颁发后单次消费、绑定 session_uuid、带 TTL。
//
// M2 简化阶段 WS 鉴权走 `?token=...`，相当于把长效 JWT 暴露在 URL 与日志里；
// M3 收紧：客户端先 POST 走鉴权拿一次性短 TTL ticket，再用 ticket 升级 WS。
package ticket

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// entry 单条票据记录。
type entry struct {
	sessionUUID string
	userID      int64
	expires     time.Time
}

// Store 内存级 WS ticket 颁发表；ticket 单次消费。
type Store struct {
	mu    sync.Mutex
	items map[string]entry
	ttl   time.Duration
}

// NewStore 构造一个 ttl 给定的 store。
func NewStore(ttl time.Duration) *Store {
	return &Store{items: map[string]entry{}, ttl: ttl}
}

// Issue 颁发一个新 ticket，绑定 session_uuid + userID。
func (s *Store) Issue(sessionUUID string, userID int64) (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	tk := hex.EncodeToString(b[:])
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[tk] = entry{sessionUUID: sessionUUID, userID: userID, expires: time.Now().Add(s.ttl)}
	s.gcLocked()
	return tk, nil
}

// Consume 校验 ticket 是否存在、未过期且 sessionUUID 匹配；匹配则立即删除。
func (s *Store) Consume(tk, sessionUUID string) (int64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.items[tk]
	if !ok || time.Now().After(e.expires) || e.sessionUUID != sessionUUID {
		return 0, false
	}
	delete(s.items, tk)
	return e.userID, true
}

// gcLocked 调用方必须持有 s.mu。
func (s *Store) gcLocked() {
	now := time.Now()
	for k, v := range s.items {
		if now.After(v.expires) {
			delete(s.items, k)
		}
	}
}
