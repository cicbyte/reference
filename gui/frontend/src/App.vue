<script setup>
import { computed, onMounted, onBeforeUnmount } from 'vue'
import { theme } from 'ant-design-vue'
import zhCN from 'ant-design-vue/es/locale/zh_CN'
import { useThemeStore } from './stores/theme'
import { useLayoutStore } from './stores/layout'
import AppLayout from './components/layout/AppLayout.vue'

const themeStore = useThemeStore()
const layoutStore = useLayoutStore()

const antTheme = computed(() => ({
  algorithm: themeStore.isDark ? theme.darkAlgorithm : theme.defaultAlgorithm,
  token: {
    colorPrimary: '#BD93F9',
    borderRadius: 8,
    colorBgContainer: themeStore.isDark ? '#1e293b' : '#ffffff',
    colorBgLayout: themeStore.isDark ? '#0f172a' : '#f8fafc',
    colorBorder: themeStore.isDark ? '#334155' : '#e2e8f0',
    colorText: themeStore.isDark ? '#f1f5f9' : '#0f172a',
    colorTextSecondary: themeStore.isDark ? '#94a3b8' : '#64748b',
    fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif',
  }
}))

onMounted(() => {
  themeStore.initializeTheme()
  // apply saved default for sidebar collapse
  const savedSidebar = localStorage.getItem('reference-sidebar-collapsed')
  if (savedSidebar === 'true') layoutStore.sidebarCollapsed = true
  window.addEventListener('resize', layoutStore.handleResize)
})
</script>

<template>
  <a-config-provider :theme="antTheme" :locale="zhCN">
    <AppLayout />
  </a-config-provider>
</template>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body, #app {
  height: 100vh;
  width: 100vw;
  overflow: hidden;
  background-color: var(--color-background);
  color: var(--color-text);
  font-family: Inter, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  -webkit-font-smoothing: antialiased;
}

::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: var(--color-border);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: var(--color-text-tertiary);
}

::selection {
  background: var(--color-primary);
  color: white;
}

:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}
</style>
