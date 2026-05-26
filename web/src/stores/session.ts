// 会话 store：会话列表 + 当前会话的 transcript（按 user 指令分组，组内挂各 agent 任务）
import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { SessionListItem, SessionTranscriptResp } from '@/api/types'

export interface ChatTask {
  task_uuid: string
  agent_uuid: string
  status: string
  output: string
  parent_task_uuid?: string
  node_id?: string
}
export interface ChatMessage {
  message_id: number
  role: string
  content: string
  tasks: ChatTask[]
}

export interface WSChunk {
  task_uuid: string
  seq: number
  type: string
  chunk: string
  ts: number
}

export const useSessionStore = defineStore('session', () => {
  const sessions = ref<SessionListItem[]>([])
  const active = ref<string>('')
  // 当前会话的对话流（只缓存 active 会话，切换时重载）
  const messages = ref<ChatMessage[]>([])

  function setSessions(list: SessionListItem[]) {
    sessions.value = list
  }
  function prependSession(s: SessionListItem) {
    sessions.value.unshift(s)
  }

  // 加载某会话的 transcript
  function loadTranscript(t: SessionTranscriptResp) {
    messages.value = t.messages.map((m) => ({
      message_id: m.message_id,
      role: m.role,
      content: m.content,
      tasks: m.tasks.map((task) => ({
        task_uuid: task.task_uuid,
        agent_uuid: task.agent_uuid,
        status: task.status,
        output: task.output,
        parent_task_uuid: task.parent_task_uuid,
        node_id: task.node_id,
      })),
    }))
  }

  function clear() {
    messages.value = []
  }

  // 本地追加一条用户指令 + 占位任务（agent_uuids 已知，task_uuid 由后端返回后补齐）
  function addUserMessage(messageID: number, content: string, taskUUIDs: string[], agentUUIDs: string[]) {
    messages.value.push({
      message_id: messageID,
      role: 'user',
      content,
      tasks: taskUUIDs.map((tu, i) => ({
        task_uuid: tu,
        agent_uuid: agentUUIDs[i] ?? '',
        status: 'running',
        output: '',
      })),
    })
  }

  // WS 增量：按 task_uuid 找任务；找不到则挂到最后一条用户指令（relay/orchestration 的后继任务）
  function appendChunk(c: WSChunk) {
    let task: ChatTask | undefined
    for (const m of messages.value) {
      task = m.tasks.find((t) => t.task_uuid === c.task_uuid)
      if (task) break
    }
    if (!task) {
      const last = messages.value[messages.value.length - 1]
      if (!last) return
      task = { task_uuid: c.task_uuid, agent_uuid: '', status: 'running', output: '' }
      last.tasks.push(task)
    }
    if (c.type === 'done') {
      task.status = 'succeeded'
    } else if (c.type === 'error') {
      task.status = 'failed'
      task.output += c.chunk
    } else {
      // stdout / stderr
      task.output += c.chunk
    }
  }

  return {
    sessions,
    active,
    messages,
    setSessions,
    prependSession,
    loadTranscript,
    clear,
    addUserMessage,
    appendChunk,
  }
})
