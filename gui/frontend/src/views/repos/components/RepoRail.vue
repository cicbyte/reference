<script setup>
/**
 * Left rail for the repos view: a flat list of repos with icon, name,
 * type/branch tags and a per-item context menu (browse / update / stats /
 * reclone / remove). Extracted from RepoListView.vue.
 */
import {
  PlusOutlined,
  SyncOutlined,
  DeleteOutlined,
  BarChartOutlined,
  MedicineBoxOutlined,
  FolderOpenOutlined,
  CloudDownloadOutlined,
  CloudServerOutlined,
  FileOutlined,
} from '@ant-design/icons-vue'

defineProps({
  repos: { type: Array, default: () => [] },
  loading: { type: Boolean, default: false },
  selectedRepo: { type: String, default: '' },
})

const emit = defineEmits(['select', 'update', 'reclone', 'stats', 'remove', 'add', 'diagnose'])
</script>

<template>
  <aside class="repo-rail">
    <div class="rail-head">
      <span>仓库列表</span>
      <div class="rail-actions">
        <a-button size="small" type="text" @click="emit('diagnose')" title="诊断修复">
          <MedicineBoxOutlined />
        </a-button>
        <a-button size="small" type="primary" @click="emit('add')" title="添加仓库">
          <PlusOutlined />
        </a-button>
      </div>
    </div>
    <div class="rail-list">
      <a-spin v-if="loading" class="rail-spin" />
      <div
        v-for="r in repos"
        :key="r.name"
        class="rail-item"
        :class="{ active: r.name === selectedRepo }"
        :title="r.source"
        @click="emit('select', r)"
      >
        <div class="rail-item-icon" :class="{ 'icon-missing': r.cacheExists === false }">
          <CloudServerOutlined v-if="r.type === 'remote' && r.cacheExists !== false" />
          <FolderOpenOutlined v-else-if="r.type === 'local' && r.cacheExists !== false" />
          <FileOutlined v-else />
        </div>
        <div class="rail-item-body">
          <div class="rail-item-name">
            {{ r.name }}
            <span v-if="r.cacheExists === false" class="missing-tag">缺失</span>
          </div>
          <div class="rail-item-meta">
            <span class="rail-type">{{ r.type === 'remote' ? '远程' : '本地' }}</span>
            <span v-if="r.branch">{{ r.branch }}</span>
          </div>
        </div>
        <a-dropdown :trigger="['contextmenu']">
          <div class="rail-item-extra" @click.stop>
            <a-button size="small" type="text" @click.stop="emit('stats', r.name)" title="统计">
              <BarChartOutlined />
            </a-button>
          </div>
          <template #overlay>
            <a-menu>
              <a-menu-item @click="emit('select', r)" :disabled="r.cacheExists === false">
                <FolderOpenOutlined /> 浏览代码
              </a-menu-item>
              <a-menu-item @click="emit('update', r.name)">
                <SyncOutlined /> 更新仓库
              </a-menu-item>
              <a-menu-item @click="emit('stats', r.name)" :disabled="r.cacheExists === false">
                <BarChartOutlined /> 代码统计
              </a-menu-item>
              <a-menu-item v-if="r.cacheExists === false && r.type === 'remote'" @click="emit('reclone', r.name)">
                <CloudDownloadOutlined /> 重新克隆
              </a-menu-item>
              <a-menu-divider />
              <a-menu-item danger @click="emit('remove', r.name)">
                <DeleteOutlined /> 移除引用
              </a-menu-item>
            </a-menu>
          </template>
        </a-dropdown>
      </div>
      <div v-if="!loading && !repos.length" class="rail-empty">
        <CloudServerOutlined class="rail-empty-icon" />
        <span>暂无仓库</span>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.repo-rail {
  width: 240px; min-width: 240px;
  display: flex; flex-direction: column;
  border-right: 1px solid var(--color-border);
  background: var(--color-surface); overflow: hidden;
}
.rail-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 12px; height: var(--navbar-height);
  border-bottom: 1px solid var(--color-border); flex-shrink: 0;
}
.rail-head > span { font-size: 12px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.06em; color: var(--color-text-tertiary); }
.rail-actions { display: flex; gap: 2px; }
.rail-list { flex: 1; overflow-y: auto; padding: var(--spacing-sm); }
.rail-list::-webkit-scrollbar { width: 5px; }
.rail-list::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 3px; }
.rail-spin { display: flex; justify-content: center; padding: var(--spacing-lg) 0; }

.rail-item {
  display: flex; align-items: center; gap: 10px;
  padding: 8px 10px; border-radius: var(--radius-md);
  cursor: pointer; transition: all var(--transition-fast); margin-bottom: 1px;
}
.rail-item:hover { background: var(--color-hover); }
.rail-item.active { background: var(--color-primary-bg); }
.rail-item-icon {
  display: flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; border-radius: var(--radius-sm);
  background: var(--color-background); color: var(--color-text-tertiary);
  font-size: 14px; flex-shrink: 0; transition: all var(--transition-fast);
}
.rail-item:hover .rail-item-icon, .rail-item.active .rail-item-icon { background: var(--color-primary); color: #fff; }
.rail-item-icon.icon-missing { background: var(--color-warning-bg); color: var(--color-warning); }
.rail-item-body { flex: 1; min-width: 0; }
.rail-item-name {
  font-size: 13px; font-weight: 500; color: var(--color-text);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.rail-item.active .rail-item-name { color: var(--color-primary); }
.rail-item-meta { display: flex; gap: 8px; margin-top: 1px; font-size: 11px; color: var(--color-text-tertiary); }
.missing-tag { font-size: 9px; font-weight: 600; padding: 0 4px; margin-left: 4px; border-radius: 3px; background: var(--color-warning-bg); color: var(--color-warning); }
.rail-item-extra { flex-shrink: 0; opacity: 0; transition: opacity var(--transition-fast); }
.rail-item:hover .rail-item-extra { opacity: 0.6; }

.rail-empty { display: flex; flex-direction: column; align-items: center; gap: 6px; padding: 40px 0; }
.rail-empty-icon { font-size: 28px; color: var(--color-text-tertiary); opacity: 0.35; }
.rail-empty span { font-size: 13px; color: var(--color-text-secondary); }
</style>
