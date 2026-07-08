<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useLayoutStore } from '../../stores/layout'
import ProjectRail from './ProjectRail.vue'
import Sidebar from './Sidebar.vue'
import Navbar from './Navbar.vue'
import MainContent from './MainContent.vue'

const route = useRoute()
const layout = useLayoutStore()

// Three-column on project-scoped pages, two-column on global pages.
const showProjectRail = computed(() => !!route.meta?.projectScoped)
</script>

<template>
  <div class="app-layout">
    <ProjectRail v-if="showProjectRail" />
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
