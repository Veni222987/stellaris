<!-- 星系管理：创建星系（拿 node_token）+ 查询本星系 agent 列表 -->
<template>
  <n-page-header>
    <template #title>星系管理</template>
  </n-page-header>
  <n-card style="margin:16px;">
    <n-flex>
      <n-input v-model:value="name" placeholder="星系名" />
      <n-button type="primary" @click="onCreate">创建星系</n-button>
    </n-flex>
    <n-table v-if="last" style="margin-top:16px;">
      <tbody>
        <tr><td>gid</td><td><code>{{ last.gid }}</code></td></tr>
        <tr><td>node_token（请妥善保存，仅本次返回）</td><td><code style="word-break:break-all;">{{ last.node_token }}</code></td></tr>
      </tbody>
    </n-table>
  </n-card>

  <n-card title="本星系 Agent" style="margin:16px;">
    <n-input v-model:value="qGid" placeholder="查询的 gid" />
    <n-button @click="loadAgents">刷新</n-button>
    <n-data-table :columns="cols" :data="agents" style="margin-top:12px;" />
  </n-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { api } from '@/api/client'
import type { GalaxyCreateResp, GalaxyAgent, GalaxyAgentsResp } from '@/api/types'

const name = ref('')
const last = ref<GalaxyCreateResp | null>(null)
const qGid = ref('')
const agents = ref<GalaxyAgent[]>([])

const cols = [
  { title: 'name', key: 'name' },
  { title: 'type', key: 'type' },
  { title: 'agent_uuid', key: 'agent_uuid' },
]

async function onCreate() {
  const { data } = await api.post<GalaxyCreateResp>('/api/galaxy/create', { name: name.value })
  last.value = data
  qGid.value = data.gid
  await loadAgents()
}

async function loadAgents() {
  if (!qGid.value) return
  const { data } = await api.get<GalaxyAgentsResp>(`/api/galaxy/${qGid.value}/agents`)
  agents.value = data.agents
}
</script>
