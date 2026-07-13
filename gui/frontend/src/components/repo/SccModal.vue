<script setup>
import { ref, watch, computed } from 'vue'
import { message } from 'ant-design-vue'
import { useI18n } from 'vue-i18n'
import { CodeOutlined, FireOutlined, FileTextOutlined } from '@ant-design/icons-vue'

const { t } = useI18n()

const props = defineProps({
  open: { type: Boolean, default: false },
  repoName: { type: String, default: '' },
})
const emit = defineEmits(['update:open'])

const loading = ref(false)
const result = ref(null)

// palette for language bars — cycled by index
const LANG_COLORS = [
  '#3b82f6', '#a855f7', '#06b6d4', '#f59e0b', '#10b981',
  '#ec4899', '#6366f1', '#14b8a6', '#f97316', '#84cc16',
]

const totalCode = computed(() =>
  (result.value?.languages || []).reduce((s, l) => s + l.code, 0),
)
const totalFiles = computed(() =>
  (result.value?.languages || []).reduce((s, l) => s + l.count, 0),
)

// languages sorted by code desc, with percentage + color attached
const sortedLangs = computed(() => {
  const langs = (result.value?.languages || []).slice().sort((a, b) => b.code - a.code)
  return langs.map((l, i) => ({
    ...l,
    pct: totalCode.value > 0 ? Math.round((l.code / totalCode.value) * 100) : 0,
    color: LANG_COLORS[i % LANG_COLORS.length],
  }))
})

// top files sorted by complexity desc, capped at 15
const topFiles = computed(() => {
  const files = (result.value?.topFiles || []).slice()
  files.sort((a, b) => (b.complexity || 0) - (a.complexity || 0))
  return files.slice(0, 15)
})

const maxComplexity = computed(() =>
  topFiles.value.reduce((m, f) => Math.max(m, f.complexity || 0), 1),
)

async function runScc() {
  if (!props.repoName) return
  loading.value = true
  result.value = null
  try {
    if (window.go?.main?.ReferenceApp) {
      result.value = await window.go.main.ReferenceApp.RunSCC(props.repoName)
    }
  } catch (e) {
    message.error(t('scc.failed') + ': ' + e)
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.open, props.repoName],
  ([open]) => {
    if (open && props.repoName) runScc()
    if (!open) result.value = null
  },
)

function fmt(n) {
  return (n || 0).toLocaleString()
}
</script>

<template>
  <a-modal
    :open="open"
    :footer="null"
    :mask-closable="true"
    destroy-on-close
    width="680px"
    wrap-class-name="scc-modal-wrap"
    @update:open="emit('update:open', $event)"
  >
    <!-- custom header strip -->
    <template #title>
      <div class="scc-head">
        <div class="scc-head-icon"><CodeOutlined /></div>
        <div class="scc-head-text">
          <span class="scc-head-title">{{ t('scc.title') }}</span>
          <span class="scc-head-sub">{{ repoName }}</span>
        </div>
      </div>
    </template>

    <a-spin :spinning="loading">
      <div v-if="result" class="scc-body">
        <!-- summary strip -->
        <div class="summary-strip">
          <div class="sum-item">
            <CodeOutlined class="sum-icon" />
            <span class="sum-val">{{ fmt(totalCode) }}</span>
            <span class="sum-lbl">{{ t('scc.totalCode') }}</span>
          </div>
          <div class="sum-sep"></div>
          <div class="sum-item">
            <FileTextOutlined class="sum-icon" />
            <span class="sum-val">{{ fmt(totalFiles) }}</span>
            <span class="sum-lbl">{{ t('scc.totalFiles') }}</span>
          </div>
          <div class="sum-sep"></div>
          <div class="sum-item">
            <span class="sum-val">{{ sortedLangs.length }}</span>
            <span class="sum-lbl">{{ t('scc.langs') }}</span>
          </div>
        </div>

        <!-- language bars -->
        <div v-if="sortedLangs.length" class="lang-section">
          <div class="section-label">{{ t('scc.languages') }}</div>
          <div v-for="lang in sortedLangs" :key="lang.name" class="lang-bar-row">
            <div class="lang-info">
              <span class="lang-dot" :style="{ background: lang.color }"></span>
              <span class="lang-name">{{ lang.name }}</span>
              <span class="lang-pct">{{ lang.pct }}%</span>
            </div>
            <div class="lang-track">
              <div class="lang-fill" :style="{ width: lang.pct + '%', background: lang.color }"></div>
            </div>
            <div class="lang-stats">
              <span>{{ fmt(lang.code) }} {{ t('scc.lines') }}</span>
              <span class="lang-files">{{ lang.count }} {{ t('scc.files') }}</span>
            </div>
          </div>
        </div>

        <!-- top files -->
        <div v-if="topFiles.length" class="files-section">
          <div class="section-label">
            <FireOutlined /> {{ t('scc.topComplexity', { n: topFiles.length }) }}
          </div>
          <div class="file-list">
            <div v-for="(f, i) in topFiles" :key="f.file" class="file-row">
              <span class="file-rank">{{ i + 1 }}</span>
              <div class="file-main">
                <div class="file-name" :title="f.file">{{ f.file }}</div>
                <div class="file-meta">
                  <a-tag class="file-lang-tag">{{ f.language }}</a-tag>
                  <span>{{ fmt(f.code) }} {{ t('scc.lines') }}</span>
                </div>
              </div>
              <div class="file-complexity">
                <div class="cx-bar">
                  <div
                    class="cx-fill"
                    :style="{ width: (maxComplexity > 0 ? (f.complexity / maxComplexity) * 100 : 0) + '%' }"
                  ></div>
                </div>
                <span class="cx-val" :class="{ high: f.complexity > 50 }">{{ f.complexity }}</span>
              </div>
            </div>
          </div>
        </div>

        <a-empty v-if="!loading && !sortedLangs.length" :description="t('scc.noData')" />
      </div>
    </a-spin>
  </a-modal>
