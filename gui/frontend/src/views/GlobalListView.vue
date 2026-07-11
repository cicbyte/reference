<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import {
  FolderOutlined,
  FolderOpenOutlined,
  WarningOutlined,
  ReloadOutlined,
  DeleteOutlined,
  CaretRightFilled,
  SwapOutlined,
  MedicineBoxOutlined,
  CopyOutlined,
} from '@ant-design/icons-vue'
import { useProjectStore } from '../stores/project'
import DiagnoseModal from '../components/repo/DiagnoseModal.vue'

const router = useRouter()
const project = useProjectStore()
const loading = ref(true)
const projects = ref([])
const diagnoseOpen = ref(false)
const diagnoseDir = ref('')

const app = window.go?.main?.ReferenceApp

// ---- grouping by parent directory ----
const expandedGroups = ref(new Set())
const groupedProjects = computed(() => {
  const groups = {}
  for (const p of projects.value) {
    // parent dir = group key (normalize separators)
    const parts = p.dir.replace(/\//g, '\\').split('\\')
    parts.pop()
    const key = parts.join('\\') || '/'
    if (!groups[key]) groups[key] = []
    groups[key].push(p)
  }
  const keys = Object.keys(groups).sort()
  return keys.map((key) => ({ key, projects: groups[key] }))
})

function toggleGroup(key) {
  const next = new Set(expandedGroups.value)
  next.has(key) ? next.delete(key) : next.add(key)
  expandedGroups.value = next
}

// auto-expand all on first load
watch(groupedProjects, (groups) => {
  if (expandedGroups.value.size === 0 && groups.length) {
    expandedGroups.value = new Set(groups.map((g) => g.key))
  }
}, { once: true })

const totalRepos = computed(() =>
  projects.value.reduce((s, p) => s + (p.repoCount || 0), 0),
)
const missingProjects = computed(() => projects.value.filter((p) => !p.exists))

const brokenProjects = computed(() =>
  projects.value.filter((p) => !p.exists || p.brokenCount > 0).length,
)

async function loadProjects() {
  loading.value = true
  try {
    if (app?.ListProjects) {
      projects.value = await app.ListProjects()
    }
  } catch (e) {
    message.error('加载失败: ' + e)
  } finally {
    loading.value = false
  }
}

onMounted(loadProjects)

function cleanupMissing() {
  const missing = missingProjects.value
  if (!missing.length) return
  const names = missing.map((p) => p.name).join('、')
  Modal.confirm({
    title: `清理 ${missing.length} 个失效项目`,
    content: `以下项目目录已不存在:${names}。将从数据库中移除这些项目的引用记录。`,
    okText: '清理',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      let ok = 0
      for (const p of missing) {
        try {
          await app.RemoveProject(p.dir, false)
          ok++
        } catch { /* skip */ }
      }
      message.success(`已清理 ${ok} 个失效项目`)
      await loadProjects()
      await project.loadProjects()
    },
  })
}

function switchTo(p) {
  project.switchTo(p.dir).then(() => {
    message.success(`已切换到 ${p.name}`)
    router.push('/')
  })
}

function onDoctor(p) {
  diagnoseDir.value = p.dir
  diagnoseOpen.value = true
}

async function onCopyPath(p) {
  try {
    await app.CopyPath(p.dir)
    message.success('路径已复制')
  } catch (e) {
    message.error('复制失败: ' + e)
  }
}

function onRemove(p, clean) {
  const label = clean ? '移除并清除 .reference' : '移除项目'
  const content = clean
    ? `将删除 ${p.name} 的所有引用记录、.reference 目录及注入的 AI 配置文件。`
    : `将删除 ${p.name} 的所有引用记录和链接,保留 .reference 目录。`
  Modal.confirm({
    title: `${label} — ${p.name}`,
    content,
    okText: label,
    okType: clean ? 'danger' : 'primary',
    cancelText: '取消',
    async onOk() {
      try {
        await app.RemoveProject(p.dir, clean)
        message.success(`${label}成功`)
        await loadProjects()
        await project.loadProjects()
      } catch (e) {
        message.error(`${label}失败: ` + e)
      }
    },
  })
}

function agentDisplayName(id) {
  const map = { claude: 'Claude', codex: 'Codex', zcode: 'ZCode', mimocode: 'MiMo', opencode: 'OpenCode' }
  return map[id] || id
}
</script>

