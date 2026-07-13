<script setup>
import { ref, onMounted, computed, watch, nextTick } from 'vue'
import { message } from 'ant-design-vue'
import {
  PlusOutlined,
  SyncOutlined,
  DeleteOutlined,
  BarChartOutlined,
  MedicineBoxOutlined,
  SearchOutlined,
  FolderOpenOutlined,
  CloudDownloadOutlined,
  CloudServerOutlined,
  FileOutlined,
  FileTextOutlined,
  CodeOutlined,
  EyeOutlined,
  ArrowLeftOutlined,
} from '@ant-design/icons-vue'
import hljs, { hljsLangForName } from '../utils/hljs-setup'
// Importing the plugin after hljs-setup (which sets window.hljs) augments our
// hljs instance with lineNumbersBlockSync / lineNumbersValue.
import 'highlightjs-line-numbers.js/src/highlightjs-line-numbers.js'
import { marked } from 'marked'
import { useProjectStore } from '../stores/project'
import FileTreeNode from '../components/repo/FileTreeNode.vue'
import AddRepoModal from '../components/repo/AddRepoModal.vue'
import SccModal from '../components/repo/SccModal.vue'
import DiagnoseModal from '../components/repo/DiagnoseModal.vue'

const project = useProjectStore()
const app = window.go?.main?.ReferenceApp

// ---- repo list ----
const repos = ref([])
const loading = ref(true)
const selectedRepo = ref('')
const addOpen = ref(false)
const sccOpen = ref(false)
const sccRepo = ref('')
const openSccName = ref('')
const diagnoseOpen = ref(false)

async function loadRepos() {
  if (!project.hasProject) { repos.value = []; loading.value = false; return }
  loading.value = true
  try {
    if (app) repos.value = await app.ListRepos()
  } catch (e) {
    console.error('Failed to load repos:', e)
  } finally {
    loading.value = false
  }
}

onMounted(loadRepos)
watch(() => project.projectEpoch, loadRepos)

const selectedRepoData = computed(() => repos.value.find((r) => r.name === selectedRepo.value))

function selectRepo(r) {
  selectedRepo.value = r.name
  selectedFile.value = ''
  fileContent.value = ''
  tree.value = {}
  expandedDirs.value = new Set()
  loadRoot(r.name)
}

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

watch(mdViewMode, (mode) => {
  if (mode === 'source' && selectedFile.value) highlightCode()
})

const codeRef = ref(null)

const isMarkdown = computed(() => /\.md$/i.test(selectedFile.value))
const filename = computed(() => selectedFile.value.split('/').pop() || '')
const hljsLang = hljsLangForName

async function loadRoot(repoName) {
  if (!app) return
  try {
    const nodes = await app.BrowseRepoList(repoName, '')
    tree.value = { '__root__': nodes }
  } catch (e) {
    message.error('加载文件树失败: ' + e)
  }
}

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
        const children = await app.BrowseRepoList(selectedRepo.value, node.path)
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
    const res = await app.BrowseRepoRead(selectedRepo.value, node.path)
    fileInfo.value = { binary: res.binary, notFound: res.notFound, lines: res.lines }
    fileContent.value = res.content
  } catch (e) {
    message.error('读取失败: ' + e)
  } finally {
    fileLoading.value = false
    if (!fileInfo.value.binary && !fileInfo.value.notFound) {
      await nextTick(); await nextTick()
      if (!isMarkdown.value || mdViewMode.value === 'source') {
        highlightCode()
      }
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
    // Wrap each line in a <tr> with its own number cell — guaranteed vertical
    // alignment because number and code share the same table row.
    if (hljs.lineNumbersBlockSync) hljs.lineNumbersBlockSync(codeRef.value)
  } catch { codeRef.value.textContent = fileContent.value }
}

