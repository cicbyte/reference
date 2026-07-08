<script setup>
import { ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'

const agents = ref([])
const selectedAgents = ref([])
const loading = ref(false)
const projectDir = ref('')

onMounted(async () => {
  try {
    if (window.go?.main?.ReferenceApp) {
      agents.value = await window.go.main.ReferenceApp.ListAgents()
    }
  } catch (e) {
    console.error('List agents failed:', e)
  }
})

async function handleInit() {
  loading.value = true
  try {
    if (window.go?.main?.ReferenceApp) {
      await window.go.main.ReferenceApp.InitProject(selectedAgents.value)
      message.success('项目初始化成功')
    }
  } catch (e) {
    message.error('初始化失败: ' + e)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="init-view">
    <div class="page-header"><h2>项目初始化</h2></div>
    <a-form layout="vertical" @finish="handleInit" style="max-width: 500px">
      <a-form-item label="选择编程助手">
        <a-checkbox-group v-model:value="selectedAgents">
          <a-checkbox v-for="agent in agents" :key="agent.id" :value="agent.id">
            {{ agent.display_name }}
          </a-checkbox>
        </a-checkbox-group>
      </a-form-item>
      <a-form-item>
        <a-button type="primary" html-type="submit" :loading="loading">初始化项目</a-button>
      </a-form-item>
    </a-form>
  </div>
</template>

<style scoped>
.init-view { max-width: 800px; }
.page-header { margin-bottom: var(--spacing-lg); }
.page-header h2 { font-size: 20px; font-weight: 600; color: var(--color-text); }
</style>
