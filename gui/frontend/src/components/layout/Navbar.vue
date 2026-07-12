<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { MinusOutlined, BorderOutlined, CloseOutlined, MenuFoldOutlined, MenuUnfoldOutlined, AppstoreOutlined, SettingOutlined } from '@ant-design/icons-vue'
import { useLayoutStore } from '../../stores/layout'
import { useThemeStore } from '../../stores/theme'
import { useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const layout = useLayoutStore()
const theme = useThemeStore()
const app = window.go?.main?.ReferenceApp

const breadcrumbItems = computed(() => {
  return [{ label: route.meta?.title || 'reference' }]
})

// rail toggle is available on all pages, fully user-controlled
const canToggleRail = computed(() => true)

function minimize() { app?.WindowMinimize() }
function maximize() { app?.WindowMaximize() }
function close() { app?.WindowClose() }
</script>

<template>
  <div class="navbar">
    <div class="navbar-left">
      <button
        v-show="canToggleRail && !layout.isMobile"
        class="sidebar-toggle"
        :class="{ 'toggle-active': layout.projectRailVisible }"
        :title="layout.projectRailVisible ? '隐藏项目栏' : '显示项目栏'"
        @click="layout.toggleProjectRailVisible()"
      >
        <AppstoreOutlined />
      </button>

      <button
        v-show="!layout.isMobile"
        class="sidebar-toggle"
        :title="layout.sidebarCollapsed ? '展开侧边栏' : '折叠侧边栏'"
        @click="layout.toggleSidebar()"
      >
        <MenuUnfoldOutlined v-if="layout.sidebarCollapsed" />
        <MenuFoldOutlined v-else />
      </button>

      <a-breadcrumb class="breadcrumb">
        <a-breadcrumb-item v-for="(item, idx) in breadcrumbItems" :key="idx">
          {{ item.label }}
        </a-breadcrumb-item>
      </a-breadcrumb>
    </div>

    <div class="navbar-right">
      <button
        class="win-btn theme-btn"
        :title="theme.isDark ? '切换到亮色' : '切换到暗色'"
        @click="theme.toggleTheme()"
      >
        <svg v-if="theme.isDark" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="5"/>
          <line x1="12" y1="1" x2="12" y2="3"/>
          <line x1="12" y1="21" x2="12" y2="23"/>
          <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/>
          <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
          <line x1="1" y1="12" x2="3" y2="12"/>
          <line x1="21" y1="12" x2="23" y2="12"/>
          <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/>
          <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
        </svg>
        <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
        </svg>
      </button>
      <button
        class="win-btn"
        :class="{ 'toggle-active': route.path === '/settings' }"
        title="设置"
        @click="router.push('/settings')"
      >
        <SettingOutlined />
      </button>
      <button class="win-btn" title="最小化" @click="minimize">
        <MinusOutlined />
      </button>
      <button class="win-btn" title="最大化" @click="maximize">
        <BorderOutlined />
      </button>
      <button class="win-btn win-btn-close" title="关闭" @click="close">
        <CloseOutlined />
      </button>
    </div>
  </div>
</template>

<style scoped>
.navbar {
  display: flex;
  align-items: center;
  height: var(--navbar-height);
  padding: 0 var(--spacing-md);
  border-bottom: 1px solid var(--color-border);
  background: var(--color-surface);
  flex-shrink: 0;
  --wails-draggable: drag;
}

.navbar-left {
  flex: 1;
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  min-width: 0;
}

.navbar-right {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 2px;
}

.sidebar-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  border-radius: var(--radius-xs);
  transition: all var(--transition-fast);
  flex-shrink: 0;
  --wails-draggable: no-drag;
}
.sidebar-toggle:hover { background: var(--color-hover); color: var(--color-primary); }
.sidebar-toggle.toggle-active { color: var(--color-primary); }

.breadcrumb { font-size: 13px; min-width: 0; --wails-draggable: no-drag; }

.win-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 28px;
  border: none;
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  border-radius: var(--radius-xs);
  transition: all var(--transition-fast);
  --wails-draggable: no-drag;
}
.win-btn:hover { background: var(--color-hover); color: var(--color-text); }
.win-btn-close:hover { background: var(--color-error); color: white; }
</style>
