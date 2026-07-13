<script setup>
import { ref, watch, computed, onMounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  FolderOpenOutlined, FileOutlined, FileTextOutlined, CodeOutlined,
  EyeOutlined, SearchOutlined, ArrowLeftOutlined,
} from '@ant-design/icons-vue'
import hljs from 'highlight.js/lib/core'
import go from 'highlight.js/lib/languages/go'
import javascript from 'highlight.js/lib/languages/javascript'
import typescript from 'highlight.js/lib/languages/typescript'
import xml from 'highlight.js/lib/languages/xml'
import css from 'highlight.js/lib/languages/css'
import json from 'highlight.js/lib/languages/json'
import yaml from 'highlight.js/lib/languages/yaml'
import markdown from 'highlight.js/lib/languages/markdown'
import python from 'highlight.js/lib/languages/python'
import bash from 'highlight.js/lib/languages/bash'
import sql from 'highlight.js/lib/languages/sql'
import rust from 'highlight.js/lib/languages/rust'
import 'highlight.js/styles/github-dark.css'
import { renderMarkdown } from '../utils/markdown'
import FileTreeNode from '../components/repo/FileTreeNode.vue'

// register only the languages we need (keeps bundle small)
hljs.registerLanguage('go', go)
hljs.registerLanguage('javascript', javascript)
hljs.registerLanguage('typescript', typescript)
hljs.registerLanguage('xml', xml)
hljs.registerLanguage('css', css)
hljs.registerLanguage('json', json)
hljs.registerLanguage('yaml', yaml)
hljs.registerLanguage('markdown', markdown)
hljs.registerLanguage('python', python)
hljs.registerLanguage('bash', bash)
hljs.registerLanguage('sql', sql)
hljs.registerLanguage('rust', rust)

const route = useRoute()
const app = window.go?.main?.ReferenceApp
const refName = computed(() => route.params.name || '')

// ---- tree state (lazy) ----
const tree = ref({})               // path → children[]
const expandedDirs = ref(new Set())
const loadingDirs = ref(new Set())
const selectedFile = ref('')

// ---- search state ----
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

const isMarkdown = computed(() => /\.md$/i.test(selectedFile.value))
const filename = computed(() => selectedFile.value.split('/').pop() || '')

// ext → hljs language name
function hljsLang(name) {
  const ext = (name.split('.').pop() || '').toLowerCase()
  const map = {
    go: 'go', js: 'javascript', mjs: 'javascript', cjs: 'javascript',
    ts: 'typescript', jsx: 'javascript', tsx: 'typescript',
    html: 'xml', htm: 'xml', vue: 'xml', svg: 'xml',
    css: 'css', scss: 'css', less: 'css',
    json: 'json', yml: 'yaml', yaml: 'yaml',
    md: 'markdown', py: 'python', rb: 'ruby',
    sh: 'bash', bash: 'bash', zsh: 'bash',
    sql: 'sql', rs: 'rust',
  }
  return map[ext] || ''
}

// ---- tree operations ----
async function loadRoot() {
  if (!app || !refName.value) return
  try {
    const nodes = await app.BrowseRepoList(refName.value, '')
    tree.value = { '__root__': nodes }
  } catch (e) {
    message.error('加载仓库失败: ' + e)
  }
}

