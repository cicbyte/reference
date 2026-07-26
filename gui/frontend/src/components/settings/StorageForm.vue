<script setup>
import { ref, onMounted, computed } from 'vue'
import { message } from 'ant-design-vue'
import {
  DatabaseOutlined, FolderOpenOutlined, ReloadOutlined,
  CloudServerOutlined, BookOutlined,
} from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'

const app = window.go?.main?.ReferenceApp
const { t } = useI18n()

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
  { key: 'repos', labelKey: 'settings.storage.reposCache', value: usage.value.repos, path: paths.value?.repos, icon: CloudServerOutlined },
  { key: 'wiki', labelKey: 'settings.storage.wikiStore', value: usage.value.wiki, path: paths.value?.wiki, icon: BookOutlined },
  { key: 'db', labelKey: 'settings.storage.database', value: usage.value.db, path: paths.value?.db, icon: DatabaseOutlined },
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
      message.success(t('common.saved'))
      if (pathChanged) {
        message.warning(t('settings.storage.pathChangedHint'), 5)
      }
    }
  } catch (e) {
    message.error(t('common.saveFailed') + ': ' + e)
  } finally {
    saving.value = false
  }
}

function openDir(dir) {
  app?.OpenInExplorer(dir)?.catch((e) => message.error(t('common.openFailed') + ': ' + e))
}

async function pickFolder(field, title) {
  if (!app) return
  try {
    // Prefer the titled PickFolder; fall back to PickProjectFolder for older backends.
    const pick = app.PickFolder || app.PickProjectFolder
    if (!pick) { message.error(t('settings.storage.backendUnsupported')); return }
    const dir = await (app.PickFolder ? app.PickFolder(title) : app.PickProjectFolder())
    if (dir) {
      form.value[field] = dir
      markDirty()
    }
  } catch (e) {
    message.error(t('settings.storage.pickFailed') + ': ' + e)
  }
}
</script>

<template>
  <div class="settings-form">
    <div class="form-header">
      <div class="form-title"><DatabaseOutlined /> {{ t('settings.storage.title') }}</div>
      <div class="form-desc">{{ t('settings.storage.desc') }}</div>
    </div>

    <a-spin :spinning="loading">
      <!-- usage cards -->
      <div class="setting-group">
        <div class="group-head">
          <div class="group-title no-margin">{{ t('settings.storage.usage') }}</div>
          <a-button size="small" type="text" :loading="usageLoading" @click="loadUsage">
            <template #icon><ReloadOutlined /></template>
            {{ t('common.refresh') }}
          </a-button>
        </div>
        <div class="usage-grid">
          <div v-for="u in usageCards" :key="u.key" class="usage-card">
            <div class="usage-icon"><component :is="u.icon" /></div>
            <div class="usage-body">
              <div class="usage-label">{{ t(u.labelKey) }}</div>
              <div class="usage-value">
                <a-spin v-if="usageLoading" size="small" />
                <span v-else>{{ formatSize(u.value) }}</span>
              </div>
            </div>
            <a-button v-if="u.path" size="small" type="text" :title="t('common.openInExplorer')" @click="openDir(u.path)">
              <FolderOpenOutlined />
            </a-button>
          </div>
        </div>
        <div class="usage-total">
          {{ t('settings.storage.total') }}：<strong>{{ formatSize(usage.total) }}</strong>
        </div>
      </div>

      <!-- storage paths -->
      <div class="setting-group">
        <div class="group-title">{{ t('settings.storage.paths') }}</div>

        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">{{ t('settings.storage.reposPath') }}</div>
            <div class="row-help">{{ t('settings.storage.reposPathHelp') }}</div>
          </div>
          <a-input-group compact class="row-control">
            <a-input
              v-model:value="form.reposPath"
              placeholder="~/.cicbyte/apps/reference/repos"
              style="width: calc(100% - 36px)"
              allow-clear @change="markDirty"
            />
            <a-button style="width: 36px" :title="t('settings.storage.pickFolder')" @click="pickFolder('reposPath', t('settings.storage.pickRepos'))">
              <FolderOpenOutlined />
            </a-button>
          </a-input-group>
        </div>

        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">{{ t('settings.storage.wikiPath') }}</div>
            <div class="row-help">{{ t('settings.storage.wikiPathHelp') }}</div>
          </div>
          <a-input-group compact class="row-control">
            <a-input
              v-model:value="form.wikiPath"
              placeholder="~/.cicbyte/apps/reference/wiki"
              style="width: calc(100% - 36px)"
              allow-clear @change="markDirty"
            />
            <a-button style="width: 36px" :title="t('settings.storage.pickFolder')" @click="pickFolder('wikiPath', t('settings.storage.pickWiki'))">
              <FolderOpenOutlined />
            </a-button>
          </a-input-group>
        </div>

        <div class="group-actions">
          <a-button type="primary" :loading="saving" :disabled="!dirty" @click="handleSave">{{ t('common.save') }}</a-button>
        </div>
      </div>

      <a-alert
        type="info" show-icon
        :message="t('settings.storage.pathChangedHint')"
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

/* path picker: input + button group */
.row-control { display: flex; width: 360px; }
.row-control :deep(.ant-input) { border-radius: var(--radius-sm) 0 0 var(--radius-sm); }
.row-control :deep(.ant-btn) { border-radius: 0 var(--radius-sm) var(--radius-sm) 0; display: flex; align-items: center; justify-content: center; }
</style>
