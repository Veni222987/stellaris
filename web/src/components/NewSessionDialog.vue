<!-- 新建会话弹窗：选 gid、模式、勾选 agent；orchestration 模式额外提供 DSL 输入框 -->
<template>
  <n-modal v-model:show="show" preset="card" title="新建会话" style="width:520px">
    <n-form>
      <n-form-item label="星系 gid">
        <n-input v-model:value="gid" placeholder="16 位十六进制" />
      </n-form-item>
      <n-form-item label="模式">
        <n-select v-model:value="mode" :options="modeOptions" />
      </n-form-item>
      <n-form-item v-if="mode === 'orchestration'" label="DSL (JSON)">
        <n-input
          v-model:value="dsl"
          type="textarea"
          :rows="8"
          placeholder='{"nodes":[{"id":"a","agent_uuid":"...","prompt":"{{user}}"}]}'
        />
      </n-form-item>
      <n-form-item label="可用 Agent">
        <n-button size="small" @click="loadAgents">刷新</n-button>
      </n-form-item>
      <n-checkbox-group v-model:value="selected">
        <n-space vertical>
          <n-checkbox v-for="a in agents" :key="a.agent_uuid" :value="a.agent_uuid">
            {{ a.name }} <n-tag size="small">{{ a.type }}</n-tag>
          </n-checkbox>
        </n-space>
      </n-checkbox-group>
    </n-form>
    <template #footer>
      <n-button type="primary" :disabled="!selected.length || !gid" @click="onCreate">创建</n-button>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useMessage } from 'naive-ui'
import { api } from '@/api/client'
import { useSessionStore } from '@/stores/session'
import type { GalaxyAgent, GalaxyAgentsResp, SessionCreateResp } from '@/api/types'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ (e: 'update:open', v: boolean): void }>()
// 双向同步：父 v-model:open 与 n-modal v-model:show
const show = ref(props.open)
watch(() => props.open, (v) => (show.value = v))
watch(show, (v) => emit('update:open', v))

const gid = ref('')
const mode = ref('parallel')
const dsl = ref('')
// M3 三种调度模式都已上线。
const modeOptions = [
  { label: 'Parallel', value: 'parallel' },
  { label: 'Relay', value: 'relay' },
  { label: 'Orchestration (DSL)', value: 'orchestration' },
]
const agents = ref<GalaxyAgent[]>([])
const selected = ref<string[]>([])
const msg = useMessage()
const store = useSessionStore()

async function loadAgents() {
  if (!gid.value) { msg.warning('请先输入 gid'); return }
  const { data } = await api.get<GalaxyAgentsResp>(`/api/galaxy/${gid.value}/agents`)
  agents.value = data.agents
}

async function onCreate() {
  const { data } = await api.post<SessionCreateResp>('/api/session/create', {
    gid: gid.value,
    mode: mode.value,
    agent_uuids: selected.value,
    dsl: mode.value === 'orchestration' ? dsl.value : undefined,
  })
  store.addLocal({
    session_uuid: data.session_uuid,
    gid: gid.value, mode: mode.value, agent_uuids: selected.value,
    title: `${mode.value} · ${selected.value.length}`,
  })
  show.value = false
}
</script>
