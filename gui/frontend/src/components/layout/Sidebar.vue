<script setup>
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useLayoutStore } from '@/stores/layout'
import { useProjectStore } from '@/stores/project'
import {
  DashboardOutlined,
  CloudDownloadOutlined,
  UnorderedListOutlined,
  ApartmentOutlined,
  DatabaseOutlined,
  CloudServerOutlined,
  DeleteOutlined,
  PieChartOutlined,
  ReadOutlined,
  SettingOutlined,
  CaretRightFilled,
  BranchesOutlined,
} from '@ant-design/icons-vue'
import logoUrl from '@/assets/logo.svg'
import { formatPath } from '@/utils/path'

const route = useRoute()
const router = useRouter()
const layout = useLayoutStore()
const project = useProjectStore()
const { t } = useI18n()

// project-scoped routes require a selected project; disable them otherwise.
const PROJECT_SCOPED_KEYS = new Set(['/', '/repos'])
const PROJECT_SCOPED_PREFIXES = []
function isDisabled(key) {
  if (!project.hasProject) {
    if (PROJECT_SCOPED_KEYS.has(key)) return true
    if (PROJECT_SCOPED_PREFIXES.some((p) => key.startsWith(p))) return true
  }
  return false
}

const menuGroups = computed(() => [
  {
    key: 'overview', label: t('sidebar.overview'), icon: DashboardOutlined,
    children: [
      { key: '/', icon: DashboardOutlined, label: t('sidebar.dashboard') },
    ],
  },
  {
    key: 'repo', label: t('sidebar.repoManage'), icon: CloudDownloadOutlined,
    children: [
      { key: '/repos', icon: UnorderedListOutlined, label: t('sidebar.refList') },
    ],
  },
  {
    key: 'global', label: t('sidebar.globalManage'), icon: ApartmentOutlined,
    children: [
      { key: '/global', icon: DatabaseOutlined, label: t('sidebar.projectList') },
      { key: '/cache', icon: CloudServerOutlined, label: t('sidebar.cache') },
      { key: '/global/stats', icon: PieChartOutlined, label: t('sidebar.stats') },
      { key: '/global/gc', icon: DeleteOutlined, label: t('sidebar.gc') },
    ],
  },
  {
    key: 'wiki', label: t('sidebar.knowledge'), icon: ReadOutlined,
    children: [
      { key: '/wiki', icon: ReadOutlined, label: t('sidebar.remoteWiki') },
      { key: '/local-wiki', icon: ReadOutlined, label: t('sidebar.localWiki') },
    ],
  },
  {
    key: 'config', label: t('sidebar.config'), icon: SettingOutlined,
    children: [
      { key: '/settings', icon: SettingOutlined, label: t('sidebar.settings') },
    ],
  },
])

const visibleMenuGroups = computed(() => menuGroups.value)

function groupOf(path) {
  for (const g of visibleMenuGroups.value) {
    if (g.children.some((c) => c.key === path)) return g.key
  }
  return null
}

const expanded = ref(new Set([groupOf(route.path)]))
const collapsedHover = ref(null)

watch(
  () => route.path,
  (p) => {
    const g = groupOf(p)
    if (g) {
      expanded.value.add(g)
      expanded.value = new Set(expanded.value)
    }
  },
  { immediate: true },
)

function toggleGroup(key) {
  const next = new Set(expanded.value)
  if (next.has(key)) next.delete(key)
  else next.add(key)
  expanded.value = next
}

function isExpanded(key) {
  return expanded.value.has(key)
}

function navigate(key) {
  if (isDisabled(key)) return
  router.push(key)
}

const collapsedWidth = 60
const expandedWidth = 200
</script>

<template>
  <div
    class="sidebar"
    :class="{ collapsed: layout.sidebarCollapsed }"
    :style="{ width: layout.sidebarCollapsed ? collapsedWidth + 'px' : expandedWidth + 'px' }"
  >
    <div class="sidebar-logo" @click="router.push('/')">
      <img :src="logoUrl" alt="reference" class="logo-img" />
      <span v-if="!layout.sidebarCollapsed" class="logo-text">reference</span>
    </div>

    <nav class="sidebar-nav">
      <template v-for="group in visibleMenuGroups" :key="group.key">
        <div
          v-if="layout.sidebarCollapsed"
          class="nav-group-collapsed"
          @mouseenter="collapsedHover = group.key"
          @mouseleave="collapsedHover = null"
        >
          <div
            class="nav-item group-head"
            :class="{ 'has-active': groupOf(route.path) === group.key }"
            @click="toggleGroup(group.key)"
          >
            <component :is="group.icon" class="nav-icon" />
          </div>
          <transition name="flyout">
            <div v-if="collapsedHover === group.key" class="flyout">
              <div class="flyout-title">{{ group.label }}</div>
              <div
                v-for="child in group.children"
                :key="child.key"
                class="nav-item flyout-item"
                :class="{ active: route.path === child.key, disabled: isDisabled(child.key) }"
                @click="navigate(child.key)"
              >
                <component :is="child.icon" class="nav-icon" />
                <span class="nav-label">{{ child.label }}</span>
              </div>
            </div>
          </transition>
        </div>

        <div v-else class="nav-group">
          <div
            class="nav-group-head"
            :class="{ 'has-active': groupOf(route.path) === group.key }"
            @click="toggleGroup(group.key)"
          >
            <component :is="group.icon" class="group-head-icon" />
            <span class="group-head-label">{{ group.label }}</span>
            <CaretRightFilled class="group-caret" :class="{ open: isExpanded(group.key) }" />
          </div>
          <transition name="expand">
            <div v-show="isExpanded(group.key)" class="nav-children">
              <div
                v-for="child in group.children"
                :key="child.key"
                class="nav-item child"
                :class="{ active: route.path === child.key, disabled: isDisabled(child.key) }"
                @click="navigate(child.key)"
              >
                <component :is="child.icon" class="nav-icon" />
                <span class="nav-label">{{ child.label }}</span>
              </div>
            </div>
          </transition>
        </div>
      </template>
    </nav>

    <div class="sidebar-bottom">
      <div v-if="project.currentName" class="sidebar-project" :title="formatPath(project.currentDir)">
        <BranchesOutlined class="sidebar-project-icon" />
        <span v-if="!layout.sidebarCollapsed" class="sidebar-project-name">{{ project.currentName }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sidebar {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-color: var(--bg-sider);
  border-right: 1px solid var(--color-border);
  transition: width var(--transition-normal);
  overflow: hidden;
  flex-shrink: 0;
}

.sidebar-logo {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-md);
  height: var(--navbar-height);
  cursor: pointer;
  flex-shrink: 0;
  border-bottom: 1px solid var(--color-border);
}

