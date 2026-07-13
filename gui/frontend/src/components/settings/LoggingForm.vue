<script setup>
import { ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { FileTextOutlined } from '@ant-design/icons-vue'

const loading = ref(true)
const saving = ref(false)
const dirty = ref(false)
const form = ref({ level: 'info', maxSize: 10, maxBackups: 30, maxAge: 30, compress: true })

const levelOptions = [
  { value: 'debug', label: 'Debug' },
  { value: 'info', label: 'Info' },
  { value: 'warn', label: 'Warn' },
  { value: 'error', label: 'Error' },
]

onMounted(async () => {
  try {
    if (window.go?.main?.ReferenceApp) {
      const cfg = await window.go.main.ReferenceApp.GetAppConfig()
      form.value = {
        level: cfg.log?.level || 'info',
        maxSize: cfg.log?.maxSize ?? 10,
        maxBackups: cfg.log?.maxBackups ?? 30,
        maxAge: cfg.log?.maxAge ?? 30,
        compress: cfg.log?.compress ?? true,
      }
    }
  } catch (e) {
    console.error('GetAppConfig failed:', e)
  } finally {
    loading.value = false
  }
})

function markDirty() { dirty.value = true }

async function handleSave() {
  saving.value = true
  try {
    if (window.go?.main?.ReferenceApp) {
      await window.go.main.ReferenceApp.SaveAppConfig({ log: { ...form.value } })
      dirty.value = false
      message.success('日志设置已保存，重启后完全生效')
    }
  } catch (e) {
    message.error('保存失败: ' + e)
  } finally {
    saving.value = false
  }
}

async function handleReset() {
  form.value = { level: 'info', maxSize: 10, maxBackups: 30, maxAge: 30, compress: true }
  dirty.value = true
}
</script>

<template>
  <div class="settings-form">
    <div class="form-header">
      <div class="form-title"><FileTextOutlined /> 日志</div>
      <div class="form-desc">控制日志级别与滚动归档策略，修改后重启应用完全生效。</div>
    </div>

    <a-spin :spinning="loading">
      <div class="setting-group">
        <div class="group-title">日志级别</div>
        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">级别</div>
            <div class="row-help">越高级别输出越少；Debug 适合排查问题</div>
          </div>
          <a-select
            v-model:value="form.level" style="width: 160px"
            :options="levelOptions" @change="markDirty"
          />
        </div>
      </div>

      <div class="setting-group">
        <div class="group-title">滚动归档</div>

        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">单文件上限</div>
            <div class="row-help">单个日志文件达到此大小后触发滚动（MB）</div>
          </div>
          <a-input-number v-model:value="form.maxSize" :min="1" :max="500" :step="1"
            style="width: 140px" @change="markDirty">
            <template #addonAfter>MB</template>
          </a-input-number>
        </div>

        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">保留份数</div>
            <div class="row-help">最多保留的归档文件数量</div>
          </div>
          <a-input-number v-model:value="form.maxBackups" :min="1" :max="365" :step="1"
            style="width: 140px" @change="markDirty">
            <template #addonAfter>份</template>
          </a-input-number>
        </div>

        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">保留天数</div>
            <div class="row-help">超过此天数的归档自动删除</div>
          </div>
          <a-input-number v-model:value="form.maxAge" :min="1" :max="3650" :step="1"
            style="width: 140px" @change="markDirty">
            <template #addonAfter>天</template>
          </a-input-number>
        </div>

        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">压缩归档</div>
            <div class="row-help">对滚动产生的归档文件进行 gzip 压缩以节省空间</div>
          </div>
          <a-switch v-model:checked="form.compress" @change="markDirty" />
        </div>

        <div class="group-actions">
          <a-button type="primary" :loading="saving" :disabled="!dirty" @click="handleSave">保存</a-button>
          <a-button :disabled="!dirty" @click="handleReset">恢复默认</a-button>
        </div>
      </div>
    </a-spin>
  </div>
</template>

<style scoped>
@import './settings-shared.css';
</style>
