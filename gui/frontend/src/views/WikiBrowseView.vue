<script setup>
import { ref, onMounted } from 'vue'

const loading = ref(true)
const wikiStatus = ref(null)
const files = ref([])

onMounted(async () => {
  try {
    if (window.go?.main?.ReferenceApp) {
      wikiStatus.value = await window.go.main.ReferenceApp.WikiStatus()
    }
  } catch (e) {
    console.error('Wiki status failed:', e)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="wiki-browse">
    <div class="page-header"><h2>知识库浏览</h2></div>
    <a-spin :spinning="loading">
      <a-descriptions v-if="wikiStatus" bordered :column="1" size="small" style="margin-bottom: 16px">
        <a-descriptions-item label="状态">{{ wikiStatus.status }}</a-descriptions-item>
        <a-descriptions-item label="路径">{{ wikiStatus.path }}</a-descriptions-item>
        <a-descriptions-item label="远程仓库">{{ wikiStatus.remote || '未设置' }}</a-descriptions-item>
      </a-descriptions>
      <a-empty v-else-if="!loading" description="知识库未初始化" />
    </a-spin>
  </div>
</template>

<style scoped>
.wiki-browse { max-width: 1000px; }
.page-header { margin-bottom: var(--spacing-lg); }
.page-header h2 { font-size: 20px; font-weight: 600; color: var(--color-text); }
</style>