<template>
  <div class="global-list">
    <div class="page-header">
      <h2>项目列表</h2>
      <div class="header-actions">
        <a-button v-if="missingProjects.length > 0" danger @click="cleanupMissing">
          <template #icon><DeleteOutlined /></template>
          清理失效项目 ({{ missingProjects.length }})
        </a-button>
        <a-button @click="loadProjects" :loading="loading">
          <template #icon><ReloadOutlined /></template>
          刷新
        </a-button>
      </div>
    </div>

    <!-- summary -->
    <div class="summary-strip">
      <div class="sum-item">
        <FolderOutlined class="sum-icon" />
        <span class="sum-val">{{ projects.length }}</span>
        <span class="sum-lbl">个项目</span>
      </div>
      <div class="sum-sep"></div>
      <div class="sum-item">
        <span class="sum-val">{{ totalRepos }}</span>
        <span class="sum-lbl">个引用</span>
      </div>
      <div class="sum-sep"></div>
      <div class="sum-item" v-if="brokenProjects > 0">
        <WarningOutlined class="sum-icon warn" />
        <span class="sum-val warn">{{ brokenProjects }}</span>
        <span class="sum-lbl">需关注</span>
      </div>
    </div>

    <a-spin :spinning="loading">
      <!-- grouped project cards -->
      <div v-if="projects.length > 0">
        <template v-for="group in groupedProjects" :key="group.key">
          <div class="group-head" @click="toggleGroup(group.key)">
            <CaretRightFilled class="group-caret" :class="{ open: expandedGroups.has(group.key) }" />
            <span class="group-label">{{ group.key }}</span>
            <span class="group-count">{{ group.projects.length }}</span>
          </div>

          <div v-if="expandedGroups.has(group.key)" class="project-grid">
            <a-dropdown
              v-for="p in group.projects"
              :key="p.dir"
              :trigger="['contextmenu']"
            >
              <div
                class="project-card"
                :class="{ 'card-warn': !p.exists || p.brokenCount > 0, 'card-missing': !p.exists }"
              >
                <div class="card-bar" :class="!p.exists ? 'bar-red' : (p.brokenCount > 0 ? 'bar-orange' : 'bar-green')"></div>
                <div class="card-body" @click="p.exists && switchTo(p)">
                  <div class="card-head">
                    <div class="card-icon" :class="!p.exists ? 'icon-missing' : ''">
                      <WarningOutlined v-if="!p.exists" />
                      <FolderOutlined v-else />
                    </div>
                    <div class="card-title-area">
                      <div class="card-title">
                        {{ p.name }}
                        <span v-if="!p.exists" class="missing-tag">目录不存在</span>
                      </div>
                      <div class="card-dir" :title="p.dir">{{ p.dir }}</div>
                    </div>
                  </div>

                  <div class="card-stats">
                    <div class="stat">
                      <span class="stat-num">{{ p.repoCount }}</span>
                      <span class="stat-label">引用</span>
                    </div>
                    <div class="stat" v-if="p.brokenCount > 0">
                      <span class="stat-num warn">{{ p.brokenCount }}</span>
                      <span class="stat-label">断链</span>
                    </div>
                    <div class="stat" v-if="p.agents && p.agents.length">
                      <span class="stat-num">{{ p.agents.length }}</span>
                      <span class="stat-label">助手</span>
                    </div>
                  </div>

                  <div class="card-agents" v-if="p.agents && p.agents.length">
                    <span v-for="a in p.agents" :key="a" class="agent-chip">{{ agentDisplayName(a) }}</span>
                  </div>
                </div>
              </div>
              <template #overlay>
                <a-menu>
                  <a-menu-item v-if="p.exists" @click="switchTo(p)">
                    <SwapOutlined /> 切换到此项目
                  </a-menu-item>
                  <a-menu-divider v-if="p.exists" />
                  <a-menu-item v-if="p.exists" @click="onDoctor(p)">
                    <MedicineBoxOutlined /> 修复断裂链接
                  </a-menu-item>
                  <a-menu-item v-if="p.exists" @click="app?.OpenInExplorer(p.dir)">
                    <FolderOpenOutlined /> 在文件管理器中打开
                  </a-menu-item>
                  <a-menu-item @click="onCopyPath(p)">
                    <CopyOutlined /> 复制路径
                  </a-menu-item>
                  <a-menu-divider />
                  <a-menu-item danger @click="onRemove(p, false)">
                    <DeleteOutlined /> 移除项目
                  </a-menu-item>
                  <a-menu-item danger @click="onRemove(p, true)">
                    <DeleteOutlined /> 移除并清除 .reference
                  </a-menu-item>
                </a-menu>
              </template>
            </a-dropdown>
          </div>
        </template>
      </div>

      <a-empty v-if="!loading && projects.length === 0" description="暂无项目">
        <template #image><FolderOutlined style="font-size: 48px; color: var(--color-text-tertiary); opacity: 0.3;" /></template>
      </a-empty>
    </a-spin>

    <DiagnoseModal v-model:open="diagnoseOpen" :project-dir="diagnoseDir" @update:open="diagnoseOpen = $event; loadProjects()" />
  </div>