function onSearchInput(val) {
  searchQuery.value = val
  if (searchTimer) clearTimeout(searchTimer)
  if (!val.trim()) { searchResults.value = []; searching.value = false; return }
  searching.value = true
  searchTimer = setTimeout(async () => {
    try {
      searchResults.value = await app.BrowseRepoSearch(selectedRepo.value, val)
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

// ---- actions ----
async function updateRepo(name) {
  try {
    await app.UpdateRepo(name)
    repos.value = await app.ListRepos()
    message.success('更新成功')
  } catch (e) { message.error('更新失败: ' + e) }
}

async function removeRepo(name) {
  try {
    await app.RemoveRepo(name, false)
    repos.value = await app.ListRepos()
    if (selectedRepo.value === name) { selectedRepo.value = ''; selectedFile.value = '' }
    message.success('已移除')
  } catch (e) { message.error('移除失败: ' + e) }
}

async function recloneRepo(name) {
  message.loading({ content: `正在重新克隆 ${name}...`, key: 'reclone', duration: 0 })
  try {
    const timeout = new Promise((_, reject) =>
      setTimeout(() => reject(new Error('克隆超时')), 180000),
    )
    await Promise.race([app.RecloneRepo(project.currentDir, name), timeout])
    message.success({ content: `${name} 重新克隆成功`, key: 'reclone' })
    repos.value = await app.ListRepos()
  } catch (e) { message.error({ content: '重新克隆失败: ' + e, key: 'reclone', duration: 5 }) }
}

const renderedMarkdown = computed(() => {
  if (!isMarkdown.value || mdViewMode.value !== 'render') return ''
  try { return marked(fileContent.value) } catch { return fileContent.value }
})
</script>

<template>
  <div class="repo-view">
    <!-- left: repo rail -->
    <aside class="repo-rail">
      <div class="rail-head">
        <span>仓库列表</span>
        <div class="rail-actions">
          <a-button size="small" type="text" @click="diagnoseOpen = true" title="诊断修复">
            <MedicineBoxOutlined />
          </a-button>
          <a-button size="small" type="primary" @click="addOpen = true" title="添加仓库">
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
          @click="selectRepo(r)"
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
              <a-button size="small" type="text" @click.stop="openSccName = r.name; sccOpen = true" title="统计">
                <BarChartOutlined />
              </a-button>
            </div>
            <template #overlay>
              <a-menu>
                <a-menu-item @click="selectRepo(r)" :disabled="r.cacheExists === false">
                  <FolderOpenOutlined /> 浏览代码
                </a-menu-item>
                <a-menu-item @click="updateRepo(r.name)">
                  <SyncOutlined /> 更新仓库
                </a-menu-item>
                <a-menu-item @click="openSccName = r.name; sccOpen = true" :disabled="r.cacheExists === false">
                  <BarChartOutlined /> 代码统计
                </a-menu-item>
                <a-menu-item v-if="r.cacheExists === false && r.type === 'remote'" @click="recloneRepo(r.name)">
                  <CloudDownloadOutlined /> 重新克隆
                </a-menu-item>
                <a-menu-divider />
                <a-menu-item danger @click="removeRepo(r.name)">
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

    <!-- right: code browser (only when repo selected) -->
    <div v-if="!selectedRepo" class="repo-placeholder">
      <FolderOpenOutlined class="rp-icon" />
      <div>从左侧选择仓库查看代码</div>
    </div>

    <div v-else-if="selectedRepoData && selectedRepoData.cacheExists === false" class="repo-missing">
      <div class="rm-icon"><FileOutlined /></div>
      <div class="rm-title">{{ selectedRepo }}</div>
      <div class="rm-hint">缓存目录不存在</div>
      <a-button v-if="selectedRepoData.type === 'remote'" type="primary" @click="recloneRepo(selectedRepo)">
        <CloudDownloadOutlined /> 重新克隆
      </a-button>
    </div>

    <div v-else class="repo-browser">
      <!-- file tree -->
      <aside class="rb-tree">
        <div class="rb-tree-head">
          <span class="rb-back" @click="selectedRepo = ''; selectedFile = ''" title="返回列表">
            <ArrowLeftOutlined />
          </span>
          <span class="rb-tree-title">{{ selectedRepo }}</span>
        </div>
        <div class="bt-search">
          <a-input :value="searchQuery" placeholder="搜索文件..." size="small" allow-clear
            @change="(e) => onSearchInput(e.target.value)">
            <template #prefix><SearchOutlined style="color: var(--color-text-tertiary)" /></template>
          </a-input>
        </div>
        <div class="bt-body">
          <div v-if="searchQuery.trim()">
            <div v-if="searching" class="bt-loading"><a-spin size="small" /></div>
            <div v-else-if="!searchResults.length" class="bt-empty">无匹配</div>
            <div v-for="r in searchResults" :key="r.path"
              class="sr-item" :class="{ active: selectedFile === r.path }"
              :title="r.path" @click="openFile(r)">
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
              :nodes="tree['__root__'] || []" :depth="0" :tree="tree"
              :expanded="expandedDirs" :selected="selectedFile" :loading-dirs="loadingDirs"
              @toggle="toggleDir" @open="openFile"
            />
          </template>
        </div>
      </aside>

      <!-- code content -->
      <div class="rb-content">
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

    <AddRepoModal v-model:open="addOpen" @added="loadRepos" />
    <SccModal v-model:open="sccOpen" :repo-name="openSccName || sccRepo" />
    <DiagnoseModal v-model:open="diagnoseOpen" :project-dir="project.currentDir" @update:open="diagnoseOpen = $event; loadRepos()" />
  </div>
</template>

<style scoped>
.repo-view { display: flex; height: 100%; width: 100%; overflow: hidden; }

/* left rail */
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

/* placeholder */
.repo-placeholder { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; color: var(--color-text-tertiary); font-size: 14px; }
.rp-icon { font-size: 48px; opacity: 0.2; }

/* missing */
.repo-missing { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; text-align: center; }
.rm-icon { font-size: 48px; color: var(--color-warning); opacity: 0.5; }
.rm-title { font-size: 18px; font-weight: 600; color: var(--color-text); }
.rm-hint { font-size: 14px; color: var(--color-text-tertiary); }

/* browser */
.repo-browser { flex: 1; min-width: 0; display: flex; overflow: hidden; }

/* tree */
.rb-tree {
  width: 260px; min-width: 260px; display: flex; flex-direction: column;
  border-right: 1px solid var(--color-border); background: var(--color-surface); overflow: hidden;
}
.rb-tree-head { display: flex; align-items: center; gap: 8px; padding: 0 12px; height: var(--navbar-height); border-bottom: 1px solid var(--color-border); flex-shrink: 0; }
.rb-back { display: flex; align-items: center; justify-content: center; width: 24px; height: 24px; cursor: pointer; border-radius: var(--radius-xs); color: var(--color-text-tertiary); transition: all var(--transition-fast); }
.rb-back:hover { background: var(--color-hover); color: var(--color-primary); }
.rb-tree-title { font-size: 13px; font-weight: 600; color: var(--color-text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.bt-search { padding: 8px 10px; flex-shrink: 0; }
.bt-body { flex: 1; overflow-y: auto; padding: 4px; }
.bt-body::-webkit-scrollbar { width: 6px; }
.bt-body::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 3px; }
.bt-loading { display: flex; justify-content: center; padding: 24px; }
.bt-empty { padding: 24px; text-align: center; font-size: 12px; color: var(--color-text-tertiary); }
.sr-item { display: flex; align-items: center; gap: 8px; padding: 5px 8px; border-radius: var(--radius-xs); cursor: pointer; transition: background var(--transition-fast); }
.sr-item:hover { background: var(--color-hover); }
.sr-item.active { background: var(--color-primary-bg); }
.sr-icon { font-size: 13px; color: var(--color-text-tertiary); flex-shrink: 0; }
.sr-body { min-width: 0; flex: 1; }
.sr-name { font-size: 13px; color: var(--color-text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.sr-mark { background: var(--color-warning-bg); color: var(--color-warning); padding: 0 2px; border-radius: 2px; }
.sr-dir { font-size: 11px; color: var(--color-text-tertiary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

/* content */
.rb-content { flex: 1; min-width: 0; display: flex; flex-direction: column; background: var(--color-background); }
.bc-bar { display: flex; align-items: center; gap: 10px; padding: 0 16px; height: var(--navbar-height); border-bottom: 1px solid var(--color-border); background: var(--color-surface); flex-shrink: 0; }
.bc-name { display: flex; align-items: center; gap: 6px; font-size: 14px; font-weight: 600; color: var(--color-text); }
.bc-path { font-size: 12px; color: var(--color-text-tertiary); font-family: 'Cascadia Code', monospace; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.bc-spacer { flex: 1; }
.bc-lines { font-size: 11px; color: var(--color-text-tertiary); }
.bc-body { flex: 1; overflow-y: auto; min-height: 0; }
.bc-body::-webkit-scrollbar { width: 8px; }
.bc-body::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 4px; }
.bc-loading { display: flex; justify-content: center; align-items: center; height: 100%; }
.bc-placeholder { display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; height: 100%; color: var(--color-text-tertiary); font-size: 14px; }
.bc-ph-icon { font-size: 48px; opacity: 0.2; }
.bc-state { display: flex; align-items: center; justify-content: center; height: 100%; color: var(--color-text-tertiary); font-size: 14px; }

.bc-code { padding: 0; }
.hljs-code-block {
  margin: 0; padding: 12px 16px;
  font-family: 'Cascadia Code', 'Fira Code', monospace; font-size: 13px; line-height: 1.6;
  overflow-x: auto; min-height: 100%;
}
.hljs-code-block code { background: transparent !important; display: block; }

/* highlightjs-line-numbers.js emits a <table.hljs-ln> with one <tr> per line;
   each row holds the number cell (.hljs-ln-n) and the code cell (.hljs-ln-code).
   Because they share a row, vertical alignment is guaranteed. */
.bc-code :deep(table.hljs-ln) { border-collapse: collapse; }
.bc-code :deep(td) { padding: 0; vertical-align: top; }
.bc-code :deep(td.hljs-ln-n) {
  width: 1px; white-space: nowrap; text-align: right; user-select: none;
  padding-right: 16px !important; opacity: 0.4; color: var(--color-text-tertiary);
}
.bc-code :deep(td.hljs-ln-code) { padding-left: 16px !important; white-space: pre; }

.bc-md { padding: 24px 32px; font-size: 14px; line-height: 1.7; color: var(--color-text); }
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
