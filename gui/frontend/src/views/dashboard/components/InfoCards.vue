<script setup>
import { computed } from 'vue'
import { ApiOutlined, DatabaseOutlined } from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'
import { agentDisplayName } from '@/utils/agents'
import { useProjectStore } from '@/stores/project'

const { t } = useI18n()

const props = defineProps({
  agents: { type: Array, default: () => [] },
  brokenCount: { type: Number, default: 0 },
  repoCount: { type: Number, default: 0 },
})

const emit = defineEmits(['navigate'])

const project = useProjectStore()

// overall health derived from individual checks
const healthLabel = computed(() => {
  if (!project.currentExists) return t('common.failed')
  if (props.brokenCount > 0) return t('dashboard.statBroken')
  if (props.repoCount === 0) return t('common.none')
  return t('dashboard.healthHealthy')
})
const healthClass = computed(() => {
  if (!project.currentExists) return 'bad'
  if (props.brokenCount > 0) return 'warn'
  return 'ok'
})
</script>

<template>
  <!-- info cards: single full-width row -->
  <div class="dash-section">
    <div class="section-head">
      <h3>{{ t('dashboard.healthChecks') }}</h3>
    </div>
    <div class="info-row">
      <!-- AI assistants -->
      <div class="side-card">
        <div class="side-card-head">
          <span class="side-title"><ApiOutlined /> {{ t('dashboard.aiTitle') }}</span>
          <a-button size="small" type="text" class="side-config" @click="emit('navigate', '/settings')">{{ t('sidebar.config') }}</a-button>
        </div>
        <div v-if="agents.length > 0" class="agent-grid">
          <div v-for="a in agents" :key="a" class="agent-chip">
            <span class="agent-dot"></span>
            <span class="agent-name">{{ agentDisplayName(a) }}</span>
          </div>
        </div>
        <div v-else class="side-muted">
          <ApiOutlined class="side-muted-icon" />
          <span>{{ t('dashboard.aiNone') }}</span>
        </div>
      </div>

      <!-- health -->
      <div class="side-card">
        <div class="side-card-head">
          <span class="side-title"><DatabaseOutlined /> {{ t('dashboard.healthTitle') }}</span>
          <span class="health-badge" :class="healthClass">{{ healthLabel }}</span>
        </div>
        <div class="check-list">
          <div class="check-row">
            <span :class="['check-dot', project.activeProject?.initialized ? 'ok' : 'off']"></span>
            <span class="check-label">{{ t('settings.project.initialized') }}</span>
          </div>
          <div class="check-row">
            <span :class="['check-dot', project.currentExists ? 'ok' : 'bad']"></span>
            <span class="check-label">{{ t('diagnose.targetExists') }}</span>
          </div>
          <div class="check-row">
            <span :class="['check-dot', brokenCount > 0 ? 'warn' : 'ok']"></span>
            <span class="check-label">{{ brokenCount > 0 ? t('globalList.brokenCount', { n: brokenCount }) : t('dashboard.healthHealthy') }}</span>
          </div>
          <div class="check-row">
            <span :class="['check-dot', repoCount > 0 ? 'ok' : 'off']"></span>
            <span class="check-label">{{ repoCount > 0 ? t('globalList.repoCount', { n: repoCount }) : t('globalStats.noRefs') }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dash-section {
  display: block;
  margin-bottom: var(--spacing-lg);
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

/* info cards in a single row */
.info-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--spacing-md);
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
