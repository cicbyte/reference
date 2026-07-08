<script setup>
import { ref } from 'vue'

const activeTab = ref('general')
</script>

<template>
  <div class="settings-view">
    <div class="page-header"><h2>设置</h2></div>
    <a-tabs v-model:activeKey="activeTab">
      <a-tab-pane key="general" tab="通用">
        <a-descriptions bordered :column="1" size="small">
          <a-descriptions-item label="配置路径">~/.cicbyte/reference/config/config.yaml</a-descriptions-item>
          <a-descriptions-item label="数据库路径">~/.cicbyte/reference/db/app.db</a-descriptions-item>
        </a-descriptions>
      </a-tab-pane>
      <a-tab-pane key="about" tab="关于">
        <a-descriptions bordered :column="1" size="small">
          <a-descriptions-item label="应用">Reference GUI</a-descriptions-item>
          <a-descriptions-item label="版本">0.1.0</a-descriptions-item>
        </a-descriptions>
      </a-tab-pane>
    </a-tabs>
  </div>
</template>

<style scoped>
.settings-view { max-width: 800px; }
.page-header { margin-bottom: var(--spacing-lg); }
.page-header h2 { font-size: 20px; font-weight: 600; color: var(--color-text); }
</style>
