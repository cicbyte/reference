<script setup>
import { ref, computed, onMounted } from 'vue'
import {
  FolderOutlined, DatabaseOutlined, CloudServerOutlined,
  ReadOutlined, HddOutlined, WarningOutlined,
} from '@ant-design/icons-vue'
import AiSparkIcon from '../components/common/AiSparkIcon.vue'

const app = window.go?.main?.ReferenceApp

// ---- raw data (each loads independently) ----
const stats = ref(null)
const projects = ref([])
const repos = ref([])
const wikiEntries = ref([])
const cacheSize = ref(null)
const wikiSize = ref(null)

const statsLoading = ref(true)
const projectsLoading = ref(true)
const reposLoading = ref(true)
const wikiLoading = ref(true)

function fmtSize(bytes) {
  if (bytes == null) return '...'
  if (!bytes) return '0 B'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + ' MB'
  return (bytes / 1024 / 1024 / 1024).toFixed(2) + ' GB'
}

const agentNameMap = { claude: 'Claude', codex: 'Codex', zcode: 'ZCode', mimocode: 'MiMo', opencode: 'OpenCode' }

// ---- ② platform distribution ----
const platformDist = computed(() => {
  const groups = {}
  for (const r of repos.value) {
    const key = r.type === 'local' ? '本地仓库' : (r.host || '未知')
    groups[key] = (groups[key] || 0) + 1
  }
  const entries = Object.entries(groups).sort((a, b) => b[1] - a[1])
  const max = Math.max(...entries.map((e) => e[1]), 1)
  return entries.map(([name, count]) => ({ name, count, pct: (count / max) * 100 }))
})

// ---- ③ AI assistant usage ----
const agentUsage = computed(() => {
  const counts = {}
  for (const p of projects.value) {
    for (const a of (p.agents || [])) {
      counts[a] = (counts[a] || 0) + 1
    }
  }
  const entries = Object.entries(counts).sort((a, b) => b[1] - a[1])
  const max = Math.max(...entries.map((e) => e[1]), 1)
  return entries.map(([id, count]) => ({ name: agentNameMap[id] || id, count, pct: (count / max) * 100 }))
})

// ---- ④ top 5 projects by repo count ----
const topProjects = computed(() => {
  return [...projects.value]
    .sort((a, b) => (b.repoCount || 0) - (a.repoCount || 0))
    .slice(0, 5)
    .map((p) => ({ name: p.name, count: p.repoCount || 0 }))
    .filter((p) => p.count > 0)
})
const topProjectMax = computed(() => Math.max(...topProjects.value.map((p) => p.count), 1))

// ---- ⑤ wiki overview ----
const wikiStats = computed(() => {
  const remote = wikiEntries.value.filter((e) => e.source === 'remote')
  const local = wikiEntries.value.filter((e) => e.source === 'local')
  const platforms = new Set()
  const repoSet = new Set()
  for (const e of wikiEntries.value) {
    if (e.platform) platforms.add(e.platform)
    if (e.repoName) repoSet.add(e.repoName)
  }
  return {
    remoteCount: remote.length,
    localCount: local.length,
    platformCount: platforms.size,
    repoCount: repoSet.size,
    platforms: [...platforms].sort().slice(0, 6),
  }
})

// ---- ⑥ top 5 cache by size (computed server-side in parallel) ----
const topCache = ref([])
const topCacheLoading = ref(true)
async function loadCacheTop() {
  try {
    const items = await app?.GetCacheTopN?.(5)
    if (items && items.length) {
      const max = Math.max(...items.map((r) => r.size), 1)
      topCache.value = items.map((r) => ({ name: r.name, size: r.size, pct: (r.size / max) * 100 }))
    }
  } catch { /* ignore */ }
  finally { topCacheLoading.value = false }
}

// ---- load all in parallel ----
onMounted(async () => {
  const tasks = []

  // stats
  tasks.push(
    app?.GlobalStats?.().then((s) => { stats.value = s }).catch(() => {}).finally(() => { statsLoading.value = false })
  )
  // projects
  tasks.push(
    app?.ListProjects?.().then((p) => { projects.value = p }).catch(() => {}).finally(() => { projectsLoading.value = false })
  )
  // repos
  tasks.push(
    app?.ListCachedRepos?.().then((r) => { repos.value = r }).catch(() => {}).finally(() => { reposLoading.value = false })
  )
  // cache top N (parallel server-side)
  tasks.push(loadCacheTop())
  // wiki
  tasks.push(
    app?.ListWikiEntries?.('all').then((w) => { wikiEntries.value = w }).catch(() => {}).finally(() => { wikiLoading.value = false })
  )

  await Promise.allSettled(tasks)

  // async dir sizes
  if (stats.value?.repos_dir) app?.GetDirSizeAsync?.(stats.value.repos_dir).then((s) => { cacheSize.value = s }).catch(() => { cacheSize.value = 0 })
  if (stats.value?.wiki_dir) app?.GetDirSizeAsync?.(stats.value.wiki_dir).then((s) => { wikiSize.value = s }).catch(() => { wikiSize.value = 0 })
})
</script>

