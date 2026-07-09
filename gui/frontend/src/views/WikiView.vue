<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { message } from 'ant-design-vue'
import {
  ReloadOutlined,
  SyncOutlined,
  ReadOutlined,
  FileTextOutlined,
  CaretRightFilled,
} from '@ant-design/icons-vue'
import { marked } from 'marked'

const app = window.go?.main?.ReferenceApp

// ---- data ----
const entries = ref([])
const loading = ref(true)
const syncing = ref(false)
const selectedKey = ref('')  // source + '|' + relPath
const content = ref('')
const contentLoading = ref(false)

const selectedEntry = computed(() =>
  entries.value.find((e) => e.source + '|' + e.relPath === selectedKey.value),
)

// ---- grouped tree: source → platform → namespace → repos ----
const expandedPlatforms = ref(new Set())
const expandedNamespaces = ref(new Set())

const groupedEntries = computed(() => {
  const tree = {}
  for (const e of entries.value) {
    const platform = e.source === 'local' ? '本地知识库' : (e.platform || '未知平台').toUpperCase()
    const ns = e.source === 'local' ? '' : (e.namespace || '未知')
    if (!tree[platform]) tree[platform] = {}
    if (!tree[platform][ns]) tree[platform][ns] = {}
    if (!tree[platform][ns][e.repoName]) tree[platform][ns][e.repoName] = []
    tree[platform][ns][e.repoName].push(e)
  }
  const platforms = Object.keys(tree).sort((a, b) => {
    if (a === '本地知识库') return 1
    if (b === '本地知识库') return -1
    return a.localeCompare(b)
  })
  return platforms.map((p) => ({
    platform: p,
    namespaces: Object.keys(tree[p]).sort().map((ns) => ({
      namespace: ns,
      repos: Object.keys(tree[p][ns]).sort().map((rn) => ({
        repoName: rn,
        files: tree[p][ns][rn].sort((a, b) => a.fileName.localeCompare(b.fileName)),
      })),
    })),
  }))
})

function togglePlatform(p) {
  const next = new Set(expandedPlatforms.value)
  next.has(p) ? next.delete(p) : next.add(p)
  expandedPlatforms.value = next
}
function toggleNamespace(key) {
  const next = new Set(expandedNamespaces.value)
  next.has(key) ? next.delete(key) : next.add(key)
  expandedNamespaces.value = next
}

