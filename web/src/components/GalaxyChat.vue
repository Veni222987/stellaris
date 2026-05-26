<!-- 星系聊天：左会话历史 / 右对话流 + 指令输入；新建会话选模式 + 星系内 agent -->
<template>
  <div class="chat">
    <!-- 会话历史侧栏 -->
    <aside class="sidebar">
      <button class="btn btn-primary new-btn" @click="openNew">+ 新建会话</button>
      <div class="sb-head eyebrow">会话历史 · {{ store.sessions.length }}</div>
      <div class="sb-list">
        <button
          v-for="s in store.sessions"
          :key="s.session_uuid"
          :class="['sb-item', { on: store.active === s.session_uuid }]"
          @click="select(s.session_uuid)"
        >
          <div class="sb-title">{{ s.title || '未命名会话' }}</div>
          <div class="sb-meta mono">
            <span class="chip" :class="modeColor(s.mode)">{{ s.mode }}</span>
            <span>{{ s.agent_uuids.length }} agents</span>
          </div>
        </button>
        <div v-if="!store.sessions.length" class="sb-empty mono">暂无会话，点击上方新建</div>
      </div>
    </aside>

    <!-- 对话主区 -->
    <section class="main">
      <template v-if="store.active">
        <div ref="scrollEl" class="stream">
          <div v-if="!store.messages.length" class="stream-empty mono">
            会话已就绪 · 在下方下达第一条指令
          </div>
          <div v-for="m in store.messages" :key="m.message_id" class="turn rise">
            <div class="user-row">
              <span class="role mono">操作员</span>
              <div class="user-bubble">{{ m.content }}</div>
            </div>
            <div class="agent-grid">
              <article v-for="t in m.tasks" :key="t.task_uuid" class="task">
                <header class="t-head">
                  <span :class="['dot', t.status]" />
                  <span class="t-agent mono">{{ agentLabel(t) }}</span>
                  <span v-if="t.node_id" class="chip plasma">{{ t.node_id }}</span>
                  <span :class="['t-status mono', statusClass(t.status)]">{{ t.status }}</span>
                </header>
                <pre v-if="t.output" class="t-out">{{ t.output }}</pre>
                <div v-else class="t-wait mono">
                  <span class="blink">▍</span> 等待回传…
                </div>
              </article>
            </div>
          </div>
        </div>

        <!-- 输入 -->
        <div class="composer">
          <textarea
            v-model="input"
            class="textarea"
            :rows="1"
            placeholder="下达指令 — Enter 发送 · Shift+Enter 换行"
            @keydown="onKeydown"
          />
          <button class="btn btn-primary send" :disabled="sending || !input.trim()" @click="onSend">
            {{ sending ? '发送中' : '发送' }}
          </button>
        </div>
      </template>

      <div v-else class="no-session">
        <div class="ns-mark"><span class="ring" /><span class="core" /></div>
        <p>选择左侧会话，或新建一个会话开始下达指令</p>
      </div>
    </section>

    <!-- 新建会话 -->
    <div v-if="showNew" class="modal-mask" @click.self="showNew = false">
      <div class="modal panel rise">
        <h2>新建会话</h2>
        <p class="sub mono">为星系 {{ detail.name }} 配置调度模式与参与的智能体</p>

        <label class="field-label">调度模式 · MODE</label>
        <div class="mode-row">
          <button
            v-for="opt in modeOptions"
            :key="opt.value"
            :class="['mode-card', { on: mode === opt.value }]"
            @click="mode = opt.value"
          >
            <span class="mode-name">{{ opt.label }}</span>
            <span class="mode-desc">{{ opt.desc }}</span>
          </button>
        </div>

        <label class="field-label" style="margin-top: 18px">
          参与智能体 · AGENTS（已选 {{ selected.length }}）
        </label>
        <div class="agent-pick">
          <template v-for="p in detail.planets" :key="p.planet_uuid">
            <div v-if="p.agents.length" class="pick-planet">
              <span class="pp-host mono"><span :class="['dot', p.status]" /> {{ p.hostname }}</span>
              <label v-for="a in p.agents" :key="a.agent_uuid" class="pick-row">
                <input type="checkbox" :value="a.agent_uuid" v-model="selected" />
                <span :class="['dot', a.status]" />
                <span class="pick-name">{{ a.name }}</span>
                <span class="chip teal">{{ a.type }}</span>
              </label>
            </div>
          </template>
          <div v-if="!hasAgents" class="mono pick-empty">该星系暂无可用智能体</div>
        </div>

        <div v-if="mode === 'orchestration'" class="dsl-block">
          <label class="field-label">编排 DSL · JSON</label>
          <textarea
            v-model="dsl"
            class="textarea"
            rows="6"
            placeholder='{"nodes":[{"id":"a","agent_uuid":"...","prompt":"{{user}}"}]}'
          />
        </div>

        <div class="modal-foot">
          <button class="btn btn-ghost btn-sm" @click="showNew = false">取消</button>
          <button class="btn btn-primary btn-sm" :disabled="!canCreate" @click="onCreate">创建会话</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useMessage } from 'naive-ui'
