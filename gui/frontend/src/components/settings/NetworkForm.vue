<script setup>
import { ref, onMounted, computed } from 'vue'
import { message } from 'ant-design-vue'
import { GlobalOutlined, CheckCircleFilled, MinusCircleFilled } from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const loading = ref(true)
const saving = ref(false)
const dirty = ref(false)
const form = ref({ proxy: '', gitProxy: '', timeout: 300 })

const proxyActive = computed(() => !!(form.value.proxy || form.value.gitProxy))

onMounted(async () => {
  try {
    if (window.go?.main?.ReferenceApp) {
      const cfg = await window.go.main.ReferenceApp.GetAppConfig()
      form.value.proxy = cfg.network?.proxy || ''
      form.value.gitProxy = cfg.network?.gitProxy || ''
      form.value.timeout = cfg.network?.timeout || 300
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
      await window.go.main.ReferenceApp.SaveAppConfig({
        network: {
          proxy: form.value.proxy,
          gitProxy: form.value.gitProxy,
          timeout: form.value.timeout,
        },
      })
      dirty.value = false
      message.success(t('settings.network.saved'))
    }
  } catch (e) {
    message.error(t('common.saveFailed') + ': ' + e)
  } finally {
    saving.value = false
  }
}

async function handleClear() {
  form.value.proxy = ''
  form.value.gitProxy = ''
  saving.value = true
  try {
    if (window.go?.main?.ReferenceApp) {
      await window.go.main.ReferenceApp.SaveAppConfig({
        network: { proxy: '', gitProxy: '', timeout: form.value.timeout },
      })
      dirty.value = false
      message.success(t('settings.network.proxyCleared'))
    }
  } catch (e) {
    message.error(t('settings.network.clearFailed') + ': ' + e)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="settings-form">
    <div class="form-header">
      <div class="form-title"><GlobalOutlined /> {{ t('settings.network.title') }}</div>
      <div class="form-desc">{{ t('settings.network.desc') }}</div>
    </div>

    <a-spin :spinning="loading">
      <!-- status banner -->
      <div class="proxy-status" :class="proxyActive ? 'on' : 'off'">
        <CheckCircleFilled v-if="proxyActive" class="ps-icon" />
        <MinusCircleFilled v-else class="ps-icon" />
        <span>{{ proxyActive ? t('settings.network.proxyOn') : t('settings.network.proxyOff') }}</span>
      </div>

      <div class="setting-group">
        <div class="group-title">{{ t('settings.network.proxy') }}</div>

        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">{{ t('settings.network.httpProxy') }}</div>
            <div class="row-help">{{ t('settings.network.httpProxyHelp') }}</div>
          </div>
          <a-input
            v-model:value="form.proxy"
            placeholder="http://127.0.0.1:7890"
            class="row-control" style="width: 320px"
            allow-clear @change="markDirty"
          />
        </div>

        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">{{ t('settings.network.gitProxy') }}</div>
            <div class="row-help">{{ t('settings.network.gitProxyHelp') }}</div>
          </div>
          <a-input
            v-model:value="form.gitProxy"
            placeholder="socks5://127.0.0.1:1080"
            class="row-control" style="width: 320px"
            allow-clear @change="markDirty"
          />
        </div>
      </div>

      <div class="setting-group">
        <div class="group-title">{{ t('settings.network.timeout') }}</div>

        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">{{ t('settings.network.timeout') }}</div>
            <div class="row-help">{{ t('settings.network.timeoutHelp') }}</div>
          </div>
          <a-input-number
            v-model:value="form.timeout" :min="30" :max="3600" :step="30"
            class="row-control" style="width: 160px" @change="markDirty"
          >
            <template #addonAfter>{{ t('settings.network.seconds') }}</template>
          </a-input-number>
        </div>

        <div class="group-actions">
          <a-button type="primary" :loading="saving" :disabled="!dirty" @click="handleSave">{{ t('common.save') }}</a-button>
          <a-button :loading="saving" :disabled="!proxyActive && !dirty" @click="handleClear">{{ t('settings.network.clearProxy') }}</a-button>
        </div>
      </div>
    </a-spin>
  </div>
</template>

<style scoped>
@import './settings-shared.css';

.proxy-status {
  display: flex; align-items: center; gap: 8px;
  padding: 10px 14px; margin-bottom: var(--spacing-md);
  border-radius: var(--radius-md); font-size: 13px;
}
.proxy-status.on { background: var(--color-success-bg); color: var(--color-success); }
.proxy-status.off { background: var(--color-surface); color: var(--color-text-tertiary); }
.ps-icon { font-size: 15px; }
</style>
