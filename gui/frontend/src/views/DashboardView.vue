<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import {
  PlusOutlined,
  MedicineBoxOutlined,
  CloudDownloadOutlined,
  DatabaseOutlined,
  CloudServerOutlined,
  ApiOutlined,
  WarningOutlined,
  LinkOutlined,
} from '@ant-design/icons-vue'
import { useProjectStore } from '../stores/project'

const router = useRouter()
const project = useProjectStore()
const repos = ref([])
const loading = ref(true)

// ---- derived stats from repos + active project ----
const remoteCount = computed(() => repos.value.filter((r) => r.type === 'remote').length)
const localCount = computed(() => repos.value.filter((r) => r.type === 'local').length)
const brokenCount = computed(() => project.activeProject?.brokenCount || 0)
const agentCount = computed(() => project.activeProject?.agents?.length || 0)
const agents = computed(() => project.activeProject?.agents || [])

// overall health derived from individual checks
const healthLabel = computed(() => {
  if (!project.currentExists) return '异常'
  if (brokenCount.value > 0) return '需修复'
  if (repos.value.length === 0) return '空闲'
  return '健康'
})
const healthClass = computed(() => {
  if (!project.currentExists) return 'bad'
  if (brokenCount.value > 0) return 'warn'
  return 'ok'
})

async function loadRepos() {
  if (!project.hasProject) {
    repos.value = []
    loading.value = false
    return
  }
  loading.value = true
  try {
    if (window.go?.main?.ReferenceApp) {
      repos.value = await window.go.main.ReferenceApp.ListRepos()
    }
  } catch (e) {
    console.error('Failed to load repos:', e)
  } finally {
    loading.value = false
  }
}

onMounted(loadRepos)
watch(() => project.projectEpoch, loadRepos)

function agentDisplayName(id) {
  const map = { claude: 'Claude', codex: 'Codex', zcode: 'ZCode', mimocode: 'MiMo', opencode: 'OpenCode' }
  return map[id] || id
}
</script>

