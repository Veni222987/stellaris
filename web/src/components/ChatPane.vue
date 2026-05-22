<!-- 中间：消息气泡区 + 输入框（Enter 发送，Shift+Enter 换行） -->
<template>
  <div style="display:flex; flex-direction:column; height:100%;">
    <div style="flex:1; overflow:auto; padding:16px;">
      <div v-for="(b, i) in bubbles" :key="i" style="margin-bottom:12px;">
        <n-tag type="success">{{ b.agent }}</n-tag>
        <n-tag size="small" :type="b.status === 'succeeded' ? 'success' : b.status === 'failed' ? 'error' : 'warning'">
          {{ b.status }}
        </n-tag>
        <div style="white-space:pre-wrap; padding:8px 0;">{{ b.content }}</div>
      </div>
    </div>
    <n-input-group>
      <n-input
        v-model:value="input"
        type="textarea"
        :autosize="{ minRows: 1, maxRows: 6 }"
        placeholder="Enter 发送，Shift+Enter 换行"
        @keydown="onKeydown"
      />
      <n-button type="primary" :loading="sending" @click="onSend">发送</n-button>
    </n-input-group>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { api } from '@/api/client'
import { useSessionStore } from '@/stores/session'

const props = defineProps<{ sessionUUID: string }>()
const input = ref('')
const sending = ref(false)
const store = useSessionStore()

// 多 agent 按 agent_uuid 分轨；每个 task 自成一个 bubble。
const bubbles = computed(() => {
  const tasks = store.tasksBySession[props.sessionUUID] || []
  return tasks.map(t => ({
    role: 'assistant',
    agent: t.agent_uuid.slice(0, 8),
    status: t.status,
    content: t.chunks.map(c => c.chunk).join(''),
  }))
})

// Enter 发送、Shift+Enter 换行；中文输入法选词时按 Enter 不能触发发送。
function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey && !e.isComposing) {
    e.preventDefault()
    onSend()
  }
}

async function onSend() {
  if (!input.value.trim()) return
  sending.value = true
  try {
    await api.post(`/api/session/${props.sessionUUID}/message`, { content: input.value })
    input.value = ''
  } finally {
    sending.value = false
  }
}
</script>
