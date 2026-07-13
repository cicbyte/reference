<script setup>
import { computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useLayoutStore } from '@/stores/layout'
import ProjectRail from './ProjectRail.vue'
import Sidebar from './Sidebar.vue'
import Navbar from './Navbar.vue'
import MainContent from './MainContent.vue'

const route = useRoute()
const layout = useLayoutStore()

// Project rail visibility is fully user-controlled via the Navbar toggle.
const showProjectRail = computed(() => layout.projectRailVisible)

// Footer: wiki file path (set by wiki views, cleared here on navigation).
// Project info is shown in the Sidebar bottom, not the global footer.

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
