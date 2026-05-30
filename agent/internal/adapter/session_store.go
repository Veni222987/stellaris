package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// SessionStore 持久化维护 (coreSessionUUID:agentName) → cliSessionID 的映射，
// 供各原生 adapter 在 Daemon 重启后仍能恢复 CLI 会话。
type SessionStore struct {
	mu       sync.RWMutex
	data     map[string]string
	filePath string
}

// NewSessionStore 从 configDir/sessions.json 加载已有映射并返回 store 实例。
func NewSessionStore(configDir string) *SessionStore {
	s := &SessionStore{
		data:     make(map[string]string),
		filePath: filepath.Join(configDir, "sessions.json"),
	}
	s.load()
	return s
}

func (s *SessionStore) key(coreSessionUUID, agentName string) string {
	return coreSessionUUID + ":" + agentName
}

// Get 返回指定 core session + agent 对应的 CLI session ID。
func (s *SessionStore) Get(coreSessionUUID, agentName string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[s.key(coreSessionUUID, agentName)]
	return v, ok
}

// Set 存入映射并异步持久化到磁盘。
func (s *SessionStore) Set(coreSessionUUID, agentName, cliSessionID string) {
	s.mu.Lock()
	s.data[s.key(coreSessionUUID, agentName)] = cliSessionID
	snapshot := make(map[string]string, len(s.data))
	for k, v := range s.data {
		snapshot[k] = v
	}
	s.mu.Unlock()
	go s.save(snapshot)
}

func (s *SessionStore) load() {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &s.data)
}

func (s *SessionStore) save(snapshot map[string]string) {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return
	}
	_ = os.WriteFile(s.filePath, data, 0o600)
}
