import { createRouter, createWebHashHistory } from 'vue-router'

// projectScoped: true  → left project rail is shown, operations target the
//                        selected project (Dashboard / repos / scc / doctor)
// projectScoped: false → global-level page (global stats / GC / wiki / settings),
//                        rail hidden, two-column layout
const routes = [
  { path: '/', name: 'dashboard', component: () => import('@/views/dashboard/index.vue'), meta: { title: '仪表盘', projectScoped: true } },
  { path: '/repos', name: 'repos', component: () => import('@/views/repos/index.vue'), meta: { title: '引用列表', projectScoped: true } },
  { path: '/global', name: 'global', component: () => import('@/views/global-list/index.vue'), meta: { title: '全局项目', projectScoped: false } },
  { path: '/cache', name: 'cache', component: () => import('@/views/cache/index.vue'), meta: { title: '仓库缓存', projectScoped: false } },
  { path: '/global/stats', name: 'global-stats', component: () => import('@/views/global-stats/index.vue'), meta: { title: '全局统计', projectScoped: false } },
  { path: '/global/gc', name: 'global-gc', component: () => import('@/views/global-gc/index.vue'), meta: { title: '垃圾回收', projectScoped: false } },
  { path: '/wiki', name: 'wiki', component: () => import('@/views/wiki/index.vue'), meta: { title: '知识库', projectScoped: false } },
  { path: '/local-wiki', name: 'local-wiki', component: () => import('@/views/local-wiki/index.vue'), meta: { title: '本地知识库', projectScoped: false } },
  { path: '/settings', name: 'settings', component: () => import('@/views/settings/index.vue'), meta: { title: '设置', projectScoped: false } },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

export default router
