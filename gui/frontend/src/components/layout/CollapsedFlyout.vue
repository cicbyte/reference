<script setup>
/**
 * Collapsed-mode project rail: a single icon button that, on hover, pops out a
 * flyout list of projects. Extracted from ProjectRail.vue to keep it under
 * the 500-line cap.
 */
import { ref } from 'vue'
import { AppstoreOutlined, FolderOutlined, WarningOutlined, PlusOutlined } from '@ant-design/icons-vue'
import { formatPath } from '../../utils/path'

defineProps({
  projects: { type: Array, default: () => [] },
  currentDir: { type: String, default: '' },
  hasProject: { type: Boolean, default: false },
})

const emit = defineEmits(['switch', 'add'])
const hover = ref(null)
</script>

<template>
  <div class="rail-collapsed-list">
    <div
      class="rail-icon-btn"
      :class="{ active: hasProject }"
      title="项目"
      @mouseenter="hover = 'projects'"
      @mouseleave="hover = null"
    >
      <AppstoreOutlined />
    </div>
    <transition name="flyout">
      <div v-if="hover === 'projects'" class="flyout">
        <div class="flyout-title">项目</div>
        <div
          v-for="p in projects"
          :key="p.dir"
          class="flyout-item"
          :class="{ active: p.dir === currentDir }"
          :title="formatPath(p.dir)"
          @click="emit('switch', p.dir)"
        >
          <WarningOutlined v-if="!p.exists" class="warn" />
          <FolderOutlined v-else />
          <span class="flyout-name">{{ p.name }}</span>
          <span class="flyout-count">{{ p.repoCount }}</span>
        </div>
        <div v-if="projects.length === 0" class="flyout-empty">暂无项目</div>
        <div class="flyout-add" @click="emit('add')">
          <PlusOutlined />
          <span>添加项目</span>
        </div>
      </div>
    </transition>
  </div>
</template>

<style scoped>
.rail-collapsed-list {
  flex: 1; overflow-y: auto; padding: var(--spacing-sm);
  display: flex; flex-direction: column; align-items: center;
}
.rail-icon-btn {
  display: flex; align-items: center; justify-content: center;
  width: 40px; height: 40px; border-radius: var(--radius-md);
  cursor: pointer; color: var(--color-text-secondary); font-size: 18px;
  transition: all var(--transition-fast); position: relative;
}
.rail-icon-btn:hover { background: var(--color-hover); color: var(--color-primary); }
.rail-icon-btn.active { background: var(--color-primary-bg); color: var(--color-primary); }

.flyout {
  position: absolute; left: calc(100% + 8px); top: 0;
  min-width: 200px; max-height: 400px; overflow-y: auto;
  background: var(--color-background); border: 1px solid var(--color-border);
  border-radius: var(--radius-md); box-shadow: var(--shadow-lg);
  padding: 6px; z-index: 1100;
}
.flyout-title {
  font-size: 10px; font-weight: 600; text-transform: uppercase;
  letter-spacing: 0.05em; color: var(--color-text-tertiary); padding: 4px 10px;
}
.flyout-item {
  display: flex; align-items: center; gap: 8px;
  padding: 7px 10px; border-radius: var(--radius-sm);
  cursor: pointer; font-size: 13px; color: var(--color-text-secondary);
  transition: all var(--transition-fast);
}
.flyout-item:hover { background: var(--color-hover); color: var(--color-primary); }
.flyout-item.active { background: var(--color-primary-bg); color: var(--color-primary); }
.flyout-item .warn { color: var(--color-warning); }
.flyout-name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.flyout-count { font-size: 11px; color: var(--color-text-tertiary); }
.flyout-empty { font-size: 12px; color: var(--color-text-tertiary); padding: 12px 10px; text-align: center; }
.flyout-add {
  display: flex; align-items: center; gap: 6px;
  padding: 7px 10px; margin-top: 4px;
  border-top: 1px solid var(--color-border-light);
  font-size: 13px; color: var(--color-text-secondary);
  cursor: pointer; border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}
.flyout-add:hover { color: var(--color-primary); background: var(--color-hover); }

.flyout-enter-active, .flyout-leave-active {
  transition: opacity var(--transition-fast), transform var(--transition-fast);
}
.flyout-enter-from, .flyout-leave-to { opacity: 0; transform: translateX(-4px); }
</style>
