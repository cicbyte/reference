<script setup>
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import { message } from 'ant-design-vue'
import {
  ReloadOutlined,
  FolderOpenOutlined,
  DeleteOutlined,
  CloudServerOutlined,
  SearchOutlined,
  FileOutlined,
  FileTextOutlined,
  CodeOutlined,
  EyeOutlined,
  ArrowLeftOutlined,
  CaretRightFilled,
} from '@ant-design/icons-vue'
import hljs from 'highlight.js/lib/core'
import 'highlight.js/styles/github-dark.css'
import { marked } from 'marked'
import FileTreeNode from '../components/repo/FileTreeNode.vue'

const app = window.go?.main?.ReferenceApp

// ---- repo list ----
const repos = ref([])
const loading = ref(true)
const selectedCachePath = ref('')

// ---- file tree ----
const tree = ref({})
const expandedDirs = ref(new Set())
const loadingDirs = ref(new Set())
const selectedFile = ref('')

// ---- search ----
const searchQuery = ref('')
const searchResults = ref([])
const searching = ref(false)
let searchTimer = null

// ---- file content ----
const fileContent = ref('')
const fileInfo = ref({ binary: false, notFound: false, lines: 0 })
const fileLoading = ref(false)
const mdViewMode = ref('render')
const codeRef = ref(null)

// ---- helpers ----
function fmtSize(bytes) {
  if (!bytes) return '…'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + ' MB'
  return (bytes / 1024 / 1024 / 1024).toFixed(2) + ' GB'
}

const selectedRepo = computed(() =>
  repos.value.find((r) => r.cachePath === selectedCachePath.value),
)

const isMarkdown = computed(() => /\.md$/i.test(selectedFile.value))
const filename = computed(() => selectedFile.value.split('/').pop() || '')

function hljsLang(name) {
  const ext = (name.split('.').pop() || '').toLowerCase()
  const map = {
    go: 'go', js: 'javascript', ts: 'typescript', html: 'xml', vue: 'xml',
    css: 'css', json: 'json', yml: 'yaml', yaml: 'yaml', md: 'markdown',
    py: 'python', sh: 'bash', sql: 'sql', rs: 'rust',
  }
  return map[ext] || ''
}

// ---- load repo list (instant — no sizes) ----
async function loadRepos() {
  loading.value = true
  try {
    if (app) {
      repos.value = await app.ListCachedRepos()
      // async-load sizes for each repo
      repos.value.forEach((r) => { fetchSize(r) })
    }
  } catch (e) {
    message.error('加载失败: ' + e)
  } finally {
    loading.value = false
  }
}

async function fetchSize(repo) {
  try {
    if (app?.GetCacheSize) {
      const size = await app.GetCacheSize(repo.cachePath)
      const idx = repos.value.findIndex((r) => r.cachePath === repo.cachePath)
      if (idx >= 0) {
        repos.value[idx] = { ...repos.value[idx], size }
      }
    }
  } catch { /* leave as … */ }
}

onMounted(loadRepos)

// ---- grouped tree: platform → namespace → repos ----
const expandedPlatforms = ref(new Set())
const expandedNamespaces = ref(new Set())

const groupedRepos = computed(() => {
  // { platform: { namespace: [repo, ...] } }
  const tree = {}
  for (const r of repos.value) {
    const platform = r.type === 'local' ? '本地仓库' : (r.host || '未知平台')
    const ns = r.type === 'local' ? '' : (r.namespace || '未知')
    if (!tree[platform]) tree[platform] = {}
    if (!tree[platform][ns]) tree[platform][ns] = []
    tree[platform][ns].push(r)
  }
  // sort platforms (本地仓库 last), namespaces alpha, repos alpha
  const platforms = Object.keys(tree).sort((a, b) => {
    if (a === '本地仓库') return 1
    if (b === '本地仓库') return -1
    return a.localeCompare(b)
  })
  return platforms.map((p) => ({
    platform: p,
    namespaces: Object.keys(tree[p]).sort().map((ns) => ({
      namespace: ns,
      repos: tree[p][ns].sort((a, b) => a.name.localeCompare(b.name)),
    })),
  }))
})

