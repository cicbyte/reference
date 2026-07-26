<script setup>
/**
 * 常规设置（显示分组的子分类）：界面语言 + 布局偏好
 *
 * 从原 DisplayForm 抽出——显示分组拆为 常规/主题/壁纸 三个子分类后，
 * 语言与布局归入「常规」，主题归 AppearanceForm，壁纸归 WallpaperSection。
 */
import { computed } from 'vue'
import { message } from 'ant-design-vue'
import { GlobalOutlined } from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'
import { useLayoutStore } from '@/stores/layout'
import { useLocaleStore } from '@/stores/locale'

const { t } = useI18n()
const layout = useLayoutStore()
const locale = useLocaleStore()

const currentLocale = computed({
  get: () => locale.currentLocale,
  set: (v) => locale.setLocale(v),
})
const railVisible = computed({
  get: () => layout.projectRailVisible,
  set: (v) => { layout.projectRailVisible = v },
})
const sidebarDefaultCollapsed = computed({
  get: () => layout.sidebarDefaultCollapsed,
  set: (v) => {
    // 仅改启动偏好（响应式 + 持久化），不碰当前运行时折叠状态——运行态由导航栏按钮控制
    layout.setSidebarDefaultCollapsed(v)
    message.success(v ? t('settings.display.collapsed') : t('settings.display.expanded'))
  },
})
</script>

<template>
  <div class="settings-form">
    <div class="form-header">
      <div class="form-title"><GlobalOutlined /> {{ t('settings.subTabs.general') }}</div>
      <div class="form-desc">{{ t('settings.subTabs.generalDesc') }}</div>
    </div>

    <div class="setting-group">
      <div class="group-title">{{ t('settings.display.language') }}</div>
      <div class="setting-row">
        <div class="row-label">
          <div class="row-title"><GlobalOutlined /> {{ t('settings.display.language') }}</div>
          <div class="row-help">{{ t('settings.display.languageHelp') }}</div>
        </div>
        <a-radio-group v-model:value="currentLocale" button-style="solid">
          <a-radio-button value="zh-CN">{{ t('settings.display.zhCN') }}</a-radio-button>
          <a-radio-button value="en">{{ t('settings.display.en') }}</a-radio-button>
        </a-radio-group>
      </div>
    </div>

    <div class="setting-group">
      <div class="group-title">{{ t('settings.display.layout') }}</div>
      <div class="setting-row">
        <div class="row-label">
          <div class="row-title">{{ t('settings.display.showRail') }}</div>
          <div class="row-help">{{ t('settings.display.showRailHelp') }}</div>
        </div>
        <a-switch v-model:checked="railVisible" />
      </div>
      <div class="setting-row">
        <div class="row-label">
          <div class="row-title">{{ t('settings.display.collapseSidebar') }}</div>
          <div class="row-help">{{ t('settings.display.collapseSidebarHelp') }}</div>
        </div>
        <a-switch v-model:checked="sidebarDefaultCollapsed" />
      </div>
    </div>
  </div>
</template>

<style scoped>
@import './settings-shared.css';
</style>
