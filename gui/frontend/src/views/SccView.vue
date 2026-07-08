<script setup>
import { ref } from 'vue'

const repoName = ref('')
const loading = ref(false)
const result = ref(null)
const repos = ref([])

async function loadRepos() {
  if (window.go?.main?.ReferenceApp) {
    repos.value = await window.go.main.ReferenceApp.ListRepos()
  }
}

async function runScc() {
  if (!repoName.value) return
  loading.value = true
  try {
    if (window.go?.main?.ReferenceApp) {
      result.value = await window.go.main.ReferenceApp.RunSCC(repoName.value)
    }
  } catch (e) {
    console.error('SCC failed:', e)
  } finally {
    loading.value = false
  }
}

loadRepos()
</script>

<template>
  <div class="scc-view">
    <div class="page-header">
      <h2>代码统计</h2>
    </div>

    <div class="scc-form">
      <a-select v-model:value="repoName" placeholder="选择仓库" style="width: 300px" @change="runScc">
        <a-select-option v-for="r in repos" :key="r.name" :value="r.name">{{ r.name }}</a-select-option>
      </a-select>
    </div>

    <a-spin :spinning="loading">
      <div v-if="result" class="scc-results">
        <a-row :gutter="16" class="stats-cards">
          <a-col :span="6" v-for="lang in result.languages" :key="lang.name">
            <a-card size="small">
              <a-statistic :title="lang.name" :value="lang.code" suffix="行" />
            </a-card>
          </a-col>
        </a-row>

        <a-table
          v-if="result.topFiles"
          :data-source="result.topFiles"
          :pagination="{ pageSize: 15 }"
          size="small"
          style="margin-top: 16px"
        >
          <a-table-column title="文件" data-index="file" ellipsis />
          <a-table-column title="语言" data-index="language" width="120" />
          <a-table-column title="代码行数" data-index="code" width="100" sorter="(a,b) => a.code - b.code" />
          <a-table-column title="复杂度" data-index="complexity" width="100" sorter="(a,b) => a.complexity - b.complexity" />
        </a-table>
      </div>
      <a-empty v-else-if="!loading" description="选择仓库查看代码统计" />
    </a-spin>
  </div>
</template>

<style scoped>
.scc-view { max-width: 1200px; }
.page-header { margin-bottom: var(--spacing-lg); }
.page-header h2 { font-size: 20px; font-weight: 600; color: var(--color-text); }
.scc-form { margin-bottom: var(--spacing-lg); }
.stats-cards { margin-bottom: var(--spacing-md); }
</style>
