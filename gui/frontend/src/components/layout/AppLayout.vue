<script setup>
import { computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { BranchesOutlined, CheckCircleOutlined, WarningOutlined, ExclamationCircleOutlined } from '@ant-design/icons-vue'
import { useLayoutStore } from '@/stores/layout'
import { useProjectStore } from '@/stores/project'
import { formatPath } from '@/utils/path'
import ProjectRail from './ProjectRail.vue'
import Sidebar from './Sidebar.vue'
import Navbar from './Navbar.vue'
import MainContent from './MainContent.vue'

const route = useRoute()
const layout = useLayoutStore()
const project = useProjectStore()
const { t } = useI18n()

// Project rail visibility is fully user-controlled via the Navbar toggle.
const showProjectRail = computed(() => layout.projectRailVisible)

// ---- Project status footer items ----
// Whenever the active project changes (switch / add / remove / diagnose),
// refresh the repo-count, health, and path items.
function syncProjectFooter() {
  const p = project.activeProject
  if (!project.hasProject || !p) {
    layout.clearFooterItem('repo-count')
    layout.clearFooterItem('health')
    layout.clearFooterItem('project-path')
    return
  }
  // repo count with broken-links breakdown
  const broken = p.brokenCount || 0
  const repoValue = broken > 0
    ? `${p.repoCount} (${t('footer.brokenShort', { n: broken })})`
    : String(p.repoCount)
  layout.setFooterItem('repo-count', t('footer.refs'), repoValue, BranchesOutlined, broken > 0 ? 'warn' : 'default')

  // health badge
  if (!p.exists) {
    layout.setFooterItem('health', t('footer.health'), t('footer.dirMissing'), ExclamationCircleOutlined, 'bad')
  } else if (broken > 0) {
    layout.setFooterItem('health', t('footer.health'), t('footer.broken', { n: broken }), WarningOutlined, 'warn')
  } else {
    layout.setFooterItem('health', t('footer.health'), t('footer.healthy'), CheckCircleOutlined, 'ok')
  }

  // project path
  layout.setFooterItem('project-path', t('footer.path'), formatPath(project.currentDir))
}

watch(() => project.projectEpoch, syncProjectFooter, { immediate: true })
watch(() => project.currentDir, syncProjectFooter)

// Clear wiki footer when leaving wiki pages.
watch(() => route.path, (path) => {
  if (!path.startsWith('/wiki') && !path.startsWith('/local-wiki')) {
    layout.clearFooterItem('wiki')
  }
})
</script>

<template>
  <div class="app-layout">
    <ProjectRail :visible="showProjectRail" />
    <Sidebar />
    <div class="app-right">
      <Navbar />
      <MainContent />
    </div>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
}

.app-right {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  /* allow the column to shrink so the footer can pin to the bottom */
  min-height: 0;
}
</style>
