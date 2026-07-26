<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { message } from 'ant-design-vue'
import {
  ReadOutlined,
  FileTextOutlined,
  DeleteOutlined,
  FolderOpenOutlined,
  EyeOutlined,
  CodeOutlined,
} from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'
import { useLayoutStore } from '@/stores/layout'
import { formatPath } from '@/utils/path'
import { renderMarkdown } from '@/utils/markdown'
import { useMermaid } from '@/composables/useMermaid'
import MermaidModal from '@/components/shared/MermaidModal.vue'
import WikiRail from './components/WikiRail.vue'

const app = window.go?.main?.ReferenceApp
const layout = useLayoutStore()
const { t } = useI18n()

// ---- shared mermaid (render + enlarge modal) ----
const { modalOpen, modalSvg, zoom, panX, panY, renderMermaid } = useMermaid()

// ---- data ----
const entries = ref([])
const loading = ref(true)
const selectedRepoKey = ref('')
const selectedFileKey = ref('')
const content = ref('')
const rawContent = ref('')
const contentLoading = ref(false)
const viewMode = ref('render')

// ---- flat repo list (local repos have no platform/namespace grouping) ----
const repos = computed(() => {
  const map = {}
  for (const e of entries.value) {
    if (!map[e.repoName]) map[e.repoName] = []
    map[e.repoName].push(e)
  }
  return Object.keys(map).sort().map((rn) => {
    const files = map[rn].sort((a, b) => {
      if (a.fileName === 'reference.md') return -1
      if (b.fileName === 'reference.md') return 1
      return a.fileName.localeCompare(b.fileName)
    })
    return { repoName: rn, files, fileCount: files.length }
  })
})

const selectedRepoFiles = computed(() => {
  if (!selectedRepoKey.value) return []
  return entries.value
    .filter((e) => e.repoName === selectedRepoKey.value)
    .sort((a, b) => {
      if (a.fileName === 'reference.md') return -1
      if (b.fileName === 'reference.md') return 1
      return a.fileName.localeCompare(b.fileName)
    })
})

const selectedEntry = computed(() =>
  entries.value.find((e) => e.source + '|' + e.relPath === selectedFileKey.value),
)

function selectRepo(repoName) {
  selectedRepoKey.value = repoName
  const files = entries.value.filter((e) => e.repoName === repoName)
  const refMd = files.find((f) => f.fileName === 'reference.md')
  const target = refMd || files[0]
  if (target) selectFile(target)
}

function selectFile(entry) {
  selectedFileKey.value = entry.source + '|' + entry.relPath
  viewMode.value = 'render'
  loadContent(entry)
  layout.setFooterItem('wiki', 'Wiki', formatPath(entry.fullPath || entry.relPath), ReadOutlined)
}

onUnmounted(() => {
  layout.clearFooterItem('wiki')
})

async function loadEntries() {
  loading.value = true
  try {
    if (app?.ListWikiEntries) {
      entries.value = await app.ListWikiEntries('local')
      entries.value.forEach((e) => fetchStatus(e))
    }
  } catch (e) {
    message.error(t('localWiki.loadFailed') + ': ' + e)
  } finally {
    loading.value = false
  }
}

async function fetchStatus(entry) {
  try {
    if (app?.CheckWikiStatus) {
      const status = await app.CheckWikiStatus(entry.source, entry.relPath)
      const idx = entries.value.findIndex((e) => e.source === entry.source && e.relPath === entry.relPath)
      if (idx >= 0) {
        entries.value[idx] = { ...entries.value[idx], status }
      }
    }
  } catch { /* leave status empty */ }
}

async function deleteEntry(entry) {
  try {
    if (app?.DeleteWikiEntry) {
      await app.DeleteWikiEntry(entry.source, entry.relPath)
      message.success(t('localWiki.deleteSuccess'))
      if (selectedFileKey.value === entry.source + '|' + entry.relPath) {
        selectedFileKey.value = ''
        content.value = ''
      }
      await loadEntries()
    }
  } catch (e) {
    message.error(t('localWiki.deleteFailed') + ': ' + e)
  }
}

function statusLabel(s) {
  return {
    ok: t('localWiki.statusOk'),
    empty: t('localWiki.statusEmpty'),
    'no-fm': t('localWiki.statusNoFm'),
  }[s] || ''
}

async function openInExplorer(entry) {
  if (!entry?.fullPath) return
  const parts = formatPath(entry.fullPath).split('/').filter(Boolean)
  parts.pop()
  const dir = parts.join('/')
  try {
    if (app?.OpenInExplorer) await app.OpenInExplorer(dir)
  } catch (e) {
    message.error(t('common.openFailed') + ': ' + e)
  }
}