async function toggleDir(node) {
  const next = new Set(expandedDirs.value)
  if (next.has(node.path)) {
    next.delete(node.path)
  } else {
    next.add(node.path)
    // lazy-load children if not already loaded
    if (!(node.path in tree.value)) {
      const ld = new Set(loadingDirs.value)
      ld.add(node.path)
      loadingDirs.value = ld
      try {
        const children = await app.BrowseRepoList(refName.value, node.path)
        tree.value = { ...tree.value, [node.path]: children }
      } catch (e) {
        message.error('读取目录失败: ' + e)
      } finally {
        const ld2 = new Set(loadingDirs.value)
        ld2.delete(node.path)
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
    const res = await app.BrowseRepoRead(refName.value, node.path)
    fileInfo.value = { binary: res.binary, notFound: res.notFound, lines: res.lines }
    fileContent.value = res.content
  } catch (e) {
    message.error('读取文件失败: ' + e)
  } finally {
    fileLoading.value = false
    // highlight AFTER loading=false so the <code> DOM is actually rendered.
    // Two nextTick calls guard against Vue batching the v-if/v-else flip.
    if (!fileInfo.value.binary && !fileInfo.value.notFound && !isMarkdown.value) {
      await nextTick()
      await nextTick()
      highlightCode()
    }
  }
}

function highlightCode() {
  if (!codeRef.value) {
    // DOM not ready yet — retry on next frame
    requestAnimationFrame(highlightCode)
    return
  }
  const lang = hljsLang(filename.value)
  try {
    if (lang && hljs.getLanguage(lang)) {
      codeRef.value.innerHTML = hljs.highlight(fileContent.value, { language: lang }).value
    } else {
      codeRef.value.textContent = fileContent.value
    }
  } catch {
    codeRef.value.textContent = fileContent.value
  }
}

// ---- search ----
function onSearchInput(val) {
  searchQuery.value = val
  if (searchTimer) clearTimeout(searchTimer)
  if (!val.trim()) {
    searchResults.value = []
    searching.value = false
    return
  }
  searching.value = true
  searchTimer = setTimeout(async () => {
    try {
      searchResults.value = await app.BrowseRepoSearch(refName.value, val)
    } catch (e) {
      searchResults.value = []
    } finally {
      searching.value = false
    }
  }, 300)
}

// highlight matched segments in search results
function highlightSegments(name) {
  const q = searchQuery.value.toLowerCase()
  const ql = q.length
  const lower = name.toLowerCase()
  const segs = []
  let i = 0
  while (i < name.length) {
    const idx = lower.indexOf(q, i)
    if (idx === -1) {
      segs.push({ text: name.slice(i), match: false })
      break
    }
    if (idx > i) segs.push({ text: name.slice(i, idx), match: false })
    segs.push({ text: name.slice(idx, idx + ql), match: true })
    i = idx + ql
  }
  return segs
}

function dirOf(path) {
  const idx = path.lastIndexOf('/')
  return idx > 0 ? path.slice(0, idx) : ''
}

function openSearchResult(r) {
  // build a fake file node for openFile
  openFile({ isDir: false, path: r.path, name: r.name })
}

const renderedMarkdown = computed(() => {
  if (!isMarkdown.value || mdViewMode.value !== 'render') return ''
  return renderMarkdown(fileContent.value)
})

// load root on mount + when refName changes (navigating between repos)
onMounted(loadRoot)
watch(refName, loadRoot)
</script>

<template>
  <div class="browse-view">
    <!-- left: file tree / search -->
    <aside class="browse-tree">
      <div class="bt-head">
        <span class="bt-back" @click="$router.push('/repos')" title="返回仓库列表">
          <ArrowLeftOutlined />
        </span>
        <FolderOpenOutlined class="bt-icon" />
        <span class="bt-title">{{ refName }}</span>
      </div>
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
        <!-- search results -->
        <div v-if="searchQuery.trim()">
          <div v-if="searching" class="bt-loading"><a-spin size="small" /></div>
          <div v-else-if="!searchResults.length" class="bt-empty">无匹配文件</div>
          <div v-else class="sr-list">
            <div
              v-for="r in searchResults" :key="r.path"
              class="sr-item"
              :class="{ active: selectedFile === r.path }"
              :title="r.path"
              @click="openSearchResult(r)"
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
            <div v-if="searchResults.length >= 500" class="sr-cap">
              结果过多，仅显示前 500 条
            </div>
          </div>
        </div>
        <!-- tree -->
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
          <div v-if="!((tree['__root__'] || []).length)" class="bt-empty">加载中...</div>
        </template>
      </div>
    </aside>

    <!-- right: file content -->
    <div class="browse-content">
      <div class="bc-bar" v-if="selectedFile">
        <span class="bc-name" :title="selectedFile">
          <FileTextOutlined class="bc-name-icon" />
          {{ filename }}
        </span>
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
          <FileOutlined class="bc-placeholder-icon" />
          <div>从左侧选择文件查看内容</div>
        </div>
        <div v-else-if="fileInfo.notFound" class="bc-state">该文件不存在</div>
        <div v-else-if="fileInfo.binary" class="bc-state">二进制文件，无法显示</div>
        <!-- markdown render -->
        <div
          v-else-if="isMarkdown && mdViewMode === 'render'"
          class="bc-md"
          v-html="renderedMarkdown"
        ></div>
        <!-- code view -->
        <div v-else class="bc-code">
          <pre class="hljs-code-block"><code ref="codeRef" class="hljs"></code></pre>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.browse-view {
  display: flex;
  height: 100%;
  width: 100%;
  overflow: hidden;
}

/* ---- left tree ---- */
.browse-tree {
  width: 280px;
  min-width: 280px;
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--color-border);
  background: var(--color-surface);
  overflow: hidden;
}
.bt-head {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
  height: var(--navbar-height);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}
.bt-back {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  cursor: pointer;
  border-radius: var(--radius-xs);
  color: var(--color-text-tertiary);
  transition: all var(--transition-fast);
}
.bt-back:hover { background: var(--color-hover); color: var(--color-primary); }
.bt-icon { color: var(--color-text-tertiary); font-size: 15px; }
.bt-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.bt-search { padding: 8px 10px; flex-shrink: 0; }
.bt-body {
  flex: 1;
  overflow-y: auto;
  padding: 4px;
}
.bt-body::-webkit-scrollbar { width: 6px; }
.bt-body::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 3px; }

