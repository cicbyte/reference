<script setup>
import { ref, computed, watch } from 'vue'
import { message } from 'ant-design-vue'
import { useI18n } from 'vue-i18n'
import {
  CheckCircleOutlined,
  WarningOutlined,
  CloseCircleOutlined,
  CloudDownloadOutlined,
  FolderOpenOutlined,
  ToolOutlined,
  ReloadOutlined,
  DeleteOutlined,
} from '@ant-design/icons-vue'
import { formatPath } from '@/utils/path'

const { t } = useI18n()

const props = defineProps({
  open: { type: Boolean, default: false },
  projectDir: { type: String, default: '' },
})
const emit = defineEmits(['update:open'])

const app = window.go?.main?.ReferenceApp
const diagnoses = ref([])
const loading = ref(false)
const fixingKey = ref('')  // refName of item currently being fixed

const okCount = computed(() => diagnoses.value.filter((d) => d.status === 'ok').length)
const issueCount = computed(() => diagnoses.value.length - okCount.value)
const autoFixableCount = computed(() =>
  diagnoses.value.filter((d) => d.status === 'broken-link' || d.status === 'missing-wiki').length
)

async function load() {
  if (!props.projectDir || !app?.DiagnoseProject) return
  loading.value = true
  try {
    const result = await app.DiagnoseProject(props.projectDir)
    // normalize to plain objects so reactivity + filter works reliably
    diagnoses.value = result.map((d) => ({
      refName: d.refName,
      linkName: d.linkName,
      type: d.type,
      remoteUrl: d.remoteUrl,
      cachePath: d.cachePath,
      localPath: d.localPath,
      branch: d.branch,
      targetExists: d.targetExists,
      linkExists: d.linkExists,
      wikiExists: d.wikiExists,
      status: d.status,
      suggestion: d.suggestion,
    }))
  } catch (e) {
    message.error(t('diagnose.diagnoseFailed') + ': ' + e)
  } finally {
    loading.value = false
  }
}

watch(() => [props.open, props.projectDir], ([open]) => {
  if (open) load()
})

function close() {
  emit('update:open', false)
}

function statusIcon(d) {
  if (d.status === 'ok') return CheckCircleOutlined
  if (d.status === 'missing-local') return CloseCircleOutlined
  return WarningOutlined
}
function statusColor(d) {
  if (d.status === 'ok') return 'var(--color-success)'
  if (d.status === 'missing-local' || d.status === 'missing-cache') return 'var(--color-error)'
  return 'var(--color-warning)'
}

async function fixLink(d) {
  fixingKey.value = d.refName
  try {
    await app.FixRepoLink(props.projectDir, d.refName)
    message.success(t('diagnose.fixLinkSuccess'))
    await load()
  } catch (e) {
    message.error(t('diagnose.fixLinkFailed') + ': ' + e)
  } finally {
    fixingKey.value = ''
  }
}

async function reclone(d) {
  fixingKey.value = d.refName
  message.loading({ content: t('diagnose.recloneProgress', { name: d.refName }), key: 'reclone', duration: 0 })
  try {
    const timeout = new Promise((_, reject) =>
      setTimeout(() => reject(new Error(t('diagnose.cloneTimeoutLong'))), 180000),
    )
    await Promise.race([app.RecloneRepo(props.projectDir, d.refName), timeout])
    message.success({ content: t('diagnose.recloneSuccess'), key: 'reclone' })
    await load()
  } catch (e) {
    message.error({ content: t('diagnose.recloneFailed') + ': ' + e, key: 'reclone', duration: 5 })
  } finally {
    fixingKey.value = ''
  }
}

async function relocate(d) {
  try {
    const dir = await app.PickProjectFolder()
    if (!dir) return
    fixingKey.value = d.refName
    await app.RelocateLocalRepo(props.projectDir, d.refName, dir)
    message.success(t('diagnose.relocatePathUpdated', { name: d.refName }))
    await load()
  } catch (e) {
    message.error(t('diagnose.relocatePathFailed') + ': ' + e)
  } finally {
    fixingKey.value = ''
  }
}