async function loadContent(entry) {
  if (!entry) return
  contentLoading.value = true
  content.value = ''
  rawContent.value = ''
  try {
    if (app?.ReadWikiEntry) {
      const raw = await app.ReadWikiEntry(entry.source, entry.relPath)
      rawContent.value = raw
      content.value = stripFrontmatter(raw)
    }
  } catch (e) {
    message.error(t('localWiki.readFailed') + ': ' + e)
  } finally {
    contentLoading.value = false
    if (viewMode.value === 'render') renderMermaid()
  }
}

function stripFrontmatter(md) {
  if (!md.startsWith('---')) return md
  const rest = md.slice(3)
  const end = rest.indexOf('\n---')
  if (end < 0) return md
  return rest.slice(end + 4).replace(/^\s*\n/, '')
}

const renderedContent = computed(() => renderMarkdown(content.value))

watch(viewMode, (mode) => {
  if (mode === 'render') renderMermaid()
})

onMounted(loadEntries)
</script>

<template>
  <div class="wiki-view">
    <!-- col 1: repo rail (flat, no platform grouping) -->
    <WikiRail
      :repos="repos"
      :selected-repo-key="selectedRepoKey"
      :loading="loading"
      :entries="entries"
      @select-repo="selectRepo"
      @reload="loadEntries"
    />

    <!-- col 2: file list for selected repo -->
    <aside v-if="selectedRepoKey" class="wiki-files">
      <div class="wf-head">
        <FileTextOutlined class="wf-icon" />
        <span class="wf-title">{{ selectedRepoKey }}</span>
      </div>
      <div class="wf-list">
        <a-dropdown
          v-for="file in selectedRepoFiles"
          :key="file.source + '|' + file.relPath"
          :trigger="['contextmenu']"
        >
          <div
            class="wf-item"
            :class="{
              active: selectedFileKey === file.source + '|' + file.relPath,
              'wf-problem': file.status && file.status !== 'ok',
            }"
            :title="file.fileName"
            @click="selectFile(file)"
          >
            <div class="wf-item-icon">
              <ReadOutlined v-if="file.fileName === 'reference.md'" />
              <FileTextOutlined v-else />
            </div>
            <div class="wf-item-body">
              <div class="wf-item-name">
                {{ file.fileName === 'reference.md' ? t('localWiki.architectureOverview') : file.fileName.replace('.md', '') }}
              </div>
              <div class="wf-item-meta">
                <span v-if="file.fileName !== 'reference.md'" class="wf-topic-tag">{{ t('localWiki.topic') }}</span>
                <span v-if="file.gitStatus && file.gitStatus !== 'committed'" class="wf-git-tag" :class="'git-' + file.gitStatus">
                  {{ file.gitStatus === 'modified' ? t('localWiki.gitModified') : t('localWiki.gitUntracked') }}
                </span>
                <span v-if="file.status && file.status !== 'ok'" class="wf-status-tag" :class="'st-' + file.status">
                  {{ statusLabel(file.status) }}
                </span>
                <span v-if="file.exploredAt">{{ file.exploredAt }}</span>
              </div>
            </div>
            <a-popconfirm
              v-if="file.status && file.status !== 'ok'"
              :title="t('localWiki.deleteConfirm', { name: file.fileName })"
              :ok-text="t('common.delete')" ok-type="danger" :cancel-text="t('common.cancel')"
              @confirm="deleteEntry(file)"
            >
              <button class="wf-delete" :title="t('common.delete')" @click.stop>
                <DeleteOutlined />
              </button>
            </a-popconfirm>
          </div>
          <template #overlay>
            <a-menu>
              <a-menu-item @click="openInExplorer(file)">
                <FolderOpenOutlined /> {{ t('common.openInExplorer') }}
              </a-menu-item>
              <a-menu-divider />
              <a-menu-item danger @click="deleteEntry(file)">
                <DeleteOutlined /> {{ t('common.delete') }}
              </a-menu-item>
            </a-menu>
          </template>
        </a-dropdown>
      </div>
    </aside>

    <!-- col 3: markdown content -->
    <div v-if="!selectedEntry" class="wiki-placeholder">
      <ReadOutlined class="wp-icon" />
      <div>{{ t('localWiki.selectFile') }}</div>
    </div>

    <div v-else class="wiki-content">
      <div class="wc-bar">
        <div class="wc-title">
          <ReadOutlined class="wc-icon" />
          <span>{{ selectedEntry.repoName }}</span>
          <span class="wc-file" v-if="selectedEntry.fileName !== 'reference.md'">
            / {{ selectedEntry.fileName.replace('.md', '') }}
          </span>
        </div>
        <div class="wc-meta">
          <span v-if="selectedEntry.gitStatus && selectedEntry.gitStatus !== 'committed'" class="wc-git-badge" :class="'git-' + selectedEntry.gitStatus">
            {{ selectedEntry.gitStatus === 'modified' ? t('localWiki.gitModified') : t('localWiki.gitUntracked') }}
          </span>
          <span v-if="selectedEntry.gitStatus === 'committed'" class="wc-git-badge git-committed">{{ t('localWiki.gitCommitted') }}</span>
          <span v-if="selectedEntry.commit" class="wc-commit">{{ selectedEntry.commit }}</span>
          <span v-if="selectedEntry.exploredAt" class="wc-date">{{ selectedEntry.exploredAt }}</span>
        </div>
        <a-radio-group v-model:value="viewMode" size="small" button-style="solid">
          <a-radio-button value="render"><EyeOutlined /> {{ t('localWiki.renderMode') }}</a-radio-button>
          <a-radio-button value="source"><CodeOutlined /> {{ t('localWiki.sourceMode') }}</a-radio-button>
        </a-radio-group>
      </div>
      <div class="wc-body">
        <a-spin v-if="contentLoading" class="wc-loading" />
        <div v-else-if="viewMode === 'render'" class="wc-md" v-html="renderedContent"></div>
        <pre v-else class="wc-source">{{ rawContent }}</pre>
      </div>
    </div>

    <!-- mermaid enlarge modal (shared) -->
    <MermaidModal
      v-model:open="modalOpen"
      :svg="modalSvg"
      v-model:zoom="zoom"
      v-model:panX="panX"
      v-model:panY="panY"
    />
  </div>