</template>

<style scoped>
/* ---- header ---- */
.scc-head { display: flex; align-items: center; gap: 12px; }
.scc-head-icon {
  display: flex; align-items: center; justify-content: center;
  width: 36px; height: 36px; border-radius: var(--radius-md);
  background: var(--color-primary-bg); color: var(--color-primary); font-size: 18px;
}
.scc-head-text { display: flex; flex-direction: column; gap: 1px; }
.scc-head-title { font-size: 16px; font-weight: 600; color: var(--color-text); line-height: 1.2; }
.scc-head-sub { font-size: 12px; color: var(--color-text-tertiary); font-family: 'Cascadia Code', monospace; }

.scc-body { display: flex; flex-direction: column; gap: var(--spacing-md); padding-top: 4px; }

/* ---- summary strip ---- */
.summary-strip {
  display: flex; align-items: center; gap: var(--spacing-md);
  padding: 14px 18px;
  background: var(--color-surface); border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
}
.sum-item { display: flex; align-items: center; gap: 8px; }
.sum-icon { font-size: 16px; color: var(--color-text-tertiary); }
.sum-val { font-size: 20px; font-weight: 700; color: var(--color-text); }
.sum-lbl { font-size: 12px; color: var(--color-text-tertiary); }
.sum-sep { width: 1px; height: 28px; background: var(--color-border); }

/* ---- section label ---- */
.section-label {
  display: flex; align-items: center; gap: 6px;
  font-size: 12px; font-weight: 600; text-transform: uppercase;
  letter-spacing: 0.05em; color: var(--color-text-tertiary);
  margin-bottom: 10px;
}

/* ---- language bars ---- */
.lang-section { display: flex; flex-direction: column; gap: 8px; }
.lang-bar-row { display: grid; grid-template-columns: 160px 1fr 110px; align-items: center; gap: 12px; }
.lang-info { display: flex; align-items: center; gap: 8px; min-width: 0; }
.lang-dot { width: 8px; height: 8px; border-radius: 2px; flex-shrink: 0; }
.lang-name { font-size: 13px; font-weight: 500; color: var(--color-text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.lang-pct { font-size: 11px; color: var(--color-text-tertiary); margin-left: auto; }
.lang-track { height: 6px; background: var(--color-surface-raised); border-radius: 3px; overflow: hidden; }
.lang-fill { height: 100%; border-radius: 3px; transition: width 0.4s ease; }
.lang-stats { display: flex; gap: 8px; font-size: 11px; color: var(--color-text-secondary); justify-content: flex-end; }
.lang-files { color: var(--color-text-tertiary); }

/* ---- files ---- */
.files-section { display: flex; flex-direction: column; }
.file-list { display: flex; flex-direction: column; gap: 2px; }
.file-row {
  display: flex; align-items: center; gap: 12px;
  padding: 8px 10px; border-radius: var(--radius-sm);
  transition: background var(--transition-fast);
}
.file-row:hover { background: var(--color-hover); }
.file-rank {
  font-size: 12px; font-weight: 700; color: var(--color-text-tertiary);
  width: 20px; text-align: center; flex-shrink: 0;
}
.file-main { flex: 1; min-width: 0; }
.file-name {
  font-size: 13px; color: var(--color-text); font-family: 'Cascadia Code', monospace;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.file-meta { display: flex; align-items: center; gap: 8px; margin-top: 2px; font-size: 11px; color: var(--color-text-tertiary); }
.file-lang-tag { margin: 0; line-height: 16px; padding: 0 6px; font-size: 10px; }

.file-complexity { display: flex; align-items: center; gap: 8px; width: 100px; flex-shrink: 0; }
.cx-bar { flex: 1; height: 4px; background: var(--color-surface-raised); border-radius: 2px; overflow: hidden; }
.cx-fill { height: 100%; background: var(--color-warning); border-radius: 2px; }
.cx-val { font-size: 12px; font-weight: 600; color: var(--color-text-secondary); min-width: 28px; text-align: right; }
.cx-val.high { color: var(--color-error); }
</style>
