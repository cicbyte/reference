<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  PlusOutlined,
  SyncOutlined,
  DeleteOutlined,
  BarChartOutlined,
  MedicineBoxOutlined,
  SearchOutlined,
  FolderOpenOutlined,
  CloudDownloadOutlined,
} from '@ant-design/icons-vue'
import { useProjectStore } from '../stores/project'
import AddRepoModal from '../components/repo/AddRepoModal.vue'
import SccModal from '../components/repo/SccModal.vue'
import DiagnoseModal from '../components/repo/DiagnoseModal.vue'

const router = useRouter()
const project = useProjectStore()
const repos = ref([])
const loading = ref(true)
const searchText = ref('')
const addOpen = ref(false)
const sccOpen = ref(false)
const sccRepo = ref('')
const diagnoseOpen = ref(false)

function openScc(name) {
  sccRepo.value = name
  sccOpen.value = true
}

const filteredRepos = computed(() => {
  if (!searchText.value) return repos.value
  const q = searchText.value.toLowerCase()
  return repos.value.filter(r => r.name.toLowerCase().includes(q) || r.source.toLowerCase().includes(q))
})

const columns = [
  { title: '名称', dataIndex: 'name', key: 'name' },
  { title: '类型', dataIndex: 'type', key: 'type', width: 80 },
  { title: '来源', dataIndex: 'source', key: 'source', ellipsis: true },
  { title: '分支', dataIndex: 'branch', key: 'branch', width: 120 },
  { title: '更新时间', dataIndex: 'commit_at', key: 'commit_at', width: 180 },
  { title: '操作', key: 'actions', width: 120, align: 'center', fixed: 'right' },
]

async function loadRepos() {
  if (!project.hasProject) { repos.value = []; loading.value = false; return }
  loading.value = true
  try {
    if (window.go?.main?.ReferenceApp) {
      repos.value = await window.go.main.ReferenceApp.ListRepos()
    }
  } catch (e) {
    console.error('Failed to load repos:', e)
  } finally {
    loading.value = false
  }
}

onMounted(loadRepos)
watch(() => project.projectEpoch, loadRepos)

async function updateRepo(name) {
  try {
    await window.go.main.ReferenceApp.UpdateRepo(name)
    repos.value = await window.go.main.ReferenceApp.ListRepos()
    message.success('更新成功')
  } catch (e) {
    message.error('更新失败: ' + e)
  }
}

async function removeRepo(name) {
  try {
    await window.go.main.ReferenceApp.RemoveRepo(name, false)
    repos.value = await window.go.main.ReferenceApp.ListRepos()
    message.success('已移除')
  } catch (e) {
    message.error('移除失败: ' + e)
  }
}

async function recloneRepo(name) {
  message.loading({ content: `正在重新克隆 ${name}...`, key: 'reclone', duration: 0 })
  try {
    await window.go.main.ReferenceApp.RecloneRepo(name)
    message.success({ content: `${name} 重新克隆成功`, key: 'reclone' })
    repos.value = await window.go.main.ReferenceApp.ListRepos()
  } catch (e) {
    message.error({ content: '重新克隆失败: ' + e, key: 'reclone', duration: 5 })
  }
}
</script>

<template>
  <div class="repo-list">
    <div class="page-header">
      <h2>仓库列表</h2>
      <div class="header-actions">
        <a-input-search
          v-model:value="searchText"
          placeholder="搜索仓库..."
          style="width: 220px"
          allow-clear
        >
          <template #prefix><SearchOutlined /></template>
        </a-input-search>
        <a-button @click="diagnoseOpen = true">
          <template #icon><MedicineBoxOutlined /></template>
          诊断修复
        </a-button>
        <a-button type="primary" @click="addOpen = true">
          <template #icon><PlusOutlined /></template>
          添加仓库
        </a-button>
      </div>
    </div>

    <a-table
      :columns="columns"
      :data-source="filteredRepos"
      :loading="loading"
      :pagination="false"
      :scroll="{ x: 900 }"
      size="middle"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'name'">
          <span>{{ record.name }}</span>
          <a-tag v-if="record.cacheExists === false" color="error" style="margin-left: 6px;">
            缓存缺失
          </a-tag>
        </template>
        <template v-if="column.key === 'type'">
          <a-tag :color="record.type === 'remote' ? 'blue' : 'green'">
            {{ record.type === 'remote' ? '远程' : '本地' }}
          </a-tag>
        </template>
        <template v-if="column.key === 'actions'">
          <a-tooltip v-if="record.cacheExists === false && record.type === 'remote'" title="重新克隆">
            <a-button size="small" type="text" @click="recloneRepo(record.name)">
              <template #icon><CloudDownloadOutlined /></template>
            </a-button>
          </a-tooltip>
          <a-tooltip title="浏览代码">
            <a-button size="small" type="text" :disabled="record.cacheExists === false" @click="router.push('/repos/browse/' + record.name)">
              <template #icon><FolderOpenOutlined /></template>
            </a-button>
          </a-tooltip>
          <a-tooltip title="更新仓库">
            <a-button size="small" type="text" @click="updateRepo(record.name)">
              <template #icon><SyncOutlined /></template>
            </a-button>
          </a-tooltip>
          <a-tooltip title="代码统计">
            <a-button size="small" type="text" :disabled="record.cacheExists === false" @click="openScc(record.name)">
              <template #icon><BarChartOutlined /></template>
            </a-button>
          </a-tooltip>
          <a-popconfirm title="确定移除此仓库引用？" @confirm="removeRepo(record.name)">
            <a-tooltip title="移除引用">
              <a-button size="small" type="text" danger>
                <template #icon><DeleteOutlined /></template>
              </a-button>
            </a-tooltip>
          </a-popconfirm>
        </template>
      </template>
    </a-table>

    <AddRepoModal v-model:open="addOpen" @added="loadRepos" />
    <SccModal v-model:open="sccOpen" :repo-name="sccRepo" />
    <DiagnoseModal v-model:open="diagnoseOpen" :project-dir="project.currentDir" @update:open="diagnoseOpen = $event; loadRepos()" />
  </div>
</template>

<style scoped>
.repo-list { width: 100%; }
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
  flex-wrap: wrap;
  gap: var(--spacing-sm);
}
.page-header h2 { font-size: 20px; font-weight: 600; color: var(--color-text); margin: 0; }
.header-actions { display: flex; gap: var(--spacing-sm); flex-wrap: wrap; }
</style>
