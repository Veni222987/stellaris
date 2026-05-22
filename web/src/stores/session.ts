// 会话 store：维护会话列表、当前选中、按 session 分组的任务输出
import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { TaskOutput } from '@/api/types'

interface SessionMeta {
  session_uuid: string
  gid: string
  agent_uuids: string[]
  mode: string
  title: string
}

export const useSessionStore = defineStore('session', () => {
  const list = ref<SessionMeta[]>([])
  const active = ref<string>('')
  const tasksBySession = ref<Record<string, TaskOutput[]>>({})

  function addLocal(s: SessionMeta) {
    list.value.unshift(s)
    active.value = s.session_uuid
  }
  function setTasks(sessionUUID: string, tasks: TaskOutput[]) {
    tasksBySession.value[sessionUUID] = tasks
  }
  // ws 推送增量：找不到 task 就先建一个 running
  function appendChunk(
    sessionUUID: string,
    c: { task_uuid: string; seq: number; type: string; chunk: string; ts: number },
  ) {
    const tasks = tasksBySession.value[sessionUUID] || []
    let t = tasks.find((x) => x.task_uuid === c.task_uuid)
    if (!t) {
      t = { task_uuid: c.task_uuid, status: 'running', agent_uuid: '', chunks: [] }
      tasks.push(t)
      tasksBySession.value[sessionUUID] = tasks
    }
    t.chunks.push({ seq: c.seq, type: c.type, chunk: c.chunk, ts: c.ts })
    if (c.type === 'done') t.status = 'succeeded'
    if (c.type === 'error') t.status = 'failed'
  }

  return { list, active, tasksBySession, addLocal, setTasks, appendChunk }
})
