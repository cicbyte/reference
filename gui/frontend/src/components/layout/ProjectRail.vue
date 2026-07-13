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
} from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'
import { useProjectStore } from '@/stores/project'
import { useLayoutStore } from '@/stores/layout'
import { formatPath } from '@/utils/path'
import { useProjectActions } from '@/composables/useProjectActions'
import DiagnoseModal from '@/components/repo/DiagnoseModal.vue'
import ProjectContextMenu from '@/components/shared/ProjectContextMenu.vue'
import CollapsedFlyout from './CollapsedFlyout.vue'

const props = defineProps({
  // When false (global-scope pages) the rail slides out to width:0 instead of
  // being removed from the DOM, so the width transition animates smoothly.
  visible: { type: Boolean, default: true },
})

const project = useProjectStore()
const layout = useLayoutStore()
const { t } = useI18n()
const { diagnoseOpen, diagnoseDir, onDoctor, onOpenInExplorer, onCopyPath, onRemove } = useProjectActions()

onMounted(async () => {
  await project.loadProjects()
  await project.detectCurrent()
})

async function onAdd() {
  const ok = await project.pickAndAdd()
  if (ok) {
    message.success(t('projectRail.switchHint', { name: project.currentName }))
  }
}

async function onSwitch(p) {
  await project.switchTo(p.dir)
}
</script>

<template>
  <div
    class="project-rail"
    :class="{ collapsed: layout.projectRailCollapsed, hidden: !props.visible }"
  >
    <div class="rail-title">
      <button
        class="rail-toggle"
        :title="layout.projectRailCollapsed ? t('projectRail.expandRail') : t('projectRail.collapseRail')"
        @click="layout.toggleProjectRail()"
      >
        <MenuUnfoldOutlined v-if="layout.projectRailCollapsed" />
        <MenuFoldOutlined v-else />
      </button>
      <span v-if="!layout.projectRailCollapsed" class="rail-title-text">{{ t('projectRail.title') }}</span>
      <button v-if="!layout.projectRailCollapsed" class="rail-refresh" :title="t('projectRail.refreshList')" @click="project.loadProjects()">
        <ReloadOutlined />
      </button>
    </div>

    <!-- collapsed: icon + flyout on hover -->
    <CollapsedFlyout
      v-if="layout.projectRailCollapsed"
      :projects="project.projects"
      :current-dir="project.currentDir"
      :has-project="project.hasProject"
      @switch="(dir) => project.switchTo(dir)"
      @add="onAdd"
    />

    <!-- expanded: full list -->
    <div v-else class="rail-list">
      <a-spin v-if="project.loading" class="rail-spin" />
      <a-dropdown
        v-for="p in project.projects"
        :key="p.dir"
        :trigger="['contextmenu']"
      >
        <div
          class="rail-item"
          :class="{ active: p.dir === project.currentDir }"
          :title="formatPath(p.dir)"
          @click="project.switchTo(p.dir)"
        >
          <div class="rail-item-icon">
            <WarningOutlined v-if="!p.exists" class="warn" />
            <FolderOutlined v-else />
          </div>
          <div class="rail-item-body">
            <div class="rail-item-name">{{ p.name }}</div>
            <div class="rail-item-meta">
              <span class="rail-count">{{ t('globalList.repoCount', { n: p.repoCount }) }}</span>
              <span v-if="p.brokenCount > 0" class="rail-broken">{{ t('globalList.brokenCount', { n: p.brokenCount }) }}</span>
            </div>
          </div>
        </div>
        <template #overlay>
          <ProjectContextMenu
            :project="p"
            @switch="onSwitch"
            @doctor="onDoctor"
            @open="onOpenInExplorer"
            @copy="onCopyPath"
            @remove="(proj) => onRemove(proj, false)"
            @clean="(proj) => onRemove(proj, true)"
          />
        </template>
      </a-dropdown>

      <div v-if="!project.loading && project.projects.length === 0" class="rail-empty">
        <FolderOutlined class="rail-empty-icon" />
        <span>{{ t('projectRail.empty') }}</span>
        <span class="rail-empty-tip">{{ t('projectRail.emptyTip') }}</span>
      </div>
    </div>

    <!-- footer: add button + aligns with global footer -->
    <div class="rail-footer">
      <button class="rail-add" :title="layout.projectRailCollapsed ? t('projectRail.addProject') : ''" @click="onAdd">
        <PlusOutlined />
        <span v-if="!layout.projectRailCollapsed">{{ t('projectRail.addProject') }}</span>
      </button>
    </div>

    <DiagnoseModal v-model:open="diagnoseOpen" :project-dir="diagnoseDir" @update:open="diagnoseOpen = $event; project.loadProjects()" />
  </div>
</template>

<style scoped>
.project-rail {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 200px;
  min-width: 0;
  border-right: 1px solid var(--color-border);
  background: var(--color-surface);
  flex-shrink: 0;
  overflow: hidden;
  /* width animates across all three states: expanded(200) → collapsed(60) → hidden(0) */
  transition: width var(--transition-normal), border-color var(--transition-normal);
}

/* collapsed (toggle button) — narrow icon-only rail */
.project-rail.collapsed {
  width: 60px;
}

/* hidden (global-scope page) — slide fully out of the layout */
.project-rail.hidden {
  width: 0;
  border-right-color: transparent;
  pointer-events: none;
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

/* ---- footer (add button, same height as global-footer) ---- */
.rail-footer {
  height: var(--footer-height);
  padding: 0 var(--spacing-sm);
  display: flex;
  align-items: center;
  border-top: 1px solid var(--color-border);
  flex-shrink: 0;
  overflow: hidden;
}

.rail-add {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  width: 100%;
  padding: 6px 12px;
  border: 1px dashed var(--color-border);
  border-radius: var(--radius-sm);
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
  padding: 6px;
}
</style>
