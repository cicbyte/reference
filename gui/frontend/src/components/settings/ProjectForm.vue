<script setup>
import { ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import AiSparkIcon from '../common/AiSparkIcon.vue'

const loading = ref(true)
const saving = ref(false)
const agents = ref([])
const selected = ref([])

onMounted(async () => {
  try {
    if (window.go?.main?.ReferenceApp) {
      agents.value = await window.go.main.ReferenceApp.ListAgents()
    }
  } catch (e) {
    console.error('List agents failed:', e)
  } finally {
    loading.value = false
  }
})

async function handleSave() {
  saving.value = true
  try {
    if (window.go?.main?.ReferenceApp) {
      await window.go.main.ReferenceApp.InitProject(selected.value)
      message.success('项目初始化配置已保存')
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
        <div class="group-title"><AiSparkIcon :size="16" /> 编程助手</div>

        <div class="row-help" style="margin-bottom: 12px;">
          选择要在当前项目注入配置的 AI 助手（可多选）。配置会写入
          <code>.reference/reference.settings.json</code>，下次运行时自动注入对应的子代理与 Skill。
        </div>

        <a-checkbox-group v-model:value="selected" style="width: 100%;">
          <div v-for="agent in agents" :key="agent.id" class="agent-row">
            <a-checkbox :value="agent.id">
              <span class="agent-name">{{ agent.display_name }}</span>
              <span class="agent-id mono">{{ agent.id }}</span>
            </a-checkbox>
          </div>
        </a-checkbox-group>

        <div class="group-actions">
          <a-button type="primary" :loading="saving" @click="handleSave">保存</a-button>
        </div>
      </div>
    </a-spin>
  </div>
</template>

<style scoped>
@import './settings-shared.css';

.agent-row {
  padding: 8px 0;
  border-bottom: 1px solid var(--color-border-light);
}
.agent-row:last-child {
  border-bottom: none;
}
.agent-name {
  font-size: 14px;
  color: var(--color-text);
  font-weight: 500;
  margin-right: 8px;
}
.agent-id {
  font-size: 12px;
  color: var(--color-text-tertiary);
}
</style>