function togglePlatform(p) {
  const next = new Set(expandedPlatforms.value)
  if (next.has(p)) next.delete(p)
  else next.add(p)
  expandedPlatforms.value = next
}

function toggleNamespace(key) {
  const next = new Set(expandedNamespaces.value)
  if (next.has(key)) next.delete(key)
  else next.add(key)
  expandedNamespaces.value = next
}

function isPlatformExpanded(p) { return expandedPlatforms.value.has(p) }
function isNamespaceExpanded(key) { return expandedNamespaces.value.has(key) }

// auto-expand all on first load
watch(groupedRepos, (groups) => {
  if (expandedPlatforms.value.size === 0 && groups.length) {
    const ps = new Set()
    const ns = new Set()
    for (const g of groups) {
      ps.add(g.platform)
      for (const n of g.namespaces) ns.add(g.platform + '/' + n.namespace)
    }
    expandedPlatforms.value = ps
    expandedNamespaces.value = ns
  }
}, { once: true })

// ---- select repo → load its root tree ----
watch(selectedCachePath, async (newPath) => {
  tree.value = {}
  expandedDirs.value = new Set()
  selectedFile.value = ''
  fileContent.value = ''
  searchQuery.value = ''
  searchResults.value = []
  if (!newPath) return
  try {
    const nodes = await app.BrowseCacheByPathList(newPath, '')
    tree.value = { '__root__': nodes }
  } catch (e) {
    message.error('加载文件树失败: ' + e)
  }
})

function selectRepo(repo) {
  selectedCachePath.value = repo.cachePath
}

// ---- tree operations ----
async function toggleDir(node) {
  const next = new Set(expandedDirs.value)
  if (next.has(node.path)) {
    next.delete(node.path)
  } else {
    next.add(node.path)
    if (!(node.path in tree.value)) {
      const ld = new Set(loadingDirs.value); ld.add(node.path)
      loadingDirs.value = ld
      try {
        const children = await app.BrowseCacheByPathList(selectedCachePath.value, node.path)
        tree.value = { ...tree.value, [node.path]: children }
      } catch (e) { message.error('读取目录失败: ' + e) }
      finally {
        const ld2 = new Set(loadingDirs.value); ld2.delete(node.path)
        loadingDirs.value = ld2
      }
    }
  }
  expandedDirs.value = next
}

async function openFile(node) {
  if (node.isDir) return
  selectedFile.value = node.path
  fileLoading.value = true
  fileInfo.value = { binary: false, notFound: false, lines: 0 }
  fileContent.value = ''
  try {
    const res = await app.BrowseCacheByPathRead(selectedCachePath.value, node.path)
    fileInfo.value = { binary: res.binary, notFound: res.notFound, lines: res.lines }
    fileContent.value = res.content
  } catch (e) {
    message.error('读取失败: ' + e)
  } finally {
    fileLoading.value = false
    if (!fileInfo.value.binary && !fileInfo.value.notFound && !isMarkdown.value) {
      await nextTick(); await nextTick()
      highlightCode()
    }
  }
}

function highlightCode() {
  if (!codeRef.value) { requestAnimationFrame(highlightCode); return }
  const lang = hljsLang(filename.value)
  try {
    if (lang && hljs.getLanguage(lang)) {
      codeRef.value.innerHTML = hljs.highlight(fileContent.value, { language: lang }).value
    } else {
      codeRef.value.textContent = fileContent.value
    }
  } catch { codeRef.value.textContent = fileContent.value }
}

// ---- search ----
function onSearchInput(val) {
  searchQuery.value = val
  if (searchTimer) clearTimeout(searchTimer)
  if (!val.trim()) { searchResults.value = []; searching.value = false; return }
  searching.value = true
  searchTimer = setTimeout(async () => {
    try {
      searchResults.value = await app.BrowseCacheByPathSearch(selectedCachePath.value, val)
    } catch { searchResults.value = [] }
    finally { searching.value = false }
  }, 300)
}

