import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useLayoutStore = defineStore('layout', () => {
  const sidebarCollapsed = ref(false)
  const windowWidth = ref(window.innerWidth)

  const isMobile = computed(() => windowWidth.value < 768)
  const isTablet = computed(() => windowWidth.value >= 768 && windowWidth.value < 1024)

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function handleResize() {
    windowWidth.value = window.innerWidth
    if (isMobile.value) {
      sidebarCollapsed.value = true
    }
  }

  return { sidebarCollapsed, windowWidth, isMobile, isTablet, toggleSidebar, handleResize }
})
