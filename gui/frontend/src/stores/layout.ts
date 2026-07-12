import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

interface FooterItem {
  key: string
  label: string
  value: string
  icon?: any
}

export const useLayoutStore = defineStore('layout', () => {
  const sidebarCollapsed = ref(false)
  const projectRailCollapsed = ref(false)
  // 'default' = follow route meta; 'show' = force show; 'hide' = force hide
  const projectRailOverride = ref('default')
  const windowWidth = ref(window.innerWidth)

  // Footer status items — views write to this, MainContent renders it.
  // Each item: { key, icon?, label, value }
  const footerItems = ref<FooterItem[]>([])

  function setFooterItem(key: string, label: string, value: string, icon?: any) {
    const idx = footerItems.value.findIndex((i) => i.key === key)
    const item: FooterItem = { key, label, value, icon }
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

  function toggleProjectRail() {
    projectRailCollapsed.value = !projectRailCollapsed.value
  }

  function toggleProjectRailVisible() {
    projectRailOverride.value = projectRailOverride.value === 'hide' ? 'default' : 'hide'
  }

  function setProjectRailVisible(visible: boolean) {
    projectRailOverride.value = visible ? 'show' : 'hide'
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
    projectRailCollapsed,
    projectRailOverride,
    windowWidth,
    footerItems,
    isMobile,
    isTablet,
    toggleSidebar,
    toggleProjectRail,
    toggleProjectRailVisible,
    setProjectRailVisible,
    handleResize,
    setFooterItem,
    clearFooterItem,
    clearAllFooterItems,
  }
})