async function fixAll() {
  const fixable = diagnoses.value.filter(
    (d) => d.status === 'broken-link' || d.status === 'missing-wiki'
  )
  if (!fixable.length) {
    message.info(t('diagnose.autoFixNone'))
    return
  }
  let ok = 0
  for (const d of fixable) {
    try {
      await app.FixRepoLink(props.projectDir, d.refName)
      ok++
    } catch { /* skip */ }
  }
  message.success(t('diagnose.autoFixDone', { success: ok, total: fixable.length }))
  await load()
}

async function removeRepo(d) {
  fixingKey.value = d.refName
  try {
    await app.RemoveRepoFromProject(props.projectDir, d.refName)
    message.success(t('diagnose.removeSuccess'))
    await load()
  } catch (e) {
    message.error(t('diagnose.removeFailed') + ': ' + e)
  } finally {
    fixingKey.value = ''
  }
}
</script>

<template>
  <a-modal
    :open="open"
    @update:open="emit('update:open', $event)"
    :footer="null"
    :mask-closable="true"
    destroy-on-close
    width="680px"
  >
    <template #title>
      <div class="diag-head">
        <div class="diag-head-icon"><ToolOutlined /></div>
        <div class="diag-head-text">
          <span class="diag-head-title">{{ t('diagnose.title') }}</span>
          <span class="diag-head-sub" v-if="!loading">
            {{ t('diagnose.summaryOk', { ok: okCount, issues: issueCount }) }}
          </span>
        </div>
      </div>
    </template>

    <a-spin :spinning="loading">
      <div v-if="diagnoses.length === 0 && !loading" class="diag-empty">
        {{ t('diagnose.projectEmpty') }}
      </div>

      <div class="diag-list">
        <div
          v-for="d in diagnoses"
          :key="d.refName"
          class="diag-item"
          :class="'st-' + d.status"
        >
          <div class="diag-item-left">
            <component
              :is="statusIcon(d)"
              :style="{ color: statusColor(d), fontSize: '18px' }"
            />
            <div class="diag-item-info">
              <div class="diag-item-name">
                {{ d.refName }}
                <a-tag :color="d.type === 'remote' ? 'blue' : 'green'" class="type-tag">
                  {{ d.type === 'remote' ? t('repos.remote') : t('repos.local') }}
                </a-tag>
              </div>
              <div class="diag-item-detail">
                <span class="detail-line" v-if="d.type === 'remote' && d.remoteUrl">
                  <span class="detail-label">remote</span>
                  <span class="mono">{{ d.remoteUrl }}</span>
                </span>
                <span class="detail-line">
                  <span class="detail-label">target</span>
                  <span class="mono">{{ formatPath(d.type === 'remote' ? d.cachePath : d.localPath) }}</span>
                  <span :class="d.targetExists ? 'ok-mark' : 'bad-mark'">
                    {{ d.targetExists ? '✓' : '✗' }}
                  </span>
                </span>
                <span class="detail-line">
                  <span class="detail-label">link</span>
                  <span>{{ d.linkExists ? t('diagnose.statusHealthy') : t('diagnose.statusBrokenLink') }}</span>
                </span>
              </div>
              <div class="diag-item-suggestion" v-if="d.suggestion">
                {{ d.suggestion }}
              </div>
            </div>
          </div>

          <div class="diag-item-action">
            <!-- broken link / missing wiki → fix button -->
            <a-button
              v-if="d.status === 'broken-link' || d.status === 'missing-wiki'"
              size="small"
              type="primary"
              :loading="fixingKey === d.refName"
              @click="fixLink(d)"
            >
              <template #icon><ToolOutlined /></template>
              {{ t('diagnose.fixLink') }}
            </a-button>
            <!-- missing cache (remote) → reclone button -->
            <a-button
              v-else-if="d.status === 'missing-cache'"
              size="small"
              :loading="fixingKey === d.refName"
              @click="reclone(d)"
            >
              <template #icon><CloudDownloadOutlined /></template>
              {{ t('diagnose.reclone') }}
            </a-button>
            <!-- missing local → relocate button -->
            <a-button
              v-else-if="d.status === 'missing-local'"
              size="small"
              :loading="fixingKey === d.refName"
              @click="relocate(d)"
            >
              <template #icon><FolderOpenOutlined /></template>
              {{ t('diagnose.relocatePickHint') }}
            </a-button>
            <!-- ok → nothing -->
            <span v-else class="ok-text">{{ t('diagnose.statusOk') }}</span>

            <!-- remove (always available) -->
            <a-popconfirm
              :title="t('globalList.removeTitle', { label: t('diagnose.removeRepo'), name: d.refName })"
              :ok-text="t('common.delete')" ok-type="danger" :cancel-text="t('common.cancel')"
              @confirm="removeRepo(d)"
            >
              <a-button size="small" type="text" danger :loading="fixingKey === d.refName" :title="t('diagnose.removeRepo')">
                <template #icon><DeleteOutlined /></template>
              </a-button>
            </a-popconfirm>
          </div>
        </div>
      </div>
    </a-spin>

    <div class="diag-footer">
      <a-button @click="load" :loading="loading" size="small">
        <template #icon><ReloadOutlined /></template>
        {{ t('common.refresh') }}
      </a-button>
      <a-button
        v-if="autoFixableCount > 0"
        type="primary"
        @click="fixAll"
        :disabled="loading"
      >
        <template #icon><ToolOutlined /></template>
        {{ t('diagnose.autoFixAllCount', { n: autoFixableCount }) }}
      </a-button>
      <div class="footer-spacer"></div>
      <a-button @click="close">{{ t('common.close') }}</a-button>
    </div>
  </a-modal>
