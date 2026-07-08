<script setup>
import { ref, onMounted } from 'vue'

const loading = ref(true)
const stats = ref(null)

onMounted(async () => {
  try {
    if (window.go?.main?.ReferenceApp) {
      stats.value = await window.go.main.ReferenceApp.GlobalStats()
    }
  } catch (e) {
    console.error('Global stats failed:', e)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="global-stats">
    <div class="page-header"><h2>全局统计</h2></div>
    <a-spin :spinning="loading">
      <a-descriptions v-if="stats" bordered :column="2">
        <a-descriptions-item label="总项目数">{{ stats.total_projects }}</a-descriptions-item>
        <a-descriptions-item label="总引用数">{{ stats.total_repos }}</a-descriptions-item>
        <a-descriptions-item label="远程仓库">{{ stats.remote_repos }}</a-descriptions-item>
        <a-descriptions-item label="本地仓库">{{ stats.local_repos }}</a-descriptions-item>
        <a-descriptions-item label="缓存大小">{{ stats.cache_size }}</a-descriptions-item>
        <a-descriptions-item label="数据库大小">{{ stats.db_size }}</a-descriptions-item>
      </a-descriptions>
    </a-spin>
  </div>
</template>

<style scoped>
.global-stats { max-width: 800px; }
.page-header { margin-bottom: var(--spacing-lg); }
.page-header h2 { font-size: 20px; font-weight: 600; color: var(--color-text); }
</style>