.logo-img {
  width: 28px;
  height: 28px;
  border-radius: 7px;
  flex-shrink: 0;
  object-fit: contain;
}

.logo-text { font-size: 16px; font-weight: 700; white-space: nowrap; color: var(--color-text); }

.sidebar-nav {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-sm);
}
.sidebar-nav::-webkit-scrollbar { width: 5px; }
.sidebar-nav::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 3px; }

.nav-group { margin-bottom: 2px; }
.nav-group-head {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: 7px 10px;
  border-radius: var(--radius-md);
  cursor: pointer;
  color: var(--color-text-tertiary);
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  transition: all var(--transition-fast);
  white-space: nowrap;
  user-select: none;
}
.nav-group-head:hover { color: var(--color-text-secondary); background: var(--color-hover); }
.nav-group-head.has-active { color: var(--color-primary); }
.group-head-icon { font-size: 14px; flex-shrink: 0; }
.group-head-label { flex: 1; }
.group-caret { font-size: 9px; transition: transform var(--transition-fast); opacity: 0.7; }
.group-caret.open { transform: rotate(90deg); }

.nav-children {
  display: flex;
  flex-direction: column;
  gap: 1px;
  padding-left: 14px;
  overflow: hidden;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: 8px 12px;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
  color: var(--color-text-secondary);
  white-space: nowrap;
}
.nav-item:hover { background: var(--color-hover); color: var(--color-primary); }
.nav-item.active { background: var(--color-primary-bg); color: var(--color-primary); font-weight: 500; }
.nav-item.disabled { opacity: 0.4; cursor: not-allowed; pointer-events: none; }

.nav-item.child { padding: 7px 12px; }
.nav-item.child::before {
  content: '';
  width: 3px; height: 3px; border-radius: 50%;
  background: var(--color-border);
  flex-shrink: 0;
  margin-left: -7px;
}
.nav-item.child.active::before { background: var(--color-primary); }

.nav-icon { font-size: 16px; flex-shrink: 0; }
.nav-label { font-size: 13px; }

.nav-group-collapsed { position: relative; }
.nav-group-collapsed .nav-item.group-head { justify-content: center; padding: 10px; margin-bottom: 2px; }
.nav-group-collapsed .nav-item.group-head.has-active { background: var(--color-primary-bg); color: var(--color-primary); }

.flyout {
  position: absolute;
  left: calc(100% + 8px);
  top: 0;
  min-width: 160px;
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
.flyout-item { padding: 7px 10px; }
.flyout-item::before { display: none; }

.flyout-enter-active, .flyout-leave-active { transition: opacity var(--transition-fast), transform var(--transition-fast); }
.flyout-enter-from, .flyout-leave-to { opacity: 0; transform: translateX(-4px); }

.expand-enter-active, .expand-leave-active { transition: all var(--transition-fast); }
.expand-enter-from, .expand-leave-to { opacity: 0; max-height: 0; }

.sidebar-bottom {
  height: var(--footer-height);
  padding: 0 10px;
  display: flex;
  align-items: center;
  border-top: 1px solid var(--color-border);
  flex-shrink: 0;
  overflow: hidden;
}
.sidebar-project {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  background: var(--color-primary-bg);
  cursor: default;
  min-width: 0;
}
.sidebar-project-icon {
  font-size: 12px;
  color: var(--color-primary);
  flex-shrink: 0;
}
.sidebar-project-name {
  font-size: 12px;
  font-weight: 500;
  color: var(--color-primary);
  font-family: 'Cascadia Code', monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.collapsed .sidebar-bottom { justify-content: center; padding: 0; }
.collapsed .sidebar-logo { justify-content: center; padding: var(--spacing-md) 0; }
.collapsed .nav-item { justify-content: center; padding: 10px; }
</style>
