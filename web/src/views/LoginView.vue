<!-- 登录页：全局单一账号（后台 .env 配置），无注册 -->
<template>
  <div class="login-wrap">
    <div class="access rise panel">
      <div class="brand">
        <div class="logo-mark">
          <span class="ring" />
          <span class="core" />
        </div>
        <div>
          <h1>STELLARIS</h1>
          <p class="eyebrow">星海调度中心 · MISSION CONTROL</p>
        </div>
      </div>

      <form class="form" @submit.prevent="onLogin">
        <div>
          <label class="field-label">操作员邮箱 · OPERATOR ID</label>
          <input
            v-model="email"
            class="input mono"
            type="email"
            placeholder="admin@stellaris.local"
            autocomplete="username"
          />
        </div>
        <div>
          <label class="field-label">通行密钥 · ACCESS KEY</label>
          <input
            v-model="password"
            class="input mono"
            type="password"
            placeholder="••••••••"
            autocomplete="current-password"
          />
        </div>
        <button class="btn btn-primary link-btn" type="submit" :disabled="loading">
          {{ loading ? '建立链路…' : '建立链路 · ESTABLISH LINK' }}
        </button>
      </form>

      <p class="hint mono">单一账号由后台 .env（ADMIN_EMAIL / ADMIN_PASSWORD）配置，不开放注册</p>
      <span class="scanline" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import type { AuthResp } from '@/api/types'

const email = ref('')
const password = ref('')
const loading = ref(false)
const auth = useAuthStore()
const router = useRouter()
const msg = useMessage()

async function onLogin() {
  if (!email.value || !password.value) {
    msg.warning('请输入邮箱与密钥')
    return
  }
  loading.value = true
  try {
    const { data } = await api.post<AuthResp>('/api/auth/login', {
      email: email.value,
      password: password.value,
    })
    auth.setToken(data.token)
    router.push('/galaxies')
  } catch (e: any) {
    msg.error(e.response?.data?.msg || '凭据无效，链路建立失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-wrap {
  height: 100%;
  display: grid;
  place-items: center;
  padding: 24px;
}
.access {
  position: relative;
  width: 420px;
  max-width: 100%;
  padding: 40px 38px 30px;
  overflow: hidden;
  box-shadow: var(--shadow);
}
.brand {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 34px;
}
.brand h1 {
  font-size: 30px;
  letter-spacing: 4px;
}
.logo-mark {
  position: relative;
  width: 46px;
  height: 46px;
  flex: none;
}
.logo-mark .ring {
  position: absolute;
  inset: 0;
  border: 1.5px solid var(--teal-line);
  border-radius: 50%;
  border-top-color: transparent;
  border-right-color: transparent;
  animation: spin 6s linear infinite;
}
.logo-mark .core {
  position: absolute;
  inset: 14px;
  border-radius: 50%;
  background: radial-gradient(circle at 35% 30%, #aefff0, var(--teal));
  box-shadow: 0 0 22px -2px var(--teal);
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
.form {
  display: flex;
  flex-direction: column;
  gap: 18px;
}
.link-btn {
  margin-top: 6px;
  padding: 13px;
  width: 100%;
}
.hint {
  margin-top: 22px;
  font-size: 11px;
  color: var(--text-faint);
  line-height: 1.6;
  text-align: center;
}
.scanline {
  position: absolute;
  left: 0;
  right: 0;
  top: 0;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--teal), transparent);
  opacity: 0.7;
  animation: scan 4.5s ease-in-out infinite;
}
@keyframes scan {
  0%,
  100% {
    transform: translateY(0);
    opacity: 0;
  }
  50% {
    transform: translateY(440px);
    opacity: 0.6;
  }
}
</style>
