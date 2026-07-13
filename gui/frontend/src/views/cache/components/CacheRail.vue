<script setup>
/**
 * Left rail for the cache view: a grouped accordion
 * (platform → namespace → repos) with per-item size and a purge button.
 * Extracted from CacheReposView.vue.
 */
import {
  ReloadOutlined,
  FolderOpenOutlined,
  DeleteOutlined,
  CloudServerOutlined,
  CaretRightFilled,
  WarningOutlined,
} from '@ant-design/icons-vue'
import { fmtSize } from '@/utils/format'

const props = defineProps({
  groupedRepos: { type: Array, default: () => [] },
  expandedPlatforms: { type: Set, default: () => new Set() },
  expandedNamespaces: { type: Set, default: () => new Set() },
  selectedCachePath: { type: String, default: '' },
  loading: { type: Boolean, default: false },
  repoCount: { type: Number, default: 0 },
})

const emit = defineEmits([
  'refresh', 'toggle-platform', 'toggle-namespace',
  'select', 'purge',
])

function isPlatformExpanded(p) { return props.expandedPlatforms.has(p) }
function isNamespaceExpanded(key) { return props.expandedNamespaces.has(key) }
</script>

<template>
  <aside class="cache-rail">
    <div class="rail-head">
      <span>仓库缓存</span>
      <button class="rail-refresh" title="刷新" @click="emit('refresh')">
        <ReloadOutlined />
      </button>
    </div>
    <div class="rail-list">
      <a-spin v-if="loading" class="rail-spin" />

      <!-- grouped: platform → namespace -->
      <template v-for="pg in groupedRepos" :key="pg.platform">
        <div class="rail-group-head" @click="emit('toggle-platform', pg.platform)">
          <CaretRightFilled class="rail-caret" :class="{ open: isPlatformExpanded(pg.platform) }" />
          <span class="rail-group-label">{{ pg.platform }}</span>
          <span class="rail-group-count">{{ pg.namespaces.reduce((s, n) => s + n.repos.length, 0) }}</span>
        </div>

        <template v-if="isPlatformExpanded(pg.platform)">
          <template v-for="ns in pg.namespaces" :key="pg.platform + '/' + ns.namespace">
            <!-- namespace sub-group (skip header for 本地仓库) -->
            <div
              v-if="pg.platform !== '本地仓库' && ns.namespace"
              class="rail-ns-head"
              @click="emit('toggle-namespace', pg.platform + '/' + ns.namespace)"
            >
              <CaretRightFilled class="rail-caret sm" :class="{ open: isNamespaceExpanded(pg.platform + '/' + ns.namespace) }" />
              <span class="rail-ns-label">{{ ns.namespace }}</span>
              <span class="rail-ns-count">{{ ns.repos.length }}</span>
            </div>

            <template v-if="pg.platform === '本地仓库' || !ns.namespace || isNamespaceExpanded(pg.platform + '/' + ns.namespace)">
              <div
                v-for="r in ns.repos"
                :key="r.cachePath"
                class="rail-item"
                :class="[
                  { active: r.cachePath === selectedCachePath },
                  pg.platform !== '本地仓库' && ns.namespace ? 'nested' : '',
                ]"
                :title="r.cachePath"
                @click="emit('select', r)"
              >
                <div class="rail-item-icon" :class="{ 'icon-local': r.type === 'local', 'icon-missing': !r.exists }">
                  <WarningOutlined v-if="!r.exists" />
                  <CloudServerOutlined v-else-if="r.type === 'remote'" />
                  <FolderOpenOutlined v-else />
                </div>
                <div class="rail-item-body">
                  <div class="rail-item-name">
                    {{ r.name }}
                    <span v-if="!r.exists" class="rail-missing-tag">路径不存在</span>
                    <span v-else-if="r.type === 'local'" class="rail-type-tag">本地</span>
                  </div>
                  <div class="rail-item-meta">
                    <span class="rail-size">{{ r.exists ? fmtSize(r.size) : '—' }}</span>
                    <span class="rail-ref">{{ r.refCount }} 引用</span>
                  </div>
                </div>
                <a-popconfirm
                  v-if="r.type === 'remote'"
                  :title="`清理 ${r.name}？`"
                  ok-text="清理" ok-type="danger" cancel-text="取消"
                  @confirm="emit('purge', r)"
                >
                  <button class="rail-purge" title="清理缓存" @click.stop>
                    <DeleteOutlined />
                  </button>
                </a-popconfirm>
              </div>
            </template>
          </template>
        </template>
      </template>

      <div v-if="!loading && !repoCount" class="rail-empty">
        <CloudServerOutlined class="rail-empty-icon" />
        <span>暂无缓存</span>
      </div>
    </div>
  </aside>
