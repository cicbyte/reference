import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

interface FooterItem {
  key: string
  label: string
  value: string
  icon?: any
  /** Optional tone for badges (health state). */
  tone?: 'ok' | 'warn' | 'bad' | 'default'
}

const SIDEBAR_DEFAULT_KEY = 'reference-sidebar-collapsed'

export const useLayoutStore = defineStore('layout', () => {
  /** 启动偏好：下次启动时侧边栏是否默认折叠（独立于运行时 sidebarCollapsed） */
  const sidebarDefaultCollapsed = ref(localStorage.getItem(SIDEBAR_DEFAULT_KEY) === 'true')
  /** 运行时折叠状态：启动按偏好初始化，之后由导航栏按钮实时切换，不回写偏好 */
  const sidebarCollapsed = ref(false)
  const projectRailCollapsed = ref(false)
  const projectRailVisible = ref(true)  // fully user-controlled
  const windowWidth = ref(window.innerWidth)

  // Footer status items — views write to this, MainContent renders it.
  // Each item: { key, icon?, label, value, tone? }
  const footerItems = ref<FooterItem[]>([])

  function setFooterItem(key: string, label: string, value: string, icon?: any, tone?: 'ok' | 'warn' | 'bad' | 'default') {
    const idx = footerItems.value.findIndex((i) => i.key === key)
    const item: FooterItem = { key, label, value, icon, tone }
    if (idx >= 0) footerItems.value[idx] = item
    else footerItems.value.push(item)
  }

  function clearFooterItem(key: string) {
    footerItems.value = footerItems.value.filter((i) => i.key !== key)
  }

  function clearAllFooterItems() {
    footerItems.value = []
  }

  const isMobile = computed(() => windowWidth.value < 768)
  const isTablet = computed(() => windowWidth.value >= 768 && windowWidth.value < 1024)

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  /** 设置启动偏好（仅持久化 + 响应式更新，不改当前运行时折叠状态） */
  function setSidebarDefaultCollapsed(v: boolean) {
    sidebarDefaultCollapsed.value = v
    localStorage.setItem(SIDEBAR_DEFAULT_KEY, String(v))
  }

  function toggleProjectRail() {
    projectRailCollapsed.value = !projectRailCollapsed.value
  }

  function toggleProjectRailVisible() {
    projectRailVisible.value = !projectRailVisible.value
  }

  function handleResize() {
    windowWidth.value = window.innerWidth
    if (isMobile.value) {
      sidebarCollapsed.value = true
      projectRailCollapsed.value = true
    }
  }

  return {
    sidebarCollapsed,
    sidebarDefaultCollapsed,
    projectRailCollapsed,
    projectRailVisible,
    windowWidth,
    footerItems,
    isMobile,
    isTablet,
    toggleSidebar,
    setSidebarDefaultCollapsed,
    toggleProjectRail,
    toggleProjectRailVisible,
    handleResize,
    setFooterItem,
    clearFooterItem,
    clearAllFooterItems,
  }
})
