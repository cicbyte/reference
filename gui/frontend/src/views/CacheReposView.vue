<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  ReloadOutlined,
  FolderOpenOutlined,
  DeleteOutlined,
  DatabaseOutlined,
  CloudServerOutlined,
} from '@ant-design/icons-vue'
import { useProjectStore } from '../stores/project'

const router = useRouter()
const project = useProjectStore()
const repos = ref([])
const loading = ref(true)
const app = window.go?.main?.ReferenceApp

const totalSize = computed(() => repos.value.reduce((s, r) => s + (r.size || 0), 0))

function fmtSize(bytes) {
  if (!bytes) return '—'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + ' MB'
  return (bytes / 1024 / 1024 / 1024).toFixed(2) + ' GB'
}

const columns = [
  { title: '仓库', dataIndex: 'name', key: 'name' },
  { title: '大小', key: 'size', width: 100, sorter: (a, b) => a.size - b.size },
  { title: '引用', key: 'refCount', width: 70, align: 'center', sorter: (a, b) => a.refCount - b.refCount },
  { title: '分支', dataIndex: 'branch', key: 'branch', width: 120 },
  { title: 'Commit', dataIndex: 'commit', key: 'commit', width: 100 },
  { title: '操作', key: 'actions', width: 100, align: 'center', fixed: 'right' },
]

async function loadRepos() {
  loading.value = true
  try {
    if (app) {
      repos.value = await app.ListCachedRepos()
    }
  } catch (e) {
    console.error('ListCachedRepos failed:', e)
    message.error('加载失败: ' + e)
  } finally {
    loading.value = false
  }
}

onMounted(loadRepos)

function browseCode(record) {
  // need a project to browse — switch to first referencing project then navigate
  const refName = record.name
  if (record.projects && record.projects.length > 0) {
    project.switchTo(record.projects[0]).then(() => {
      router.push('/repos/browse/' + refName)
    })
  } else {
    message.warning('该缓存没有被任何项目引用，无法浏览')
  }
}

async function purge(record) {
  try {
    await app.PurgeCachedRepo(record.cachePath)
    message.success(`已清理 ${record.name} (${fmtSize(record.size)})`)
    await loadRepos()
  } catch (e) {
    message.error('清理失败: ' + e)
  }
}
</script>

<template>
  <div class="cache-view">
    <div class="page-header">
      <h2>仓库缓存</h2>
      <a-button @click="loadRepos" :loading="loading">
        <template #icon><ReloadOutlined /></template>
        刷新
      </a-button>
    </div>

    <!-- summary strip -->
    <div class="summary-strip">
      <div class="sum-item">
        <DatabaseOutlined class="sum-icon" />
        <span class="sum-val">{{ repos.length }}</span>
        <span class="sum-lbl">个缓存仓库</span>
      </div>
      <div class="sum-sep"></div>
      <div class="sum-item">
        <CloudServerOutlined class="sum-icon" />
        <span class="sum-val">{{ fmtSize(totalSize) }}</span>
        <span class="sum-lbl">总占用</span>
      </div>
    </div>

    <a-table
      :columns="columns"
      :data-source="repos"
      :loading="loading"
      :pagination="false"
      :scroll="{ x: 700 }"
      size="middle"
      :row-key="(r) => r.cachePath"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'name'">
          <div class="repo-name-cell">
            <CloudServerOutlined class="repo-icon" />
            <div>
              <div class="repo-name">{{ record.name }}</div>
              <div class="repo-path" :title="record.cachePath">{{ record.cachePath }}</div>
            </div>
          </div>
        </template>
        <template v-if="column.key === 'size'">
          <span class="size-cell">{{ fmtSize(record.size) }}</span>
        </template>
        <template v-if="column.key === 'refCount'">
          <a-tooltip :title="record.projects && record.projects.length ? record.projects.join('\n') : '无引用'">
            <a-tag :color="record.refCount > 0 ? 'blue' : 'orange'">{{ record.refCount }}</a-tag>
          </a-tooltip>
        </template>
        <template v-if="column.key === 'commit'">
          <span class="mono-text" v-if="record.commit">{{ record.commit }}</span>
          <span v-else class="muted">—</span>
        </template>
        <template v-if="column.key === 'actions'">
          <a-tooltip title="浏览代码">
            <a-button size="small" type="text" :disabled="!record.refCount" @click="browseCode(record)">
              <template #icon><FolderOpenOutlined /></template>
            </a-button>
          </a-tooltip>
          <a-popconfirm
            :title="`确认清理 ${record.name} (${fmtSize(record.size)})？`"
            ok-text="清理"
            ok-type="danger"
            cancel-text="取消"
            @confirm="purge(record)"
          >
            <a-tooltip title="清理缓存">
              <a-button size="small" type="text" danger>
                <template #icon><DeleteOutlined /></template>
              </a-button>
            </a-tooltip>
          </a-popconfirm>
        </template>
      </template>

      <template #emptyText>
        <a-empty description="暂无缓存仓库">
          <template #image><DatabaseOutlined style="font-size: 48px; color: var(--color-text-tertiary); opacity: 0.3;" /></template>
        </a-empty>
      </template>
    </a-table>
  </div>
</template>

<style scoped>
.cache-view { width: 100%; }

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
}
.page-header h2 { font-size: 20px; font-weight: 600; color: var(--color-text); margin: 0; }

.summary-strip {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: 14px 18px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  margin-bottom: var(--spacing-lg);
}
.sum-item { display: flex; align-items: center; gap: 8px; }
.sum-icon { font-size: 16px; color: var(--color-text-tertiary); }
.sum-val { font-size: 20px; font-weight: 700; color: var(--color-text); }
.sum-lbl { font-size: 12px; color: var(--color-text-tertiary); }
.sum-sep { width: 1px; height: 28px; background: var(--color-border); }

.repo-name-cell { display: flex; align-items: center; gap: 10px; }
.repo-icon { font-size: 16px; color: var(--color-text-tertiary); flex-shrink: 0; }
.repo-name { font-size: 14px; font-weight: 500; color: var(--color-text); }
.repo-path {
  font-size: 11px; color: var(--color-text-tertiary);
  font-family: 'Cascadia Code', monospace;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
  max-width: 360px;
}

.size-cell { font-weight: 600; color: var(--color-text); }
.mono-text { font-family: 'Cascadia Code', monospace; font-size: 12px; color: var(--color-text-secondary); }
.muted { color: var(--color-text-tertiary); }
</style>
