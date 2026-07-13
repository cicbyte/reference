<script setup>
/**
 * Col-1 repo rail for the LOCAL wiki view.
 *
 * FLAT list (no platform / namespace grouping) — just repo names with a
 * refresh action. Extracted verbatim from the original LocalWikiView.vue.
 */
import {
  ReloadOutlined,
  ReadOutlined,
  FileMarkdownOutlined,
} from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'

defineProps({
  repos: { type: Array, default: () => [] },
  selectedRepoKey: { type: String, default: '' },
  loading: { type: Boolean, default: false },
  entries: { type: Array, default: () => [] },
})

const emit = defineEmits(['select-repo', 'reload'])

const { t } = useI18n()
</script>

<template>
  <aside class="wiki-rail">
    <div class="rail-head">
      <span>{{ t('localWiki.title') }}</span>
      <button class="rail-btn" :title="t('common.refresh')" @click="emit('reload')">
        <ReloadOutlined />
      </button>
    </div>
    <div class="rail-list">
      <a-spin v-if="loading" class="rail-spin" />
      <div
        v-for="repo in repos"
        :key="repo.repoName"
        class="rail-item"
        :class="{ active: selectedRepoKey === repo.repoName }"
        :title="repo.repoName"
        @click="emit('select-repo', repo.repoName)"
      >
        <div class="rail-item-icon">
          <ReadOutlined v-if="repo.fileCount === 1" />
          <FileMarkdownOutlined v-else />
        </div>
        <div class="rail-item-body">
          <div class="rail-item-name">{{ repo.repoName }}</div>
          <div class="rail-item-meta">{{ t('localWiki.fileCount', { n: repo.fileCount }) }}</div>
        </div>
      </div>
      <div v-if="!loading && !entries.length" class="rail-empty">
        <ReadOutlined class="rail-empty-icon" />
        <span>{{ t('localWiki.empty') }}</span>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.wiki-rail {
  width: 220px; min-width: 220px;
  display: flex; flex-direction: column;
  border-right: 1px solid var(--color-border);
  background: var(--color-surface); overflow: hidden;
}
.rail-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 12px; height: var(--navbar-height);
  border-bottom: 1px solid var(--color-border); flex-shrink: 0;
}
.rail-head > span {
  font-size: 12px; font-weight: 600; text-transform: uppercase;
  letter-spacing: 0.06em; color: var(--color-text-tertiary);
}
.rail-btn {
  display: flex; align-items: center; justify-content: center;
  width: 26px; height: 26px; border: none; background: transparent;
  color: var(--color-text-tertiary); cursor: pointer;
  border-radius: var(--radius-xs); transition: all var(--transition-fast);
}
.rail-btn:hover { background: var(--color-hover); color: var(--color-primary); }
.rail-list { flex: 1; overflow-y: auto; padding: var(--spacing-sm); }
.rail-list::-webkit-scrollbar { width: 5px; }
.rail-list::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 3px; }
.rail-spin { display: flex; justify-content: center; padding: var(--spacing-lg) 0; }

.rail-item {
  display: flex; align-items: center; gap: 10px;
  padding: 7px 10px; border-radius: var(--radius-md);
  cursor: pointer; transition: all var(--transition-fast); margin-bottom: 1px;
}
.rail-item:hover { background: var(--color-hover); }
.rail-item.active { background: var(--color-primary-bg); }
.rail-item-icon {
  display: flex; align-items: center; justify-content: center;
  width: 26px; height: 26px; border-radius: var(--radius-sm);
  background: var(--color-background); color: var(--color-text-tertiary);
  font-size: 13px; flex-shrink: 0; transition: all var(--transition-fast);
}
.rail-item:hover .rail-item-icon, .rail-item.active .rail-item-icon { background: var(--color-primary); color: #fff; }
.rail-item-body { flex: 1; min-width: 0; }
.rail-item-name {
  font-size: 13px; font-weight: 500; color: var(--color-text);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.rail-item.active .rail-item-name { color: var(--color-primary); }
.rail-item-meta { margin-top: 1px; font-size: 11px; color: var(--color-text-tertiary); }
.rail-empty {
  display: flex; flex-direction: column; align-items: center; gap: 4px; padding: 40px 0;
}
.rail-empty-icon { font-size: 28px; color: var(--color-text-tertiary); opacity: 0.35; }
.rail-empty span { font-size: 13px; color: var(--color-text-secondary); }
</style>
