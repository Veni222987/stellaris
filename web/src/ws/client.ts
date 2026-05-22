// session 维度的 WebSocket 客户端：先 POST 拿一次性 ticket，再 connect WS
import { api } from '@/api/client'
import type { WSTicketResp } from '@/api/types'

export interface WSChunk {
  task_uuid: string
  seq: number
  type: string
  chunk: string
  ts: number
}

// openSessionWS 先 POST 拿一次性 ticket，再连接 ws，避免长效 JWT 暴露在 URL。
export async function openSessionWS(sessionUUID: string, onChunk: (c: WSChunk) => void): Promise<WebSocket> {
  const { data } = await api.post<WSTicketResp>(`/api/session/${sessionUUID}/ws-ticket`)
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  const url = `${proto}://${location.host}/ws/session/${sessionUUID}?ticket=${encodeURIComponent(data.ticket)}`
  const ws = new WebSocket(url)
  ws.onmessage = (ev) => {
    try {
      onChunk(JSON.parse(ev.data) as WSChunk)
    } catch (e) {
      console.error('[ws] parse error', e)
    }
  }
  return ws
}
