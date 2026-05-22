<!-- 登录 / 注册页：调用 /api/auth/login | /api/auth/register -->
<template>
  <n-flex justify="center" align="center" style="height:100vh">
    <n-card title="Stellaris 登录" style="width:380px">
      <n-form @submit.prevent="onSubmit">
        <n-form-item label="邮箱"><n-input v-model:value="email" /></n-form-item>
        <n-form-item label="密码"><n-input v-model:value="password" type="password" /></n-form-item>
        <n-space>
          <n-button type="primary" attr-type="submit" :loading="loading">登录</n-button>
          <n-button @click="onRegister" :loading="loading">注册</n-button>
        </n-space>
      </n-form>
    </n-card>
  </n-flex>
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

async function call(path: '/api/auth/login' | '/api/auth/register') {
  loading.value = true
  try {
    const { data } = await api.post<AuthResp>(path, { email: email.value, password: password.value })
    auth.setToken(data.token)
    router.push('/chat')
  } catch (e: any) {
    msg.error(e.response?.data?.msg || e.message)
  } finally {
    loading.value = false
  }
}
const onSubmit = () => call('/api/auth/login')
const onRegister = () => call('/api/auth/register')
</script>
