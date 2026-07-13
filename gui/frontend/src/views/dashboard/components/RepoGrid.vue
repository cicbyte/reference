<script setup>
import { LinkOutlined } from '@ant-design/icons-vue'

defineProps({
  loading: { type: Boolean, default: false },
  repos: { type: Array, default: () => [] },
})

const emit = defineEmits(['navigate'])
</script>

<template>
  <!-- repo grid: full width -->
  <a-spin :spinning="loading" class="dash-section">
    <div class="section-head">
      <h3>引用仓库</h3>
      <span class="section-count">{{ repos.length }}</span>
    </div>

    <div v-if="repos.length > 0" class="repo-grid">
      <div
        v-for="repo in repos"
        :key="repo.name"
        class="repo-card"
        :class="'type-' + repo.type"
        @click="emit('navigate', '/repos')"
      >
        <div class="repo-type-bar"></div>
        <div class="repo-card-inner">
          <div class="repo-card-header">
            <span class="repo-name">{{ repo.name }}</span>
            <a-tag :color="repo.type === 'remote' ? 'blue' : 'green'" class="repo-tag">
              {{ repo.type === 'remote' ? '远程' : '本地' }}
            </a-tag>
          </div>
          <div class="repo-card-body">
            <div class="repo-source" :title="repo.source">{{ repo.source }}</div>
            <div class="repo-meta">
              <span v-if="repo.branch"><LinkOutlined /> {{ repo.branch }}</span>
              <span v-if="repo.commit_at">· {{ repo.commit_at }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
    <a-empty v-else-if="!loading" description="暂无引用仓库">
      <a-button type="primary" @click="emit('navigate', '/repos')">添加第一个仓库</a-button>
    </a-empty>
  </a-spin>
</template>

<style scoped>
.dash-section {
  display: block;
  margin-bottom: var(--spacing-lg);
}

.section-head {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-md);
}
.section-head h3 {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
  margin: 0;
}
.section-count {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-tertiary);
  background: var(--color-surface-raised);
  padding: 1px 8px;
  border-radius: 999px;
}

/* ---- repo cards ---- */
.repo-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: var(--spacing-md);
}
.repo-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  overflow: hidden;
  transition: all var(--transition-fast);
  cursor: pointer;
}
.repo-card:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-md);
  transform: translateY(-1px);
}
.repo-type-bar {
  height: 3px;
}
.type-remote .repo-type-bar { background: #3b82f6; }
.type-local .repo-type-bar { background: var(--color-success); }
.repo-card-inner { padding: var(--spacing-md); }
.repo-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-sm);
}
.repo-name {
  font-weight: 600;
  font-size: 15px;
  color: var(--color-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.repo-tag { margin: 0; flex-shrink: 0; }
.repo-card-body {
  font-size: 13px;
  color: var(--color-text-secondary);
}
.repo-source {
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.repo-meta {
  display: flex;
  gap: 6px;
  align-items: center;
  font-size: 12px;
  color: var(--color-text-tertiary);
}
</style>
