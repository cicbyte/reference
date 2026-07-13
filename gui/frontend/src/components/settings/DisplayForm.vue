<script setup>
import { computed } from 'vue'
import { message } from 'ant-design-vue'
import { BgColorsOutlined, GlobalOutlined } from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'
import { useThemeStore } from '@/stores/theme'
import { useLayoutStore } from '@/stores/layout'
import { useLocaleStore } from '@/stores/locale'

const { t } = useI18n()
const theme = useThemeStore()
const layout = useLayoutStore()
const locale = useLocaleStore()

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

      <!-- visual theme picker -->
      <div class="theme-picker">
        <div
          class="theme-card"
          :class="{ active: themeMode === 'dark' }"
          @click="themeMode = 'dark'"
        >
          <div class="theme-preview theme-preview-dark">
            <div class="tp-bar tp-bar-rail"></div>
            <div class="tp-bar tp-bar-side"></div>
            <div class="tp-bar tp-bar-main"></div>
          </div>
          <div class="theme-card-label">{{ t('settings.display.dark') }}</div>
        </div>
        <div
          class="theme-card"
          :class="{ active: themeMode === 'light' }"
          @click="themeMode = 'light'"
        >
          <div class="theme-preview theme-preview-light">
            <div class="tp-bar tp-bar-rail"></div>
            <div class="tp-bar tp-bar-side"></div>
            <div class="tp-bar tp-bar-main"></div>
          </div>
          <div class="theme-card-label">{{ t('settings.display.light') }}</div>
        </div>
      </div>
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

.theme-picker {
  display: flex; gap: 14px; flex-wrap: wrap;
}
.theme-card {
  cursor: pointer; border: 2px solid var(--color-border-light);
  border-radius: var(--radius-md); padding: 8px;
  transition: all var(--transition-fast);
}
.theme-card:hover { border-color: var(--color-primary); }
.theme-card.active {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-bg);
}
.theme-preview {
  width: 150px; height: 92px; border-radius: var(--radius-sm);
  display: flex; overflow: hidden; padding: 8px; gap: 4px;
}
.theme-preview-dark { background: #1e1f2c; }
.theme-preview-light { background: #f5f5f7; }
.tp-bar { border-radius: 2px; height: 100%; }
.tp-bar-rail { width: 20px; }
.tp-bar-side { width: 30px; }
.tp-bar-main { flex: 1; }
.theme-preview-dark .tp-bar { background: #313344; }
.theme-preview-dark .tp-bar-main { background: #262737; }
.theme-preview-light .tp-bar { background: #e4e4e8; }
.theme-preview-light .tp-bar-main { background: #ffffff; border: 1px solid #e4e4e8; }
.theme-card-label {
  text-align: center; font-size: 13px; font-weight: 500;
  color: var(--color-text); margin-top: 8px;
}
</style>