<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <h1>{{ project.hasProject ? project.currentName : 'Reference' }}</h1>
      <p class="subtitle">
        {{ project.hasProject ? project.currentDir : '本地代码仓库引用管理器' }}
      </p>
    </div>

    <!-- no project: empty state -->
    <a-empty
      v-if="!project.hasProject"
      description="请从左侧选择一个项目，或添加新项目"
    >
      <template #image><CloudDownloadOutlined style="font-size: 48px; color: var(--color-text-tertiary); opacity: 0.4;" /></template>
    </a-empty>

    <template v-else>
      <!-- project dir missing warning -->
      <a-alert
        v-if="!project.currentExists"
        type="warning"
        show-icon
        message="该项目目录不存在"
        description="目录可能已被移动或删除。数据库记录仍然保留，你可以修复断链或移除该项目。"
        style="margin-bottom: var(--spacing-lg)"
      />

      <!-- stat cards row -->
      <div class="stat-row">
        <div class="stat-card" @click="router.push('/repos')">
          <div class="stat-icon stat-blue"><DatabaseOutlined /></div>
          <div class="stat-body">
            <div class="stat-value">{{ repos.length }}</div>
            <div class="stat-label">引用仓库</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon stat-cyan"><CloudServerOutlined /></div>
          <div class="stat-body">
            <div class="stat-value">{{ remoteCount }} <span class="stat-sub">/ {{ localCount }}</span></div>
            <div class="stat-label">远程 / 本地</div>
          </div>
        </div>
        <div class="stat-card" :class="{ 'stat-warn': brokenCount > 0 }" @click="router.push('/doctor')">
          <div class="stat-icon" :class="brokenCount > 0 ? 'stat-orange' : 'stat-green'">
            <WarningOutlined v-if="brokenCount > 0" />
            <LinkOutlined v-else />
          </div>
          <div class="stat-body">
            <div class="stat-value">{{ brokenCount }}</div>
            <div class="stat-label">{{ brokenCount > 0 ? '断裂链接' : '链接正常' }}</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon stat-purple"><ApiOutlined /></div>
          <div class="stat-body">
            <div class="stat-value">{{ agentCount }}</div>
            <div class="stat-label">AI 助手</div>
          </div>
        </div>
      </div>

      <!-- quick actions -->
      <div class="quick-actions">
        <a-button type="primary" @click="router.push('/repos')">
          <template #icon><PlusOutlined /></template>
          添加仓库
        </a-button>
        <a-button @click="router.push('/repos')">
          <template #icon><CloudDownloadOutlined /></template>
          仓库列表
        </a-button>
        <a-button @click="router.push('/doctor')">
          <template #icon><MedicineBoxOutlined /></template>
          运行诊断
        </a-button>
      </div>

      <!-- info cards: single full-width row -->
      <div class="dash-section">
        <div class="section-head">
          <h3>项目信息</h3>
        </div>
        <div class="info-row">
          <!-- AI assistants -->
          <div class="side-card">
            <div class="side-card-head">
              <span class="side-title"><ApiOutlined /> AI 助手</span>
              <a-button size="small" type="text" class="side-config" @click="router.push('/settings')">配置</a-button>
            </div>
            <div v-if="agents.length > 0" class="agent-grid">
              <div v-for="a in agents" :key="a" class="agent-chip">
                <span class="agent-dot"></span>
                <span class="agent-name">{{ agentDisplayName(a) }}</span>
              </div>
            </div>
            <div v-else class="side-muted">
              <ApiOutlined class="side-muted-icon" />
              <span>未配置 AI 助手</span>
            </div>
          </div>

          <!-- health -->
          <div class="side-card">
            <div class="side-card-head">
              <span class="side-title"><DatabaseOutlined /> 健康状态</span>
              <span class="health-badge" :class="healthClass">{{ healthLabel }}</span>
            </div>
            <div class="check-list">
              <div class="check-row">
                <span :class="['check-dot', project.activeProject?.initialized ? 'ok' : 'off']"></span>
                <span class="check-label">已初始化</span>
              </div>
              <div class="check-row">
                <span :class="['check-dot', project.currentExists ? 'ok' : 'bad']"></span>
                <span class="check-label">目录存在</span>
              </div>
              <div class="check-row">
                <span :class="['check-dot', brokenCount > 0 ? 'warn' : 'ok']"></span>
                <span class="check-label">{{ brokenCount > 0 ? `${brokenCount} 个断链` : '链接完整' }}</span>
              </div>
              <div class="check-row">
                <span :class="['check-dot', repos.length > 0 ? 'ok' : 'off']"></span>
                <span class="check-label">{{ repos.length > 0 ? `${repos.length} 个引用` : '暂无引用' }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- repo grid: full width -->
      <a-spin :spinning="loading" class="dash-section">
        <div class="section-head">
          <h3>引用仓库</h3>
          <span class="section-count">{{ repos.length }}</span>
        </div>

        <div v-if="repos.length > 0" class="repo-grid">
          <div v-for="repo in repos" :key="repo.name" class="repo-card" :class="'type-' + repo.type">
            <div class="repo-type-bar"></div>
            <div class="repo-card-inner">
              <div class="repo-card-header">
                <span class="repo-name">{{ repo.name }}</span>
                <a-tag :color="repo.type === 'remote' ? 'blue' : 'green'" class="repo-tag">
                  {{ repo.type === 'remote' ? '远程' : '本地' }}
                </a-tag>
              </div>
              <div class="repo-card-body">
                <div class="repo-source" :title="repo.source">{{ repo.source }}</div>
                <div class="repo-meta">
                  <span v-if="repo.branch"><LinkOutlined /> {{ repo.branch }}</span>
                  <span v-if="repo.commit_at">· {{ repo.commit_at }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        <a-empty v-else-if="!loading" description="暂无引用仓库">
          <a-button type="primary" @click="router.push('/repos')">添加第一个仓库</a-button>
        </a-empty>
      </a-spin>
    </template>
  </div>
</template>

<style scoped>
.dashboard {
  width: 100%;
}

.dashboard-header {
  margin-bottom: var(--spacing-lg);
}
.dashboard-header h1 {
  font-size: 24px;
  font-weight: 700;
  color: var(--color-text);
  margin-bottom: var(--spacing-xs);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.subtitle {
  color: var(--color-text-tertiary);
  font-size: 12px;
  font-family: 'Cascadia Code', monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ---- stat cards ---- */
.stat-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-lg);
}
.stat-card {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-md);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
  cursor: pointer;
}
.stat-card:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-sm);
  transform: translateY(-1px);
}
.stat-card.stat-warn {
  border-color: var(--color-warning);
}
.stat-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  height: 42px;
  border-radius: var(--radius-md);
  font-size: 20px;
  flex-shrink: 0;
}
.stat-blue   { background: rgba(59, 130, 246, 0.12); color: #3b82f6; }
.stat-cyan   { background: rgba(6, 182, 212, 0.12); color: #06b6d4; }
.stat-green  { background: rgba(22, 163, 74, 0.12); color: var(--color-success); }
.stat-orange { background: rgba(217, 119, 6, 0.12); color: var(--color-warning); }
.stat-purple { background: rgba(168, 85, 247, 0.12); color: #a855f7; }
.stat-body { min-width: 0; }
.stat-value {
  font-size: 22px;
  font-weight: 700;
  color: var(--color-text);
  line-height: 1.1;
}
.stat-sub {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text-tertiary);
}
.stat-label {
  font-size: 12px;
  color: var(--color-text-secondary);
  margin-top: 2px;
}

/* ---- quick actions ---- */
.quick-actions {
  display: flex;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-lg);
}

/* ---- stacked full-width sections ---- */
.dash-section {
  display: block;
  margin-bottom: var(--spacing-lg);
}

/* info cards in a single row */
.info-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--spacing-md);
}

