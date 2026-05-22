<!-- 右栏：任务历史，按 task_uuid 列出状态与 chunk 计数；relay/orchestration 模式展示 parent + node_id -->
<template>
  <div style="padding:12px; overflow:auto; height:100%;">
    <div v-for="t in tasks" :key="t.task_uuid" style="margin-bottom:16px;">
      <n-tag :type="t.status === 'succeeded' ? 'success' : t.status === 'failed' ? 'error' : 'warning'">
        {{ t.status }}
      </n-tag>
      <span style="margin-left:8px; font-family:monospace; font-size:12px;">{{ t.task_uuid.slice(0, 8) }}</span>
      <div style="margin-left:8px; font-size:12px; color:#888;">
        agent={{ t.agent_uuid.slice(0, 8) }} · chunks={{ t.chunks.length }}
      </div>
      <div v-if="t.parent_task_uuid" style="margin-left:8px; font-size:11px; color:#666;">
        ↑ parent {{ t.parent_task_uuid.slice(0, 8) }}
      </div>
      <div v-if="t.node_id" style="margin-left:8px; font-size:11px; color:#888;">
        node: {{ t.node_id }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useSessionStore } from '@/stores/session'

const props = defineProps<{ sessionUUID: string }>()
const store = useSessionStore()
const tasks = computed(() => store.tasksBySession[props.sessionUUID] || [])
</script>
