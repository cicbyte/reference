<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { PlusOutlined, SyncOutlined, MedicineBoxOutlined, CloudDownloadOutlined } from '@ant-design/icons-vue'
import { useProjectStore } from '../stores/project'

const router = useRouter()
const project = useProjectStore()
const repos = ref([])
const loading = ref(true)

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
</script>

<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <h1>Reference</h1>
      <p class="subtitle">
        {{ project.hasProject ? `当前项目：${project.currentName}` : '本地代码仓库引用管理器' }}
      </p>
    </div>

    <a-empty
      v-if="!project.hasProject"
      description="请从左侧选择一个项目，或添加新项目"
    >
      <template #image><CloudDownloadOutlined style="font-size: 48px; color: var(--color-text-tertiary); opacity: 0.4;" /></template>
    </a-empty>

    <template v-else>

    <div class="quick-actions">
      <a-button type="primary" @click="router.push('/repos/add')">
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

    <a-spin :spinning="loading">
      <div v-if="repos.length > 0" class="repo-grid">
        <div v-for="repo in repos" :key="repo.name" class="repo-card">
          <div class="repo-card-header">
            <a-tag :color="repo.type === 'remote' ? 'blue' : 'green'">
              {{ repo.type === 'remote' ? '远程' : '本地' }}
            </a-tag>
            <span class="repo-name">{{ repo.name }}</span>
          </div>
          <div class="repo-card-body">
            <div class="repo-source">{{ repo.source }}</div>
            <div class="repo-meta" v-if="repo.branch">
              <span>分支: {{ repo.branch }}</span>
            </div>
            <div class="repo-meta" v-if="repo.commit_at">
              <span>更新: {{ repo.commit_at }}</span>
            </div>
          </div>
        </div>
      </div>
      <a-empty v-else-if="!loading" description="暂无引用仓库">
        <a-button type="primary" @click="router.push('/repos/add')">添加第一个仓库</a-button>
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
  margin-bottom: var(--spacing-xl);
}

.dashboard-header h1 {
  font-size: 28px;
  font-weight: 700;
  color: var(--color-text);
  margin-bottom: var(--spacing-xs);
}

.subtitle {
  color: var(--color-text-secondary);
  font-size: 14px;
}

.quick-actions {
  display: flex;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-xl);
}

.repo-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: var(--spacing-md);
}

.repo-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  transition: all var(--transition-fast);
  cursor: pointer;
}

.repo-card:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-md);
}

.repo-card-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-sm);
}

.repo-name {
  font-weight: 600;
  font-size: 15px;
  color: var(--color-text);
}

.repo-card-body {
  font-size: 13px;
  color: var(--color-text-secondary);
}

.repo-source {
  margin-bottom: var(--spacing-xs);
  word-break: break-all;
}

.repo-meta {
  margin-top: 2px;
}
</style>
