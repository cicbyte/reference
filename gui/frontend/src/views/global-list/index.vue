<script setup>
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
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
import { useI18n } from 'vue-i18n'
import { useProjectStore } from '@/stores/project'
import DiagnoseModal from '@/components/repo/DiagnoseModal.vue'
import { formatPath } from '@/utils/path'
import { agentDisplayName } from '@/utils/agents'
import { useProjectActions } from '@/composables/useProjectActions'

const { t } = useI18n()
const router = useRouter()
const project = useProjectStore()
const loading = ref(true)
const projects = ref([])

const app = window.go?.main?.ReferenceApp

async function loadProjects() {
  loading.value = true
  try {
    if (app?.ListProjects) {
      projects.value = await app.ListProjects()
    }
  } catch (e) {
    message.error(t('wiki.loadFailed') + ': ' + e)
  } finally {
    loading.value = false
  }
}

const {
  diagnoseOpen,
  diagnoseDir,
  onDoctor,
  onCopyPath,
  onRemove,
} = useProjectActions(loadProjects)

// ---- grouping by parent directory ----
const expandedGroups = ref(new Set())
const groupedProjects = computed(() => {
  const groups = {}
  for (const p of projects.value) {
    // parent dir = group key (normalize to forward slashes)
    const parts = formatPath(p.dir).split('/').filter(Boolean)
    parts.pop()
    const key = parts.join('/') || '/'
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

const totalRepos = computed(() =>
  projects.value.reduce((s, p) => s + (p.repoCount || 0), 0),
)
const missingProjects = computed(() => projects.value.filter((p) => !p.exists))

const brokenProjects = computed(() =>
  projects.value.filter((p) => !p.exists || p.brokenCount > 0).length,
)

// ---- cleanup modal: pick which invalid projects to remove ----
const cleanupOpen = ref(false)
const cleanupSelected = ref([])

function openCleanup() {
  cleanupSelected.value = missingProjects.value.map((p) => p.dir)
  cleanupOpen.value = true
}
async function doCleanup() {
  const dirs = cleanupSelected.value
  if (!dirs.length) return
  let ok = 0
  for (const dir of dirs) {
    try {
      await app.RemoveProject(dir, false)
      ok++
    } catch { /* skip */ }
  }
  cleanupOpen.value = false
  message.success(t('globalList.cleanupSuccess', { n: ok }))
  await loadProjects()
  await project.loadProjects()
}

// ---- scroll anchoring: left directory rail ↔ right project groups ----
const scrollRef = ref(null)
const groupEls = ref({})
const activeGroup = ref('')

function setGroupEl(key, el) {
  if (el) groupEls.value[key] = el
}
/** 缩短目录显示：取最后两段（title 仍给全路径） */
function shortDir(key) {
  if (key === '/' || !key) return '/'
  const parts = key.split('/').filter(Boolean)
  return parts.slice(-2).join('/') || key
}
function scrollToGroup(key) {
  const container = scrollRef.value
  const el = groupEls.value[key]
  if (!container || !el) return
  container.scrollTo({ top: el.offsetTop - 8, behavior: 'smooth' })
}
function onMainScroll() {
  const container = scrollRef.value
  if (!container || !groupedProjects.value.length) return
  // 找当前滚到顶部之上的最后一个分组（即视口顶部可见的分组）
  let current = groupedProjects.value[0].key
  for (const g of groupedProjects.value) {
    const el = groupEls.value[g.key]
    if (el && el.offsetTop <= container.scrollTop + 12) current = g.key
  }
  activeGroup.value = current
}

watch(groupedProjects, (groups) => {
  if (!groups.length) return
  // auto-expand all groups on first load
  if (expandedGroups.value.size === 0) {
    expandedGroups.value = new Set(groups.map((g) => g.key))
  }
  nextTick(() => onMainScroll())
})

onMounted(loadProjects)

function switchTo(p) {
  project.switchTo(p.dir).then(() => {
    message.success(t('globalList.switchTo', { name: p.name }))
    router.push('/')
  })
}
</script>

<template>
  <div class="global-list">
    <!-- left rail: directory list -->
    <aside class="list-rail">
      <div class="rail-head">
        <span class="rail-title">{{ t('globalList.directory') }}</span>
        <span class="rail-count">{{ groupedProjects.length }}</span>
      </div>
      <div class="rail-scroll">
        <div
          v-for="g in groupedProjects"
          :key="g.key"
          class="rail-item"
          :class="{ active: activeGroup === g.key }"
          :title="g.key"
          @click="scrollToGroup(g.key)"
        >
          <span class="rail-item-label">{{ shortDir(g.key) }}</span>
          <span class="rail-item-count">{{ g.projects.length }}</span>
        </div>
        <div v-if="!loading && groupedProjects.length === 0" class="rail-empty">
          {{ t('globalList.empty') }}
        </div>
      </div>
    </aside>

    <!-- right main: fixed header + scrollable project cards -->
    <section class="list-main">
      <div class="list-main-head">
        <div class="head-stats">
          <div class="stat-item">
            <FolderOutlined class="stat-icon" />
            <span class="stat-num">{{ projects.length }}</span>
            <span class="stat-label">{{ t('globalList.summaryProjects') }}</span>
          </div>
          <div class="stat-sep"></div>
          <div class="stat-item">
            <span class="stat-num">{{ totalRepos }}</span>
            <span class="stat-label">{{ t('globalList.summaryRepos') }}</span>
          </div>
          <template v-if="brokenProjects > 0">
            <div class="stat-sep"></div>
            <div class="stat-item warn">
              <WarningOutlined class="stat-icon" />
              <span class="stat-num">{{ brokenProjects }}</span>
              <span class="stat-label">{{ t('globalList.summaryBroken') }}</span>
            </div>
          </template>
        </div>
        <div class="head-actions">
          <a-button size="small" danger @click="openCleanup">
            <template #icon><DeleteOutlined /></template>
            {{ t('globalList.cleanup') }}
          </a-button>
          <a-button size="small" @click="loadProjects" :loading="loading">
            <template #icon><ReloadOutlined /></template>
            {{ t('common.refresh') }}
          </a-button>
        </div>
      </div>

      <div ref="scrollRef" class="list-scroll" @scroll.passive="onMainScroll">
        <a-spin :spinning="loading">
          <!-- grouped project cards -->
          <div v-if="projects.length > 0">
            <template v-for="group in groupedProjects" :key="group.key">
              <div class="group-head" :ref="(el) => setGroupEl(group.key, el)" @click="toggleGroup(group.key)">
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
                            <span v-if="!p.exists" class="missing-tag">{{ t('footer.dirMissing') }}</span>
                          </div>
                          <div class="card-dir" :title="formatPath(p.dir)">{{ formatPath(p.dir) }}</div>
                        </div>
                      </div>

                      <div class="card-stats">
                        <div class="stat">
                          <span class="stat-num">{{ p.repoCount }}</span>
                          <span class="stat-label">{{ t('globalList.summaryRepos') }}</span>
                        </div>
                        <div class="stat" v-if="p.brokenCount > 0">
                          <span class="stat-num warn">{{ p.brokenCount }}</span>
                          <span class="stat-label">{{ t('globalList.summaryBroken') }}</span>
                        </div>
                        <div class="stat" v-if="p.agents && p.agents.length">
                          <span class="stat-num">{{ p.agents.length }}</span>
                          <span class="stat-label">{{ t('globalList.agentChip') }}</span>
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
                        <SwapOutlined /> {{ t('globalList.switchToHere') }}
                      </a-menu-item>
                      <a-menu-divider v-if="p.exists" />
                      <a-menu-item v-if="p.exists" @click="onDoctor(p)">
                        <MedicineBoxOutlined /> {{ t('globalList.fixLinks') }}
                      </a-menu-item>
                      <a-menu-item v-if="p.exists" @click="app?.OpenInExplorer(p.dir)">
                        <FolderOpenOutlined /> {{ t('common.openInExplorer') }}
                      </a-menu-item>
                      <a-menu-item @click="onCopyPath(p)">
                        <CopyOutlined /> {{ t('common.copy') }}
                      </a-menu-item>
                      <a-menu-divider />
                      <a-menu-item danger @click="onRemove(p, false)">
                        <DeleteOutlined /> {{ t('globalList.removeProject') }}
                      </a-menu-item>
                      <a-menu-item danger @click="onRemove(p, true)">
                        <DeleteOutlined /> {{ t('globalList.removeClean') }}
                      </a-menu-item>
                    </a-menu>
                  </template>
                </a-dropdown>
              </div>
            </template>
          </div>

          <a-empty v-if="!loading && projects.length === 0" :description="t('globalList.empty')">
            <template #image><FolderOutlined style="font-size: 48px; color: var(--color-text-tertiary); opacity: 0.3;" /></template>
          </a-empty>
        </a-spin>
      </div>
    </section>

    <!-- cleanup modal -->
    <a-modal
      v-model:open="cleanupOpen"
      :title="t('globalList.cleanup')"
      :ok-text="t('globalList.cleanupSelected', { n: cleanupSelected.length })"
      ok-type="danger"
      :cancel-text="t('common.cancelText')"
      :ok-button-props="{ disabled: cleanupSelected.length === 0 }"
      @ok="doCleanup"
    >
      <div v-if="missingProjects.length === 0" class="cleanup-empty">
        {{ t('globalList.noMissing') }}
      </div>
      <template v-else>
        <p class="cleanup-hint">{{ t('globalList.cleanupHint') }}</p>
        <a-checkbox-group v-model:value="cleanupSelected" class="cleanup-list">
          <div v-for="p in missingProjects" :key="p.dir" class="cleanup-item">
            <a-checkbox :value="p.dir">
              <span class="cleanup-name">{{ p.name }}</span>
              <span class="cleanup-dir">{{ formatPath(p.dir) }}</span>
            </a-checkbox>
          </div>
        </a-checkbox-group>
      </template>
    </a-modal>

    <DiagnoseModal v-model:open="diagnoseOpen" :project-dir="diagnoseDir" @update:open="diagnoseOpen = $event; loadProjects()" />
  </div>
</template>

<style scoped>
.global-list {
  display: flex;
  height: calc(100% + 2 * var(--spacing-lg));
  overflow: hidden;
  margin: calc(-1 * var(--spacing-lg));
}

/* ---- left rail: directory list ---- */
.list-rail {
  width: 200px;
  min-width: 200px;
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--color-border);
  background: var(--color-surface);
  flex-shrink: 0;
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
.rail-title {
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-text-tertiary);
}
.rail-count {
  font-size: 10px;
  font-weight: 600;
  color: var(--color-text-tertiary);
  background: var(--color-surface-raised);
  padding: 0 6px;
  border-radius: 999px;
}
.rail-scroll {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-sm);
}
.rail-scroll::-webkit-scrollbar { width: 5px; }
.rail-scroll::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 3px; }
.rail-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 10px;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
  margin-bottom: 1px;
}
.rail-item:hover { background: var(--color-hover); }
.rail-item.active { background: var(--color-primary-bg); }
.rail-item-label {
  flex: 1;
  font-size: 12px;
  color: var(--color-text-secondary);
  font-family: 'Cascadia Code', monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.rail-item.active .rail-item-label { color: var(--color-primary); }
.rail-item-count { font-size: 10px; color: var(--color-text-tertiary); flex-shrink: 0; }
.rail-empty { padding: 24px 12px; text-align: center; font-size: 12px; color: var(--color-text-tertiary); }

/* ---- right main ---- */
.list-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.list-main-head {
  height: var(--navbar-height);
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: 0 var(--spacing-md);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}
.head-stats { display: flex; align-items: center; gap: var(--spacing-sm); }
.stat-item { display: flex; align-items: center; gap: 5px; }
.stat-icon { font-size: 14px; color: var(--color-text-tertiary); }
.stat-item.warn .stat-icon { color: var(--color-warning); }
.stat-num { font-size: 15px; font-weight: 700; color: var(--color-text); line-height: 1; }
.stat-item.warn .stat-num { color: var(--color-warning); }
.stat-label { font-size: 11px; color: var(--color-text-tertiary); }
.stat-sep { width: 1px; height: 16px; background: var(--color-border); }
.head-actions { margin-left: auto; display: flex; gap: var(--spacing-sm); }

.list-scroll {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-lg);
  position: relative;
}
.list-scroll::-webkit-scrollbar { width: 8px; }
.list-scroll::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 4px; }

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
  background: var(--bg-card);
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
  border-radius: 999px; background: var(--color-primary-bg);
  color: var(--color-primary); border: 1px solid var(--color-primary-border);
}

/* cleanup modal */
.cleanup-hint { margin: 0 0 12px; color: var(--color-text-secondary); font-size: 13px; }
.cleanup-list { display: flex; flex-direction: column; gap: 2px; width: 100%; }
.cleanup-item { padding: 6px 8px; border-radius: var(--radius-sm); }
.cleanup-item:hover { background: var(--color-hover); }
.cleanup-name { font-weight: 600; color: var(--color-text); margin-right: 8px; }
.cleanup-dir { font-size: 11px; color: var(--color-text-tertiary); font-family: 'Cascadia Code', monospace; }
.cleanup-empty { text-align: center; color: var(--color-text-tertiary); padding: 24px; }
</style>
