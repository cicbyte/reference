<script setup>
import { computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { BranchesOutlined, CheckCircleOutlined, WarningOutlined, ExclamationCircleOutlined } from '@ant-design/icons-vue'
import { useLayoutStore } from '@/stores/layout'
import { useProjectStore } from '@/stores/project'
import { useWallpaperStore } from '@/stores/wallpaper'
import { formatPath } from '@/utils/path'
import ProjectRail from './ProjectRail.vue'
import Sidebar from './Sidebar.vue'
import Navbar from './Navbar.vue'
import MainContent from './MainContent.vue'

const route = useRoute()
const layout = useLayoutStore()
const project = useProjectStore()
const wallpaper = useWallpaperStore()
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

// 壁纸：加载持久化设置 + 拉取壁纸列表 + 注入 CSS 变量（壁纸层/遮罩层见模板）
onMounted(() => {
  wallpaper.init()
})
</script>

<template>
  <div class="app-layout">
    <!-- 壁纸层（z:0）：壁纸图 + 自身模糊 + 微饱和 -->
    <div
      v-if="wallpaper.enabled && wallpaper.hasWallpaper"
      class="wp-layer"
      :style="{ backgroundImage: `url(${wallpaper.activeUrl})`, filter: `blur(${wallpaper.blur}px) saturate(1.1)` }"
    />
    <!-- 遮罩层（z:1）：跟随明暗压色，保证文字可读 -->
    <div
      v-if="wallpaper.enabled && wallpaper.hasWallpaper"
      class="wp-mask"
      :style="{ background: wallpaper.maskColor }"
    />
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
  position: relative;
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
}

/* 壁纸层（最底）+ 遮罩层（跟随明暗压暗/压亮）。
   面板层（project-rail/sidebar/app-right）的 z-index 提升与真毛玻璃由
   derivation.css 全局规则按 data-wallpaper/data-wp-glass 触发（非 scoped 可选子组件根）。 */
.wp-layer,
.wp-mask {
  position: absolute;
  inset: 0;
  pointer-events: none;
}
.wp-layer {
  z-index: 0;
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
}
.wp-mask {
  z-index: 1;
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