</template>

<style scoped>
.diag-head { display: flex; align-items: center; gap: 12px; }
.diag-head-icon {
  display: flex; align-items: center; justify-content: center;
  width: 36px; height: 36px; border-radius: var(--radius-md);
  background: var(--color-primary-bg); color: var(--color-primary); font-size: 18px;
}
.diag-head-text { display: flex; flex-direction: column; gap: 1px; flex: 1; }
.diag-head-title { font-size: 16px; font-weight: 600; color: var(--color-text); line-height: 1.2; }
.diag-head-sub { font-size: 12px; color: var(--color-text-tertiary); }

.diag-empty {
  padding: 40px; text-align: center; font-size: 14px; color: var(--color-text-tertiary);
}

.diag-list {
  max-height: 420px; overflow-y: auto;
  display: flex; flex-direction: column; gap: 1px;
}

.diag-item {
  display: flex; align-items: flex-start; justify-content: space-between;
  gap: 12px; padding: 12px 14px; border-radius: var(--radius-md);
  background: var(--color-surface); transition: background var(--transition-fast);
}
.diag-item.st-broken-link,
.diag-item.st-missing-wiki { border-left: 3px solid var(--color-warning); }
.diag-item.st-missing-cache,
.diag-item.st-missing-local { border-left: 3px solid var(--color-error); }
.diag-item.st-ok { opacity: 0.65; }

.diag-item-left { display: flex; gap: 10px; flex: 1; min-width: 0; }
.diag-item-info { flex: 1; min-width: 0; }
.diag-item-name {
  font-size: 14px; font-weight: 600; color: var(--color-text);
  display: flex; align-items: center; gap: 6px;
}
.type-tag { margin: 0; line-height: 18px; font-size: 10px; padding: 0 5px; }

.diag-item-detail { margin-top: 4px; }
.detail-line {
  display: flex; align-items: center; gap: 6px;
  font-size: 11px; color: var(--color-text-tertiary); line-height: 1.6;
}
.detail-label {
  width: 40px; flex-shrink: 0; text-transform: uppercase;
  font-size: 9px; font-weight: 600; letter-spacing: 0.05em;
}
.mono { font-family: 'Cascadia Code', monospace; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ok-mark { color: var(--color-success); font-weight: 700; }
.bad-mark { color: var(--color-error); font-weight: 700; }

.diag-item-suggestion {
  font-size: 12px; color: var(--color-warning); margin-top: 4px;
}

.diag-item-action { flex-shrink: 0; }
.ok-text { font-size: 12px; color: var(--color-success); }

.diag-footer {
  display: flex; align-items: center; gap: var(--spacing-sm);
  margin-top: var(--spacing-md); padding-top: var(--spacing-md);
  border-top: 1px solid var(--color-border-light);
}
.footer-spacer { flex: 1; }
</style>
