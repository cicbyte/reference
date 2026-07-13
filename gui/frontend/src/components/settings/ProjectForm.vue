<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { message } from 'ant-design-vue'
import {
  CheckOutlined, FolderOpenOutlined, ExclamationCircleFilled,
} from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'
import AiSparkIcon from '@/components/common/AiSparkIcon.vue'
import { useProjectStore } from '@/stores/project'
import { formatPath, joinPath } from '@/utils/path'

const { t } = useI18n()
const project = useProjectStore()
const app = window.go?.main?.ReferenceApp

const loading = ref(true)
const saving = ref(false)
const agents = ref([])
const selected = ref([])
const initialized = ref(false)
const initialAgents = ref([])

const selectedCount = computed(() => selected.value.length)
const hasProject = computed(() => !!project.currentDir)

const dirty = computed(() => {
  const orig = new Set(initialAgents.value)
  const cur = new Set(selected.value)
  if (orig.size !== cur.size) return true
  for (const a of cur) if (!orig.has(a)) return true
  return false
})

// first-letter badge color palette (deterministic per agent id)
const badgeColors = {
  claude: '#d97757',
  zcode: '#3b82f6',
  mimocode: '#10b981',
  opencode: '#8b5cf6',
  codex: '#0ea5e9',
}
function badgeLetter(name) {
  return (name || '?').trim().charAt(0).toUpperCase()
}
function badgeColor(id) {
  return badgeColors[id] || '#6b7280'
}
function toggle(id) {
  const i = selected.value.indexOf(id)
  if (i >= 0) selected.value = selected.value.filter((x) => x !== id)
  else selected.value = [...selected.value, id]
}

async function loadCurrent() {
  if (!app) { loading.value = false; return }
  loading.value = true
  try {
    if (!agents.value.length) agents.value = await app.ListAgents()
    if (hasProject.value) {
      const projects = await app.ListProjects()
      const cur = projects.find((p) => p.dir === project.currentDir)
      if (cur) {
        initialAgents.value = [...(cur.agents || [])]
        selected.value = [...(cur.agents || [])]
        initialized.value = cur.initialized
      } else {
        initialAgents.value = []
        selected.value = []
        initialized.value = false
      }
    } else {
      initialAgents.value = []
      selected.value = []
      initialized.value = false
    }
  } catch (e) {
    console.error('load agents failed:', e)
  } finally {
    loading.value = false
  }
}

onMounted(loadCurrent)
// reload when the user switches project while this form is mounted
watch(() => project.currentDir, loadCurrent)

async function handleSave() {
  if (!hasProject.value) {
    message.warning(t('common.selectProjectFirst'))
    return
  }
  saving.value = true
  try {
    await app.InitProject(selected.value)
    initialAgents.value = [...selected.value]
    initialized.value = true
    message.success(t('settings.project.injected', { n: selected.value.length, name: project.currentName }))
  } catch (e) {
    message.error(t('settings.project.injectFailed') + ': ' + e)
  } finally {
    saving.value = false
  }
}

function selectAll() { selected.value = agents.value.map((a) => a.id) }
function clearAll() { selected.value = [] }
</script>

<template>
  <div class="settings-form">
    <div class="form-header">
      <div class="form-title"><AiSparkIcon :size="18" /> {{ t('settings.project.title') }}</div>
      <div class="form-desc">{{ t('settings.project.desc') }}</div>
    </div>

    <a-spin :spinning="loading">
      <!-- context banner -->
      <div class="ctx-banner" :class="{ 'no-project': !hasProject }">
        <div class="ctx-info">
          <div class="ctx-row">
            <span class="ctx-label">{{ t('settings.project.targetProject') }}</span>
            <span class="ctx-value" :title="formatPath(project.currentDir)">
              {{ project.currentName || t('settings.project.noProject') }}
            </span>
          </div>
          <div class="ctx-row" v-if="hasProject">
            <span class="ctx-label">{{ t('settings.project.writeLocation') }}</span>
            <span class="ctx-path mono">{{ joinPath(project.currentDir, '.reference/reference.settings.json') }}</span>
          </div>
        </div>
        <span class="ctx-state" :class="{ on: initialized }" v-if="hasProject">
          <CheckOutlined v-if="initialized" />
          {{ initialized ? t('settings.project.initialized') : t('settings.project.uninitialized') }}
        </span>
        <span class="ctx-state warn" v-else>
          <ExclamationCircleFilled /> {{ t('settings.project.notSelected') }}
        </span>
      </div>

      <div class="setting-group">
        <div class="group-head">
          <div class="group-title no-margin">{{ t('settings.project.availableAgents') }}</div>
          <div class="head-actions">
            <span class="sel-count">{{ t('settings.project.selected', { selected: selectedCount, total: agents.length }) }}</span>
            <a-button size="small" type="text" @click="selectAll" :disabled="!agents.length">{{ t('settings.project.selectAll') }}</a-button>
            <a-button size="small" type="text" @click="clearAll" :disabled="!selected.length">{{ t('settings.project.clearAll') }}</a-button>
          </div>
        </div>

        <div class="agent-grid">
          <div
            v-for="agent in agents" :key="agent.id"
            class="agent-card"
            :class="{ active: selected.includes(agent.id) }"
            @click="toggle(agent.id)"
          >
            <div class="agent-badge" :style="{ background: badgeColor(agent.id) }">
              {{ badgeLetter(agent.displayName) }}
            </div>
            <div class="agent-body">
              <div class="agent-name">{{ agent.displayName }}</div>
              <div class="agent-meta">
                <span class="meta-path mono">{{ agent.baseDir }}</span>
                <span class="meta-files">{{ t('settings.project.files', { n: agent.fileCount }) }}</span>
              </div>
            </div>
            <div class="agent-check" :class="{ checked: selected.includes(agent.id) }">
              <CheckOutlined v-if="selected.includes(agent.id)" />
            </div>
          </div>
        </div>

        <div v-if="!agents.length && !loading" class="agent-empty">
          {{ t('settings.project.noAgents') }}
        </div>

        <div class="group-actions">
          <a-button
            type="primary"
            :loading="saving"
            :disabled="!hasProject || !dirty"
            @click="handleSave"
          >
            {{ t('settings.project.injectToProject') }}
          </a-button>
          <span v-if="hasProject && !dirty" class="saved-hint">
            <CheckOutlined /> {{ t('settings.project.upToDate') }}
          </span>
          <span v-else-if="!hasProject" class="saved-hint warn">
            {{ t('common.selectProjectFirst') }}
          </span>
        </div>
      </div>
    </a-spin>
  </div>