import { api } from '@/api/client'
import { useSessionStore } from '@/stores/session'
import { openSessionWS } from '@/ws/client'
import type {
  GalaxyDetailResp,
  SessionCreateResp,
  SessionListResp,
  SessionMessageResp,
  SessionTranscriptResp,
} from '@/api/types'

const props = defineProps<{ gid: string; detail: GalaxyDetailResp }>()
const store = useSessionStore()
const msg = useMessage()

const input = ref('')
const sending = ref(false)
const scrollEl = ref<HTMLElement | null>(null)
let currentWS: WebSocket | null = null

// 新建会话表单
const showNew = ref(false)
const mode = ref('parallel')
const dsl = ref('')
const selected = ref<string[]>([])
const modeOptions = [
  { label: 'Parallel', value: 'parallel', desc: '全员并行 fan-out' },
  { label: 'Relay', value: 'relay', desc: '按顺序接力传递' },
  { label: 'Orchestration', value: 'orchestration', desc: 'DSL 编排 DAG' },
]

const hasAgents = computed(() => props.detail.planets.some((p) => p.agents.length))
const canCreate = computed(
  () => selected.value.length > 0 && (mode.value !== 'orchestration' || dsl.value.trim().length > 0),
)
const activeMeta = computed(() => store.sessions.find((s) => s.session_uuid === store.active))

function modeColor(m: string) {
  if (m === 'relay') return 'amber'
  if (m === 'orchestration') return 'plasma'
  return 'teal'
}
function statusClass(s: string) {
  if (s === 'succeeded') return 'ok'
  if (s === 'failed' || s === 'error') return 'bad'
  return 'warn'
}

// agent uuid → 名称（取自星系详情）
const agentNameMap = computed(() => {
  const map: Record<string, string> = {}
  for (const p of props.detail.planets) for (const a of p.agents) map[a.agent_uuid] = a.name
  return map
})
function agentLabel(t: { agent_uuid: string }) {
  if (!t.agent_uuid) return '智能体 ····'
  return agentNameMap.value[t.agent_uuid] || t.agent_uuid.slice(0, 10)
}

async function loadSessions() {
  const { data } = await api.get<SessionListResp>('/api/session/list', { params: { gid: props.gid } })
  store.setSessions(data.sessions ?? [])
}

function openNew() {
  mode.value = 'parallel'
  dsl.value = ''
  selected.value = []
  showNew.value = true
}

async function onCreate() {
  if (!canCreate.value) return
  try {
    const { data } = await api.post<SessionCreateResp>('/api/session/create', {
      gid: props.gid,
      mode: mode.value,
      agent_uuids: selected.value,
      dsl: mode.value === 'orchestration' ? dsl.value : undefined,
    })
    store.prependSession({
      session_uuid: data.session_uuid,
      gid: props.gid,
      mode: mode.value,
      agent_uuids: [...selected.value],
      title: '',
      created_at: Date.now() / 1000,
    })
    showNew.value = false
    await select(data.session_uuid)
  } catch (e: any) {
    msg.error(e.response?.data?.msg || '创建会话失败')
  }
}

