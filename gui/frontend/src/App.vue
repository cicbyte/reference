<script setup>
import { computed, onMounted, onBeforeUnmount } from 'vue'
import zhCN from 'ant-design-vue/es/locale/zh_CN'
import enUS from 'ant-design-vue/es/locale/en_US'
import { useThemeStore } from './stores/theme'
import { useLayoutStore } from './stores/layout'
import { useLocaleStore } from './stores/locale'
import AppLayout from './components/layout/AppLayout.vue'

const themeStore = useThemeStore()
const layoutStore = useLayoutStore()
const localeStore = useLocaleStore()

// Ant Design component locale follows the app language.
const antLocale = computed(() => localeStore.isEn ? enUS : zhCN)

// antd 主题 token 完全委托给 theme store（跟随预设/自定义/明暗）
const antTheme = computed(() => themeStore.antdTheme)

onMounted(() => {
  themeStore.initializeTheme()
  // apply saved default for sidebar collapse (preference lives in the layout store)
  if (layoutStore.sidebarDefaultCollapsed) layoutStore.sidebarCollapsed = true
  window.addEventListener('resize', layoutStore.handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', layoutStore.handleResize)
})
</script>

<template>
  <a-config-provider :theme="antTheme" :locale="antLocale">
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
