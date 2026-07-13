<script setup>
import {
  DatabaseOutlined,
  CloudServerOutlined,
  WarningOutlined,
  LinkOutlined,
  ApiOutlined,
} from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineProps({
  repos: { type: Array, default: () => [] },
  remoteCount: { type: Number, default: 0 },
  localCount: { type: Number, default: 0 },
  brokenCount: { type: Number, default: 0 },
  agentCount: { type: Number, default: 0 },
})

const emit = defineEmits(['navigate', 'diagnose'])
</script>

<template>
  <!-- stat cards row -->
  <div class="stat-row">
    <div class="stat-card" @click="emit('navigate', '/repos')">
      <div class="stat-icon stat-blue"><DatabaseOutlined /></div>
      <div class="stat-body">
        <div class="stat-value">{{ repos.length }}</div>
        <div class="stat-label">{{ t('dashboard.statRepos') }}</div>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon stat-cyan"><CloudServerOutlined /></div>
      <div class="stat-body">
        <div class="stat-value">{{ remoteCount }} <span class="stat-sub">/ {{ localCount }}</span></div>
        <div class="stat-label">{{ t('dashboard.statRemoteLocal') }}</div>
      </div>
    </div>
    <div class="stat-card" :class="{ 'stat-warn': brokenCount > 0 }" @click="emit('diagnose')">
      <div class="stat-icon" :class="brokenCount > 0 ? 'stat-orange' : 'stat-green'">
        <WarningOutlined v-if="brokenCount > 0" />
        <LinkOutlined v-else />
      </div>
      <div class="stat-body">
        <div class="stat-value">{{ brokenCount }}</div>
        <div class="stat-label">{{ brokenCount > 0 ? t('dashboard.statBroken') : t('dashboard.healthHealthy') }}</div>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon stat-purple"><ApiOutlined /></div>
      <div class="stat-body">
        <div class="stat-value">{{ agentCount }}</div>
        <div class="stat-label">{{ t('dashboard.statAgents') }}</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
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
</style>
