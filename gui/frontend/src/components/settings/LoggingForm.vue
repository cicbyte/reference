<script setup>
import { ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { FileTextOutlined } from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

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
      message.success(t('settings.logging.saved'))
    }
  } catch (e) {
    message.error(t('common.saveFailed') + ': ' + e)
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
      <div class="form-title"><FileTextOutlined /> {{ t('settings.logging.title') }}</div>
      <div class="form-desc">{{ t('settings.logging.desc') }}</div>
    </div>

    <a-spin :spinning="loading">
      <div class="setting-group">
        <div class="group-title">{{ t('settings.logging.level') }}</div>
        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">{{ t('settings.logging.level') }}</div>
            <div class="row-help">{{ t('settings.logging.levelHelp') }}</div>
          </div>
          <a-select
            v-model:value="form.level" style="width: 160px"
            :options="levelOptions" @change="markDirty"
          />
        </div>
      </div>

      <div class="setting-group">
        <div class="group-title">{{ t('settings.logging.rolling') }}</div>

        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">{{ t('settings.logging.maxSize') }}</div>
            <div class="row-help">{{ t('settings.logging.maxSizeHelp') }}</div>
          </div>
          <a-input-number v-model:value="form.maxSize" :min="1" :max="500" :step="1"
            style="width: 140px" @change="markDirty">
            <template #addonAfter>{{ t('settings.logging.mb') }}</template>
          </a-input-number>
        </div>

        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">{{ t('settings.logging.maxBackups') }}</div>
            <div class="row-help">{{ t('settings.logging.maxBackupsHelp') }}</div>
          </div>
          <a-input-number v-model:value="form.maxBackups" :min="1" :max="365" :step="1"
            style="width: 140px" @change="markDirty">
            <template #addonAfter>{{ t('settings.logging.copies') }}</template>
          </a-input-number>
        </div>

        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">{{ t('settings.logging.maxAge') }}</div>
            <div class="row-help">{{ t('settings.logging.maxAgeHelp') }}</div>
          </div>
          <a-input-number v-model:value="form.maxAge" :min="1" :max="3650" :step="1"
            style="width: 140px" @change="markDirty">
            <template #addonAfter>{{ t('settings.logging.days') }}</template>
          </a-input-number>
        </div>

        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">{{ t('settings.logging.compress') }}</div>
            <div class="row-help">{{ t('settings.logging.compressHelp') }}</div>
          </div>
          <a-switch v-model:checked="form.compress" @change="markDirty" />
        </div>

        <div class="group-actions">
          <a-button type="primary" :loading="saving" :disabled="!dirty" @click="handleSave">{{ t('common.save') }}</a-button>
          <a-button :disabled="!dirty" @click="handleReset">{{ t('settings.logging.reset') }}</a-button>
        </div>
      </div>
    </a-spin>
  </div>
</template>

<style scoped>
@import './settings-shared.css';
</style>
