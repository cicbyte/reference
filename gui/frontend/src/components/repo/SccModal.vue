<script setup>
import { ref, watch } from 'vue'
import { message } from 'ant-design-vue'

const props = defineProps({
  open: { type: Boolean, default: false },
  repoName: { type: String, default: '' },
})
const emit = defineEmits(['update:open'])

const loading = ref(false)
const result = ref(null)

async function runScc() {
  if (!props.repoName) return
  loading.value = true
  result.value = null
  try {
    if (window.go?.main?.ReferenceApp) {
      result.value = await window.go.main.ReferenceApp.RunSCC(props.repoName)
    }
  } catch (e) {
    message.error('统计失败: ' + e)
  } finally {
    loading.value = false
  }
}

// auto-run whenever the modal opens with a repo name
watch(
  () => [props.open, props.repoName],
  ([open]) => {
    if (open && props.repoName) runScc()
    if (!open) result.value = null
  },
)

const columns = [
  { title: '文件', dataIndex: 'file', key: 'file', ellipsis: true },
  { title: '语言', dataIndex: 'language', key: 'language', width: 100 },
  { title: '代码', dataIndex: 'code', key: 'code', width: 90, sorter: (a, b) => a.code - b.code },
  { title: '复杂度', dataIndex: 'complexity', key: 'complexity', width: 90, sorter: (a, b) => a.complexity - b.complexity },
]
</script>

<template>
  <a-modal
    :open="open"
    :title="'代码统计 — ' + repoName"
    :footer="null"
    :mask-closable="true"
    destroy-on-close
    width="780px"
    @update:open="emit('update:open', $event)"
  >
    <a-spin :spinning="loading">
      <div v-if="result" class="scc-body">
        <!-- language stats -->
        <div v-if="result.languages && result.languages.length" class="lang-row">
          <div v-for="lang in result.languages" :key="lang.name" class="lang-chip">
            <div class="lang-name">{{ lang.name }}</div>
            <div class="lang-code">{{ lang.code.toLocaleString() }} 行</div>
            <div class="lang-files">{{ lang.count }} 文件</div>
          </div>
        </div>

        <!-- top files -->
        <div v-if="result.topFiles && result.topFiles.length" class="files-section">
          <div class="files-title">复杂度 Top 文件</div>
          <a-table
            :data-source="result.topFiles"
            :columns="columns"
            :pagination="{ pageSize: 10, size: 'small' }"
            size="small"
            :row-key="(r) => r.file"
          />
        </div>

        <a-empty v-if="!loading && (!result.languages || !result.languages.length)" description="无统计数据" />
      </div>
      <div v-else-if="!loading" class="scc-empty">
        <a-empty description="选择仓库查看代码统计" />
      </div>
    </a-spin>
  </a-modal>
</template>

<style scoped>
.scc-body {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.lang-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-sm);
}
.lang-chip {
  display: flex;
  flex-direction: column;
  padding: 10px 14px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  min-width: 110px;
}
.lang-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
}
.lang-code {
  font-size: 18px;
  font-weight: 700;
  color: var(--color-primary);
  margin-top: 2px;
}
.lang-files {
  font-size: 11px;
  color: var(--color-text-tertiary);
  margin-top: 2px;
}

.files-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-secondary);
  margin-bottom: var(--spacing-sm);
}
</style>
