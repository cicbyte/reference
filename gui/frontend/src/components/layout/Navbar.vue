<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { MinusOutlined, BorderOutlined, CloseOutlined, MenuFoldOutlined, MenuUnfoldOutlined } from '@ant-design/icons-vue'
import { useLayoutStore } from '../../stores/layout'

const route = useRoute()
const layout = useLayoutStore()
const app = window.go?.main?.ReferenceApp

const breadcrumbItems = computed(() => {
  return [{ label: route.meta?.title || 'reference' }]
})

function minimize() { app?.WindowMinimize() }
function maximize() { app?.WindowMaximize() }
function close() { app?.WindowClose() }
</script>

<template>
  <div class="navbar">
    <div class="navbar-left">
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
