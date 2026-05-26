<!-- 星系概览：planet 列表（状态 + 心跳）每个内嵌 agent 列表（状态） -->
<template>
  <div class="overview">
    <div class="ov-head">
      <div class="stats">
        <div class="stat">
          <span class="num mono">{{ planets.length }}</span>
          <span class="lbl eyebrow">行星 PLANETS</span>
        </div>
        <div class="stat">
          <span class="num mono teal">{{ onlinePlanets }}</span>
          <span class="lbl eyebrow">在线 ONLINE</span>
        </div>
        <div class="stat">
          <span class="num mono">{{ totalAgents }}</span>
          <span class="lbl eyebrow">智能体 AGENTS</span>
        </div>
      </div>
      <button class="btn btn-ghost btn-sm" @click="$emit('refresh')">↻ 刷新</button>
    </div>

    <div v-if="loading && !planets.length" class="muted">加载中…</div>
    <div v-else-if="!planets.length" class="empty panel">
      <p>该星系暂无行星接入。</p>
      <p class="mono sub">在行星主机上用 node_token 执行 orbit 接入后将在此显示。</p>
    </div>

    <div v-else class="planets">
      <section
        v-for="(p, i) in planets"
        :key="p.planet_uuid"
        class="planet panel rise"
        :style="{ animationDelay: i * 50 + 'ms' }"
      >
        <header class="p-head">
          <div class="p-id">
            <span :class="['dot', p.status]" />
            <h3>{{ p.hostname || '未命名行星' }}</h3>
            <span class="chip">{{ p.os || 'unknown os' }}</span>
          </div>
          <div class="p-meta mono">
            <span>{{ p.ip || '—' }}</span>
            <span class="sep">·</span>
            <span :class="statusClass(p.status)">{{ p.status }}</span>
            <span class="sep">·</span>
            <span>{{ heartbeat(p.last_heartbeat) }}</span>
          </div>
        </header>

        <div v-if="p.agents.length" class="agents">
          <div v-for="a in p.agents" :key="a.agent_uuid" class="agent">
            <span :class="['dot', a.status]" />
            <span class="a-name">{{ a.name }}</span>
            <span class="chip teal">{{ a.type }}</span>
            <span class="a-uuid mono">{{ a.agent_uuid.slice(0, 12) }}</span>
            <span :class="['a-status mono', statusClass(a.status)]">{{ a.status }}</span>
          </div>
        </div>
        <p v-else class="no-agent mono">该行星无注册智能体</p>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { GalaxyDetailResp } from '@/api/types'

const props = defineProps<{ detail: GalaxyDetailResp | null; loading: boolean }>()
defineEmits<{ (e: 'refresh'): void }>()

const planets = computed(() => props.detail?.planets ?? [])
const onlinePlanets = computed(() => planets.value.filter((p) => p.status === 'online').length)
const totalAgents = computed(() => planets.value.reduce((n, p) => n + p.agents.length, 0))

function statusClass(s: string) {
  if (s === 'online' || s === 'succeeded') return 'ok'
  if (s === 'failed' || s === 'error') return 'bad'
  if (s === 'running' || s === 'pending') return 'warn'
  return 'dim'
}

function heartbeat(ts: number) {
  if (!ts) return '无心跳'
  const diff = Date.now() / 1000 - ts
  if (diff < 60) return `${Math.floor(diff)}s 前`
  if (diff < 3600) return `${Math.floor(diff / 60)}m 前`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h 前`
  return `${Math.floor(diff / 86400)}d 前`
}
</script>

<style scoped>
.overview {
  flex: 1;
  overflow: auto;
  padding: 26px 32px 48px;
  max-width: 1080px;
  width: 100%;
  margin: 0 auto;
}
.ov-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 22px;
}
.stats {
  display: flex;
  gap: 34px;
}
.stat {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.num {
  font-size: 28px;
  font-weight: 600;
  color: var(--text-hi);
  line-height: 1;
}
.num.teal {
  color: var(--teal);
}
.lbl {
  font-size: 10px;
}
.planets {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.planet {
  padding: 20px 22px;
}
.p-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}
.p-id {
  display: flex;
  align-items: center;
  gap: 11px;
}
.p-id h3 {
  font-size: 17px;
}
.p-meta {
  display: flex;
  align-items: center;
  gap: 9px;
  font-size: 12px;
  color: var(--text-dim);
}
.p-meta .sep {
  color: var(--text-faint);
}
.agents {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  border-top: 1px solid var(--hairline);
  padding-top: 14px;
}
.agent {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 9px 10px;
  border-radius: var(--r-sm);
  transition: background 0.15s;
}
.agent:hover {
  background: var(--panel-hi);
}
.a-name {
  color: var(--text-hi);
  font-weight: 500;
  font-size: 14px;
}
.a-uuid {
  font-size: 11px;
  color: var(--text-faint);
  flex: 1;
}
.a-status {
  font-size: 11px;
}
.no-agent {
  margin: 14px 0 0;
  font-size: 12px;
  color: var(--text-faint);
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
.dim {
  color: var(--text-faint);
}
.empty {
  padding: 44px;
  text-align: center;
  color: var(--text-dim);
}
.empty .sub {
  font-size: 12px;
  color: var(--text-faint);
  margin-top: 8px;
}
.muted {
  color: var(--text-dim);
  padding: 20px;
}
</style>
