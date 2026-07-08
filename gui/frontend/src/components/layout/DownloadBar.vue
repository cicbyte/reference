<script setup>
import { ref, computed } from 'vue'
import { CloudDownloadOutlined, DownloadOutlined, FileTextOutlined, UnorderedListOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'

const app = window.go?.main?.GitMateApp

const modalOpen = ref(false)
const query = ref('')
const releases = ref([])
const loading = ref(false)
const error = ref('')
const currentRepo = ref('')

// Left tab: selected release index
const activeIdx = ref(0)
// Right sub-tab: 'log' (body) or 'files' (assets)
const subTab = ref('files')
// Multi-selected asset URLs within the active release
const selectedUrls = ref([])

const activeRelease = computed(() => releases.value[activeIdx.value])

function fmtSize(bytes) {
  if (!bytes) return '—'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + ' MB'
  return (bytes / 1024 / 1024 / 1024).toFixed(2) + ' GB'
}

function openModal() {
  modalOpen.value = true
}

function selectRelease(idx) {
  activeIdx.value = idx
  selectedUrls.value = []
  // default to files view if assets exist, else log
  subTab.value = releases.value[idx]?.assets?.length ? 'files' : 'log'
}

async function search() {
  const q = query.value.trim()
  if (!q) return
  const parts = q.split('/')
  if (parts.length !== 2 || !parts[0] || !parts[1]) {
    error.value = '请输入 owner/repo 格式'
    return
  }
  if (!app) return
  loading.value = true
  error.value = ''
  releases.value = []
  currentRepo.value = q
  try {
    releases.value = await app.DownloadReleases(parts[0], parts[1])
    if (releases.value.length) selectRelease(0)
  } catch (e) {
    error.value = String(e?.message || e)
  } finally {
    loading.value = false
  }
}

// batch download selected assets
function downloadSelected() {
  if (!selectedUrls.value.length) {
    message.warning('请先选择文件')
    return
  }
  selectedUrls.value.forEach((url, i) => {
    // slight stagger to avoid the browser blocking simultaneous opens
    setTimeout(() => window.open(url, '_blank'), i * 250)
  })
  message.success(`开始下载 ${selectedUrls.value.length} 个文件`)
}

// toggle one asset selection
function toggleAsset(url) {
  const i = selectedUrls.value.indexOf(url)
  if (i >= 0) selectedUrls.value.splice(i, 1)
  else selectedUrls.value.push(url)
}

// select/deselect all assets of active release
function toggleAll(e) {
  const assets = activeRelease.value?.assets || []
  if (e.target.checked) {
    selectedUrls.value = assets.map((a) => a.browserDownloadURL)
  } else {
    selectedUrls.value = []
  }
}

const allChecked = computed(() => {
  const assets = activeRelease.value?.assets || []
  return assets.length > 0 && assets.every((a) => selectedUrls.value.includes(a.browserDownloadURL))
})
</script>

<template>
  <!-- Trigger: an icon button meant for the Navbar right cluster -->
  <button class="dl-trigger" title="下载 GitHub Release" @click="openModal">
    <CloudDownloadOutlined />
  </button>

  <a-modal
    v-model:open="modalOpen"
    :title="currentRepo ? `下载 — ${currentRepo}` : '下载 GitHub Release'"
    :footer="null"
    width="900px"
    destroy-on-close
  >
    <!-- Search bar -->
    <div class="search-row">
      <a-input-search
        v-model:value="query"
        placeholder="owner/repo，如 cicbyte/ktconsole"
        enter-button="查询"
        :loading="loading"
        @search="search"
      />
    </div>

    <!-- States -->
    <div v-if="loading" class="modal-state"><a-spin /></div>
    <div v-else-if="error" class="modal-state error">{{ error }}</div>
    <div v-else-if="!releases.length" class="modal-state muted">
      输入 owner/repo 查询 Release 列表
    </div>

    <!-- Split layout: left release tabs + right detail -->
    <div v-else class="split">
      <!-- Left: vertical release tabs (tag names) -->
      <aside class="release-tabs">
        <div
          v-for="(r, i) in releases"
          :key="r.tagName"
          class="tab-item"
          :class="{ active: i === activeIdx }"
          @click="selectRelease(i)"
          :title="r.tagName"
        >
          <span class="tab-name">{{ r.tagName }}</span>
          <span class="tab-count" v-if="r.assets?.length">{{ r.assets.length }}</span>
        </div>
      </aside>

      <!-- Right: detail of active release -->
      <section class="release-detail" v-if="activeRelease">
        <div class="detail-head">
          <a-tag color="green">{{ activeRelease.tagName }}</a-tag>
          <span class="date">{{ activeRelease.publishedAt }}</span>
        </div>

        <!-- sub-tabs: log / files -->
        <div class="sub-tabs">
          <button
            class="sub-tab"
            :class="{ active: subTab === 'files' }"
            @click="subTab = 'files'"
          >
            <UnorderedListOutlined /> 文件列表
            <span v-if="activeRelease.assets?.length" class="sub-count">({{ activeRelease.assets.length }})</span>
          </button>
          <button
            class="sub-tab"
            :class="{ active: subTab === 'log' }"
            @click="subTab = 'log'"
          >
            <FileTextOutlined /> 更新日志
          </button>

          <div class="sub-actions" v-if="subTab === 'files' && activeRelease.assets?.length">
            <a-checkbox :checked="allChecked" @change="toggleAll">全选</a-checkbox>
            <a-button
              type="primary"
              size="small"
              :disabled="!selectedUrls.length"
              @click="downloadSelected"
            >
              <template #icon><DownloadOutlined /></template>
              下载选中<template v-if="selectedUrls.length"> ({{ selectedUrls.length }})</template>
            </a-button>
          </div>
        </div>

        <!-- files view -->
        <div v-if="subTab === 'files'" class="files-view">
          <div v-if="activeRelease.assets?.length" class="file-list">
            <label
              v-for="a in activeRelease.assets"
              :key="a.browserDownloadURL"
              class="file-item"
              :class="{ checked: selectedUrls.includes(a.browserDownloadURL) }"
            >
              <input
                type="checkbox"
                :checked="selectedUrls.includes(a.browserDownloadURL)"
                @change="toggleAsset(a.browserDownloadURL)"
              />
              <span class="fname">{{ a.name }}</span>
              <span class="fsize">{{ fmtSize(a.size) }}</span>
              <a :href="a.browserDownloadURL" target="_blank" class="fdl" @click.stop title="单独下载">
                <DownloadOutlined />
              </a>
            </label>
          </div>
          <div v-else class="empty">此版本无附件</div>
        </div>

        <!-- log view -->
        <div v-else class="log-view">
          <pre v-if="activeRelease.body" class="log-body">{{ activeRelease.body }}</pre>
          <div v-else class="empty">无更新日志</div>
        </div>
      </section>
    </div>
  </a-modal>
</template>

<style scoped>
.dl-trigger {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 28px;
  border: none;
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  border-radius: var(--radius-xs);
  transition: all var(--transition-fast);
  font-size: 16px;
}
.dl-trigger:hover {
  background: var(--color-hover);
  color: var(--color-primary);
}

.search-row { margin-bottom: var(--spacing-md); }

.modal-state {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 200px;
  font-size: 13px;
}
.modal-state.error { color: var(--color-error); word-break: break-all; }
.modal-state.muted { color: var(--color-text-tertiary); }

/* Split layout */
.split {
  display: flex;
  gap: var(--spacing-md);
  height: 440px;
}

.release-tabs {
  width: 200px;
  flex-shrink: 0;
  overflow-y: auto;
  border-right: 1px solid var(--color-border);
  padding-right: var(--spacing-sm);
}
.tab-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 13px;
  color: var(--color-text-secondary);
  transition: all var(--transition-fast);
  margin-bottom: 2px;
}
.tab-item:hover { background: var(--color-hover); color: var(--color-text); }
.tab-item.active { background: var(--color-primary-bg); color: var(--color-primary); font-weight: 600; }
.tab-name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.tab-count {
  font-size: 11px;
  background: var(--color-surface-raised);
  color: var(--color-text-tertiary);
  border-radius: 10px;
  padding: 0 6px;
  flex-shrink: 0;
}

