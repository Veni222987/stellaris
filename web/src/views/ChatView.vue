<!-- 三栏聊天主界面：左会话列表 / 中聊天 / 右任务历史；切换 active 时重连 ws -->
<template>
  <n-layout has-sider style="height:100vh;">
    <n-layout-sider :width="260" bordered>
      <SessionList @new="dlgOpen = true" />
    </n-layout-sider>
    <n-layout>
      <n-layout-content>
        <ChatPane v-if="store.active" :session-u-u-i-d="store.active" />
        <n-empty v-else description="请新建或选择一个会话" style="margin-top:120px;" />
      </n-layout-content>
    </n-layout>
    <n-layout-sider :width="320" bordered>
      <TaskHistoryPane v-if="store.active" :session-u-u-i-d="store.active" />
    </n-layout-sider>
  </n-layout>
  <NewSessionDialog v-model:open="dlgOpen" />
</template>

<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import SessionList from '@/components/SessionList.vue'
import ChatPane from '@/components/ChatPane.vue'
import TaskHistoryPane from '@/components/TaskHistoryPane.vue'
import NewSessionDialog from '@/components/NewSessionDialog.vue'
import { useSessionStore } from '@/stores/session'
import { openSessionWS } from '@/ws/client'

const store = useSessionStore()
const dlgOpen = ref(false)
let currentWS: WebSocket | null = null

// 切换 active 时：关闭旧连接，按需异步拿 ticket 再开新连接
watch(() => store.active, async (uuid) => {
  if (currentWS) { currentWS.close(); currentWS = null }
  if (!uuid) return
  currentWS = await openSessionWS(uuid, (c) => store.appendChunk(uuid, c))
}, { immediate: true })

onUnmounted(() => { currentWS?.close() })
</script>