function highlightSegments(name) {
  const q = searchQuery.value.toLowerCase()
  const lower = name.toLowerCase()
  const segs = []
  let i = 0
  while (i < name.length) {
    const idx = lower.indexOf(q, i)
    if (idx === -1) { segs.push({ text: name.slice(i), match: false }); break }
    if (idx > i) segs.push({ text: name.slice(i, idx), match: false })
    segs.push({ text: name.slice(idx, idx + q.length), match: true })
    i = idx + q.length
  }
  return segs
}

function dirOf(path) {
  const idx = path.lastIndexOf('/')
  return idx > 0 ? path.slice(0, idx) : ''
}

// ---- purge ----
async function purge(repo) {
  try {
    await app.PurgeCachedRepo(repo.cachePath)
    message.success(`已清理 ${repo.name}`)
    if (selectedCachePath.value === repo.cachePath) {
      selectedCachePath.value = ''
    }
    await loadRepos()
  } catch (e) {
    message.error('清理失败: ' + e)
  }
}

const renderedMarkdown = computed(() => {
  if (!isMarkdown.value || mdViewMode.value !== 'render') return ''
  try { return marked(fileContent.value) } catch { return fileContent.value }
})
</script>

<template>
  <div class="cache-view">
    <!-- left: repo rail -->
    <aside class="cache-rail">
      <div class="rail-head">
        <span>仓库缓存</span>
        <button class="rail-refresh" title="刷新" @click="loadRepos">
          <ReloadOutlined />
        </button>
      </div>
      <div class="rail-list">
        <a-spin v-if="loading" class="rail-spin" />

        <!-- grouped: platform → namespace → repos -->
        <template v-for="pg in groupedRepos" :key="pg.platform">
          <div class="rail-group-head" @click="togglePlatform(pg.platform)">
            <CaretRightFilled class="rail-caret" :class="{ open: isPlatformExpanded(pg.platform) }" />
            <span class="rail-group-label">{{ pg.platform }}</span>
            <span class="rail-group-count">{{ pg.namespaces.reduce((s, n) => s + n.repos.length, 0) }}</span>
          </div>

          <template v-if="isPlatformExpanded(pg.platform)">
            <template v-for="ns in pg.namespaces" :key="pg.platform + '/' + ns.namespace">
              <!-- namespace sub-group (skip header for 本地仓库 or single-child) -->
              <div
                v-if="pg.platform !== '本地仓库' && ns.namespace"
                class="rail-ns-head"
                @click="toggleNamespace(pg.platform + '/' + ns.namespace)"
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
                  @click="selectRepo(r)"
                >
                  <div class="rail-item-icon" :class="{ 'icon-local': r.type === 'local' }">
                    <CloudServerOutlined v-if="r.type === 'remote'" />
                    <FolderOpenOutlined v-else />
                  </div>
                  <div class="rail-item-body">
                    <div class="rail-item-name">
                      {{ r.name }}
                      <span v-if="r.type === 'local'" class="rail-type-tag">本地</span>
                    </div>
                    <div class="rail-item-meta">
                      <span class="rail-size">{{ fmtSize(r.size) }}</span>
                      <span class="rail-ref">{{ r.refCount }} 引用</span>
                    </div>
                  </div>
                  <a-popconfirm
                    v-if="r.type === 'remote'"
                    :title="`清理 ${r.name}？`"
                    ok-text="清理" ok-type="danger" cancel-text="取消"
                    @confirm="purge(r)"
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

        <div v-if="!loading && !repos.length" class="rail-empty">
          <CloudServerOutlined class="rail-empty-icon" />
          <span>暂无缓存</span>
        </div>
      </div>
    </aside>

    <!-- right: file tree + code viewer -->
    <div v-if="!selectedCachePath" class="cache-placeholder">
      <CloudServerOutlined class="cp-icon" />
      <div>从左侧选择一个缓存仓库查看代码</div>
    </div>

    <div v-else class="cache-browser">
      <!-- file tree -->
      <aside class="cb-tree">
        <div class="bt-search">
          <a-input
            :value="searchQuery"
            placeholder="搜索文件名..."
            size="small"
            allow-clear
            @change="(e) => onSearchInput(e.target.value)"
          >
            <template #prefix><SearchOutlined style="color: var(--color-text-tertiary)" /></template>
          </a-input>
        </div>
        <div class="bt-body">
          <div v-if="searchQuery.trim()">
            <div v-if="searching" class="bt-loading"><a-spin size="small" /></div>
            <div v-else-if="!searchResults.length" class="bt-empty">无匹配</div>
            <div v-for="r in searchResults" :key="r.path"
              class="sr-item" :class="{ active: selectedFile === r.path }"
              :title="r.path" @click="openFile(r)"
            >
              <FileTextOutlined class="sr-icon" />
              <div class="sr-body">
                <div class="sr-name">
                  <template v-for="(seg, i) in highlightSegments(r.name)" :key="i">
                    <mark v-if="seg.match" class="sr-mark">{{ seg.text }}</mark>
                    <span v-else>{{ seg.text }}</span>
                  </template>
                </div>
                <div class="sr-dir" v-if="dirOf(r.path)">{{ dirOf(r.path) }}</div>
              </div>
            </div>
          </div>
          <template v-else>
            <FileTreeNode
              :nodes="tree['__root__'] || []"
              :depth="0"
              :tree="tree"
              :expanded="expandedDirs"
              :selected="selectedFile"
              :loading-dirs="loadingDirs"
              @toggle="toggleDir"
              @open="openFile"
            />
          </template>
        </div>
      </aside>

      <!-- code content -->
      <div class="cb-content">
        <div class="bc-bar" v-if="selectedFile">
          <span class="bc-name"><FileTextOutlined /> {{ filename }}</span>
          <span class="bc-path">{{ selectedFile }}</span>
          <div class="bc-spacer"></div>
          <span class="bc-lines" v-if="fileInfo.lines">{{ fileInfo.lines }} 行</span>
          <a-radio-group v-if="isMarkdown" v-model:value="mdViewMode" size="small" button-style="solid">
            <a-radio-button value="render"><EyeOutlined /> 渲染</a-radio-button>
            <a-radio-button value="source"><CodeOutlined /> 源码</a-radio-button>
          </a-radio-group>
        </div>
        <div class="bc-body">
          <div v-if="fileLoading" class="bc-loading"><a-spin tip="读取中..." /></div>
          <div v-else-if="!selectedFile" class="bc-placeholder">
            <FileOutlined class="bc-ph-icon" />
            <div>从左侧选择文件</div>
          </div>
          <div v-else-if="fileInfo.notFound" class="bc-state">文件不存在</div>
          <div v-else-if="fileInfo.binary" class="bc-state">二进制文件</div>
          <div v-else-if="isMarkdown && mdViewMode === 'render'" class="bc-md" v-html="renderedMarkdown"></div>
          <div v-else class="bc-code">
            <pre class="hljs-code-block"><code ref="codeRef" class="hljs"></code></pre>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.cache-view {
  display: flex;
  height: 100%;
  width: 100%;
  overflow: hidden;
}

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

