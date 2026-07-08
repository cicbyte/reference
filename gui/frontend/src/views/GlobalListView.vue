<script setup>
import { ref, onMounted } from 'vue'

const loading = ref(true)
const projects = ref([])

onMounted(async () => {
  try {
    if (window.go?.main?.ReferenceApp) {
      projects.value = await window.go.main.ReferenceApp.GlobalList()
    }
  } catch (e) {
    console.error('Global list failed:', e)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="global-list">
    <div class="page-header"><h2>全局项目列表</h2></div>
    <a-spin :spinning="loading">
      <a-table :data-source="projects" :pagination="false" size="middle">
        <a-table-column title="项目" data-index="project_dir" ellipsis />
        <a-table-column title="引用数" data-index="repo_count" width="100" />
      </a-table>
      <a-empty v-if="!loading && projects.length === 0" description="暂无项目" />
    </a-spin>
  </div>
</template>

<style scoped>
.global-list { width: 100%; }
.page-header { margin-bottom: var(--spacing-lg); }
.page-header h2 { font-size: 20px; font-weight: 600; color: var(--color-text); }
</style>
