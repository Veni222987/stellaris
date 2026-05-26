// 路由配置与登录守卫
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: () => import('@/views/LoginView.vue') },
    { path: '/', redirect: '/galaxies' },
    { path: '/galaxies', component: () => import('@/views/GalaxiesView.vue'), meta: { auth: true } },
    { path: '/galaxy/:gid', component: () => import('@/views/GalaxyDetailView.vue'), meta: { auth: true } },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.auth && !auth.token) return { path: '/login' }
  if (to.path === '/login' && auth.token) return { path: '/galaxies' }
})

export default router
