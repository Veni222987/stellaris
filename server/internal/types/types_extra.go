// types_extra.go 存放手工新增（不走 goctl）的请求/响应类型，避免被 goctl 重新生成时覆盖。
package types

type GalaxyAgentsReq struct {
	GID string `path:"gid"`
}

type GalaxyItem struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

type GalaxyListResp struct {
	Galaxies []GalaxyItem `json:"galaxies"`
}

type GalaxyAgent struct {
	AgentUUID string `json:"agent_uuid"`
	Type      string `json:"type"`
	Name      string `json:"name"`
}

type GalaxyAgentsResp struct {
	Agents []GalaxyAgent `json:"agents"`
}

// ===== 星系详情：planets 内嵌 agents，均带状态 =====

type GalaxyDetailReq struct {
	GID string `path:"gid"`
}

type GalaxyDetailAgent struct {
	AgentUUID string `json:"agent_uuid"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Status    string `json:"status"`
}

type GalaxyPlanet struct {
	PlanetUUID    string              `json:"planet_uuid"`
	Hostname      string              `json:"hostname"`
	IP            string              `json:"ip"`
	OS            string              `json:"os"`
	Status        string              `json:"status"`
	LastHeartbeat int64               `json:"last_heartbeat"` // unix 秒，0 表示从未心跳
	Agents        []GalaxyDetailAgent `json:"agents"`
}

type GalaxyDetailResp struct {
	GID     string         `json:"gid"`
	Name    string         `json:"name"`
	Planets []GalaxyPlanet `json:"planets"`
}

// ===== 会话列表 =====

type SessionListReq struct {
	GID string `form:"gid,optional"`
}

type SessionListItem struct {
	SessionUUID string   `json:"session_uuid"`
	GID         string   `json:"gid"`
	Mode        string   `json:"mode"`
	AgentUUIDs  []string `json:"agent_uuids"`
	Title       string   `json:"title"`
	CreatedAt   int64    `json:"created_at"`
}

type SessionListResp struct {
	Sessions []SessionListItem `json:"sessions"`
}

// ===== 会话 transcript：按 user 指令分组，每组挂各 agent 任务输出 =====

type SessionTranscriptReq struct {
	SessionUUID string `path:"session_uuid"`
}

type TranscriptTask struct {
	TaskUUID       string `json:"task_uuid"`
	Status         string `json:"status"`
	AgentUUID      string `json:"agent_uuid"`
	ParentTaskUUID string `json:"parent_task_uuid,omitempty"`
	NodeID         string `json:"node_id,omitempty"`
	Output         string `json:"output"`
}

type TranscriptMessage struct {
	MessageID int64            `json:"message_id"`
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	CreatedAt int64            `json:"created_at"`
	Tasks     []TranscriptTask `json:"tasks"`
}

type SessionTranscriptResp struct {
	SessionUUID string              `json:"session_uuid"`
	Mode        string              `json:"mode"`
	AgentUUIDs  []string            `json:"agent_uuids"`
	Messages    []TranscriptMessage `json:"messages"`
}

type AgentRegisterReq struct {
	Type             string                 `json:"type"`
	Name             string                 `json:"name"`
	Models           []string               `json:"models"`
	CapabilitiesJSON map[string]interface{} `json:"capabilities"`
}

type AgentRegisterResp struct {
	AgentUUID string `json:"agent_uuid"`
}

type AgentUnregisterReq struct {
	AgentUUID string `json:"agent_uuid"`
}

type AgentUnregisterResp struct {
	OK bool `json:"ok"`
}
