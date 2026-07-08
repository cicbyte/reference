<script setup>
import { ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import {
  FolderOutlined,
  PlusOutlined,
  WarningOutlined,
  ReloadOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  AppstoreOutlined,
} from '@ant-design/icons-vue'
import { useProjectStore } from '../../stores/project'
import { useLayoutStore } from '../../stores/layout'

const project = useProjectStore()
const layout = useLayoutStore()
const collapsedHover = ref(null)

onMounted(async () => {
  await project.loadProjects()
  await project.detectCurrent()
})

async function onAdd() {
  const ok = await project.pickAndAdd()
  if (ok) {
    message.success(`已切换到 ${project.currentName}`)
  }
}

const collapsedWidth = 60
const expandedWidth = 200
</script>

<template>
  <div
    class="project-rail"
    :class="{ collapsed: layout.projectRailCollapsed }"
    :style="{ width: layout.projectRailCollapsed ? collapsedWidth + 'px' : expandedWidth + 'px' }"
  >
    <div class="rail-title">
      <button
        class="rail-toggle"
        :title="layout.projectRailCollapsed ? '展开项目栏' : '折叠项目栏'"
        @click="layout.toggleProjectRail()"
      >
        <MenuUnfoldOutlined v-if="layout.projectRailCollapsed" />
        <MenuFoldOutlined v-else />
      </button>
      <span v-if="!layout.projectRailCollapsed" class="rail-title-text">项目</span>
      <button v-if="!layout.projectRailCollapsed" class="rail-refresh" title="刷新列表" @click="project.loadProjects()">
        <ReloadOutlined />
      </button>
    </div>

    <!-- collapsed: icon + flyout on hover -->
    <div
      v-if="layout.projectRailCollapsed"
      class="rail-collapsed-list"
    >
      <div
        class="rail-icon-btn"
        :class="{ active: !!project.hasProject }"
        title="项目"
        @mouseenter="collapsedHover = 'projects'"
        @mouseleave="collapsedHover = null"
      >
        <AppstoreOutlined />
      </div>
      <transition name="flyout">
        <div v-if="collapsedHover === 'projects'" class="flyout">
          <div class="flyout-title">项目</div>
          <div
            v-for="p in project.projects"
            :key="p.dir"
            class="flyout-item"
            :class="{ active: p.dir === project.currentDir }"
            :title="p.dir"
            @click="project.switchTo(p.dir)"
          >
            <WarningOutlined v-if="!p.exists" class="warn" />
            <FolderOutlined v-else />
            <span class="flyout-name">{{ p.name }}</span>
            <span class="flyout-count">{{ p.repoCount }}</span>
          </div>
          <div v-if="project.projects.length === 0" class="flyout-empty">暂无项目</div>
          <div class="flyout-add" @click="onAdd">
            <PlusOutlined />
            <span>添加项目</span>
          </div>
        </div>
      </transition>
    </div>

    <!-- expanded: full list -->
    <div v-else class="rail-list">
      <a-spin v-if="project.loading" class="rail-spin" />
      <div
        v-for="p in project.projects"
        :key="p.dir"
        class="rail-item"
        :class="{ active: p.dir === project.currentDir }"
        :title="p.dir"
        @click="project.switchTo(p.dir)"
      >
        <div class="rail-item-icon">
          <WarningOutlined v-if="!p.exists" class="warn" />
          <FolderOutlined v-else />
        </div>
        <div class="rail-item-body">
          <div class="rail-item-name">{{ p.name }}</div>
          <div class="rail-item-meta">
            <span class="rail-count">{{ p.repoCount }} 个引用</span>
            <span v-if="p.brokenCount > 0" class="rail-broken">{{ p.brokenCount }} 断链</span>
          </div>
        </div>
      </div>

      <div v-if="!project.loading && project.projects.length === 0" class="rail-empty">
        <FolderOutlined class="rail-empty-icon" />
        <span>暂无项目</span>
        <span class="rail-empty-tip">点击下方添加</span>
      </div>
    </div>

    <!-- add button (sits above the footer spacer) -->
    <div class="rail-add-area">
      <button class="rail-add" :title="layout.projectRailCollapsed ? '添加项目' : ''" @click="onAdd">
        <PlusOutlined />
        <span v-if="!layout.projectRailCollapsed">添加项目</span>
      </button>
    </div>

    <!-- footer spacer: aligns with MainContent's global-footer (48px) -->
    <div class="rail-footer-spacer"></div>
  </div>
</template>

<style scoped>
.project-rail {
  display: flex;
  flex-direction: column;
  height: 100vh;
  border-right: 1px solid var(--color-border);
  background: var(--color-surface);
  flex-shrink: 0;
  overflow: hidden;
  /* width animation — driven by the inline :style binding */
  transition: width var(--transition-normal);
}

/* ---- title bar ---- */
.rail-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: 0 var(--spacing-sm) 0 6px;
  height: var(--navbar-height);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}
.collapsed .rail-title {
  justify-content: center;
  padding: 0;
}

