<script setup>
import { ref, onMounted, computed } from 'vue'
import { message } from 'ant-design-vue'
import {
  DatabaseOutlined, FolderOpenOutlined, ReloadOutlined,
  CloudServerOutlined, BookOutlined,
} from '@ant-design/icons-vue'

const loading = ref(true)
const saving = ref(false)
const dirty = ref(false)
const form = ref({ reposPath: '', wikiPath: '' })
const original = ref({ reposPath: '', wikiPath: '' })
const paths = ref(null)

// usage stats (async, independent)
const usage = ref({ repos: -1, wiki: -1, db: -1, total: 0 })
const usageLoading = ref(false)

function formatSize(bytes) {
  if (bytes < 0) return '—'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + ' MB'
  return (bytes / 1024 / 1024 / 1024).toFixed(2) + ' GB'
}

const usageCards = computed(() => [
  { key: 'repos', label: '仓库缓存', value: usage.value.repos, path: paths.value?.repos, icon: CloudServerOutlined },
  { key: 'wiki', label: '知识库', value: usage.value.wiki, path: paths.value?.wiki, icon: BookOutlined },
  { key: 'db', label: '数据库', value: usage.value.db, path: paths.value?.db, icon: DatabaseOutlined },
])

onMounted(async () => {
  try {
    if (window.go?.main?.ReferenceApp) {
      const app = window.go.main.ReferenceApp
      const cfg = await app.GetAppConfig()
      form.value.reposPath = cfg.reposPath || ''
      form.value.wikiPath = cfg.wikiPath || ''
      original.value = { ...form.value }
      paths.value = cfg.paths
      loadUsage()
    }
  } catch (e) {
    console.error('GetAppConfig failed:', e)
  } finally {
    loading.value = false
  }
})

async function loadUsage() {
  if (!window.go?.main?.ReferenceApp || !paths.value) return
  usageLoading.value = true
  usage.value = { repos: -1, wiki: -1, db: -1, total: 0 }
  const app = window.go.main.ReferenceApp
  const targets = [
    { key: 'repos', path: paths.value.repos },
    { key: 'wiki', path: paths.value.wiki },
    { key: 'db', path: paths.value.db },
  ]
  // parallel async size queries
  const results = await Promise.allSettled(
    targets.map((t) => app.GetDirSizeAsync(t.path)),
  )
  const next = { repos: 0, wiki: 0, db: 0, total: 0 }
  results.forEach((r, i) => {
    if (r.status === 'fulfilled') {
      next[targets[i].key] = r.value || 0
      next.total += r.value || 0
    }
  })
  usage.value = next
  usageLoading.value = false
}

function markDirty() { dirty.value = true }

async function handleSave() {
  saving.value = true
  try {
    if (window.go?.main?.ReferenceApp) {
      await window.go.main.ReferenceApp.SaveAppConfig({
        reposPath: form.value.reposPath,
        wikiPath: form.value.wikiPath,
      })
      const pathChanged =
        form.value.reposPath !== original.value.reposPath ||
        form.value.wikiPath !== original.value.wikiPath
      original.value = { ...form.value }
      dirty.value = false
      message.success('保存成功')
      if (pathChanged) {
        message.warning('存储路径已变更，重启应用后完全生效', 5)
      }
    }
  } catch (e) {
    message.error('保存失败: ' + e)
  } finally {
    saving.value = false
  }
}

function openDir(dir) {
  window.go.main.ReferenceApp.OpenInExplorer(dir).catch((e) => message.error('打开失败: ' + e))
}
</script>

<template>
  <div class="settings-form">
    <div class="form-header">
      <div class="form-title"><DatabaseOutlined /> 存储</div>
      <div class="form-desc">管理仓库缓存与知识库的存储位置和磁盘占用。</div>
    </div>

    <a-spin :spinning="loading">
      <!-- usage cards -->
      <div class="setting-group">
        <div class="group-head">
          <div class="group-title no-margin">磁盘占用</div>
          <a-button size="small" type="text" :loading="usageLoading" @click="loadUsage">
            <template #icon><ReloadOutlined /></template>
            刷新
          </a-button>
        </div>
        <div class="usage-grid">
          <div v-for="u in usageCards" :key="u.key" class="usage-card">
            <div class="usage-icon"><component :is="u.icon" /></div>
            <div class="usage-body">
              <div class="usage-label">{{ u.label }}</div>
              <div class="usage-value">
                <a-spin v-if="usageLoading" size="small" />
                <span v-else>{{ formatSize(u.value) }}</span>
              </div>
            </div>
            <a-button v-if="u.path" size="small" type="text" title="在文件管理器中打开" @click="openDir(u.path)">
              <FolderOpenOutlined />
            </a-button>
          </div>
        </div>
        <div class="usage-total">
          合计：<strong>{{ formatSize(usage.total) }}</strong>
        </div>
      </div>

      <!-- storage paths -->
      <div class="setting-group">
        <div class="group-title">存储路径</div>

        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">仓库缓存目录</div>
            <div class="row-help">全局仓库克隆缓存位置，留空使用默认值</div>
          </div>
          <a-input
            v-model:value="form.reposPath"
            placeholder="~/.cicbyte/reference/repos"
            class="row-control" style="width: 340px"
            allow-clear @change="markDirty"
          />
        </div>

        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">知识库目录</div>
            <div class="row-help">远程仓库知识文件存储位置，留空使用默认值</div>
          </div>
          <a-input
            v-model:value="form.wikiPath"
            placeholder="~/.cicbyte/reference/wiki"
            class="row-control" style="width: 340px"
            allow-clear @change="markDirty"
          />
        </div>

        <div class="group-actions">
          <a-button type="primary" :loading="saving" :disabled="!dirty" @click="handleSave">保存</a-button>
        </div>
      </div>

      <a-alert
        type="info" show-icon
        message="修改存储路径后，需重启应用以触发数据迁移（已有缓存的路径记录会自动更新）。"
      />
    </a-spin>
  </div>
</template>

<style scoped>
@import './settings-shared.css';

.group-head {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 12px;
}
.group-title.no-margin { margin-bottom: 0; }

.usage-grid {
  display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px;
}
.usage-card {
  display: flex; align-items: center; gap: 10px;
  padding: 12px; border-radius: var(--radius-sm);
  background: var(--color-background); border: 1px solid var(--color-border-light);
}
.usage-icon {
  display: flex; align-items: center; justify-content: center;
  width: 36px; height: 36px; border-radius: var(--radius-sm);
  background: var(--color-primary-bg); color: var(--color-primary);
  font-size: 18px; flex-shrink: 0;
}
.usage-body { flex: 1; min-width: 0; }
.usage-label { font-size: 11px; color: var(--color-text-tertiary); }
.usage-value { font-size: 16px; font-weight: 600; color: var(--color-text); margin-top: 2px; }
.usage-total {
  margin-top: 12px; padding-top: 10px; border-top: 1px solid var(--color-border-light);
  font-size: 13px; color: var(--color-text-secondary); text-align: right;
}
.usage-total strong { color: var(--color-text); font-size: 15px; }
</style>