<template>
  <div class="stats-view">
    <div class="page-header"><h2>全局统计</h2></div>

    <!-- ① overview cards -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="sc-icon sc-blue"><FolderOutlined /></div>
        <div class="sc-body">
          <div class="sc-value">{{ stats?.total_projects ?? '...' }}</div>
          <div class="sc-label">项目总数</div>
        </div>
        <div class="sc-extra" v-if="stats?.deleted_projects > 0">
          <span class="extra-warn"><WarningOutlined /> {{ stats.deleted_projects }} 失效</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="sc-icon sc-cyan"><CloudServerOutlined /></div>
        <div class="sc-body">
          <div class="sc-value">{{ stats?.total_repos ?? '...' }}</div>
          <div class="sc-label">缓存仓库</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="sc-icon sc-purple"><HddOutlined /></div>
        <div class="sc-body">
          <div class="sc-value">{{ fmtSize(cacheSize) }}</div>
          <div class="sc-label">仓库缓存</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="sc-icon sc-green"><ReadOutlined /></div>
        <div class="sc-body">
          <div class="sc-value">{{ fmtSize(wikiSize) }}</div>
          <div class="sc-label">知识库</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="sc-icon sc-orange"><DatabaseOutlined /></div>
        <div class="sc-body">
          <div class="sc-value">{{ fmtSize(stats?.db_size) }}</div>
          <div class="sc-label">数据库</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="sc-icon sc-blue"><FolderOutlined /></div>
        <div class="sc-body">
          <div class="sc-value">{{ stats?.existing_projects ?? '...' }}</div>
          <div class="sc-label">活跃项目</div>
        </div>
      </div>
    </div>

    <!-- ②③ platform + agent distribution -->
    <div class="dual-row">
      <div class="panel">
        <div class="panel-head"><CloudServerOutlined /> 仓库平台分布</div>
        <a-spin v-if="reposLoading" size="small" />
        <div v-else-if="platformDist.length" class="bar-list">
          <div v-for="item in platformDist" :key="item.name" class="bar-row">
            <span class="bar-label">{{ item.name }}</span>
            <div class="bar-track">
              <div class="bar-fill sc-cyan" :style="{ width: item.pct + '%' }"></div>
            </div>
            <span class="bar-count">{{ item.count }}</span>
          </div>
        </div>
        <div v-else class="panel-empty">暂无数据</div>
      </div>

      <div class="panel">
        <div class="panel-head"><AiSparkIcon :size="16" /> AI 助手使用分布</div>
        <a-spin v-if="projectsLoading" size="small" />
        <div v-else-if="agentUsage.length" class="bar-list">
          <div v-for="item in agentUsage" :key="item.name" class="bar-row">
            <span class="bar-label">{{ item.name }}</span>
            <div class="bar-track">
              <div class="bar-fill sc-purple" :style="{ width: item.pct + '%' }"></div>
            </div>
            <span class="bar-count">{{ item.count }} 项目</span>
          </div>
        </div>
        <div v-else class="panel-empty">暂无配置</div>
      </div>
    </div>

    <!-- ④⑤ top projects + wiki overview -->
    <div class="dual-row">
      <div class="panel">
        <div class="panel-head"><FolderOutlined /> 引用数 Top 5 项目</div>
        <a-spin v-if="projectsLoading" size="small" />
        <div v-else-if="topProjects.length" class="bar-list">
          <div v-for="item in topProjects" :key="item.name" class="bar-row">
            <span class="bar-label">{{ item.name }}</span>
            <div class="bar-track">
              <div class="bar-fill sc-blue" :style="{ width: (item.count / topProjectMax * 100) + '%' }"></div>
            </div>
            <span class="bar-count">{{ item.count }}</span>
          </div>
        </div>
        <div v-else class="panel-empty">暂无引用</div>
      </div>

      <div class="panel">
        <div class="panel-head"><ReadOutlined /> 知识库概况</div>
        <a-spin v-if="wikiLoading" size="small" />
        <div v-else class="kv-list">
          <div class="kv-row"><span>远程知识文件</span><span class="kv-val">{{ wikiStats.remoteCount }}</span></div>
          <div class="kv-row"><span>本地知识文件</span><span class="kv-val">{{ wikiStats.localCount }}</span></div>
          <div class="kv-row"><span>覆盖仓库数</span><span class="kv-val">{{ wikiStats.repoCount }}</span></div>
          <div class="kv-row"><span>平台数</span><span class="kv-val">{{ wikiStats.platformCount }}</span></div>
          <div class="kv-platforms" v-if="wikiStats.platforms.length">
            <span class="kv-platform-label">平台:</span>
            <span v-for="p in wikiStats.platforms" :key="p" class="kv-platform-chip">{{ p }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- ⑥ cache top 5 -->
    <div class="panel full">
      <div class="panel-head"><HddOutlined /> 缓存占用 Top 5</div>
      <a-spin v-if="topCacheLoading" size="small" />
      <div v-else-if="topCache.length" class="bar-list">
        <div v-for="item in topCache" :key="item.name" class="bar-row">
          <span class="bar-label">{{ item.name }}</span>
          <div class="bar-track">
            <div class="bar-fill sc-orange" :style="{ width: item.pct + '%' }"></div>
          </div>
          <span class="bar-count">{{ fmtSize(item.size) }}</span>
        </div>
      </div>
      <div v-else class="panel-empty">暂无缓存数据</div>
    </div>
  </div>
</template>

<style scoped>
.stats-view { width: 100%; display: flex; flex-direction: column; gap: var(--spacing-lg); }
.page-header h2 { font-size: 20px; font-weight: 600; color: var(--color-text); margin: 0; }

/* ① overview cards */
.stats-grid {
  display: grid; grid-template-columns: repeat(auto-fill, minmax(180px, 1fr)); gap: var(--spacing-md);
}
.stat-card {
  display: flex; align-items: center; gap: var(--spacing-md);
  padding: 16px 18px; background: var(--color-surface);
  border: 1px solid var(--color-border); border-radius: var(--radius-md);
  transition: all var(--transition-fast); position: relative;
}
.stat-card:hover { border-color: var(--color-primary); box-shadow: var(--shadow-sm); transform: translateY(-1px); }
.sc-icon {
  display: flex; align-items: center; justify-content: center;
  width: 42px; height: 42px; border-radius: var(--radius-md); font-size: 20px; flex-shrink: 0;
}
.sc-blue { background: rgba(59,130,246,0.12); color: #3b82f6; }
.sc-cyan { background: rgba(6,182,212,0.12); color: #06b6d4; }
.sc-green { background: rgba(22,163,74,0.12); color: var(--color-success); }
.sc-purple { background: rgba(168,85,247,0.12); color: #a855f7; }
.sc-orange { background: rgba(249,115,22,0.12); color: #f97316; }
.sc-body { flex: 1; min-width: 0; }
.sc-value { font-size: 22px; font-weight: 700; color: var(--color-text); line-height: 1.1; }
.sc-label { font-size: 12px; color: var(--color-text-secondary); margin-top: 2px; }
.sc-extra { position: absolute; top: 6px; right: 10px; }
.extra-warn { font-size: 11px; color: var(--color-warning); display: inline-flex; align-items: center; gap: 3px; }

/* ②③④⑤ panels */
.dual-row { display: grid; grid-template-columns: 1fr 1fr; gap: var(--spacing-md); }
@media (max-width: 900px) { .dual-row { grid-template-columns: 1fr; } }
.panel {
  background: var(--color-surface); border: 1px solid var(--color-border);
  border-radius: var(--radius-md); padding: 16px 18px;
}
.panel.full { width: 100%; }
.panel-head {
  display: flex; align-items: center; gap: 8px;
  font-size: 14px; font-weight: 600; color: var(--color-text); margin-bottom: 14px;
}
.panel-empty { padding: 20px; text-align: center; font-size: 13px; color: var(--color-text-tertiary); }

/* bar chart rows */
.bar-list { display: flex; flex-direction: column; gap: 10px; }
.bar-row { display: flex; align-items: center; gap: 10px; }
.bar-label {
  width: 100px; flex-shrink: 0; font-size: 13px; color: var(--color-text);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.bar-track { flex: 1; height: 8px; background: var(--color-surface-raised); border-radius: 4px; overflow: hidden; }
.bar-fill {
  height: 100%; border-radius: 4px;
  background: var(--color-primary);
  transition: width 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}
.bar-fill.sc-cyan { background: #06b6d4; }
.bar-fill.sc-purple { background: #a855f7; }
.bar-fill.sc-blue { background: #3b82f6; }
.bar-fill.sc-orange { background: #f97316; }
.bar-count { width: 60px; flex-shrink: 0; text-align: right; font-size: 12px; font-weight: 600; color: var(--color-text-secondary); }

/* kv list (wiki overview) */
.kv-list { display: flex; flex-direction: column; gap: 8px; }
.kv-row {
  display: flex; justify-content: space-between; align-items: center;
  font-size: 13px; color: var(--color-text-secondary);
}
.kv-val { font-size: 18px; font-weight: 700; color: var(--color-text); }
.kv-platforms { display: flex; flex-wrap: wrap; align-items: center; gap: 4px; margin-top: 4px; }
.kv-platform-label { font-size: 12px; color: var(--color-text-tertiary); margin-right: 4px; }
.kv-platform-chip {
  font-size: 10px; font-weight: 600; padding: 1px 7px;
  border-radius: 999px; background: var(--color-surface-raised); color: var(--color-text-secondary);
}
</style>
