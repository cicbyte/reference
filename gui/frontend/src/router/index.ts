import { createRouter, createWebHashHistory } from 'vue-router'

// projectScoped: true  → left project rail is shown, operations target the
//                        selected project (Dashboard / repos / scc / doctor)
// projectScoped: false → global-level page (global stats / GC / wiki / settings),
//                        rail hidden, two-column layout
// titleKey: i18n key for the breadcrumb title (translated at render time)
const routes = [
  { path: '/', name: 'dashboard', component: () => import('@/views/dashboard/index.vue'), meta: { titleKey: 'sidebar.dashboard', projectScoped: true } },
  { path: '/repos', name: 'repos', component: () => import('@/views/repos/index.vue'), meta: { titleKey: 'sidebar.refList', projectScoped: true } },
  { path: '/global', name: 'global', component: () => import('@/views/global-list/index.vue'), meta: { titleKey: 'sidebar.projectList', projectScoped: false } },
  { path: '/cache', name: 'cache', component: () => import('@/views/cache/index.vue'), meta: { titleKey: 'sidebar.cache', projectScoped: false } },
  { path: '/global/stats', name: 'global-stats', component: () => import('@/views/global-stats/index.vue'), meta: { titleKey: 'sidebar.stats', projectScoped: false } },
  { path: '/global/gc', name: 'global-gc', component: () => import('@/views/global-gc/index.vue'), meta: { titleKey: 'sidebar.gc', projectScoped: false } },
  { path: '/wiki', name: 'wiki', component: () => import('@/views/wiki/index.vue'), meta: { titleKey: 'sidebar.remoteWiki', projectScoped: false } },
  { path: '/local-wiki', name: 'local-wiki', component: () => import('@/views/local-wiki/index.vue'), meta: { titleKey: 'sidebar.localWiki', projectScoped: false } },
  { path: '/settings', name: 'settings', component: () => import('@/views/settings/index.vue'), meta: { titleKey: 'sidebar.settings', projectScoped: false } },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

export default router
