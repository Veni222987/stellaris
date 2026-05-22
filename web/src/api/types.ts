// 后端 REST 接口的请求 / 响应类型定义
export interface LoginReq { email: string; password: string }
export interface AuthResp { token: string }

export interface GalaxyCreateResp { gid: string; node_token: string }

export interface GalaxyAgent { agent_uuid: string; type: string; name: string }
export interface GalaxyAgentsResp { agents: GalaxyAgent[] }

export interface SessionCreateReq {
  gid: string
  mode: string
  agent_uuids: string[]
  dsl?: string
}
export interface SessionCreateResp { session_uuid: string }

export interface SessionMessageResp { message_id: number; task_uuids: string[] }

export interface TaskChunk { seq: number; type: string; chunk: string; ts: number }
export interface TaskOutput {
  task_uuid: string
  status: string
  agent_uuid: string
  chunks: TaskChunk[]
  parent_task_uuid?: string
  node_id?: string
}
export interface SessionHistoryResp { tasks: TaskOutput[] }

export interface WSTicketResp { ticket: string }
