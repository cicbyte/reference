<script setup>
/**
 * Shared code browser: file tree + search + syntax-highlighted viewer.
 *
 * Extracted from the near-identical blocks in RepoListView.vue /
 * CacheReposView.vue. The only real difference between those two was which
 * backend API to call and which key (repo name vs cache path) drives it —
 * both are injected via the `api` prop below.
 *
 * Props:
 *   api.listDir(subPath)   → Promise<BrowserFileNode[]>   (root when subPath='')
 *   api.readFile(relPath)  → Promise<{content,binary,notFound,lines}>
 *   api.search(query)      → Promise<BrowserFileNode[]>
 *   showBack               → show a back button + title in the tree header
 *   title                  → text shown next to the back button
 *
 * The parent must call `loadRoot()` (exposed via defineExpose) whenever the
 * underlying key changes (e.g. user selects another repo).
 */
import { ref, computed, watch, nextTick, onUnmounted } from 'vue'
import { message } from 'ant-design-vue'
import {
  SearchOutlined, FileOutlined, FileTextOutlined,
  CodeOutlined, EyeOutlined, ArrowLeftOutlined,
} from '@ant-design/icons-vue'
import hljs, { hljsLangForName } from '@/utils/hljs-setup'
import 'highlightjs-line-numbers.js/src/highlightjs-line-numbers.js'
import { renderMarkdown } from '@/utils/markdown'
import FileTreeNode from '@/components/repo/FileTreeNode.vue'

const props = defineProps({
  api: { type: Object, required: true },
  showBack: { type: Boolean, default: false },
  title: { type: String, default: '' },
})
const emit = defineEmits(['back'])

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

watch(mdViewMode, (mode) => {
  if (mode === 'source' && selectedFile.value) highlightCode()
})

const isMarkdown = computed(() => /\.md$/i.test(selectedFile.value))
const filename = computed(() => selectedFile.value.split('/').pop() || '')
const hljsLang = hljsLangForName

const renderedMarkdown = computed(() => {
  if (!isMarkdown.value || mdViewMode.value !== 'render') return ''
  return renderMarkdown(fileContent.value)
})

async function loadRoot() {
  try {
    const nodes = await props.api.listDir('')
    tree.value = { __root__: nodes }
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
        const children = await props.api.listDir(node.path)
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
    const res = await props.api.readFile(node.path)
    fileInfo.value = { binary: res.binary, notFound: res.notFound, lines: res.lines }
    fileContent.value = res.content
  } catch (e) {
    message.error('读取失败: ' + e)
  } finally {
    fileLoading.value = false
    if (!fileInfo.value.binary && !fileInfo.value.notFound) {
      await nextTick(); await nextTick()
      if (!isMarkdown.value || mdViewMode.value === 'source') highlightCode()
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
    if (hljs.lineNumbersBlockSync) hljs.lineNumbersBlockSync(codeRef.value)
  } catch { codeRef.value.textContent = fileContent.value }
}

function onSearchInput(val) {
  searchQuery.value = val
  if (searchTimer) clearTimeout(searchTimer)
  if (!val.trim()) { searchResults.value = []; searching.value = false; return }
  searching.value = true
  searchTimer = setTimeout(async () => {
    try { searchResults.value = await props.api.search(val) }
    catch { searchResults.value = [] }
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

function clearSelection() {
  selectedFile.value = ''
  fileContent.value = ''
  tree.value = {}
  expandedDirs.value = new Set()
}

onUnmounted(() => { if (searchTimer) clearTimeout(searchTimer) })

defineExpose({ loadRoot, clearSelection })
</script>

<template>
  <div class="code-browser">
    <!-- file tree -->
    <aside class="cb-tree">
      <div class="cb-tree-head">
        <span v-if="showBack" class="cb-back" title="返回列表" @click="emit('back')">
          <ArrowLeftOutlined />
        </span>
        <span class="cb-tree-title">{{ title }}</span>
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
</template>

<style scoped>
.code-browser { flex: 1; min-width: 0; display: flex; overflow: hidden; }

/* tree */
.cb-tree {
  width: 260px; min-width: 260px; display: flex; flex-direction: column;
  border-right: 1px solid var(--color-border); background: var(--color-surface); overflow: hidden;
}
.cb-tree-head { display: flex; align-items: center; gap: 8px; padding: 0 12px; height: var(--navbar-height); border-bottom: 1px solid var(--color-border); flex-shrink: 0; }
.cb-back { display: flex; align-items: center; justify-content: center; width: 24px; height: 24px; cursor: pointer; border-radius: var(--radius-xs); color: var(--color-text-tertiary); transition: all var(--transition-fast); }
.cb-back:hover { background: var(--color-hover); color: var(--color-primary); }
.cb-tree-title { font-size: 13px; font-weight: 600; color: var(--color-text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
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
.cb-content { flex: 1; min-width: 0; display: flex; flex-direction: column; background: var(--color-background); }
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
