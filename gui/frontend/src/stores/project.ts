import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface ProjectItem {
  dir: string
  name: string
  exists: boolean
  initialized: boolean
  agents: string[]
  repoCount: number
  brokenCount: number
}

export interface ProjectInfo {
  dir: string
  name: string
  exists: boolean
}

/**
 * Multi-project state.
 *
 * `currentDir` is the explicitly selected project; project-level views watch
 * `projectEpoch` to reload whenever the user switches / adds / removes a
 * project. Backend methods (ListRepos / AddRepo / SCC …) read the active
 * project from the Go side (set via SwitchProject), so views don't need to
 * pass the dir on every call.
 */
export const useProjectStore = defineStore('project', () => {
  const app = window.go?.main?.ReferenceApp

  const currentDir = ref('')
  const currentName = ref('')
  const currentExists = ref(true)
  const projects = ref<ProjectItem[]>([])
  const loading = ref(false)
  /** Incremented after each switch / add / remove — views watch it to reload. */
  const projectEpoch = ref(0)

  const hasProject = computed(() => currentDir.value !== '')
  const activeProject = computed(() =>
    projects.value.find((p) => p.dir === currentDir.value),
  )

  /** Load the full project list from the DB (GlobalList). */
  async function loadProjects() {
    if (!app?.ListProjects) return
    loading.value = true
    try {
      projects.value = (await app.ListProjects()) || []
    } catch (e) {
      console.error('ListProjects failed:', e)
      projects.value = []
    } finally {
      loading.value = false
    }
  }

  /** Detect the backend's current project (fall-back chain). */
  async function detectCurrent() {
    if (!app?.GetCurrentProject) return
    try {
      const info: ProjectInfo = await app.GetCurrentProject()
      currentDir.value = info?.dir || ''
      currentName.value = info?.name || ''
      currentExists.value = info?.exists ?? true
    } catch {
      // not running in Wails
    }
  }

  /** Switch the active project to `dir`; bumps epoch so views reload. */
  async function switchTo(dir: string): Promise<boolean> {
    if (!app?.SwitchProject || !dir) return false
    try {
      const info: ProjectInfo = await app.SwitchProject(dir)
      currentDir.value = info.dir
      currentName.value = info.name
      currentExists.value = info.exists
      projectEpoch.value++
      return true
    } catch (e) {
      console.error('SwitchProject failed:', e)
      return false
    }
  }

  /** Open a native directory picker; on success switch to it + refresh list. */
  async function pickAndAdd(): Promise<boolean> {
    if (!app?.PickProjectFolder) return false
    try {
      const dir = await app.PickProjectFolder()
      if (!dir) return false // cancelled
      await loadProjects() // the dir may already be in DB; refresh either way
      return await switchTo(dir)
    } catch (e) {
      console.error('PickProjectFolder failed:', e)
      return false
    }
  }

  return {
    currentDir,
    currentName,
    currentExists,
    projects,
    loading,
    projectEpoch,
    hasProject,
    activeProject,
    loadProjects,
    detectCurrent,
    switchTo,
    pickAndAdd,
  }
})
