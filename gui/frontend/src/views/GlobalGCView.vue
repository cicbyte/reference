<script setup>
import { ref } from 'vue'
import { message } from 'ant-design-vue'

const loading = ref(false)
const dryRun = ref(true)
const result = ref(null)

async function runGC() {
  loading.value = true
  try {
    if (window.go?.main?.ReferenceApp) {
      result.value = await window.go.main.ReferenceApp.GlobalGC(dryRun.value)
      if (!dryRun.value) message.success('清理完成')
    }
  } catch (e) {
    message.error('GC 失败: ' + e)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="global-gc">
    <div class="page-header"><h2>垃圾回收</h2></div>
    <a-space direction="vertical" style="width: 100%">
      <a-alert message="清理过期的数据库记录和孤立的缓存目录" type="info" show-icon />
      <a-checkbox v-model:checked="dryRun">预览模式（不实际删除）</a-checkbox>
      <a-button type="primary" @click="runGC" :loading="loading">
        {{ dryRun ? '预览清理' : '执行清理' }}
      </a-button>
      <a-card v-if="result" title="清理结果" size="small">
        <pre>{{ JSON.stringify(result, null, 2) }}</pre>
      </a-card>
    </a-space>
  </div>
</template>

<style scoped>
.global-gc { width: 100%; }
.page-header { margin-bottom: var(--spacing-lg); }
.page-header h2 { font-size: 20px; font-weight: 600; color: var(--color-text); }
</style>
