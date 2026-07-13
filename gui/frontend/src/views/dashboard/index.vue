<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import {
  PlusOutlined,
  MedicineBoxOutlined,
  CloudDownloadOutlined,
} from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'
import { useProjectStore } from '@/stores/project'
import { formatPath } from '@/utils/path'
import DiagnoseModal from '@/components/repo/DiagnoseModal.vue'
import StatCards from './components/StatCards.vue'
import InfoCards from './components/InfoCards.vue'
import RepoGrid from './components/RepoGrid.vue'

const { t } = useI18n()
const router = useRouter()
const project = useProjectStore()
const repos = ref([])
const loading = ref(true)
const diagnoseOpen = ref(false)

// ---- derived stats from repos + active project ----
const remoteCount = computed(() => repos.value.filter((r) => r.type === 'remote').length)
const localCount = computed(() => repos.value.filter((r) => r.type === 'local').length)
const brokenCount = computed(() => project.activeProject?.brokenCount || 0)
const agentCount = computed(() => project.activeProject?.agents?.length || 0)
const agents = computed(() => project.activeProject?.agents || [])

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

function navigate(path) {
  router.push(path)
}
</script>

<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <h1>{{ project.hasProject ? project.currentName : 'Reference' }}</h1>
      <p class="subtitle">
        {{ project.hasProject ? formatPath(project.currentDir) : t('dashboard.subtitle') }}
      </p>
    </div>

    <!-- no project: empty state -->
    <a-empty
      v-if="!project.hasProject"
      :description="t('common.selectProjectFirst')"
    >
      <template #image><CloudDownloadOutlined style="font-size: 48px; color: var(--color-text-tertiary); opacity: 0.4;" /></template>
    </a-empty>

    <template v-else>
      <!-- project dir missing warning -->
      <a-alert
        v-if="!project.currentExists"
        type="warning"
        show-icon
        :message="t('dashboard.notInitialized')"
        :description="t('dashboard.initHint')"
        style="margin-bottom: var(--spacing-lg)"
      />

      <StatCards
        :repos="repos"
        :remote-count="remoteCount"
        :local-count="localCount"
        :broken-count="brokenCount"
        :agent-count="agentCount"
        @navigate="navigate"
        @diagnose="diagnoseOpen = true"
      />

      <!-- quick actions -->
      <div class="quick-actions">
        <a-button type="primary" @click="router.push('/repos')">
          <template #icon><PlusOutlined /></template>
          {{ t('dashboard.quickAddRepo') }}
        </a-button>
        <a-button @click="router.push('/repos')">
          <template #icon><CloudDownloadOutlined /></template>
          {{ t('dashboard.quickRepoList') }}
        </a-button>
        <a-button @click="diagnoseOpen = true">
          <template #icon><MedicineBoxOutlined /></template>
          {{ t('dashboard.quickDiagnose') }}
        </a-button>
      </div>

      <InfoCards
        :agents="agents"
        :broken-count="brokenCount"
        :repo-count="repos.length"
        @navigate="navigate"
      />

      <RepoGrid :loading="loading" :repos="repos" @navigate="navigate" />
    </template>

    <DiagnoseModal v-model:open="diagnoseOpen" :project-dir="project.currentDir" />
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

/* ---- quick actions ---- */
.quick-actions {
  display: flex;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-lg);
}
</style>
