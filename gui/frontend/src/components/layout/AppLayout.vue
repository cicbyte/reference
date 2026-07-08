<script setup>
import { useLayoutStore } from '../../stores/layout'
import Sidebar from './Sidebar.vue'
import Navbar from './Navbar.vue'
import MainContent from './MainContent.vue'

const layout = useLayoutStore()
</script>

<template>
  <div class="app-layout">
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