/* ---- placeholder ---- */
.cache-placeholder {
  flex: 1;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 12px; color: var(--color-text-tertiary); font-size: 14px;
}
.cp-icon { font-size: 48px; opacity: 0.25; }

/* ---- browser (tree + content) ---- */
.cache-browser {
  flex: 1;
  min-width: 0;
  display: flex;
  overflow: hidden;
}

/* tree */
.cb-tree {
  width: 260px; min-width: 260px;
  display: flex; flex-direction: column;
  border-right: 1px solid var(--color-border);
  background: var(--color-surface);
  overflow: hidden;
}
.bt-search { padding: 8px 10px; flex-shrink: 0; }
.bt-body { flex: 1; overflow-y: auto; padding: 4px; }
.bt-body::-webkit-scrollbar { width: 6px; }
.bt-body::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 3px; }
.bt-loading { display: flex; justify-content: center; padding: 24px; }
.bt-empty { padding: 24px; text-align: center; font-size: 12px; color: var(--color-text-tertiary); }

.sr-item {
  display: flex; align-items: center; gap: 8px;
  padding: 5px 8px; border-radius: var(--radius-xs);
  cursor: pointer; transition: background var(--transition-fast);
}
.sr-item:hover { background: var(--color-hover); }
.sr-item.active { background: var(--color-primary-bg); }
.sr-icon { font-size: 13px; color: var(--color-text-tertiary); flex-shrink: 0; }
.sr-body { min-width: 0; flex: 1; }
.sr-name { font-size: 13px; color: var(--color-text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.sr-mark { background: var(--color-warning-bg); color: var(--color-warning); padding: 0 2px; border-radius: 2px; }
.sr-dir { font-size: 11px; color: var(--color-text-tertiary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

/* content */
.cb-content {
  flex: 1; min-width: 0;
  display: flex; flex-direction: column;
  background: var(--color-background);
}
.bc-bar {
  display: flex; align-items: center; gap: 10px;
  padding: 0 16px; height: var(--navbar-height);
  border-bottom: 1px solid var(--color-border);
  background: var(--color-surface); flex-shrink: 0;
}
.bc-name { display: flex; align-items: center; gap: 6px; font-size: 14px; font-weight: 600; color: var(--color-text); }
.bc-path {
  font-size: 12px; color: var(--color-text-tertiary);
  font-family: 'Cascadia Code', monospace;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.bc-spacer { flex: 1; }
.bc-lines { font-size: 11px; color: var(--color-text-tertiary); }

.bc-body { flex: 1; overflow-y: auto; min-height: 0; }
.bc-body::-webkit-scrollbar { width: 8px; }
.bc-body::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 4px; }

.bc-loading { display: flex; justify-content: center; align-items: center; height: 100%; }
.bc-placeholder {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 12px; height: 100%; color: var(--color-text-tertiary); font-size: 14px;
}
.bc-ph-icon { font-size: 48px; opacity: 0.25; }
.bc-state { display: flex; align-items: center; justify-content: center; height: 100%; color: var(--color-text-tertiary); }

.bc-code { padding: 0; }
.hljs-code-block {
  margin: 0; padding: 16px 0;
  font-family: 'Cascadia Code', 'Fira Code', monospace;
  font-size: 13px; line-height: 1.6; overflow-x: auto;
}
.hljs-code-block code { display: block; padding: 0 20px; background: transparent !important; white-space: pre; }

/* markdown */
.bc-md {
  padding: 24px 32px; font-size: 14px; line-height: 1.7;
  color: var(--color-text); max-width: 860px;
}
.bc-md :deep(h1) { font-size: 24px; font-weight: 700; margin: 24px 0 12px; }
.bc-md :deep(h2) { font-size: 20px; font-weight: 600; margin: 20px 0 10px; border-bottom: 1px solid var(--color-border); padding-bottom: 6px; }
.bc-md :deep(h3) { font-size: 16px; font-weight: 600; margin: 16px 0 8px; }
.bc-md :deep(p) { margin: 8px 0; }
.bc-md :deep(code) { font-family: 'Cascadia Code', monospace; font-size: 12px; padding: 2px 6px; background: var(--color-surface-raised); border-radius: 3px; }
.bc-md :deep(pre) { padding: 12px 16px; background: var(--color-surface); border: 1px solid var(--color-border); border-radius: var(--radius-md); overflow-x: auto; margin: 12px 0; }
.bc-md :deep(pre code) { background: none; padding: 0; }
.bc-md :deep(a) { color: var(--color-primary); }
.bc-md :deep(ul), .bc-md :deep(ol) { padding-left: 24px; margin: 8px 0; }
.bc-md :deep(li) { margin: 4px 0; }
.bc-md :deep(table) { border-collapse: collapse; width: 100%; margin: 12px 0; }
.bc-md :deep(th), .bc-md :deep(td) { border: 1px solid var(--color-border); padding: 6px 12px; text-align: left; }
.bc-md :deep(th) { background: var(--color-surface); font-weight: 600; }
.bc-md :deep(blockquote) { border-left: 3px solid var(--color-primary); padding-left: 16px; margin: 12px 0; color: var(--color-text-secondary); }
</style>
