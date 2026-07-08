<script setup>
import { ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'

const proxyUrl = ref('')
const currentProxy = ref('')
const loading = ref(false)

onMounted(async () => {
  try {
    if (window.go?.main?.ReferenceApp) {
      const info = await window.go.main.ReferenceApp.GetProxyInfo()
      currentProxy.value = info?.proxy || '未设置'
    }
  } catch (e) {
    console.error('Get proxy failed:', e)
  }
})

async function setProxy() {
  loading.value = true
  try {
    if (window.go?.main?.ReferenceApp) {
      await window.go.main.ReferenceApp.SetProxy(proxyUrl.value)
      currentProxy.value = proxyUrl.value || '未设置'
      message.success('代理设置成功')
    }
  } catch (e) {
    message.error('设置失败: ' + e)
  } finally {
    loading.value = false
  }
}

async function clearProxy() {
  try {
    if (window.go?.main?.ReferenceApp) {
      await window.go.main.ReferenceApp.ClearProxy()
      currentProxy.value = '未设置'
      proxyUrl.value = ''
      message.success('代理已清除')
    }
  } catch (e) {
    message.error('清除失败: ' + e)
  }
}
</script>

<template>
  <div class="proxy-view">
    <div class="page-header"><h2>代理设置</h2></div>
    <a-descriptions bordered :column="1" size="small" style="margin-bottom: 16px; max-width: 500px">
      <a-descriptions-item label="当前代理">{{ currentProxy }}</a-descriptions-item>
    </a-descriptions>
    <a-form layout="inline" @finish="setProxy" style="max-width: 500px">
      <a-form-item>
        <a-input v-model:value="proxyUrl" placeholder="http://proxy:port 或 socks5://proxy:port" style="width: 300px" />
      </a-form-item>
      <a-form-item>
        <a-button type="primary" html-type="submit" :loading="loading">设置</a-button>
      </a-form-item>
      <a-form-item>
        <a-button danger @click="clearProxy">清除</a-button>
      </a-form-item>
    </a-form>
  </div>
</template>

<style scoped>
.proxy-view { max-width: 800px; }
.page-header { margin-bottom: var(--spacing-lg); }
.page-header h2 { font-size: 20px; font-weight: 600; color: var(--color-text); }
</style>
