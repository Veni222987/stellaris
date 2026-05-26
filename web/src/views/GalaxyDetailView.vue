<!-- 星系详情：顶部信息 + 横向切换 概览 / 聊天 -->
<template>
  <div class="page">
    <AppHeader>
      <template #crumbs>
        <span class="crumb-link" @click="router.push('/galaxies')">星系</span>
        <span class="sep">/</span>
        <span class="crumb-cur">{{ detail?.name || gid }}</span>
        <span class="gid mono">{{ gid }}</span>
      </template>
      <template #actions>
        <div class="tabs">
          <button :class="['tab', { on: tab === 'overview' }]" @click="tab = 'overview'">概览</button>
          <button :class="['tab', { on: tab === 'chat' }]" @click="tab = 'chat'">聊天</button>
          <span class="tab-ind" :style="{ transform: tab === 'chat' ? 'translateX(100%)' : 'none' }" />
        </div>
      </template>
    </AppHeader>

    <main class="body">
      <GalaxyOverview
        v-show="tab === 'overview'"
        :detail="detail"
        :loading="loading"
        @refresh="loadDetail"
      />
      <GalaxyChat v-if="detail" v-show="tab === 'chat'" :gid="gid" :detail="detail" />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppHeader from '@/components/AppHeader.vue'
import GalaxyOverview from '@/components/GalaxyOverview.vue'
import GalaxyChat from '@/components/GalaxyChat.vue'
import { api } from '@/api/client'
import type { GalaxyDetailResp } from '@/api/types'

const route = useRoute()
const router = useRouter()
const gid = computed(() => route.params.gid as string)
const detail = ref<GalaxyDetailResp | null>(null)
const loading = ref(false)
const tab = ref<'overview' | 'chat'>('overview')

async function loadDetail() {
  loading.value = true
  try {
    const { data } = await api.get<GalaxyDetailResp>(`/api/galaxy/${gid.value}/detail`)
    detail.value = data
  } finally {
    loading.value = false
  }
}

onMounted(loadDetail)
</script>

<style scoped>
.page {
  height: 100%;
  display: flex;
  flex-direction: column;
}
.body {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.crumb-link {
  font-family: var(--display);
  color: var(--text-dim);
  cursor: pointer;
  font-size: 14px;
}
.crumb-link:hover {
  color: var(--text-hi);
}
.sep {
  color: var(--text-faint);
}
.crumb-cur {
  font-family: var(--display);
  color: var(--text-hi);
  font-size: 14px;
  font-weight: 600;
}
.gid {
  font-size: 11px;
  color: var(--text-faint);
  margin-left: 4px;
}
.tabs {
  position: relative;
  display: flex;
  background: rgba(4, 7, 14, 0.6);
  border: 1px solid var(--hairline);
  border-radius: var(--r);
  padding: 3px;
}
.tab {
  position: relative;
  z-index: 1;
  font-family: var(--display);
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.5px;
  padding: 7px 22px;
  border: none;
  background: transparent;
  color: var(--text-dim);
  cursor: pointer;
  transition: color 0.2s;
}
.tab.on {
  color: var(--ink);
}
.tab-ind {
  position: absolute;
  z-index: 0;
  top: 3px;
  bottom: 3px;
  left: 3px;
  width: calc(50% - 3px);
  border-radius: 7px;
  background: linear-gradient(180deg, #5cf0d8, var(--teal));
  box-shadow: 0 0 16px -2px var(--teal);
  transition: transform 0.28s cubic-bezier(0.22, 1, 0.36, 1);
}
</style>