.bt-loading { display: flex; justify-content: center; padding: 24px; }
.bt-empty { padding: 24px; text-align: center; font-size: 12px; color: var(--color-text-tertiary); }

/* search results */
.sr-list { display: flex; flex-direction: column; }
.sr-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 5px 8px;
  border-radius: var(--radius-xs);
  cursor: pointer;
  transition: background var(--transition-fast);
}
.sr-item:hover { background: var(--color-hover); }
.sr-item.active { background: var(--color-primary-bg); }
.sr-icon { font-size: 13px; color: var(--color-text-tertiary); flex-shrink: 0; }
.sr-body { min-width: 0; flex: 1; }
.sr-name { font-size: 13px; color: var(--color-text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.sr-mark { background: var(--color-warning-bg); color: var(--color-warning); padding: 0 2px; border-radius: 2px; }
.sr-dir { font-size: 11px; color: var(--color-text-tertiary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.sr-cap { padding: 8px; font-size: 11px; color: var(--color-text-tertiary); text-align: center; }

/* ---- right content ---- */
.browse-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  background: var(--color-background);
}
.bc-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 16px;
  height: var(--navbar-height);
  border-bottom: 1px solid var(--color-border);
  background: var(--color-surface);
  flex-shrink: 0;
}
.bc-name {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}
.bc-name-icon { color: var(--color-text-tertiary); }
.bc-path {
  font-size: 12px;
  color: var(--color-text-tertiary);
  font-family: 'Cascadia Code', monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.bc-spacer { flex: 1; }
.bc-lines { font-size: 11px; color: var(--color-text-tertiary); }

.bc-body { flex: 1; overflow-y: auto; min-height: 0; }
.bc-body::-webkit-scrollbar { width: 8px; }
.bc-body::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 4px; }

.bc-loading { display: flex; justify-content: center; align-items: center; height: 100%; }
.bc-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  height: 100%;
  color: var(--color-text-tertiary);
  font-size: 14px;
}
.bc-placeholder-icon { font-size: 48px; opacity: 0.3; }
.bc-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--color-text-tertiary);
  font-size: 14px;
}

/* code block */
.bc-code { padding: 0; }
.hljs-code-block {
  margin: 0;
  padding: 16px 0;
  font-family: 'Cascadia Code', 'Fira Code', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.6;
  overflow-x: auto;
}
.hljs-code-block code {
  display: block;
  padding: 0 20px;
  background: transparent !important;
  white-space: pre;
}

/* markdown */
.bc-md {
  padding: 24px 32px;
  font-size: 14px;
  line-height: 1.7;
  color: var(--color-text);
  max-width: 860px;
}
.bc-md :deep(h1) { font-size: 24px; font-weight: 700; margin: 24px 0 12px; }
.bc-md :deep(h2) { font-size: 20px; font-weight: 600; margin: 20px 0 10px; border-bottom: 1px solid var(--color-border); padding-bottom: 6px; }
.bc-md :deep(h3) { font-size: 16px; font-weight: 600; margin: 16px 0 8px; }
.bc-md :deep(p) { margin: 8px 0; }
.bc-md :deep(code) {
  font-family: 'Cascadia Code', monospace;
  font-size: 12px;
  padding: 2px 6px;
  background: var(--color-surface-raised);
  border-radius: 3px;
}
.bc-md :deep(pre) {
  padding: 12px 16px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  overflow-x: auto;
  margin: 12px 0;
}
.bc-md :deep(pre code) { background: none; padding: 0; }
.bc-md :deep(a) { color: var(--color-primary); }
.bc-md :deep(ul), .bc-md :deep(ol) { padding-left: 24px; margin: 8px 0; }
.bc-md :deep(li) { margin: 4px 0; }
.bc-md :deep(table) { border-collapse: collapse; width: 100%; margin: 12px 0; }
.bc-md :deep(th), .bc-md :deep(td) { border: 1px solid var(--color-border); padding: 6px 12px; text-align: left; }
.bc-md :deep(th) { background: var(--color-surface); font-weight: 600; }
.bc-md :deep(blockquote) { border-left: 3px solid var(--color-primary); padding-left: 16px; margin: 12px 0; color: var(--color-text-secondary); }
</style>
