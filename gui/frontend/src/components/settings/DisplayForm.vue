<script setup>
import { computed } from 'vue'
import { message } from 'ant-design-vue'
import { BgColorsOutlined, GlobalOutlined } from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'
import AppearanceForm from '@/components/settings/AppearanceForm.vue'
import { useLayoutStore } from '@/stores/layout'
import { useLocaleStore } from '@/stores/locale'

const { t } = useI18n()
const layout = useLayoutStore()
const locale = useLocaleStore()

const railVisible = computed({
  get: () => layout.projectRailVisible,
  set: (v) => { layout.projectRailVisible = v },
})

const sidebarDefaultCollapsed = computed({
  get: () => localStorage.getItem('reference-sidebar-collapsed') === 'true',
  set: (v) => {
    localStorage.setItem('reference-sidebar-collapsed', String(v))
    layout.sidebarCollapsed = v
    message.success(v ? t('settings.display.collapsed') : t('settings.display.expanded'))
  },
})

const currentLocale = computed({
  get: () => locale.currentLocale,
  set: (v) => locale.setLocale(v),
})
</script>

<template>
  <div class="settings-form">
    <div class="form-header">
      <div class="form-title"><BgColorsOutlined /> {{ t('settings.display.title') }}</div>
      <div class="form-desc">{{ t('settings.display.desc') }}</div>
    </div>

    <div class="setting-group">
      <div class="group-title">{{ t('settings.display.appearance') }}</div>
      <AppearanceForm />
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
