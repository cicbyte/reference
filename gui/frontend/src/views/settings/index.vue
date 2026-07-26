<script setup>
import { ref, computed } from 'vue'
import {
  BgColorsOutlined,
  DatabaseOutlined,
  GlobalOutlined,
  FileTextOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'
import AiSparkIcon from '@/components/common/AiSparkIcon.vue'
import DisplayForm from '@/components/settings/DisplayForm.vue'
import StorageForm from '@/components/settings/StorageForm.vue'
import NetworkForm from '@/components/settings/NetworkForm.vue'
import LoggingForm from '@/components/settings/LoggingForm.vue'
import ProjectForm from '@/components/settings/ProjectForm.vue'
import AboutForm from '@/components/settings/AboutForm.vue'

const { t } = useI18n()

const tabs = computed(() => [
  { key: 'display', icon: BgColorsOutlined, label: t('settings.display.title') },
  { key: 'storage', icon: DatabaseOutlined, label: t('settings.storage.title') },
  { key: 'network', icon: GlobalOutlined, label: t('settings.network.title') },
  { key: 'logging', icon: FileTextOutlined, label: t('settings.logging.title') },
  { key: 'project', icon: AiSparkIcon, label: t('settings.project.title') },
  { key: 'about', icon: InfoCircleOutlined, label: t('settings.about.title') },
])
const activeTab = ref('display')
</script>

<template>
  <div class="settings-view">
    <!-- Left: group menu -->
    <aside class="settings-sider">
      <div class="sider-title">{{ t('sidebar.settings') }}</div>
      <ul class="group-menu">
        <li v-for="tab in tabs" :key="tab.key">
          <a
            class="menu-item"
            :class="{ active: activeTab === tab.key }"
            @click="activeTab = tab.key"
          >
            <component :is="tab.icon" />
            <span>{{ tab.label }}</span>
          </a>
        </li>
      </ul>
    </aside>

    <!-- Right: form content -->
    <div class="settings-content">
      <DisplayForm v-if="activeTab === 'display'" />
      <StorageForm v-else-if="activeTab === 'storage'" />
      <NetworkForm v-else-if="activeTab === 'network'" />
      <LoggingForm v-else-if="activeTab === 'logging'" />
      <ProjectForm v-else-if="activeTab === 'project'" />
      <AboutForm v-else-if="activeTab === 'about'" />
    </div>
  </div>
</template>

<style scoped>
.settings-view {
  display: flex;
  height: calc(100% + 2 * var(--spacing-lg));
  overflow: hidden;
  background: var(--color-background);
  /* 穿透 .content-wrapper 的 padding，让左侧设置列紧贴 header / 菜单栏 / footer */
  margin: calc(-1 * var(--spacing-lg));
}

/* Left group menu */
.settings-sider {
  display: flex;
  flex-direction: column;
  min-width: 180px;
  width: 180px;
  border-right: 1px solid var(--color-border);
  padding: 10px;
  user-select: none;
  flex-shrink: 0;
  background: var(--color-surface);
}

.sider-title {
  padding: 8px 12px 14px;
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-text-tertiary);
}

.group-menu {
  display: flex;
  flex-direction: column;
  list-style: none;
  margin: 0;
  padding: 0;
}

.group-menu li {
  margin-bottom: 2px;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 12px;
  width: 100%;
  text-decoration: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  border-radius: var(--radius-md);
  font-weight: 500;
  font-size: 13px;
  transition: all var(--transition-fast);
}

.menu-item .anticon,
.menu-item svg {
  font-size: 15px;
  opacity: 0.85;
  flex-shrink: 0;
}

.menu-item:hover {
  background: var(--color-hover);
  color: var(--color-text);
}

.menu-item.active {
  background: var(--color-primary-bg);
  color: var(--color-primary);
  font-weight: 600;
}

/* Right content area */
.settings-content {
  flex: 1;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
  padding: 24px 32px;
  background: var(--bg-content);
}
.settings-content::-webkit-scrollbar { width: 8px; }
.settings-content::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 4px; }
</style>
