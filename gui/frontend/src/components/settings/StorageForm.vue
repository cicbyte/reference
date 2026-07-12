<script setup>
import { ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { DatabaseOutlined } from '@ant-design/icons-vue'

const loading = ref(true)
const saving = ref(false)
const form = ref({ reposPath: '', wikiPath: '' })
const original = ref({ reposPath: '', wikiPath: '' })

onMounted(async () => {
  try {
    if (window.go?.main?.ReferenceApp) {
      const cfg = await window.go.main.ReferenceApp.GetAppConfig()
      form.value.reposPath = cfg.reposPath || ''
      form.value.wikiPath = cfg.wikiPath || ''
      original.value = { ...form.value }
    }
  } catch (e) {
    console.error('GetAppConfig failed:', e)
  } finally {
    loading.value = false
  }
})

async function handleSave() {
  saving.value = true
  try {
    if (window.go?.main?.ReferenceApp) {
      await window.go.main.ReferenceApp.SaveAppConfig({
        reposPath: form.value.reposPath,
        wikiPath: form.value.wikiPath,
      })
      const pathChanged =
        form.value.reposPath !== original.value.reposPath ||
        form.value.wikiPath !== original.value.wikiPath
      original.value = { ...form.value }
      message.success('保存成功')
      if (pathChanged) {
        message.warning('存储路径已变更，重启应用后完全生效', 5)
      }
    }
  } catch (e) {
    message.error('保存失败: ' + e)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="settings-form">
    <a-spin :spinning="loading">
      <div class="setting-group">
        <div class="group-title"><DatabaseOutlined /> 存储路径</div>

        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">仓库缓存目录</div>
            <div class="row-help">全局仓库克隆缓存位置，留空使用默认值</div>
          </div>
          <a-input
            v-model:value="form.reposPath"
            placeholder="~/.cicbyte/reference/repos"
            class="row-control"
            style="width: 320px"
            allow-clear
          />
        </div>

        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">知识库目录</div>
            <div class="row-help">远程仓库知识文件存储位置，留空使用默认值</div>
          </div>
          <a-input
            v-model:value="form.wikiPath"
            placeholder="~/.cicbyte/reference/wiki"
            class="row-control"
            style="width: 320px"
            allow-clear
          />
        </div>

        <div class="group-actions">
          <a-button type="primary" :loading="saving" @click="handleSave">保存</a-button>
        </div>
      </div>

      <a-alert
        type="info"
        show-icon
        message="修改存储路径后，需重启应用以触发数据迁移（已有缓存的路径记录会自动更新）。"
        style="max-width: 100%"
      />
    </a-spin>
  </div>
</template>

<style scoped>
@import './settings-shared.css';
</style>
