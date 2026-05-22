<!-- 左栏：会话列表 + 新建按钮 -->
<template>
  <div style="padding:12px; display:flex; flex-direction:column; gap:8px;">
    <n-button type="primary" block @click="$emit('new')">+ 新建会话</n-button>
    <n-list bordered>
      <n-list-item
        v-for="s in store.list"
        :key="s.session_uuid"
        @click="store.active = s.session_uuid"
        :style="{ cursor:'pointer', background: store.active === s.session_uuid ? 'var(--n-color-target)' : '' }"
      >
        <n-thing :title="s.title" :description="`mode=${s.mode} · ${s.agent_uuids.length} agents`" />
      </n-list-item>
    </n-list>
  </div>
</template>

<script setup lang="ts">
import { useSessionStore } from '@/stores/session'
defineEmits<{ (e: 'new'): void }>()
const store = useSessionStore()
</script>
