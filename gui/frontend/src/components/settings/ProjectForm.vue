<script setup>
import { ref, onMounted, computed } from 'vue'
import { message } from 'ant-design-vue'
import { RobotOutlined } from '@ant-design/icons-vue'
import AiSparkIcon from '../common/AiSparkIcon.vue'
import { useProjectStore } from '../../stores/project'

const project = useProjectStore()
const app = window.go?.main?.ReferenceApp

const loading = ref(true)
const saving = ref(false)
const agents = ref([])
const selected = ref([])
const initialized = ref(false)

const selectedCount = computed(() => selected.value.length)
const dirty = computed(() => {
  const orig = new Set(initialAgents.value)
  const cur = new Set(selected.value)
  if (orig.size !== cur.size) return true
  for (const a of cur) if (!orig.has(a)) return true
  return false
})
const initialAgents = ref([])

onMounted(async () => {
  if (!app) { loading.value = false; return }
  try {
    agents.value = await app.ListAgents()
    // pre-select from current project's settings
    const projects = await app.ListProjects()
    const cur = projects.find((p) => p.dir === project.currentDir)
    if (cur) {
      initialAgents.value = [...(cur.agents || [])]
      selected.value = [...(cur.agents || [])]
      initialized.value = cur.initialized
    }
  } catch (e) {
    console.error('List agents failed:', e)
  } finally {
    loading.value = false
  }
})

async function handleSave() {
  if (!project.currentDir) {
    message.warning('请先选择一个项目')
    return
  }
  saving.value = true
  try {
    await app.InitProject(selected.value)
    initialAgents.value = [...selected.value]
    initialized.value = true
    message.success('已注入到当前项目')
  } catch (e) {
    message.error('保存失败: ' + e)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="settings-form">
    <div class="form-header">
      <div class="form-title"><AiSparkIcon :size="18" /> AI 助手</div>
      <div class="form-desc">选择要注入到当前项目的 AI 编程助手配置。</div>
    </div>

    <a-spin :spinning="loading">
      <!-- context banner -->
      <div class="ctx-banner">
        <div class="ctx-label">目标项目</div>
        <div class="ctx-value" :title="project.currentDir">
          {{ project.currentName || '未选择' }}
        </div>
        <span class="ctx-state" :class="{ on: initialized }">
          {{ initialized ? '已初始化' : '未初始化' }}
        </span>
      </div>

      <div class="setting-group">
        <div class="group-head">
          <div class="group-title no-margin">编程助手</div>
          <span class="sel-count">已选 {{ selectedCount }} / {{ agents.length }}</span>
        </div>
        <div class="row-help" style="margin-bottom: 12px;">
          配置会写入 <code>.reference/reference.settings.json</code>，下次运行时自动注入对应的子代理与 Skill。
        </div>

        <div class="agent-grid">
          <label
            v-for="agent in agents" :key="agent.id"
            class="agent-card"
            :class="{ active: selected.includes(agent.id) }"
          >
            <input type="checkbox" :value="agent.id" v-model="selected" class="agent-check" />
            <div class="agent-card-body">
              <div class="agent-card-name">{{ agent.display_name }}</div>
              <div class="agent-card-id mono">{{ agent.id }}</div>
            </div>
            <RobotOutlined class="agent-card-icon" />
          </label>
        </div>

        <div class="group-actions">
          <a-button type="primary" :loading="saving" :disabled="!dirty" @click="handleSave">
            注入到当前项目
          </a-button>
          <span v-if="!dirty && initialized" class="saved-hint">已是最新配置</span>
        </div>
      </div>
    </a-spin>
  </div>
</template>

<style scoped>
@import './settings-shared.css';

.ctx-banner {
  display: flex; align-items: center; gap: 12px;
  padding: 12px 14px; margin-bottom: var(--spacing-md);
  border-radius: var(--radius-md); border: 1px solid var(--color-border);
  background: var(--color-surface);
}
.ctx-label { font-size: 12px; color: var(--color-text-tertiary); }
.ctx-value { font-size: 14px; font-weight: 600; color: var(--color-text); flex: 1;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ctx-state {
  font-size: 11px; padding: 2px 8px; border-radius: 4px;
  background: var(--color-surface-raised); color: var(--color-text-tertiary);
}
.ctx-state.on { background: var(--color-success-bg); color: var(--color-success); }

.group-head {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 12px;
}
.group-title.no-margin { margin-bottom: 0; }
.sel-count { font-size: 12px; color: var(--color-text-tertiary); }

.agent-grid {
  display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 10px;
}
.agent-card {
  display: flex; align-items: center; gap: 10px;
  padding: 12px 14px; cursor: pointer;
  border: 1px solid var(--color-border-light); border-radius: var(--radius-sm);
  background: var(--color-background); transition: all var(--transition-fast);
  position: relative;
}
.agent-card:hover { border-color: var(--color-primary); }
.agent-card.active {
  border-color: var(--color-primary); background: var(--color-primary-bg);
}
.agent-check { accent-color: var(--color-primary); width: 15px; height: 15px; flex-shrink: 0; }
.agent-card-body { flex: 1; min-width: 0; }
.agent-card-name { font-size: 13px; font-weight: 600; color: var(--color-text); }
.agent-card-id { font-size: 11px; color: var(--color-text-tertiary); margin-top: 2px; }
.agent-card-icon {
  font-size: 20px; color: var(--color-text-tertiary); opacity: 0.4; flex-shrink: 0;
}
.agent-card.active .agent-card-icon { color: var(--color-primary); opacity: 1; }

.saved-hint { font-size: 12px; color: var(--color-text-tertiary); }
</style>
