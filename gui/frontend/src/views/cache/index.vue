<script setup>
/**
 * Cache view. Left rail (CacheRail) shows cached repos grouped by
 * platform → namespace. The right pane is the shared CodeBrowser wired to the
 * BrowseCacheByPath* backend API for the currently selected cache path.
 *
 * The browser/tree/search/highlight/markdown logic lives entirely in the
 * shared CodeBrowser component — this view only supplies the API adapter and
 * calls `browserRef.loadRoot()` whenever the selected cache path changes.
 */
import { ref, computed, onMounted, watch } from 'vue'
import { message } from 'ant-design-vue'
import { useI18n } from 'vue-i18n'
import {
  CloudServerOutlined,
  WarningOutlined,
} from '@ant-design/icons-vue'
import { formatPath } from '@/utils/path'
import CodeBrowser from '@/components/shared/CodeBrowser.vue'
import CacheRail from './components/CacheRail.vue'

const { t } = useI18n()
const app = window.go?.main?.ReferenceApp

// Stable sentinel key for the "local repos" platform group (kept stable
// across language switches; only the display label is translated).
const LOCAL_PLATFORM_KEY = '__local__'

// ---- repo list ----
const repos = ref([])
const loading = ref(true)
const selectedCachePath = ref('')

const selectedRepo = computed(() =>
  repos.value.find((r) => r.cachePath === selectedCachePath.value),
)

async function loadRepos() {
  loading.value = true
  try {
    if (app) {
      repos.value = await app.ListCachedRepos()
      // async-load sizes for each repo
      repos.value.forEach((r) => { fetchSize(r) })
    }
  } catch (e) {
    message.error(t('cache.loadFailed') + ': ' + e)
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
    const platform = r.type === 'local' ? LOCAL_PLATFORM_KEY : (r.host || '')
    const ns = r.type === 'local' ? '' : (r.namespace || '')
    if (!tree[platform]) tree[platform] = {}
    if (!tree[platform][ns]) tree[platform][ns] = []
    tree[platform][ns].push(r)
  }
  // sort platforms (local last), namespaces alpha, repos alpha
  const platforms = Object.keys(tree).sort((a, b) => {
    if (a === LOCAL_PLATFORM_KEY) return 1
    if (b === LOCAL_PLATFORM_KEY) return -1
    return a.localeCompare(b)
  })
  return platforms.map((p) => ({
    platform: p,
    platformLabel: p === LOCAL_PLATFORM_KEY
      ? t('cache.localRepos')
      : (p || t('cache.unknownPlatform')),
    isLocal: p === LOCAL_PLATFORM_KEY,
    namespaces: Object.keys(tree[p]).sort().map((ns) => ({
      namespace: ns,
      namespaceLabel: ns || t('common.unknown'),
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
const browserRef = ref(null)

// Shared CodeBrowser API adapter — delegates to BrowseCacheByPath* backend
// calls, bound to the currently selected cache path.
const browserApi = computed(() => ({
  listDir: (sub) => app.BrowseCacheByPathList(selectedCachePath.value, sub),
  readFile: (rel) => app.BrowseCacheByPathRead(selectedCachePath.value, rel),
  search: (q) => app.BrowseCacheByPathSearch(selectedCachePath.value, q),
}))

function selectRepo(repo) {
  selectedCachePath.value = repo.cachePath
  // CodeBrowser 自行 watch(api) 响应 selectedCachePath 变化并重新加载根，无需手动调用
}

// ---- purge ----
async function purge(repo) {
  try {
    await app.PurgeCachedRepo(repo.cachePath)
    message.success(t('cache.purged', { name: repo.name }))
    if (selectedCachePath.value === repo.cachePath) {
      selectedCachePath.value = ''
      browserRef.value?.clearSelection()
    }
    await loadRepos()
  } catch (e) {
    message.error(t('cache.purgeFailed') + ': ' + e)
  }
}
</script>

<template>
  <div class="cache-view">
    <CacheRail
      :grouped-repos="groupedRepos"
      :expanded-platforms="expandedPlatforms"
      :expanded-namespaces="expandedNamespaces"
      :selected-cache-path="selectedCachePath"
      :loading="loading"
      :repo-count="repos.length"
      @refresh="loadRepos"
      @toggle-platform="togglePlatform"
      @toggle-namespace="toggleNamespace"
      @select="selectRepo"
      @purge="purge"
    />

    <!-- placeholder: nothing selected -->
    <div v-if="!selectedCachePath" class="cache-placeholder">
      <CloudServerOutlined class="cp-icon" />
      <div>{{ t('cache.selectToBrowse') }}</div>
    </div>

    <!-- selected repo path doesn't exist on disk -->
    <div v-else-if="selectedRepo && !selectedRepo.exists" class="cache-missing">
      <WarningOutlined class="cm-icon" />
      <div class="cm-title">{{ t('cache.pathNotExistTitle') }}</div>
      <div class="cm-path">{{ formatPath(selectedRepo.cachePath) }}</div>
      <div class="cm-hint" v-if="selectedRepo.type === 'local'">
        {{ t('cache.localHint') }}
      </div>
      <div class="cm-hint" v-else>
        {{ t('cache.cacheHint') }}
      </div>
    </div>

    <!-- code browser -->
    <CodeBrowser
      v-else
      ref="browserRef"
      :api="browserApi"
      show-back
      :title="selectedRepo?.name || ''"
      @back="selectedCachePath = ''; browserRef?.clearSelection()"
    />
  </div>
</template>

<style scoped>
.cache-view {
  display: flex;
  height: calc(100% + 2 * var(--spacing-lg));
  overflow: hidden;
  margin: calc(-1 * var(--spacing-lg));
}

/* ---- placeholder ---- */
.cache-placeholder {
  flex: 1;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 12px; color: var(--color-text-tertiary); font-size: 14px;
}
.cp-icon { font-size: 48px; opacity: 0.25; }

/* ---- missing path state ---- */
.cache-missing {
  flex: 1;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 10px; padding: 0 32px; text-align: center;
}
.cm-icon { font-size: 48px; color: var(--color-warning); opacity: 0.7; }
.cm-title { font-size: 18px; font-weight: 600; color: var(--color-text); }
.cm-path {
  font-size: 13px; color: var(--color-text-tertiary);
  font-family: 'Cascadia Code', monospace;
  max-width: 500px; word-break: break-all;
}
.cm-hint { font-size: 13px; color: var(--color-text-secondary); max-width: 420px; line-height: 1.6; }
</style>
