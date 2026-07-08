<script setup>
import { ref } from 'vue'
import { message } from 'ant-design-vue'
import { SyncOutlined } from '@ant-design/icons-vue'

const loading = ref(false)

async function syncWiki() {
  loading.value = true
  try {
    if (window.go?.main?.ReferenceApp) {
      await window.go.main.ReferenceApp.WikiSync()
      message.success('知识库同步成功')
    }
  } catch (e) {
    message.error('同步失败: ' + e)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="wiki-sync">
    <div class="page-header"><h2>知识库同步</h2></div>
    <a-space direction="vertical">
      <p>执行 pull + commit + push 同步知识库</p>
      <a-button type="primary" @click="syncWiki" :loading="loading">
        <template #icon><SyncOutlined /></template>
        同步知识库
      </a-button>
    </a-space>
  </div>
</template>

<style scoped>
.wiki-sync { max-width: 800px; }
.page-header { margin-bottom: var(--spacing-lg); }
.page-header h2 { font-size: 20px; font-weight: 600; color: var(--color-text); }
</style>
