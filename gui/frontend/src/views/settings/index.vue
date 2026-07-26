<script setup>
import { ref, computed } from 'vue'
import {
  BgColorsOutlined,
  DatabaseOutlined,
  GlobalOutlined,
  FileTextOutlined,
  InfoCircleOutlined,
  PictureOutlined,
  SettingOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'
import AiSparkIcon from '@/components/common/AiSparkIcon.vue'
import GeneralForm from '@/components/settings/GeneralForm.vue'
import AppearanceForm from '@/components/settings/AppearanceForm.vue'
import WallpaperSection from '@/components/settings/WallpaperSection.vue'
import AnimationForm from '@/components/settings/AnimationForm.vue'
import StorageForm from '@/components/settings/StorageForm.vue'
import NetworkForm from '@/components/settings/NetworkForm.vue'
import LoggingForm from '@/components/settings/LoggingForm.vue'
import ProjectForm from '@/components/settings/ProjectForm.vue'
import AboutForm from '@/components/settings/AboutForm.vue'

const { t } = useI18n()

// 三级设置：分组 → 子分类（仅 display 有，便于后续给其他分组扩展）→ 内容。
// 子分类内联在 tabs；新增子项只需 push 一条 + 在 contentComponent 加分支。
const tabs = computed(() => [
  {
    key: 'display',
    icon: BgColorsOutlined,
    label: t('settings.display.title'),
    subs: [
      { key: 'general', icon: SettingOutlined, label: t('settings.subTabs.general') },
      { key: 'theme', icon: BgColorsOutlined, label: t('settings.subTabs.theme') },
      { key: 'wallpaper', icon: PictureOutlined, label: t('settings.subTabs.wallpaper') },
      { key: 'animation', icon: ThunderboltOutlined, label: t('settings.subTabs.animation') },
    ],
  },
  { key: 'storage', icon: DatabaseOutlined, label: t('settings.storage.title') },
  { key: 'network', icon: GlobalOutlined, label: t('settings.network.title') },
  { key: 'logging', icon: FileTextOutlined, label: t('settings.logging.title') },
  { key: 'project', icon: AiSparkIcon, label: t('settings.project.title') },
  { key: 'about', icon: InfoCircleOutlined, label: t('settings.about.title') },
])

const activeTab = ref('display')
const activeSub = ref('general')

const currentTab = computed(() => tabs.value.find((tb) => tb.key === activeTab.value))
const currentSubs = computed(() => currentTab.value?.subs ?? [])

// 无子分类的分组 → 固定表单映射
const FORM_MAP = {
  storage: StorageForm,
  network: NetworkForm,
  logging: LoggingForm,
  project: ProjectForm,
  about: AboutForm,
}
const contentComponent = computed(() => {
  if (activeTab.value === 'display') {
    if (activeSub.value === 'theme') return AppearanceForm
    if (activeSub.value === 'wallpaper') return WallpaperSection
    if (activeSub.value === 'animation') return AnimationForm
    return GeneralForm
  }
  return FORM_MAP[activeTab.value] || GeneralForm
})
</script>

<template>
  <div class="settings-view">
    <!-- 第一栏：分组菜单 -->
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

    <!-- 第二栏：子分类菜单（仅当当前分组有子项时展开这一列） -->
    <aside v-if="currentSubs.length" class="settings-subsider">
      <div class="subsider-title">{{ currentTab?.label }}</div>
      <ul class="submenu">
        <li v-for="sub in currentSubs" :key="sub.key">
          <a
            class="submenu-item"
            :class="{ active: activeSub === sub.key }"
            @click="activeSub = sub.key"
          >
            <component :is="sub.icon" />
            <span>{{ sub.label }}</span>
          </a>
        </li>
      </ul>
    </aside>

    <!-- 第三栏：表单内容 -->
    <div class="settings-content">
      <component :is="contentComponent" />
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

/* 第一栏：分组菜单 */
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

/* 第二栏：子分类菜单 */
.settings-subsider {
  display: flex;
  flex-direction: column;
  min-width: 150px;
  width: 150px;
  border-right: 1px solid var(--color-border);
  padding: 10px;
  user-select: none;
  flex-shrink: 0;
  background: var(--color-surface);
}
.subsider-title {
  padding: 8px 12px 14px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-text-tertiary);
}
.submenu {
  display: flex;
  flex-direction: column;
  list-style: none;
  margin: 0;
  padding: 0;
}
.submenu li {
  margin-bottom: 2px;
}
.submenu-item {
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 8px 12px;
  width: 100%;
  text-decoration: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  border-radius: var(--radius-md);
  font-weight: 500;
  font-size: 13px;
  transition: all var(--transition-fast);
}
.submenu-item .anticon,
.submenu-item svg {
  font-size: 14px;
  opacity: 0.85;
  flex-shrink: 0;
}
.submenu-item:hover {
  background: var(--color-hover);
  color: var(--color-text);
}
.submenu-item.active {
  background: var(--color-primary-bg);
  color: var(--color-primary);
  font-weight: 600;
}

/* 第三栏：内容 */
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
