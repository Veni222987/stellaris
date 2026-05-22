// 路由配置与登录守卫
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: () => import('@/views/LoginView.vue') },
    { path: '/', redirect: '/chat' },
    { path: '/chat', component: () => import('@/views/ChatView.vue'), meta: { auth: true } },
    { path: '/galaxy', component: () => import('@/views/GalaxyView.vue'), meta: { auth: true } },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.auth && !auth.token) return { path: '/login' }
  if (to.path === '/login' && auth.token) return { path: '/chat' }
})

export default router
