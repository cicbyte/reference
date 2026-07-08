import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
  { path: '/', name: 'dashboard', component: () => import('../views/DashboardView.vue'), meta: { title: 'Dashboard' } },
  { path: '/repos', name: 'repos', component: () => import('../views/RepoListView.vue'), meta: { title: '仓库列表' } },
  { path: '/repos/add', name: 'repo-add', component: () => import('../views/RepoAddView.vue'), meta: { title: '添加仓库' } },
  { path: '/scc', name: 'scc', component: () => import('../views/SccView.vue'), meta: { title: '代码统计' } },
  { path: '/doctor', name: 'doctor', component: () => import('../views/DoctorView.vue'), meta: { title: '诊断修复' } },
  { path: '/global', name: 'global', component: () => import('../views/GlobalListView.vue'), meta: { title: '全局项目' } },
  { path: '/global/stats', name: 'global-stats', component: () => import('../views/GlobalStatsView.vue'), meta: { title: '全局统计' } },
  { path: '/global/gc', name: 'global-gc', component: () => import('../views/GlobalGCView.vue'), meta: { title: '垃圾回收' } },
  { path: '/wiki', name: 'wiki', component: () => import('../views/WikiBrowseView.vue'), meta: { title: '知识库' } },
  { path: '/wiki/sync', name: 'wiki-sync', component: () => import('../views/WikiSyncView.vue'), meta: { title: '知识同步' } },
  { path: '/settings', name: 'settings', component: () => import('../views/SettingsView.vue'), meta: { title: '设置' } },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

export default router
