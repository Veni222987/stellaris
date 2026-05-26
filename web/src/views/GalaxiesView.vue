<!-- 星系列表：卡片网格 + 顶部创建星系，点击进入详情 -->
<template>
  <div class="page">
    <AppHeader>
      <template #crumbs>
        <span class="eyebrow">星系总览 · GALAXY REGISTRY</span>
      </template>
    </AppHeader>

    <main class="body">
      <!-- 创建栏 -->
      <section class="create panel rise">
        <div class="create-head">
          <h2>部署新星系</h2>
          <p class="sub mono">创建后将返回一次性 node_token，用于行星 orbit 接入</p>
        </div>
        <div class="create-row">
          <input
            v-model="name"
            class="input"
            placeholder="星系名称，如 production-cluster"
            @keydown.enter="onCreate"
          />
          <button class="btn btn-primary" :disabled="creating || !name.trim()" @click="onCreate">
            {{ creating ? '部署中…' : '+ 创建星系' }}
          </button>
        </div>
      </section>

      <!-- 列表 -->
      <div class="list-head">
        <span class="eyebrow">我的星系 · {{ galaxies.length }} ACTIVE</span>
        <button class="btn btn-ghost btn-sm" @click="loadGalaxies">↻ 刷新</button>
      </div>

      <div v-if="galaxies.length" class="grid">
        <button
          v-for="(g, i) in galaxies"
          :key="g.gid"
          class="card rise"
          :style="{ animationDelay: i * 50 + 'ms' }"
          @click="open(g.gid)"
        >
          <div class="card-orbit"><span class="planet" /></div>
          <div class="card-body">
            <h3>{{ g.name }}</h3>
            <span class="gid mono">GID · {{ g.gid }}</span>
          </div>
          <span class="enter mono">进入 →</span>
        </button>
      </div>
      <div v-else class="empty panel">
        <p>尚无星系。在上方部署你的第一个星系。</p>
      </div>
    </main>

    <!-- node_token 模态 -->
    <div v-if="lastToken" class="modal-mask" @click.self="lastToken = null">
      <div class="modal panel rise">
        <h2>星系已部署</h2>
        <p class="sub mono">node_token 仅本次显示，请立即妥善保存</p>
        <label class="field-label">GID</label>
        <div class="token-box mono">{{ lastToken.gid }}</div>
        <label class="field-label">NODE TOKEN</label>
        <div class="token-box mono token">{{ lastToken.node_token }}</div>
        <div class="modal-foot">
          <button class="btn btn-ghost btn-sm" @click="copyToken">复制 token</button>
          <button class="btn btn-primary btn-sm" @click="open(lastToken.gid)">进入星系 →</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import AppHeader from '@/components/AppHeader.vue'
import { api } from '@/api/client'
import type { GalaxyCreateResp, GalaxyItem, GalaxyListResp } from '@/api/types'

const router = useRouter()
const msg = useMessage()
const name = ref('')
const creating = ref(false)
const galaxies = ref<GalaxyItem[]>([])
const lastToken = ref<GalaxyCreateResp | null>(null)

async function loadGalaxies() {
  const { data } = await api.get<GalaxyListResp>('/api/galaxy/list')
  galaxies.value = data.galaxies ?? []
}

async function onCreate() {
  if (!name.value.trim()) return
  creating.value = true
  try {
    const { data } = await api.post<GalaxyCreateResp>('/api/galaxy/create', { name: name.value.trim() })
    lastToken.value = data
    name.value = ''
    await loadGalaxies()
  } catch (e: any) {
    msg.error(e.response?.data?.msg || '创建失败')
  } finally {
    creating.value = false
  }
}

function copyToken() {
  if (!lastToken.value) return
  navigator.clipboard?.writeText(lastToken.value.node_token)
  msg.success('node_token 已复制')
}

function open(gid: string) {
  router.push(`/galaxy/${gid}`)
}

onMounted(loadGalaxies)
</script>

<style scoped>
.page {
  height: 100%;
  display: flex;
  flex-direction: column;
}
.body {
  flex: 1;
  overflow: auto;
  padding: 28px 32px 48px;
  max-width: 1080px;
  width: 100%;
  margin: 0 auto;
}
.create {
  padding: 24px 26px;
  margin-bottom: 30px;
}
.create-head h2 {
  font-size: 19px;
}
.sub {
  font-size: 12px;
  color: var(--text-dim);
  margin: 6px 0 0;
}
.create-row {
  display: flex;
  gap: 12px;
  margin-top: 18px;
}
.create-row .input {
  flex: 1;
}
.list-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}
.card {
  position: relative;
  text-align: left;
  display: flex;
  align-items: center;
  gap: 18px;
  padding: 22px;
  border-radius: var(--r-lg);
  border: 1px solid var(--hairline);
  background: var(--panel);
  backdrop-filter: blur(10px);
  cursor: pointer;
  transition: border-color 0.2s, transform 0.2s, box-shadow 0.2s;
  font-family: inherit;
  color: inherit;
  overflow: hidden;
}
.card:hover {
  border-color: var(--teal-line);
  transform: translateY(-3px);
  box-shadow: var(--glow-teal);
}
.card-orbit {
  position: relative;
  width: 48px;
  height: 48px;
  flex: none;
  border: 1px dashed var(--hairline-hi);
  border-radius: 50%;
}
.card-orbit .planet {
  position: absolute;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  top: -7px;
  left: 50%;
  margin-left: -7px;
  background: radial-gradient(circle at 35% 30%, #c9b3ff, var(--plasma));
  box-shadow: 0 0 14px -1px var(--plasma);
}
.card:hover .card-orbit {
  animation: orbit 8s linear infinite;
}
@keyframes orbit {
  to {
    transform: rotate(360deg);
  }
}
.card-body {
  flex: 1;
  min-width: 0;
}
.card-body h3 {
  font-size: 18px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.gid {
  font-size: 11px;
  color: var(--text-dim);
  margin-top: 5px;
  display: block;
}
.enter {
  font-size: 12px;
  color: var(--teal);
  opacity: 0;
  transition: opacity 0.2s;
  flex: none;
}
.card:hover .enter {
  opacity: 1;
}
.empty {
  padding: 48px;
  text-align: center;
  color: var(--text-dim);
}
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
  width: 520px;
  max-width: 100%;
  padding: 30px;
  box-shadow: var(--shadow);
}
.modal h2 {
  font-size: 20px;
}
.token-box {
  background: rgba(4, 7, 14, 0.7);
  border: 1px solid var(--hairline);
  border-radius: var(--r);
  padding: 11px 14px;
  font-size: 13px;
  color: var(--text-hi);
  margin-bottom: 16px;
  word-break: break-all;
}
.token-box.token {
  color: var(--amber);
}
.modal-foot {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 6px;
}
</style>
