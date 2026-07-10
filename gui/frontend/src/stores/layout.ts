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
    windowWidth,
    footerItems,
    isMobile,
    isTablet,
    toggleSidebar,
    toggleProjectRail,
    handleResize,
    setFooterItem,
    clearFooterItem,
    clearAllFooterItems,
  }
})
