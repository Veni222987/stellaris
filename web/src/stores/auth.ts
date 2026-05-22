// 鉴权 store：保存 / 清除 token，自动同步 localStorage
import { defineStore } from 'pinia'
import { ref } from 'vue'

const TOKEN_KEY = 'stellaris-token'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem(TOKEN_KEY) || '')

  function setToken(t: string) {
    token.value = t
    localStorage.setItem(TOKEN_KEY, t)
  }
  function clear() {
    token.value = ''
    localStorage.removeItem(TOKEN_KEY)
  }

  return { token, setToken, clear }
})