</template>

<style scoped>
.wiki-view { display: flex; height: calc(100% + 2 * var(--spacing-lg)); overflow: hidden; margin: calc(-1 * var(--spacing-lg)); }

/* col 2: file list */
.wiki-files {
  width: 200px; min-width: 200px;
  display: flex; flex-direction: column;
  border-right: 1px solid var(--color-border);
  background: var(--color-surface); overflow: hidden;
}
.wf-head {
  display: flex; align-items: center; gap: 8px;
  padding: 0 12px; height: var(--navbar-height);
  border-bottom: 1px solid var(--color-border); flex-shrink: 0;
}
.wf-icon { font-size: 14px; color: var(--color-text-tertiary); }
.wf-title {
  font-size: 13px; font-weight: 600; color: var(--color-text);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.wf-list { flex: 1; overflow-y: auto; padding: var(--spacing-sm); }
.wf-list::-webkit-scrollbar { width: 5px; }
.wf-list::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 3px; }
.wf-item {
  display: flex; align-items: center; gap: 8px;
  padding: 7px 10px; border-radius: var(--radius-md);
  cursor: pointer; transition: all var(--transition-fast); margin-bottom: 1px;
}
.wf-item:hover { background: var(--color-hover); }
.wf-item.active { background: var(--color-primary-bg); }
.wf-item-icon {
  display: flex; align-items: center; justify-content: center;
  width: 24px; height: 24px; border-radius: var(--radius-sm);
  background: var(--color-background); color: var(--color-text-tertiary);
  font-size: 12px; flex-shrink: 0; transition: all var(--transition-fast);
}
.wf-item:hover .wf-item-icon, .wf-item.active .wf-item-icon { background: var(--color-primary); color: #fff; }
.wf-item-body { flex: 1; min-width: 0; }
.wf-item-name {
  font-size: 13px; font-weight: 500; color: var(--color-text);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.wf-item.active .wf-item-name { color: var(--color-primary); }
.wf-item-meta { display: flex; gap: 6px; margin-top: 1px; font-size: 11px; color: var(--color-text-tertiary); }
.wf-topic-tag {
  font-size: 9px; font-weight: 600; padding: 0 4px;
  border-radius: 3px; background: var(--color-primary-bg); color: var(--color-primary);
}
.wf-problem { border-left: 2px solid var(--color-warning); }
.wf-git-tag { font-size: 9px; font-weight: 600; padding: 0 4px; border-radius: 3px; }
.wf-git-tag.git-modified { background: var(--color-warning-bg); color: var(--color-warning); }
.wf-git-tag.git-untracked { background: var(--color-error-bg); color: var(--color-error); }
.wf-status-tag { font-size: 9px; font-weight: 600; padding: 0 4px; border-radius: 3px; }
.wf-status-tag.st-empty { background: var(--color-error-bg); color: var(--color-error); }
.wf-status-tag.st-no-fm { background: var(--color-warning-bg); color: var(--color-warning); }
.wf-delete {
  display: flex; align-items: center; justify-content: center;
  width: 22px; height: 22px; border: none; background: transparent;
  color: var(--color-text-tertiary); cursor: pointer; border-radius: var(--radius-xs);
  opacity: 0; transition: all var(--transition-fast); flex-shrink: 0;
}
.wf-item:hover .wf-delete { opacity: 0.5; }
.wf-delete:hover { opacity: 1 !important; background: var(--color-error-bg); color: var(--color-error); }

/* col 3: content */
.wiki-placeholder {
  flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 12px; color: var(--color-text-tertiary); font-size: 14px;
}
.wp-icon { font-size: 48px; opacity: 0.2; }
.wiki-content {
  flex: 1; min-width: 0; display: flex; flex-direction: column; background: var(--color-background);
}
.wc-bar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 20px; height: var(--navbar-height);
  border-bottom: 1px solid var(--color-border);
  background: var(--color-surface); flex-shrink: 0; gap: 12px;
}
.wc-title { display: flex; align-items: center; gap: 6px; font-size: 15px; font-weight: 600; color: var(--color-text); }
.wc-icon { color: var(--color-text-tertiary); }
.wc-file { font-size: 13px; font-weight: 400; color: var(--color-text-secondary); }
.wc-meta { display: flex; align-items: center; gap: 8px; flex-shrink: 0; font-size: 12px; color: var(--color-text-tertiary); }
.wc-commit { font-family: 'Cascadia Code', monospace; }
.wc-date { font-family: 'Cascadia Code', monospace; }
.wc-git-badge { font-size: 10px; font-weight: 600; padding: 1px 6px; border-radius: 3px; }
.wc-git-badge.git-committed { background: var(--color-success-bg); color: var(--color-success); }
.wc-git-badge.git-modified { background: var(--color-warning-bg); color: var(--color-warning); }
.wc-git-badge.git-untracked { background: var(--color-error-bg); color: var(--color-error); }
.wc-body { flex: 1; overflow-y: auto; min-height: 0; }
.wc-body::-webkit-scrollbar { width: 8px; }
.wc-body::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 4px; }
.wc-loading { display: flex; justify-content: center; padding: 40px; }
.wc-md {
  padding: 24px 32px; font-size: 14px; line-height: 1.7;
  color: var(--color-text);
}
.wc-md :deep(h1) { font-size: 24px; font-weight: 700; margin: 24px 0 12px; }
.wc-md :deep(h2) { font-size: 20px; font-weight: 600; margin: 20px 0 10px; border-bottom: 1px solid var(--color-border); padding-bottom: 6px; }
.wc-md :deep(h3) { font-size: 16px; font-weight: 600; margin: 16px 0 8px; }
.wc-md :deep(p) { margin: 8px 0; }
.wc-md :deep(code) { font-family: 'Cascadia Code', monospace; font-size: 12px; padding: 2px 6px; background: var(--color-surface-raised); border-radius: 3px; }
.wc-md :deep(pre) { padding: 12px 16px; background: var(--color-surface); border: 1px solid var(--color-border); border-radius: var(--radius-md); overflow-x: auto; margin: 12px 0; }
.wc-md :deep(pre code) { background: none; padding: 0; }
.wc-md :deep(a) { color: var(--color-primary); }
.wc-md :deep(ul), .wc-md :deep(ol) { padding-left: 24px; margin: 8px 0; }
.wc-md :deep(li) { margin: 4px 0; }
.wc-md :deep(table) { border-collapse: collapse; width: 100%; margin: 12px 0; }
.wc-md :deep(th), .wc-md :deep(td) { border: 1px solid var(--color-border); padding: 6px 12px; text-align: left; }
.wc-md :deep(th) { background: var(--color-surface); font-weight: 600; }
.wc-md :deep(blockquote) { border-left: 3px solid var(--color-primary); padding-left: 16px; margin: 12px 0; color: var(--color-text-secondary); }

.wc-source {
  margin: 0; padding: 16px 24px;
  font-family: 'Cascadia Code', 'Fira Code', monospace;
  font-size: 13px; line-height: 1.6;
  color: var(--color-text); white-space: pre-wrap; word-break: break-word;
}

/* rendered mermaid blocks (created by useMermaid after v-html patch) */
.mermaid-rendered {
  display: flex; justify-content: center;
  margin: 16px 0; padding: 16px;
  background: var(--color-surface); border: 1px solid var(--color-border);
  border-radius: var(--radius-md); overflow-x: auto;
}
.mermaid-rendered svg { max-width: 100%; height: auto; }
</style>
