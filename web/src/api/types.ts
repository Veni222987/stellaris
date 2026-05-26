// 后端 REST 接口的请求 / 响应类型定义
export interface LoginReq { email: string; password: string }
export interface AuthResp { token: string }

export interface GalaxyCreateResp { gid: string; node_token: string }
export interface GalaxyItem { gid: string; name: string }
export interface GalaxyListResp { galaxies: GalaxyItem[] }

// 星系详情：planets 内嵌 agents，均带状态
export interface GalaxyDetailAgent {
  agent_uuid: string
  type: string
  name: string
  status: string
}
export interface GalaxyPlanet {
  planet_uuid: string
  hostname: string
  ip: string
  os: string
  status: string
  last_heartbeat: number
  agents: GalaxyDetailAgent[]
}
export interface GalaxyDetailResp {
  gid: string
  name: string
  planets: GalaxyPlanet[]
}

export interface SessionCreateReq {
  gid: string
  mode: string
  agent_uuids: string[]
  dsl?: string
}
export interface SessionCreateResp { session_uuid: string }

export interface SessionMessageResp { message_id: number; task_uuids: string[] }

// 会话列表
export interface SessionListItem {
  session_uuid: string
  gid: string
  mode: string
  agent_uuids: string[]
  title: string
  created_at: number
}
export interface SessionListResp { sessions: SessionListItem[] }

// 会话 transcript：按 user 指令分组
export interface TranscriptTask {
  task_uuid: string
  status: string
  agent_uuid: string
  parent_task_uuid?: string
  node_id?: string
  output: string
}
export interface TranscriptMessage {
  message_id: number
  role: string
  content: string
  created_at: number
  tasks: TranscriptTask[]
}
export interface SessionTranscriptResp {
  session_uuid: string
  mode: string
  agent_uuids: string[]
  messages: TranscriptMessage[]
}

export interface WSTicketResp { ticket: string }