.rail-toggle {
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
}
.rail-toggle:hover {
  background: var(--color-hover);
  color: var(--color-primary);
}

.rail-title-text {
  flex: 1;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-text-tertiary);
  white-space: nowrap;
}

.rail-refresh {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  color: var(--color-text-tertiary);
  cursor: pointer;
  border-radius: var(--radius-xs);
  transition: all var(--transition-fast);
  flex-shrink: 0;
}
.rail-refresh:hover {
  background: var(--color-hover);
  color: var(--color-primary);
}

/* ---- expanded list ---- */
.rail-list {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-sm);
}
.rail-list::-webkit-scrollbar {
  width: 5px;
}
.rail-list::-webkit-scrollbar-thumb {
  background: var(--color-border);
  border-radius: 3px;
}

.rail-spin {
  display: flex;
  justify-content: center;
  padding: var(--spacing-lg) 0;
}

.rail-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
  margin-bottom: 2px;
}
.rail-item:hover {
  background: var(--color-hover);
}
.rail-item.active {
  background: var(--color-primary-bg);
}

.rail-item-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: var(--radius-sm);
  background: var(--color-background);
  color: var(--color-text-tertiary);
  font-size: 14px;
  flex-shrink: 0;
  transition: all var(--transition-fast);
}
.rail-item:hover .rail-item-icon,
.rail-item.active .rail-item-icon {
  background: var(--color-primary);
  color: #fff;
}
.rail-item-icon .warn {
  color: var(--color-warning);
}

.rail-item-body {
  flex: 1;
  min-width: 0;
}

.rail-item-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.rail-item.active .rail-item-name {
  color: var(--color-primary);
}

.rail-item-meta {
  display: flex;
  gap: 8px;
  margin-top: 2px;
  font-size: 11px;
  color: var(--color-text-tertiary);
}
.rail-broken {
  color: var(--color-warning);
}

.rail-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 40px 0;
}
.rail-empty-icon {
  font-size: 28px;
  color: var(--color-text-tertiary);
  opacity: 0.35;
}
.rail-empty span:nth-child(2) {
  font-size: 13px;
  color: var(--color-text-secondary);
}
.rail-empty-tip {
  font-size: 11px;
  color: var(--color-text-tertiary);
}

/* ---- add area (above footer spacer) ---- */
.rail-add-area {
  padding: var(--spacing-sm);
  flex-shrink: 0;
}

.rail-add {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  width: 100%;
  padding: 8px 12px;
  border: 1px dashed var(--color-border);
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  font-size: 13px;
  transition: all var(--transition-fast);
}
.rail-add:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
  background: var(--color-primary-bg);
}
.collapsed .rail-add {
  padding: 10px;
}

/* ---- footer spacer (aligns with global-footer) ---- */
.rail-footer-spacer {
  height: var(--footer-height);
  border-top: 1px solid var(--color-border);
  flex-shrink: 0;
}

/* ---- collapsed mode ---- */
.rail-collapsed-list {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-sm);
  display: flex;
  flex-direction: column;
  align-items: center;
}

.rail-icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  cursor: pointer;
  color: var(--color-text-secondary);
  font-size: 18px;
  transition: all var(--transition-fast);
  position: relative;
}
.rail-icon-btn:hover {
  background: var(--color-hover);
  color: var(--color-primary);
}
.rail-icon-btn.active {
  background: var(--color-primary-bg);
  color: var(--color-primary);
}

/* ---- flyout (collapsed hover popover) ---- */
.flyout {
  position: absolute;
  left: calc(100% + 8px);
  top: 0;
  min-width: 200px;
  max-height: 400px;
  overflow-y: auto;
  background: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  padding: 6px;
  z-index: 1100;
}
.flyout-title {
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-tertiary);
  padding: 4px 10px;
}
.flyout-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 10px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 13px;
  color: var(--color-text-secondary);
  transition: all var(--transition-fast);
}
.flyout-item:hover {
  background: var(--color-hover);
  color: var(--color-primary);
}
.flyout-item.active {
  background: var(--color-primary-bg);
  color: var(--color-primary);
}
.flyout-item .warn {
  color: var(--color-warning);
}
.flyout-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.flyout-count {
  font-size: 11px;
  color: var(--color-text-tertiary);
}
.flyout-empty {
  font-size: 12px;
  color: var(--color-text-tertiary);
  padding: 12px 10px;
  text-align: center;
}
.flyout-add {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 10px;
  margin-top: 4px;
  border-top: 1px solid var(--color-border-light);
  font-size: 13px;
  color: var(--color-text-secondary);
  cursor: pointer;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}
.flyout-add:hover {
  color: var(--color-primary);
  background: var(--color-hover);
}

/* flyout enter/leave animation */
.flyout-enter-active,
.flyout-leave-active {
  transition: opacity var(--transition-fast), transform var(--transition-fast);
}
.flyout-enter-from,
.flyout-leave-to {
  opacity: 0;
  transform: translateX(-4px);
}
</style>
