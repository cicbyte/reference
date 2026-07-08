<script setup>
import { ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { GlobalOutlined } from '@ant-design/icons-vue'

const loading = ref(true)
const saving = ref(false)
const form = ref({ proxy: '', gitProxy: '', timeout: 300 })

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
    <a-spin :spinning="loading">
      <div class="setting-group">
        <div class="group-title"><GlobalOutlined /> 代理与网络</div>

        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">HTTP 代理</div>
            <div class="row-help">HTTP/HTTPS 请求使用的代理地址</div>
          </div>
          <a-input
            v-model:value="form.proxy"
            placeholder="http://127.0.0.1:7890"
            class="row-control"
            style="width: 320px"
            allow-clear
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
            class="row-control"
            style="width: 320px"
            allow-clear
          />
        </div>

        <div class="setting-row">
          <div class="row-label">
            <div class="row-title">超时时间（秒）</div>
            <div class="row-help">克隆 / 拉取操作的超时阈值</div>
          </div>
          <a-input-number
            v-model:value="form.timeout"
            :min="30"
            :max="3600"
            :step="30"
            class="row-control"
            style="width: 160px"
          />
        </div>

        <div class="group-actions">
          <a-button type="primary" :loading="saving" @click="handleSave">保存</a-button>
          <a-button :loading="saving" @click="handleClear">清除代理</a-button>
        </div>
      </div>
    </a-spin>
  </div>
</template>

<style scoped>
@import './settings-shared.css';
</style>
