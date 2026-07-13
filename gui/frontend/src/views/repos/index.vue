<script setup>
/**
 * Repos view. Left rail (RepoRail) lists the project's repos; the right pane
 * is the shared CodeBrowser, wired to the BrowseRepo* backend API for the
 * currently selected repo.
 *
 * The browser/tree/search/highlight/markdown logic lives entirely in the
 * shared CodeBrowser component — this view only supplies the API adapter and
 * calls `browserRef.loadRoot()` whenever the selected repo changes.
 */
import { ref, onMounted, computed, watch } from 'vue'
import { message } from 'ant-design-vue'
import {
  FolderOpenOutlined,
  FileOutlined,
  CloudDownloadOutlined,
} from '@ant-design/icons-vue'
import { useProjectStore } from '../../stores/project'
import CodeBrowser from '../../components/shared/CodeBrowser.vue'
import AddRepoModal from '../../components/repo/AddRepoModal.vue'
import SccModal from '../../components/repo/SccModal.vue'
import DiagnoseModal from '../../components/repo/DiagnoseModal.vue'
import RepoRail from './components/RepoRail.vue'

const project = useProjectStore()
const app = window.go?.main?.ReferenceApp

// ---- repo list ----
const repos = ref([])
const loading = ref(true)
const selectedRepo = ref('')
const addOpen = ref(false)
const sccOpen = ref(false)
const sccRepo = ref('')
const openSccName = ref('')
const diagnoseOpen = ref(false)

async function loadRepos() {
  if (!project.hasProject) { repos.value = []; loading.value = false; return }
  loading.value = true
  try {
    if (app) repos.value = await app.ListRepos()
  } catch (e) {
    console.error('Failed to load repos:', e)
  } finally {
    loading.value = false
  }
}

onMounted(loadRepos)
watch(() => project.projectEpoch, loadRepos)

const selectedRepoData = computed(() => repos.value.find((r) => r.name === selectedRepo.value))

const browserRef = ref(null)

// Shared CodeBrowser API adapter — delegates to BrowseRepo* backend calls,
// bound to the currently selected repo name.
const browserApi = computed(() => ({
  listDir: (sub) => app.BrowseRepoList(selectedRepo.value, sub),
  readFile: (rel) => app.BrowseRepoRead(selectedRepo.value, rel),
  search: (q) => app.BrowseRepoSearch(selectedRepo.value, q),
}))

function selectRepo(r) {
  selectedRepo.value = r.name
  // The underlying repo key changed; reload the file tree root for it.
  browserRef.value?.loadRoot()
}

// ---- actions ----
async function updateRepo(name) {
  try {
    await app.UpdateRepo(name)
    repos.value = await app.ListRepos()
    message.success('更新成功')
  } catch (e) { message.error('更新失败: ' + e) }
}

async function removeRepo(name) {
  try {
    await app.RemoveRepo(name, false)
    repos.value = await app.ListRepos()
    if (selectedRepo.value === name) {
      selectedRepo.value = ''
      browserRef.value?.clearSelection()
    }
    message.success('已移除')
  } catch (e) { message.error('移除失败: ' + e) }
}

async function recloneRepo(name) {
  message.loading({ content: `正在重新克隆 ${name}...`, key: 'reclone', duration: 0 })
  try {
    const timeout = new Promise((_, reject) =>
      setTimeout(() => reject(new Error('克隆超时')), 180000),
    )
    await Promise.race([app.RecloneRepo(project.currentDir, name), timeout])
    message.success({ content: `${name} 重新克隆成功`, key: 'reclone' })
    repos.value = await app.ListRepos()
  } catch (e) { message.error({ content: '重新克隆失败: ' + e, key: 'reclone', duration: 5 }) }
}
</script>

<template>
  <div class="repo-view">
    <RepoRail
      :repos="repos"
      :loading="loading"
      :selected-repo="selectedRepo"
      @select="selectRepo"
      @update="updateRepo"
      @reclone="recloneRepo"
      @stats="(n) => { openSccName = n; sccOpen = true }"
      @remove="removeRepo"
      @add="addOpen = true"
      @diagnose="diagnoseOpen = true"
    />

    <!-- placeholder: nothing selected -->
    <div v-if="!selectedRepo" class="repo-placeholder">
      <FolderOpenOutlined class="rp-icon" />
      <div>从左侧选择仓库查看代码</div>
    </div>

    <!-- selected repo's cache directory is missing -->
    <div v-else-if="selectedRepoData && selectedRepoData.cacheExists === false" class="repo-missing">
      <div class="rm-icon"><FileOutlined /></div>
      <div class="rm-title">{{ selectedRepo }}</div>
      <div class="rm-hint">缓存目录不存在</div>
      <a-button v-if="selectedRepoData.type === 'remote'" type="primary" @click="recloneRepo(selectedRepo)">
        <CloudDownloadOutlined /> 重新克隆
      </a-button>
    </div>

    <!-- code browser -->
    <CodeBrowser
      v-else
      ref="browserRef"
      :api="browserApi"
      show-back
      :title="selectedRepo"
      @back="selectedRepo = ''; browserRef?.clearSelection()"
    />

    <AddRepoModal v-model:open="addOpen" @added="loadRepos" />
    <SccModal v-model:open="sccOpen" :repo-name="openSccName || sccRepo" />
    <DiagnoseModal v-model:open="diagnoseOpen" :project-dir="project.currentDir" @update:open="diagnoseOpen = $event; loadRepos()" />
  </div>
</template>

<style scoped>
.repo-view { display: flex; height: 100%; width: 100%; overflow: hidden; }

/* placeholder */
.repo-placeholder { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; color: var(--color-text-tertiary); font-size: 14px; }
.rp-icon { font-size: 48px; opacity: 0.2; }

/* missing */
.repo-missing { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; text-align: center; }
.rm-icon { font-size: 48px; color: var(--color-warning); opacity: 0.5; }
.rm-title { font-size: 18px; font-weight: 600; color: var(--color-text); }
.rm-hint { font-size: 14px; color: var(--color-text-tertiary); }
</style>