</template>

<style scoped>
@import './settings-shared.css';

/* context banner */
.ctx-banner {
  display: flex; align-items: center; gap: 16px;
  padding: 14px 16px; margin-bottom: var(--spacing-md);
  border-radius: var(--radius-md); border: 1px solid var(--color-border);
  background: var(--color-surface);
}
.ctx-banner.no-project { border-style: dashed; }
.ctx-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4px; }
.ctx-row { display: flex; align-items: baseline; gap: 10px; min-width: 0; }
.ctx-label {
  font-size: 11px; color: var(--color-text-tertiary);
  width: 56px; flex-shrink: 0;
}
.ctx-value {
  font-size: 14px; font-weight: 600; color: var(--color-text);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.ctx-path {
  font-size: 11px; color: var(--color-text-tertiary);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.ctx-state {
  display: inline-flex; align-items: center; gap: 4px; flex-shrink: 0;
  font-size: 11px; padding: 3px 10px; border-radius: 12px;
  background: var(--color-surface-raised); color: var(--color-text-tertiary);
}
.ctx-state.on { background: var(--color-success-bg); color: var(--color-success); }
.ctx-state.warn { background: var(--color-warning-bg); color: var(--color-warning); }

/* group head */
.group-head {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 12px;
}
.group-title.no-margin { margin-bottom: 0; }
.head-actions { display: flex; align-items: center; gap: 4px; }
.sel-count { font-size: 12px; color: var(--color-text-tertiary); margin-right: 6px; }

/* agent grid */
.agent-grid {
  display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 10px;
}
.agent-card {
  display: flex; align-items: center; gap: 12px;
  padding: 12px 14px; cursor: pointer;
  border: 1px solid var(--color-border-light); border-radius: var(--radius-md);
  background: var(--color-background);
  transition: all var(--transition-fast);
  position: relative; overflow: hidden;
}
.agent-card::before {
  content: ''; position: absolute; left: 0; top: 0; bottom: 0; width: 3px;
  background: var(--color-primary); opacity: 0; transition: opacity var(--transition-fast);
}
.agent-card:hover { border-color: var(--color-primary); }
.agent-card.active {
  border-color: var(--color-primary); background: var(--color-primary-bg);
}
.agent-card.active::before { opacity: 1; }

.agent-badge {
  display: flex; align-items: center; justify-content: center;
  width: 36px; height: 36px; border-radius: var(--radius-sm);
  color: #fff; font-size: 16px; font-weight: 700; flex-shrink: 0;
  font-family: 'Cascadia Code', 'Fira Code', monospace;
}
.agent-body { flex: 1; min-width: 0; }
.agent-name { font-size: 13px; font-weight: 600; color: var(--color-text); }
.agent-meta {
  display: flex; align-items: center; gap: 8px; margin-top: 3px;
}
.meta-path { font-size: 11px; color: var(--color-text-tertiary); }
.meta-files {
  font-size: 10px; color: var(--color-text-tertiary);
  padding: 0 6px; border: 1px solid var(--color-border-light); border-radius: 8px;
}

.agent-check {
  display: flex; align-items: center; justify-content: center;
  width: 18px; height: 18px; border-radius: 50%;
  border: 1.5px solid var(--color-border); flex-shrink: 0;
  color: #fff; font-size: 10px; transition: all var(--transition-fast);
}
.agent-check.checked {
  background: var(--color-primary); border-color: var(--color-primary);
}

.agent-empty {
  padding: 32px; text-align: center; color: var(--color-text-tertiary); font-size: 13px;
}

.group-actions { align-items: center; }
.saved-hint { font-size: 12px; color: var(--color-text-tertiary); display: inline-flex; align-items: center; gap: 4px; }
.saved-hint.warn { color: var(--color-warning); }
</style>