.section-head {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-md);
}
.section-head h3 {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
  margin: 0;
}
.section-count {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-tertiary);
  background: var(--color-surface-raised);
  padding: 1px 8px;
  border-radius: 999px;
}

/* ---- repo cards ---- */
.repo-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: var(--spacing-md);
}
.repo-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  overflow: hidden;
  transition: all var(--transition-fast);
  cursor: pointer;
}
.repo-card:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-md);
  transform: translateY(-1px);
}
.repo-type-bar {
  height: 3px;
}
.type-remote .repo-type-bar { background: #3b82f6; }
.type-local .repo-type-bar { background: var(--color-success); }
.repo-card-inner { padding: var(--spacing-md); }
.repo-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-sm);
}
.repo-name {
  font-weight: 600;
  font-size: 15px;
  color: var(--color-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.repo-tag { margin: 0; flex-shrink: 0; }
.repo-card-body {
  font-size: 13px;
  color: var(--color-text-secondary);
}
.repo-source {
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.repo-meta {
  display: flex;
  gap: 6px;
  align-items: center;
  font-size: 12px;
  color: var(--color-text-tertiary);
}

/* ---- sidebar cards ---- */
.side-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: 14px 16px;
}
.side-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}
.side-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}
.side-title .anticon { color: var(--color-text-tertiary); font-size: 15px; }
.side-config {
  padding: 0 4px;
  height: auto;
  font-size: 12px;
  color: var(--color-text-tertiary);
}
.side-config:hover { color: var(--color-primary); }

/* agent chips */
.agent-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.agent-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px 4px 10px;
  border-radius: 999px;
  background: rgba(168, 85, 247, 0.1);
  border: 1px solid rgba(168, 85, 247, 0.2);
  font-size: 12.5px;
  font-weight: 500;
  color: #a855f7;
  transition: all var(--transition-fast);
}
.agent-chip:hover {
  background: rgba(168, 85, 247, 0.15);
}
.agent-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #a855f7;
  flex-shrink: 0;
}
.side-muted {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--color-text-tertiary);
}
.side-muted-icon {
  font-size: 16px;
  opacity: 0.5;
}

/* health badge + checks */
.health-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 10px;
  border-radius: 999px;
}
.health-badge.ok   { background: var(--color-success-bg); color: var(--color-success); }
.health-badge.warn { background: var(--color-warning-bg); color: var(--color-warning); }
.health-badge.bad  { background: var(--color-error-bg); color: var(--color-error); }

.check-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.check-row {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
}
.check-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.check-dot.ok   { background: var(--color-success); box-shadow: 0 0 0 3px var(--color-success-bg); }
.check-dot.warn { background: var(--color-warning); box-shadow: 0 0 0 3px var(--color-warning-bg); }
.check-dot.bad  { background: var(--color-error); box-shadow: 0 0 0 3px var(--color-error-bg); }
.check-dot.off  { background: var(--color-text-tertiary); box-shadow: 0 0 0 3px var(--color-surface-raised); opacity: 0.4; }
.check-label {
  color: var(--color-text-secondary);
}
</style>
