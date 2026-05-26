<!-- 通用顶栏：品牌标识 + 面包屑 slot + 退出 -->
<template>
  <header class="topbar">
    <div class="left" @click="router.push('/galaxies')">
      <span class="mark">
        <span class="ring" />
        <span class="core" />
      </span>
      <span class="wordmark">STELLARIS</span>
    </div>

    <nav class="crumbs">
      <slot name="crumbs" />
    </nav>

    <div class="right">
      <slot name="actions" />
      <button class="btn btn-ghost btn-sm" @click="logout">退出 · DISCONNECT</button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

function logout() {
  auth.clear()
  router.push('/login')
}
</script>

<style scoped>
.topbar {
  height: 60px;
  flex: none;
  display: flex;
  align-items: center;
  gap: 18px;
  padding: 0 24px;
  border-bottom: 1px solid var(--hairline);
  background: rgba(7, 11, 21, 0.7);
  backdrop-filter: blur(12px);
}
.left {
  display: flex;
  align-items: center;
  gap: 11px;
  cursor: pointer;
  flex: none;
}
.mark {
  position: relative;
  width: 26px;
  height: 26px;
}
.mark .ring {
  position: absolute;
  inset: 0;
  border: 1.4px solid var(--teal-line);
  border-radius: 50%;
  border-top-color: transparent;
  border-right-color: transparent;
  animation: spin 6s linear infinite;
}
.mark .core {
  position: absolute;
  inset: 8px;
  border-radius: 50%;
  background: radial-gradient(circle at 35% 30%, #aefff0, var(--teal));
  box-shadow: 0 0 14px -1px var(--teal);
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
.wordmark {
  font-family: var(--display);
  font-weight: 700;
  letter-spacing: 3px;
  color: var(--text-hi);
  font-size: 16px;
}
.crumbs {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}
.right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: none;
}
</style>
