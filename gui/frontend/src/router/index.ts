import { createRouter, createWebHashHistory } from 'vue-router'

// projectScoped: true  → left project rail is shown, operations target the
//                        selected project (Dashboard / repos / scc / doctor)
// projectScoped: false → global-level page (global stats / GC / wiki / settings),
//                        rail hidden, two-column layout
const routes = [
  { path: '/', name: 'dashboard', component: () => import('../views/DashboardView.vue'), meta: { title: 'Dashboard', projectScoped: true } },
  { path: '/repos', name: 'repos', component: () => import('../views/RepoListView.vue'), meta: { title: '仓库列表', projectScoped: true } },
  { path: '/repos/browse/:name', name: 'repo-browse', component: () => import('../views/BrowseRepoView.vue'), meta: { title: '代码浏览', projectScoped: true } },
  { path: '/doctor', name: 'doctor', component: () => import('../views/DoctorView.vue'), meta: { title: '诊断修复', projectScoped: true } },
  { path: '/global', name: 'global', component: () => import('../views/GlobalListView.vue'), meta: { title: '全局项目', projectScoped: false } },
  { path: '/cache', name: 'cache', component: () => import('../views/CacheReposView.vue'), meta: { title: '仓库缓存', projectScoped: false } },
  { path: '/global/stats', name: 'global-stats', component: () => import('../views/GlobalStatsView.vue'), meta: { title: '全局统计', projectScoped: false } },
  { path: '/global/gc', name: 'global-gc', component: () => import('../views/GlobalGCView.vue'), meta: { title: '垃圾回收', projectScoped: false } },
  { path: '/wiki', name: 'wiki', component: () => import('../views/WikiBrowseView.vue'), meta: { title: '知识库', projectScoped: false } },
  { path: '/wiki/sync', name: 'wiki-sync', component: () => import('../views/WikiSyncView.vue'), meta: { title: '知识同步', projectScoped: false } },
  { path: '/settings', name: 'settings', component: () => import('../views/SettingsView.vue'), meta: { title: '设置', projectScoped: false } },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

export default router
