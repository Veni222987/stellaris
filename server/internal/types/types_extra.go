// types_extra.go 存放手工新增（不走 goctl）的请求/响应类型，避免被 goctl 重新生成时覆盖。
package types

type GalaxyAgentsReq struct {
	GID string `path:"gid"`
}

type GalaxyAgent struct {
	AgentUUID string `json:"agent_uuid"`
	Type      string `json:"type"`
	Name      string `json:"name"`
}

type GalaxyAgentsResp struct {
	Agents []GalaxyAgent `json:"agents"`
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
