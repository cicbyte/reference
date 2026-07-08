<script setup>
import { ref } from 'vue'
import {
  SettingOutlined,
  GlobalOutlined,
  DatabaseOutlined,
  RobotOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons-vue'
import GeneralForm from '../components/settings/GeneralForm.vue'
import NetworkForm from '../components/settings/NetworkForm.vue'
import StorageForm from '../components/settings/StorageForm.vue'
import ProjectForm from '../components/settings/ProjectForm.vue'
import AboutForm from '../components/settings/AboutForm.vue'

const tabs = [
  { key: 'general', icon: SettingOutlined, label: '通用' },
  { key: 'storage', icon: DatabaseOutlined, label: '存储' },
  { key: 'network', icon: GlobalOutlined, label: '网络' },
  { key: 'project', icon: RobotOutlined, label: '项目初始化' },
  { key: 'about', icon: InfoCircleOutlined, label: '关于' },
]
const activeTab = ref('general')
</script>

<template>
  <div class="settings-view">
    <!-- Left: group menu -->
    <aside class="settings-sider">
      <div class="sider-title">
        <SettingOutlined />
        <span>设置</span>
      </div>
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
      <GeneralForm v-if="activeTab === 'general'" />
      <StorageForm v-else-if="activeTab === 'storage'" />
      <NetworkForm v-else-if="activeTab === 'network'" />
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
  min-width: 200px;
  width: 200px;
  border-right: 1px solid var(--color-border);
  padding: 10px;
  user-select: none;
  flex-shrink: 0;
  background: var(--color-surface);
}

.sider-title {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px 14px;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

.group-menu {
  display: flex;
  flex-direction: column;
  list-style: none;
  margin: 0;
  padding: 0;
}

.group-menu li {
  margin-bottom: 4px;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  width: 100%;
  text-decoration: none;
  color: var(--color-text);
  cursor: pointer;
  border-radius: var(--radius-md);
  font-weight: 500;
  font-size: 14px;
  border: 1px solid transparent;
  transition: all var(--transition-fast);
}

.menu-item .anticon {
  font-size: 15px;
  opacity: 0.85;
}

.menu-item:hover {
  background: var(--color-hover);
  color: var(--color-primary);
}

.menu-item.active {
  background: var(--color-primary-bg);
  color: var(--color-primary);
  border-color: var(--color-border);
}

/* Right content area */
.settings-content {
  flex: 1;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
  padding: 20px 24px;
  background: var(--color-background);
}

@media (max-width: 768px) {
  .settings-sider {
    min-width: 64px;
    width: 64px;
  }
  .sider-title span,
  .menu-item span {
    display: none;
  }
  .menu-item {
    justify-content: center;
  }
}
</style>
