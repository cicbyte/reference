<script setup>
/**
 * Col-1 repo rail for the REMOTE wiki view.
 *
 * Grouped accordion: platform -> namespace -> repo, with sync + refresh
 * actions in the header. Extracted verbatim from the original WikiView.vue
 * so all grouping / caret / active-state behavior is preserved.
 */
import {
  ReloadOutlined,
  CloudSyncOutlined,
  ReadOutlined,
  CaretRightFilled,
  FileMarkdownOutlined,
} from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'

// Sentinel matching the one used by the parent (views/wiki/index.vue) to
// tag the platform grouping key for local entries.
const LOCAL_PLATFORM_KEY = '__local__'

defineProps({
  groupedRepos: { type: Array, default: () => [] },
  expandedPlatforms: { type: Object, default: () => new Set() },
  expandedNamespaces: { type: Object, default: () => new Set() },
  selectedRepoKey: { type: String, default: '' },
  loading: { type: Boolean, default: false },
  syncing: { type: Boolean, default: false },
  entries: { type: Array, default: () => [] },
})

const emit = defineEmits([
  'select-repo',
  'toggle-platform',
  'toggle-namespace',
  'sync',
  'reload',
])

const { t } = useI18n()

// Platform display label: local sentinel → localized "Local Knowledge Base";
// any other platform string is shown verbatim (already uppercased upstream).
function platformLabel(p) {
  return p === LOCAL_PLATFORM_KEY ? t('localWiki.title') : p
}
</script>

<template>
  <aside class="wiki-rail">
    <div class="rail-head">
      <span>{{ t('wiki.title') }}</span>
      <div class="rail-actions">
        <button class="rail-btn" :title="t('wiki.syncTooltip')" :disabled="syncing" @click="emit('sync')">
          <CloudSyncOutlined :spin="syncing" />
        </button>
        <button class="rail-btn" :title="t('common.refresh')" @click="emit('reload')">
          <ReloadOutlined />
        </button>
      </div>
    </div>
    <div class="rail-list">
      <a-spin v-if="loading" class="rail-spin" />

      <template v-for="pg in groupedRepos" :key="pg.platform">
        <div class="rail-group-head" @click="emit('toggle-platform', pg.platform)">
          <CaretRightFilled class="rail-caret" :class="{ open: expandedPlatforms.has(pg.platform) }" />
          <span class="rail-group-label">{{ platformLabel(pg.platform) }}</span>
          <span class="rail-group-count">{{ pg.namespaces.reduce((s, n) => s + n.repos.length, 0) }}</span>
        </div>

        <template v-if="expandedPlatforms.has(pg.platform)">
          <template v-for="ns in pg.namespaces" :key="pg.platform + '/' + ns.namespace">
            <div
              v-if="pg.platform !== LOCAL_PLATFORM_KEY && ns.namespace"
              class="rail-ns-head"
              @click="emit('toggle-namespace', pg.platform + '/' + ns.namespace)"
            >
              <CaretRightFilled class="rail-caret sm" :class="{ open: expandedNamespaces.has(pg.platform + '/' + ns.namespace) }" />
              <span class="rail-ns-label">{{ ns.namespace }}</span>
            </div>

            <template v-if="pg.platform === LOCAL_PLATFORM_KEY || !ns.namespace || expandedNamespaces.has(pg.platform + '/' + ns.namespace)">
              <div
                v-for="repo in ns.repos"
                :key="pg.platform + '/' + ns.namespace + '/' + repo.repoName"
                class="rail-item"
                :class="{ active: selectedRepoKey === (ns.repos[0]?.source || 'remote') + '|' + repo.repoName }"
                :title="repo.repoName"
                @click="emit('select-repo', ns.repos[0]?.source || 'remote', repo.repoName)"
              >
                <div class="rail-item-icon">
                  <ReadOutlined v-if="repo.fileCount === 1" />
                  <FileMarkdownOutlined v-else />
                </div>
                <div class="rail-item-body">
                  <div class="rail-item-name">{{ repo.repoName }}</div>
                  <div class="rail-item-meta">
                    <span>{{ t('wiki.fileCount', { n: repo.fileCount }) }}</span>
                  </div>
                </div>
              </div>
            </template>
          </template>
        </template>
      </template>

      <div v-if="!loading && !entries.length" class="rail-empty">
        <ReadOutlined class="rail-empty-icon" />
        <span>{{ t('wiki.empty') }}</span>
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
.rail-actions { display: flex; gap: 2px; }
.rail-btn {
  display: flex; align-items: center; justify-content: center;
  width: 26px; height: 26px; border: none; background: transparent;
  color: var(--color-text-tertiary); cursor: pointer;
  border-radius: var(--radius-xs); transition: all var(--transition-fast);
}
.rail-btn:hover { background: var(--color-hover); color: var(--color-primary); }
.rail-btn:disabled { opacity: 0.4; cursor: not-allowed; }

.rail-list { flex: 1; overflow-y: auto; padding: var(--spacing-sm); }
.rail-list::-webkit-scrollbar { width: 5px; }
.rail-list::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 3px; }
.rail-spin { display: flex; justify-content: center; padding: var(--spacing-lg) 0; }

.rail-group-head {
  display: flex; align-items: center; gap: 6px;
  padding: 6px 8px 4px; cursor: pointer; user-select: none;
}
.rail-group-label {
  flex: 1; font-size: 11px; font-weight: 600; text-transform: uppercase;
  letter-spacing: 0.05em; color: var(--color-text-tertiary);
}
.rail-group-count {
  font-size: 10px; font-weight: 600; color: var(--color-text-tertiary);
  background: var(--color-surface-raised); padding: 0 6px; border-radius: 999px;
}
.rail-ns-head {
  display: flex; align-items: center; gap: 5px;
  padding: 4px 8px 2px 20px; cursor: pointer; user-select: none;
}
.rail-ns-label {
  flex: 1; font-size: 12px; font-weight: 500; color: var(--color-text-secondary);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.rail-caret {
  font-size: 9px; color: var(--color-text-tertiary); flex-shrink: 0;
  transition: transform var(--transition-fast);
}
.rail-caret.sm { font-size: 8px; }
.rail-caret.open { transform: rotate(90deg); }

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
.rail-item:hover .rail-item-icon,
.rail-item.active .rail-item-icon { background: var(--color-primary); color: #fff; }
.rail-item-body { flex: 1; min-width: 0; }
.rail-item-name {
  font-size: 13px; font-weight: 500; color: var(--color-text);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.rail-item.active .rail-item-name { color: var(--color-primary); }
.rail-item-meta { display: flex; gap: 6px; margin-top: 1px; font-size: 11px; color: var(--color-text-tertiary); }

.rail-empty {
  display: flex; flex-direction: column; align-items: center; gap: 4px; padding: 40px 0;
}
.rail-empty-icon { font-size: 28px; color: var(--color-text-tertiary); opacity: 0.35; }
.rail-empty span { font-size: 13px; color: var(--color-text-secondary); }
</style>
