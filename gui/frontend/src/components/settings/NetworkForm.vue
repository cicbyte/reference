<script setup>
import { ref, onMounted, computed } from 'vue'
import { message } from 'ant-design-vue'
import { GlobalOutlined, CheckCircleFilled, MinusCircleFilled } from '@ant-design/icons-vue'

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
      message.success('网络设置已保存')
    }
  } catch (e) {
    message.error('保存失败: ' + e)
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
      message.success('代理已清除')
    }
  } catch (e) {
    message.error('清除失败: ' + e)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="settings-form">
    <div class="form-header">
      <div class="form-title"><GlobalOutlined /> 网络</div>
      <div class="form-desc">配置代理与超时，用于克隆和拉取远程仓库。</div>
    </div>

    <a-spin :spinning="loading">
      <!-- status banner -->
      <div class="proxy-status" :class="proxyActive ? 'on' : 'off'">
        <CheckCircleFilled v-if="proxyActive" class="ps-icon" />
        <MinusCircleFilled v-else class="ps-icon" />
        <span>{{ proxyActive ? '代理已启用' : '未配置代理，使用直连' }}</span>
      </div>

      <div class="setting-group">
        <div class="group-title">代理</div>

        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">HTTP 代理</div>
            <div class="row-help">HTTP/HTTPS 请求使用的代理地址</div>
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
            <div class="row-title">Git 专用代理</div>
            <div class="row-help">Git 操作专用代理，为空则回退到 HTTP 代理</div>
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
        <div class="group-title">超时</div>

        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">超时时间</div>
            <div class="row-help">克隆 / 拉取操作的超时阈值</div>
          </div>
          <a-input-number
            v-model:value="form.timeout" :min="30" :max="3600" :step="30"
            class="row-control" style="width: 160px" @change="markDirty"
          >
            <template #addonAfter>秒</template>
          </a-input-number>
        </div>

        <div class="group-actions">
          <a-button type="primary" :loading="saving" :disabled="!dirty" @click="handleSave">保存</a-button>
          <a-button :loading="saving" :disabled="!proxyActive && !dirty" @click="handleClear">清除代理</a-button>
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
