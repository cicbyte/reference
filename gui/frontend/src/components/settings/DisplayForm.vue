<script setup>
import { computed } from 'vue'
import { message } from 'ant-design-vue'
import { useThemeStore } from '../../stores/theme'
import { useLayoutStore } from '../../stores/layout'

const theme = useThemeStore()
const layout = useLayoutStore()

const themeMode = computed({
  get: () => theme.currentTheme,
  set: (v) => theme.setTheme(v),
})

const railVisible = computed({
  get: () => layout.projectRailVisible,
  set: (v) => { layout.projectRailVisible = v },
})

const sidebarDefaultCollapsed = computed({
  get: () => localStorage.getItem('reference-sidebar-collapsed') === 'true',
  set: (v) => {
    localStorage.setItem('reference-sidebar-collapsed', String(v))
    layout.sidebarCollapsed = v
    message.success(v ? '已设为默认折叠' : '已设为默认展开')
  },
})
</script>

<template>
  <div class="settings-form">
    <div class="setting-group">
      <div class="group-title">外观主题</div>

      <div class="setting-row">
        <div class="row-label">
          <div class="row-title">主题模式</div>
          <div class="row-help">选择深色或浅色界面主题</div>
        </div>
        <a-radio-group v-model:value="themeMode" button-style="solid">
          <a-radio-button value="dark">深色</a-radio-button>
          <a-radio-button value="light">浅色</a-radio-button>
        </a-radio-group>
      </div>
    </div>

    <div class="setting-group">
      <div class="group-title">布局</div>

      <div class="setting-row">
        <div class="row-label">
          <div class="row-title">默认显示项目栏</div>
          <div class="row-help">控制启动时是否显示左侧项目滑轨</div>
        </div>
        <a-switch v-model:checked="railVisible" />
      </div>

      <div class="setting-row">
        <div class="row-label">
          <div class="row-title">默认折叠功能侧栏</div>
          <div class="row-help">启动时自动收起功能菜单,仅显示图标</div>
        </div>
        <a-switch v-model:checked="sidebarDefaultCollapsed" />
      </div>
    </div>
  </div>
</template>

<style scoped>
@import './settings-shared.css';
</style>
