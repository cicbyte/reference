<script setup>
import { ref, onMounted } from 'vue'
import {
  FolderOutlined,
  DatabaseOutlined,
  CloudServerOutlined,
  ReadOutlined,
  HddOutlined,
  WarningOutlined,
} from '@ant-design/icons-vue'

const loading = ref(true)
const stats = ref(null)
const cacheSize = ref(null)  // null = loading
const wikiSize = ref(null)
const app = window.go?.main?.ReferenceApp

function fmtSize(bytes) {
  if (bytes == null) return '...'
  if (!bytes) return '0 B'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + ' MB'
  return (bytes / 1024 / 1024 / 1024).toFixed(2) + ' GB'
}

onMounted(async () => {
  try {
    if (app?.GlobalStats) {
      stats.value = await app.GlobalStats()
      // async fetch directory sizes
      if (app.GetDirSizeAsync && stats.value.repos_dir) {
        app.GetDirSizeAsync(stats.value.repos_dir).then((s) => { cacheSize.value = s }).catch(() => { cacheSize.value = 0 })
      }
      if (app.GetDirSizeAsync && stats.value.wiki_dir) {
        app.GetDirSizeAsync(stats.value.wiki_dir).then((s) => { wikiSize.value = s }).catch(() => { wikiSize.value = 0 })
      }
    }
  } catch (e) {
    console.error('Global stats failed:', e)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="stats-view">
    <div class="page-header">
      <h2>全局统计</h2>
    </div>

    <a-spin :spinning="loading">
      <div v-if="stats" class="stats-grid">
        <!-- projects -->
        <div class="stat-card">
          <div class="sc-icon sc-blue"><FolderOutlined /></div>
          <div class="sc-body">
            <div class="sc-value">{{ stats.total_projects }}</div>
            <div class="sc-label">项目总数</div>
          </div>
          <div class="sc-extra" v-if="stats.deleted_projects > 0">
            <span class="extra-warn"><WarningOutlined /> {{ stats.deleted_projects }} 失效</span>
          </div>
        </div>

        <!-- repos -->
        <div class="stat-card">
          <div class="sc-icon sc-cyan"><CloudServerOutlined /></div>
          <div class="sc-body">
            <div class="sc-value">{{ stats.total_repos }}</div>
            <div class="sc-label">缓存仓库</div>
          </div>
        </div>

        <!-- cache size -->
        <div class="stat-card">
          <div class="sc-icon sc-purple"><HddOutlined /></div>
          <div class="sc-body">
            <div class="sc-value">{{ fmtSize(cacheSize) }}</div>
            <div class="sc-label">仓库缓存</div>
          </div>
          <div class="sc-path" :title="stats.repos_dir">{{ stats.repos_dir }}</div>
        </div>

        <!-- wiki size -->
        <div class="stat-card">
          <div class="sc-icon sc-green"><ReadOutlined /></div>
          <div class="sc-body">
            <div class="sc-value">{{ fmtSize(wikiSize) }}</div>
            <div class="sc-label">知识库</div>
          </div>
          <div class="sc-path" :title="stats.wiki_dir">{{ stats.wiki_dir }}</div>
        </div>

        <!-- db size -->
        <div class="stat-card">
          <div class="sc-icon sc-orange"><DatabaseOutlined /></div>
          <div class="sc-body">
            <div class="sc-value">{{ fmtSize(stats.db_size) }}</div>
            <div class="sc-label">数据库</div>
          </div>
        </div>

        <!-- existing projects -->
        <div class="stat-card">
          <div class="sc-icon sc-blue"><FolderOutlined /></div>
          <div class="sc-body">
            <div class="sc-value">{{ stats.existing_projects }}</div>
            <div class="sc-label">活跃项目</div>
          </div>
        </div>
      </div>
    </a-spin>
  </div>
</template>

<style scoped>
.stats-view { width: 100%; }

.page-header { margin-bottom: var(--spacing-lg); }
.page-header h2 { font-size: 20px; font-weight: 600; color: var(--color-text); margin: 0; }

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: var(--spacing-md);
}

.stat-card {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: 18px 20px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
  position: relative;
  flex-wrap: wrap;
}
.stat-card:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-sm);
  transform: translateY(-1px);
}

.sc-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md);
  font-size: 22px;
  flex-shrink: 0;
}
.sc-blue   { background: rgba(59, 130, 246, 0.12); color: #3b82f6; }
.sc-cyan   { background: rgba(6, 182, 212, 0.12); color: #06b6d4; }
.sc-green  { background: rgba(22, 163, 74, 0.12); color: var(--color-success); }
.sc-purple { background: rgba(168, 85, 247, 0.12); color: #a855f7; }
.sc-orange { background: rgba(249, 115, 22, 0.12); color: #f97316; }

.sc-body { flex: 1; min-width: 0; }
.sc-value {
  font-size: 24px;
  font-weight: 700;
  color: var(--color-text);
  line-height: 1.1;
}
.sc-label {
  font-size: 12px;
  color: var(--color-text-secondary);
  margin-top: 2px;
}

.sc-extra {
  position: absolute;
  top: 8px;
  right: 12px;
}
.extra-warn {
  font-size: 11px;
  color: var(--color-warning);
  display: inline-flex;
  align-items: center;
  gap: 3px;
}

.sc-path {
  width: 100%;
  margin-top: 6px;
  font-size: 10px;
  color: var(--color-text-tertiary);
  font-family: 'Cascadia Code', monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
