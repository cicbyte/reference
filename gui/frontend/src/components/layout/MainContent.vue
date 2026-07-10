<script setup>
import { BranchesOutlined } from '@ant-design/icons-vue'
import { useLayoutStore } from '../../stores/layout'

const layout = useLayoutStore()
</script>

<template>
  <div class="main-content">
    <div class="content-wrapper">
      <router-view v-slot="{ Component }">
        <transition mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </div>

    <div class="global-footer">
      <div class="status-bar">
        <template v-for="item in layout.footerItems" :key="item.key">
          <span v-if="item.value" class="status-item">
            <BranchesOutlined v-if="item.icon === 'branch'" />
            <component :is="item.icon" v-else-if="item.icon" />
            <span class="status-label" v-if="item.label">{{ item.label }}:</span>
            <span class="status-value" :title="item.value">{{ item.value }}</span>
          </span>
          <span v-if="item.value" class="status-sep"></span>
        </template>
        <span class="status-spacer"></span>
        <span class="status-item">
          <BranchesOutlined />
          <span class="status-value">reference</span>
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  background: var(--color-background);
}

.content-wrapper {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-lg);
  min-height: 0;
}

.global-footer {
  display: flex;
  align-items: center;
  padding: 0 var(--spacing-lg);
  height: var(--footer-height);
  border-top: 1px solid var(--color-border);
  background: var(--color-surface);
  flex-shrink: 0;
  overflow: hidden;
}

.status-bar {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  width: 100%;
  font-size: 12px;
  color: var(--color-text-secondary);
  user-select: none;
}

.status-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
}

.status-label {
  color: var(--color-text-tertiary);
  flex-shrink: 0;
}

.status-value {
  color: var(--color-text);
  font-weight: 500;
  font-family: 'Cascadia Code', monospace;
  font-size: 11px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 500px;
}

.status-sep {
  color: var(--color-text-tertiary);
  opacity: 0.4;
}

.status-sep::after {
  content: '·';
}

.status-spacer {
  flex: 1;
}
</style>