</template>

<style scoped>
/* ---- left rail ---- */
.cache-rail {
  width: 240px;
  min-width: 240px;
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--color-border);
  background: var(--color-surface);
  overflow: hidden;
}
.rail-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 12px;
  height: var(--navbar-height);
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-text-tertiary);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}
.rail-refresh {
  display: flex; align-items: center; justify-content: center;
  width: 24px; height: 24px; border: none; background: transparent;
  color: var(--color-text-tertiary); cursor: pointer;
  border-radius: var(--radius-xs); transition: all var(--transition-fast);
}
.rail-refresh:hover { background: var(--color-hover); color: var(--color-primary); }
.rail-list { flex: 1; overflow-y: auto; padding: var(--spacing-sm); }
.rail-list::-webkit-scrollbar { width: 5px; }
.rail-list::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 3px; }
.rail-spin { display: flex; justify-content: center; padding: var(--spacing-lg) 0; }

/* ---- group headers ---- */
.rail-group-head {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 8px 4px;
  cursor: pointer;
  user-select: none;
}
.rail-group-label {
  flex: 1;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-tertiary);
}
.rail-group-count {
  font-size: 10px;
  font-weight: 600;
  color: var(--color-text-tertiary);
  background: var(--color-surface-raised);
  padding: 0 6px;
  border-radius: 999px;
}
.rail-ns-head {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 4px 8px 2px 20px;
  cursor: pointer;
  user-select: none;
}
.rail-ns-label {
  flex: 1;
  font-size: 12px;
  font-weight: 500;
  color: var(--color-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.rail-ns-count {
  font-size: 10px;
  color: var(--color-text-tertiary);
}
.rail-caret {
  font-size: 9px;
  color: var(--color-text-tertiary);
  flex-shrink: 0;
  transition: transform var(--transition-fast);
}
.rail-caret.sm { font-size: 8px; }
.rail-caret.open { transform: rotate(90deg); }

.rail-item.nested { padding-left: 34px; }

.rail-item {
  display: flex; align-items: center; gap: 10px;
  padding: 8px 10px; border-radius: var(--radius-md);
  cursor: pointer; transition: all var(--transition-fast); margin-bottom: 2px;
}
.rail-item:hover { background: var(--color-hover); }
.rail-item.active { background: var(--color-primary-bg); }

.rail-item-icon {
  display: flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; border-radius: var(--radius-sm);
  background: var(--color-background); color: var(--color-text-tertiary);
  font-size: 14px; flex-shrink: 0; transition: all var(--transition-fast);
}
.rail-item:hover .rail-item-icon,
.rail-item.active .rail-item-icon { background: var(--color-primary); color: #fff; }
.rail-item-icon.icon-local { background: rgba(22, 163, 74, 0.12); color: var(--color-success); }
.rail-item:hover .icon-local,
.rail-item.active .icon-local { background: var(--color-primary); color: #fff; }
.rail-item-icon.icon-missing { background: var(--color-warning-bg); color: var(--color-warning); }
.rail-item:hover .icon-missing,
.rail-item.active .icon-missing { background: var(--color-warning); color: #fff; }

.rail-item-body { flex: 1; min-width: 0; }
.rail-item-name {
  font-size: 13px; font-weight: 500; color: var(--color-text);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.rail-type-tag {
  font-size: 9px; font-weight: 600; padding: 0 5px; margin-left: 4px;
  border-radius: 3px; background: var(--color-success-bg); color: var(--color-success);
  vertical-align: middle;
}
.rail-missing-tag {
  font-size: 9px; font-weight: 600; padding: 0 5px; margin-left: 4px;
  border-radius: 3px; background: var(--color-warning-bg); color: var(--color-warning);
  vertical-align: middle;
}
.rail-item.active .rail-item-name { color: var(--color-primary); }
.rail-item-meta {
  display: flex; gap: 8px; margin-top: 2px; font-size: 11px;
  color: var(--color-text-tertiary);
}

.rail-purge {
  display: flex; align-items: center; justify-content: center;
  width: 22px; height: 22px; border: none; background: transparent;
  color: var(--color-text-tertiary); cursor: pointer; border-radius: var(--radius-xs);
  opacity: 0; transition: all var(--transition-fast); flex-shrink: 0;
}
.rail-item:hover .rail-purge { opacity: 0.4; }
.rail-purge:hover { opacity: 1 !important; background: var(--color-error-bg); color: var(--color-error); }

.rail-empty {
  display: flex; flex-direction: column; align-items: center; gap: 6px; padding: 40px 0;
}
.rail-empty-icon { font-size: 28px; color: var(--color-text-tertiary); opacity: 0.35; }
.rail-empty span { font-size: 13px; color: var(--color-text-secondary); }
</style>