// auto-expand on first load
watch(groupedEntries, (groups) => {
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

// ---- load ----
async function loadEntries() {
  loading.value = true
  try {
    if (app?.ListWikiEntries) {
      entries.value = await app.ListWikiEntries('all')
    }
  } catch (e) {
    message.error('加载失败: ' + e)
  } finally {
    loading.value = false
  }
}

async function loadContent(entry) {
  if (!entry) return
  contentLoading.value = true
  content.value = ''
  try {
    if (app?.ReadWikiEntry) {
      const raw = await app.ReadWikiEntry(entry.source, entry.relPath)
      // strip frontmatter before rendering
      content.value = stripFrontmatter(raw)
    }
  } catch (e) {
    message.error('读取失败: ' + e)
  } finally {
    contentLoading.value = false
  }
}

function stripFrontmatter(md) {
  if (!md.startsWith('---')) return md
  const rest = md.slice(3)
  const end = rest.indexOf('\n---')
  if (end < 0) return md
  return rest.slice(end + 4).replace(/^\s*\n/, '')
}

const renderedContent = computed(() => {
  if (!content.value) return ''
  try { return marked(content.value) } catch { return content.value }
})

function selectEntry(entry) {
  selectedKey.value = entry.source + '|' + entry.relPath
  loadContent(entry)
}

async function doSync() {
  syncing.value = true
  try {
    if (app?.WikiSync) {
      await app.WikiSync()
      message.success('同步成功')
      await loadEntries()
    }
  } catch (e) {
    message.error('同步失败: ' + e)
  } finally {
    syncing.value = false
  }
}

onMounted(loadEntries)
</script>

<template>
  <div class="wiki-view">
    <!-- left: repo rail -->
    <aside class="wiki-rail">
      <div class="rail-head">
        <span>知识库</span>
        <div class="rail-actions">
          <button class="rail-btn" title="同步(pull+commit+push)" :disabled="syncing" @click="doSync">
            <SyncOutlined :spin="syncing" />
          </button>
          <button class="rail-btn" title="刷新" @click="loadEntries">
            <ReloadOutlined />
          </button>
        </div>
      </div>
      <div class="rail-list">
        <a-spin v-if="loading" class="rail-spin" />

        <template v-for="pg in groupedEntries" :key="pg.platform">
          <div class="rail-group-head" @click="togglePlatform(pg.platform)">
            <CaretRightFilled class="rail-caret" :class="{ open: expandedPlatforms.has(pg.platform) }" />
            <span class="rail-group-label">{{ pg.platform }}</span>
            <span class="rail-group-count">{{ pg.namespaces.reduce((s, n) => s + n.repos.reduce((s2, r) => s2 + r.files.length, 0), 0) }}</span>
          </div>

          <template v-if="expandedPlatforms.has(pg.platform)">
            <template v-for="ns in pg.namespaces" :key="pg.platform + '/' + ns.namespace">
              <div
                v-if="pg.platform !== '本地知识库' && ns.namespace"
                class="rail-ns-head"
                @click="toggleNamespace(pg.platform + '/' + ns.namespace)"
              >
                <CaretRightFilled class="rail-caret sm" :class="{ open: expandedNamespaces.has(pg.platform + '/' + ns.namespace) }" />
                <span class="rail-ns-label">{{ ns.namespace }}</span>
              </div>

              <template v-if="pg.platform === '本地知识库' || !ns.namespace || expandedNamespaces.has(pg.platform + '/' + ns.namespace)">
                <template v-for="repo in ns.repos" :key="pg.platform + '/' + ns.namespace + '/' + repo.repoName">
                  <div
                    v-for="file in repo.files"
                    :key="file.source + '|' + file.relPath"
                    class="rail-item"
                    :class="{
                      active: selectedKey === file.source + '|' + file.relPath,
                      nested: pg.platform !== '本地知识库' && ns.namespace,
                    }"
                    :title="file.relPath"
                    @click="selectEntry(file)"
                  >
                    <div class="rail-item-icon">
                      <ReadOutlined v-if="file.fileName === 'reference.md'" />
                      <FileTextOutlined v-else />
                    </div>
                    <div class="rail-item-body">
                      <div class="rail-item-name">
                        {{ file.fileName === 'reference.md' ? repo.repoName : file.fileName.replace('.md', '') }}
                      </div>
                      <div class="rail-item-meta">
                        <span v-if="file.exploredAt">{{ file.exploredAt }}</span>
                        <span v-if="file.fileName !== 'reference.md'" class="rail-topic-tag">主题</span>
                      </div>
                    </div>
                  </div>
                </template>
              </template>
            </template>
          </template>
        </template>

        <div v-if="!loading && !entries.length" class="rail-empty">
          <ReadOutlined class="rail-empty-icon" />
          <span>暂无知识文件</span>
          <span class="rail-empty-tip">使用 AI 子代理探索仓库后生成</span>
        </div>
      </div>
    </aside>

    <!-- right: markdown content -->
    <div v-if="!selectedEntry" class="wiki-placeholder">
      <ReadOutlined class="wp-icon" />
      <div>从左侧选择一个知识文件查看内容</div>
    </div>

    <div v-else class="wiki-content">
      <!-- info bar -->
      <div class="wc-bar">
        <div class="wc-title">
          <ReadOutlined class="wc-icon" />
          <span>{{ selectedEntry.repoName }}</span>
          <span class="wc-file" v-if="selectedEntry.fileName !== 'reference.md'">
            / {{ selectedEntry.fileName }}
          </span>
        </div>
        <div class="wc-meta">
          <span v-if="selectedEntry.namespace">{{ selectedEntry.namespace }}/</span>
          <span class="wc-platform">{{ selectedEntry.platform }}</span>
          <span v-if="selectedEntry.commit" class="wc-commit">{{ selectedEntry.commit }}</span>
          <span v-if="selectedEntry.exploredAt" class="wc-date">{{ selectedEntry.exploredAt }}</span>
        </div>
      </div>

      <!-- rendered markdown -->
      <div class="wc-body">
        <a-spin v-if="contentLoading" class="wc-loading" />
        <div v-else class="wc-md" v-html="renderedContent"></div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.wiki-view {
  display: flex;
  height: 100%;
  width: 100%;
  overflow: hidden;
}

/* ---- left rail ---- */
.wiki-rail {
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
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}
.rail-head > span {
  font-size: 12px; font-weight: 600; text-transform: uppercase;
  letter-spacing: 0.06em; color: var(--color-text-tertiary);
}
.rail-actions { display: flex; gap: 2px; }
.rail-btn {
  display: flex; align-items: center; justify-content: center;
  width: 26px; height: 26px; border: none; background: transparent;
  color: var(--color-text-tertiary); cursor: pointer;
  border-radius: var(--radius-xs); transition: all var(--transition-fast);
}
.rail-btn:hover { background: var(--color-hover); color: var(--color-primary); }
.rail-btn:disabled { opacity: 0.4; cursor: not-allowed; }

.rail-list { flex: 1; overflow-y: auto; padding: var(--spacing-sm); }
.rail-list::-webkit-scrollbar { width: 5px; }
.rail-list::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 3px; }
.rail-spin { display: flex; justify-content: center; padding: var(--spacing-lg) 0; }