async function select(uuid: string) {
  store.active = uuid
}

async function onSend() {
  const content = input.value.trim()
  if (!content || !store.active) return
  sending.value = true
  try {
    const { data } = await api.post<SessionMessageResp>(`/api/session/${store.active}/message`, { content })
    store.addUserMessage(data.message_id, content, data.task_uuids, activeMeta.value?.agent_uuids ?? [])
    input.value = ''
    scrollToBottom()
  } catch (e: any) {
    msg.error(e.response?.data?.msg || '发送失败')
  } finally {
    sending.value = false
  }
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey && !e.isComposing) {
    e.preventDefault()
    onSend()
  }
}

function scrollToBottom() {
  nextTick(() => {
    if (scrollEl.value) scrollEl.value.scrollTop = scrollEl.value.scrollHeight
  })
}

// 切换会话：关旧 WS → 载 transcript → 开新 WS
watch(
  () => store.active,
  async (uuid) => {
    if (currentWS) {
      currentWS.close()
      currentWS = null
    }
    store.clear()
    if (!uuid) return
    try {
      const { data } = await api.get<SessionTranscriptResp>(`/api/session/${uuid}/transcript`)
      store.loadTranscript(data)
      scrollToBottom()
    } catch {
      /* 新会话无 transcript，忽略 */
    }
    currentWS = await openSessionWS(uuid, (c) => {
      store.appendChunk(c)
      scrollToBottom()
    })
  },
)

onMounted(loadSessions)
onUnmounted(() => currentWS?.close())
</script>