</template>

<style scoped>
.global-list { width: 100%; }

.page-header {
  display: flex; justify-content: space-between; align-items: center;
  margin-bottom: var(--spacing-lg);
}
.page-header h2 { font-size: 20px; font-weight: 600; color: var(--color-text); margin: 0; }
.header-actions { display: flex; gap: var(--spacing-sm); }

/* summary */
.summary-strip {
  display: flex; align-items: center; gap: var(--spacing-md);
  padding: 14px 18px;
  background: var(--color-surface); border: 1px solid var(--color-border);
  border-radius: var(--radius-md); margin-bottom: var(--spacing-lg);
}
.sum-item { display: flex; align-items: center; gap: 8px; }
.sum-icon { font-size: 16px; color: var(--color-text-tertiary); }
.sum-icon.warn { color: var(--color-warning); }
.sum-val { font-size: 20px; font-weight: 700; color: var(--color-text); }
.sum-val.warn { color: var(--color-warning); }
.sum-lbl { font-size: 12px; color: var(--color-text-tertiary); }
.sum-sep { width: 1px; height: 28px; background: var(--color-border); }

/* group headers */
.group-head {
  display: flex; align-items: center; gap: 6px;
  padding: 10px 4px 6px; cursor: pointer; user-select: none;
}
.group-label {
  flex: 1; font-size: 12px; font-weight: 600;
  color: var(--color-text-secondary);
  font-family: 'Cascadia Code', monospace;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.group-count {
  font-size: 10px; font-weight: 600; color: var(--color-text-tertiary);
  background: var(--color-surface-raised); padding: 0 6px; border-radius: 999px;
}
.group-caret {
  font-size: 9px; color: var(--color-text-tertiary); flex-shrink: 0;
  transition: transform var(--transition-fast);
}
.group-caret.open { transform: rotate(90deg); }

/* project cards */
.project-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: var(--spacing-md);
}
.project-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  overflow: hidden;
  transition: all var(--transition-fast);
  display: flex;
}
.project-card:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-md);
  transform: translateY(-1px);
}
.card-warn { border-color: var(--color-warning); }
.card-missing { opacity: 0.75; }
.card-missing .card-title-area .card-title { color: var(--color-text-secondary); }
.missing-tag {
  font-size: 10px; font-weight: 600; padding: 1px 6px; margin-left: 6px;
  border-radius: 3px; background: var(--color-error-bg); color: var(--color-error);
  vertical-align: middle;
}

.card-bar { width: 3px; flex-shrink: 0; }
.bar-green { background: var(--color-success); }
.bar-orange { background: var(--color-warning); }
.bar-red { background: var(--color-error); }

.card-body {
  flex: 1; padding: 14px 16px; cursor: pointer;
}
.card-head { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.card-icon {
  display: flex; align-items: center; justify-content: center;
  width: 36px; height: 36px; border-radius: var(--radius-md);
  background: var(--color-primary-bg); color: var(--color-primary);
  font-size: 16px; flex-shrink: 0;
}
.card-icon.icon-missing { background: var(--color-warning-bg); color: var(--color-warning); }
.card-title-area { flex: 1; min-width: 0; }
.card-title {
  font-size: 15px; font-weight: 600; color: var(--color-text);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.card-dir {
  font-size: 11px; color: var(--color-text-tertiary);
  font-family: 'Cascadia Code', monospace;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}

.card-stats { display: flex; gap: 16px; margin-bottom: 8px; }
.stat { display: flex; flex-direction: column; }
.stat-num { font-size: 18px; font-weight: 700; color: var(--color-text); line-height: 1.2; }
.stat-num.warn { color: var(--color-warning); }
.stat-label { font-size: 11px; color: var(--color-text-tertiary); }

.card-agents { display: flex; flex-wrap: wrap; gap: 4px; }
.agent-chip {
  font-size: 10px; font-weight: 600; padding: 1px 7px;
  border-radius: 999px; background: rgba(168, 85, 247, 0.1);
  color: #a855f7; border: 1px solid rgba(168, 85, 247, 0.2);
}

/* card actions */
.card-actions {
  display: flex; flex-direction: column; gap: 2px;
  padding: 8px 6px; border-left: 1px solid var(--color-border-light);
  flex-shrink: 0;
}
.card-btn {
  display: flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; border: none; background: transparent;
  color: var(--color-text-tertiary); cursor: pointer;
  border-radius: var(--radius-xs); transition: all var(--transition-fast);
}
.card-btn:hover { background: var(--color-hover); color: var(--color-primary); }
</style>
