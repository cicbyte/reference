<script setup>
import { ref, onMounted } from 'vue'
import { SettingOutlined } from '@ant-design/icons-vue'

const loading = ref(true)
const paths = ref(null)

onMounted(async () => {
  try {
    if (window.go?.main?.ReferenceApp) {
      const cfg = await window.go.main.ReferenceApp.GetAppConfig()
      paths.value = cfg.paths
    }
  } catch (e) {
    console.error('GetAppConfig failed:', e)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="settings-form">
    <a-spin :spinning="loading">
      <div class="setting-group">
        <div class="group-title"><SettingOutlined /> 路径信息</div>
        <a-descriptions bordered :column="1" size="small">
          <a-descriptions-item label="配置文件">{{ paths?.config || '—' }}</a-descriptions-item>
          <a-descriptions-item label="数据库">{{ paths?.db || '—' }}</a-descriptions-item>
          <a-descriptions-item label="日志目录">{{ paths?.logDir || '—' }}</a-descriptions-item>
          <a-descriptions-item label="仓库缓存（实际）">{{ paths?.repos || '—' }}</a-descriptions-item>
          <a-descriptions-item label="知识库（实际）">{{ paths?.wiki || '—' }}</a-descriptions-item>
        </a-descriptions>
      </div>
    </a-spin>
  </div>
</template>

<style scoped>
@import './settings-shared.css';

.a-descriptions :deep(.ant-descriptions-item-label) {
  width: 140px;
}
</style>
