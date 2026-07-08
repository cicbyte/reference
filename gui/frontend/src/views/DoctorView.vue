<script setup>
import { ref, onMounted, watch } from 'vue'
import { CheckCircleOutlined, ExclamationCircleOutlined, CloseCircleOutlined, SyncOutlined } from '@ant-design/icons-vue'
import { useProjectStore } from '../stores/project'

const project = useProjectStore()
const loading = ref(false)
const result = ref(null)

async function runDoctor() {
  if (!project.hasProject) { result.value = null; return }
  loading.value = true
  try {
    if (window.go?.main?.ReferenceApp) {
      result.value = await window.go.main.ReferenceApp.RunDoctor()
    }
  } catch (e) {
    console.error('Doctor failed:', e)
  } finally {
    loading.value = false
  }
}

onMounted(runDoctor)
watch(() => project.projectEpoch, runDoctor)

function statusIcon(status) {
  if (status === 'pass' || status === 'ok') return CheckCircleOutlined
  if (status === 'warn' || status === 'warning') return ExclamationCircleOutlined
  return CloseCircleOutlined
}

function statusColor(status) {
  if (status === 'pass' || status === 'ok') return 'var(--color-success)'
  if (status === 'warn' || status === 'warning') return 'var(--color-warning)'
  return 'var(--color-error)'
}
</script>

<template>
  <div class="doctor-view">
    <div class="page-header">
      <h2>诊断修复</h2>
      <a-button @click="runDoctor" :loading="loading">
        <template #icon><SyncOutlined /></template>
        重新检查
      </a-button>
    </div>

    <a-spin :spinning="loading">
      <div v-if="result">
        <a-alert
          :message="result.summary"
          :type="result.summary.includes('通过') ? 'success' : 'warning'"
          show-icon
          style="margin-bottom: 16px"
        />

        <a-list :data-source="result.checks" item-layout="horizontal">
          <template #renderItem="{ item }">
            <a-list-item>
              <a-list-item-meta>
                <template #avatar>
                  <component :is="statusIcon(item.status)" :style="{ color: statusColor(item.status), fontSize: '20px' }" />
                </template>
                <template #title>
                  <span>{{ item.group }} — {{ item.name }}</span>
                </template>
                <template #description>
                  <span>{{ item.details }}</span>
                </template>
              </a-list-item-meta>
            </a-list-item>
          </template>
        </a-list>
      </div>
    </a-spin>
  </div>
</template>

<style scoped>
.doctor-view { width: 100%; }
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
}
.page-header h2 { font-size: 20px; font-weight: 600; color: var(--color-text); }
</style>
