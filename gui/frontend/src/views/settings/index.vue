<script setup>
import { ref } from 'vue'
import {
  BgColorsOutlined,
  DatabaseOutlined,
  GlobalOutlined,
  FileTextOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons-vue'
import AiSparkIcon from '@/components/common/AiSparkIcon.vue'
import DisplayForm from '@/components/settings/DisplayForm.vue'
import StorageForm from '@/components/settings/StorageForm.vue'
import NetworkForm from '@/components/settings/NetworkForm.vue'
import LoggingForm from '@/components/settings/LoggingForm.vue'
import ProjectForm from '@/components/settings/ProjectForm.vue'
import AboutForm from '@/components/settings/AboutForm.vue'

const tabs = [
  { key: 'display', icon: BgColorsOutlined, label: '显示' },
  { key: 'storage', icon: DatabaseOutlined, label: '存储' },
  { key: 'network', icon: GlobalOutlined, label: '网络' },
  { key: 'logging', icon: FileTextOutlined, label: '日志' },
  { key: 'project', icon: AiSparkIcon, label: 'AI 助手' },
  { key: 'about', icon: InfoCircleOutlined, label: '关于' },
]
const activeTab = ref('display')
</script>

<template>
  <div class="settings-view">
    <!-- Left: group menu -->
    <aside class="settings-sider">
      <div class="sider-title">设置</div>
      <ul class="group-menu">
        <li v-for="t in tabs" :key="t.key">
          <a
            class="menu-item"
            :class="{ active: activeTab === t.key }"
            @click="activeTab = t.key"
          >
            <component :is="t.icon" />
            <span>{{ t.label }}</span>
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
  height: 100%;
  overflow: hidden;
  background: var(--color-background);
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
  background: var(--color-background);
}
.settings-content::-webkit-scrollbar { width: 8px; }
.settings-content::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 4px; }
</style>