<style scoped>
.chat {
  flex: 1;
  min-height: 0;
  display: flex;
}
/* 侧栏 */
.sidebar {
  width: 280px;
  flex: none;
  border-right: 1px solid var(--hairline);
  display: flex;
  flex-direction: column;
  padding: 16px 14px;
  background: rgba(7, 11, 21, 0.4);
}
.new-btn {
  width: 100%;
}
.sb-head {
  margin: 20px 6px 10px;
}
.sb-list {
  flex: 1;
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.sb-item {
  text-align: left;
  border: 1px solid transparent;
  background: transparent;
  border-radius: var(--r);
  padding: 11px 12px;
  cursor: pointer;
  transition: background 0.16s, border-color 0.16s;
  font-family: inherit;
  color: inherit;
}
.sb-item:hover {
  background: var(--panel-hi);
}
.sb-item.on {
  background: var(--teal-soft);
  border-color: var(--teal-line);
}
.sb-title {
  color: var(--text-hi);
  font-size: 13px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 7px;
}
.sb-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 10px;
  color: var(--text-dim);
}
.sb-empty {
  text-align: center;
  color: var(--text-faint);
  font-size: 12px;
  padding: 30px 10px;
}
/* 主区 */
.main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.stream {
  flex: 1;
  overflow: auto;
  padding: 26px 28px;
}
.stream-empty {
  text-align: center;
  color: var(--text-faint);
  margin-top: 80px;
  font-size: 13px;
}
.turn {
  margin-bottom: 28px;
}
.user-row {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 6px;
  margin-bottom: 16px;
}
.role {
  font-size: 10px;
  letter-spacing: 1.5px;
  text-transform: uppercase;
  color: var(--teal);
}
.user-bubble {
  max-width: 75%;
  background: linear-gradient(180deg, rgba(65, 227, 201, 0.16), rgba(65, 227, 201, 0.06));
  border: 1px solid var(--teal-line);
  border-radius: 14px 14px 4px 14px;
  padding: 12px 16px;
  color: var(--text-hi);
  white-space: pre-wrap;
  line-height: 1.55;
}
.agent-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 12px;
}
.task {
  border: 1px solid var(--hairline);
  background: var(--panel);
  border-radius: 4px 14px 14px 14px;
  overflow: hidden;
}
.t-head {
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 10px 14px;
  border-bottom: 1px solid var(--hairline);
  background: var(--panel-hi);
}
.t-agent {
  color: var(--text-hi);
  font-size: 12px;
  flex: 1;
  font-weight: 500;
}
.t-status {
  font-size: 11px;
}
.t-out {
  margin: 0;
  padding: 14px;
  font-family: var(--mono);
  font-size: 12.5px;
  line-height: 1.6;
  color: var(--text);
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 420px;
  overflow: auto;
}
.t-wait {
  padding: 14px;
  font-size: 12px;
  color: var(--text-dim);
}
.blink {
  color: var(--teal);
  animation: blink 1s step-start infinite;
}
@keyframes blink {
  50% {
    opacity: 0;
  }
}
.ok {
  color: var(--ok);
}
.bad {
  color: var(--danger);
}
.warn {
  color: var(--amber);
}
/* 输入 */
.composer {
  flex: none;
  border-top: 1px solid var(--hairline);
  padding: 16px 24px;
  display: flex;
  gap: 12px;
  align-items: flex-end;
  background: rgba(7, 11, 21, 0.5);
}
.composer .textarea {
  flex: 1;
  max-height: 160px;
}
.send {
  height: 44px;
}
.no-session {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 22px;
  color: var(--text-dim);
}
.ns-mark {
  position: relative;
  width: 70px;
  height: 70px;
}
.ns-mark .ring {
  position: absolute;
  inset: 0;
  border: 1.5px dashed var(--hairline-hi);
  border-radius: 50%;
  animation: spin 16s linear infinite;
}
.ns-mark .core {
  position: absolute;
  inset: 26px;
  border-radius: 50%;
  background: radial-gradient(circle at 35% 30%, #aefff0, var(--teal));
  box-shadow: 0 0 20px -2px var(--teal);
  opacity: 0.5;
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
/* 模态 */
.modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(2, 4, 9, 0.72);
  backdrop-filter: blur(4px);
  display: grid;
  place-items: center;
  z-index: 50;
  padding: 20px;
}
.modal {
  width: 560px;
  max-width: 100%;
  max-height: 88vh;
  overflow: auto;
  padding: 28px 30px;
  box-shadow: var(--shadow);
}
.modal h2 {
  font-size: 20px;
}
.sub {
  font-size: 12px;
  color: var(--text-dim);
  margin: 6px 0 22px;
}
.mode-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}
.mode-card {
  display: flex;
  flex-direction: column;
  gap: 5px;
  text-align: left;
  padding: 12px 14px;
  border: 1px solid var(--hairline);
  border-radius: var(--r);
  background: rgba(4, 7, 14, 0.5);
  cursor: pointer;
  transition: border-color 0.18s, background 0.18s;
  font-family: inherit;
  color: inherit;
}
.mode-card.on {
  border-color: var(--teal-line);
  background: var(--teal-soft);
}
.mode-name {
  font-family: var(--display);
  font-weight: 600;
  font-size: 14px;
  color: var(--text-hi);
}
.mode-desc {
  font-size: 11px;
  color: var(--text-dim);
}
.agent-pick {
  border: 1px solid var(--hairline);
  border-radius: var(--r);
  padding: 8px;
  max-height: 260px;
  overflow: auto;
  background: rgba(4, 7, 14, 0.4);
}
.pick-planet {
  margin-bottom: 8px;
}
.pp-host {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 11px;
  color: var(--text-dim);
  padding: 6px 8px;
}
.pick-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px 8px 22px;
  border-radius: var(--r-sm);
  cursor: pointer;
}
.pick-row:hover {
  background: var(--panel-hi);
}
.pick-row input {
  accent-color: var(--teal);
  width: 15px;
  height: 15px;
}
.pick-name {
  color: var(--text-hi);
  font-size: 13px;
}
.pick-empty {
  text-align: center;
  color: var(--text-faint);
  font-size: 12px;
  padding: 24px;
}
.dsl-block {
  margin-top: 18px;
}
.modal-foot {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 24px;
}
</style>
