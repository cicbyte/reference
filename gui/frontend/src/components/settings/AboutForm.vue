<script setup>
import { ref, onMounted } from 'vue'
import { InfoCircleOutlined, FileTextOutlined } from '@ant-design/icons-vue'

const loading = ref(true)
const version = ref(null)
const log = ref(null)

onMounted(async () => {
  try {
    if (window.go?.main?.ReferenceApp) {
      const app = window.go.main.ReferenceApp
      version.value = await app.GetVersionInfo()
      const cfg = await app.GetAppConfig()
      log.value = cfg.log
    }
  } catch (e) {
    console.error('Load about info failed:', e)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="settings-form">
    <a-spin :spinning="loading">
      <div class="setting-group">
        <div class="group-title"><InfoCircleOutlined /> 关于</div>
        <a-descriptions bordered :column="1" size="small">
          <a-descriptions-item label="应用">reference</a-descriptions-item>
          <a-descriptions-item label="版本">{{ version?.version || 'dev' }}</a-descriptions-item>
          <a-descriptions-item label="Commit">{{ version?.commit || 'unknown' }}</a-descriptions-item>
          <a-descriptions-item label="构建时间">{{ version?.buildTime || 'unknown' }}</a-descriptions-item>
        </a-descriptions>
      </div>

      <div class="setting-group">
        <div class="group-title"><FileTextOutlined /> 日志配置（只读）</div>
        <div class="row-help" style="margin-bottom: 12px;">
          日志参数通过配置文件 <code>config.yaml</code> 管理，当前页面仅展示。
        </div>
        <a-descriptions bordered :column="2" size="small">
          <a-descriptions-item label="级别">{{ log?.level || '—' }}</a-descriptions-item>
          <a-descriptions-item label="单文件上限">{{ log?.maxSize ? log.maxSize + ' MB' : '—' }}</a-descriptions-item>
          <a-descriptions-item label="保留份数">{{ log?.maxBackups ?? '—' }}</a-descriptions-item>
          <a-descriptions-item label="保留天数">{{ log?.maxAge ?? '—' }}</a-descriptions-item>
          <a-descriptions-item label="压缩">{{ log?.compress ? '是' : '否' }}</a-descriptions-item>
        </a-descriptions>
      </div>
    </a-spin>
  </div>
</template>

<style scoped>
@import './settings-shared.css';

.a-descriptions :deep(.ant-descriptions-item-label) {
  width: 120px;
}
</style>