/* group headers */
.rail-group-head {
  display: flex; align-items: center; gap: 6px;
  padding: 6px 8px 4px; cursor: pointer; user-select: none;
}
.rail-group-label {
  flex: 1; font-size: 11px; font-weight: 600; text-transform: uppercase;
  letter-spacing: 0.05em; color: var(--color-text-tertiary);
}
.rail-group-count {
  font-size: 10px; font-weight: 600; color: var(--color-text-tertiary);
  background: var(--color-surface-raised); padding: 0 6px; border-radius: 999px;
}
.rail-ns-head {
  display: flex; align-items: center; gap: 5px;
  padding: 4px 8px 2px 20px; cursor: pointer; user-select: none;
}
.rail-ns-label {
  flex: 1; font-size: 12px; font-weight: 500; color: var(--color-text-secondary);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.rail-caret {
  font-size: 9px; color: var(--color-text-tertiary); flex-shrink: 0;
  transition: transform var(--transition-fast);
}
.rail-caret.sm { font-size: 8px; }
.rail-caret.open { transform: rotate(90deg); }

/* repo items */
.rail-item {
  display: flex; align-items: center; gap: 10px;
  padding: 7px 10px; border-radius: var(--radius-md);
  cursor: pointer; transition: all var(--transition-fast); margin-bottom: 1px;
}
.rail-item.nested { padding-left: 34px; }
.rail-item:hover { background: var(--color-hover); }
.rail-item.active { background: var(--color-primary-bg); }
.rail-item-icon {
  display: flex; align-items: center; justify-content: center;
  width: 26px; height: 26px; border-radius: var(--radius-sm);
  background: var(--color-background); color: var(--color-text-tertiary);
  font-size: 13px; flex-shrink: 0; transition: all var(--transition-fast);
}
.rail-item:hover .rail-item-icon,
.rail-item.active .rail-item-icon { background: var(--color-primary); color: #fff; }
.rail-item-body { flex: 1; min-width: 0; }
.rail-item-name {
  font-size: 13px; font-weight: 500; color: var(--color-text);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.rail-item.active .rail-item-name { color: var(--color-primary); }
.rail-item-meta {
  display: flex; gap: 6px; margin-top: 1px; font-size: 11px;
  color: var(--color-text-tertiary);
}
.rail-topic-tag {
  font-size: 9px; font-weight: 600; padding: 0 4px;
  border-radius: 3px; background: var(--color-primary-bg); color: var(--color-primary);
}

.rail-empty {
  display: flex; flex-direction: column; align-items: center; gap: 4px; padding: 40px 0;
}
.rail-empty-icon { font-size: 28px; color: var(--color-text-tertiary); opacity: 0.35; }
.rail-empty span:nth-child(2) { font-size: 13px; color: var(--color-text-secondary); }
.rail-empty-tip { font-size: 11px; color: var(--color-text-tertiary) !important; }

/* ---- placeholder ---- */
.wiki-placeholder {
  flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 12px; color: var(--color-text-tertiary); font-size: 14px;
}
.wp-icon { font-size: 48px; opacity: 0.2; }

/* ---- content ---- */
.wiki-content {
  flex: 1; min-width: 0; display: flex; flex-direction: column;
  background: var(--color-background);
}
.wc-bar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 20px; height: var(--navbar-height);
  border-bottom: 1px solid var(--color-border);
  background: var(--color-surface); flex-shrink: 0; gap: 12px;
}
.wc-title {
  display: flex; align-items: center; gap: 6px;
  font-size: 15px; font-weight: 600; color: var(--color-text);
}
.wc-icon { color: var(--color-text-tertiary); }
.wc-file { font-size: 13px; font-weight: 400; color: var(--color-text-secondary); }
.wc-meta {
  display: flex; align-items: center; gap: 8px; flex-shrink: 0;
  font-size: 12px; color: var(--color-text-tertiary);
}
.wc-platform { color: var(--color-text-secondary); }
.wc-commit { font-family: 'Cascadia Code', monospace; }
.wc-date { font-family: 'Cascadia Code', monospace; }

.wc-body { flex: 1; overflow-y: auto; min-height: 0; }
.wc-body::-webkit-scrollbar { width: 8px; }
.wc-body::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 4px; }
.wc-loading { display: flex; justify-content: center; padding: 40px; }

.wc-md {
  padding: 24px 32px; font-size: 14px; line-height: 1.7;
  color: var(--color-text); max-width: 860px;
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
</style>