.release-detail {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.detail-head {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-sm);
}
.date { font-size: 12px; color: var(--color-text-tertiary); }

/* sub-tabs */
.sub-tabs {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  border-bottom: 1px solid var(--color-border);
  padding-bottom: 8px;
  margin-bottom: var(--spacing-sm);
}
.sub-tab {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 10px;
  border: none;
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  font-size: 13px;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}
.sub-tab:hover { color: var(--color-primary); }
.sub-tab.active { color: var(--color-primary); font-weight: 600; }
.sub-count { font-size: 11px; color: var(--color-text-tertiary); }
.sub-actions {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

/* files view */
.files-view { flex: 1; overflow-y: auto; }
.file-list { display: flex; flex-direction: column; gap: 2px; }
.file-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: 8px 10px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 13px;
  transition: background var(--transition-fast);
}
.file-item:hover { background: var(--color-hover); }
.file-item.checked { background: var(--color-primary-bg); }
.file-item input { cursor: pointer; }
.fname { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--color-text); }
.fsize { color: var(--color-text-tertiary); font-size: 12px; flex-shrink: 0; }
.fdl { color: var(--color-text-tertiary); padding: 2px 4px; flex-shrink: 0; }
.fdl:hover { color: var(--color-primary); }

/* log view */
.log-view { flex: 1; overflow-y: auto; }
.log-body {
  margin: 0;
  font-family: 'Cascadia Code', 'Fira Code', monospace;
  font-size: 12px;
  line-height: 1.7;
  color: var(--color-text-secondary);
  white-space: pre-wrap;
  word-break: break-word;
  padding: var(--spacing-sm);
  background: var(--color-background);
  border-radius: var(--radius-sm);
}

.empty { color: var(--color-text-tertiary); font-size: 13px; padding: 40px 0; text-align: center; }
</style>
