<script setup>
/**
 * Project right-click context menu overlay content.
 *
 * Extracted from the near-identical `<a-menu>` blocks in ProjectRail.vue and
 * GlobalListView.vue. Renders the standard item set (switch / doctor / open /
 * copy / remove / remove+clean); the caller binds the `<a-dropdown>` trigger
 * and passes the project + optional flags.
 *
 * Usage (inside an `<a-dropdown :trigger="['contextmenu']">`):
 *   <template #overlay>
 *     <ProjectContextMenu :project="p" :show-switch="true"
 *       @switch="..." @doctor="..." @open="..." @copy="..." @remove="..." />
 *   </template>
 */
import {
  SwapOutlined, MedicineBoxOutlined, FolderOpenOutlined,
  CopyOutlined, DeleteOutlined,
} from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineProps({
  project: { type: Object, required: true },
  showSwitch: { type: Boolean, default: true },
  showExistsGuard: { type: Boolean, default: false },
})

defineEmits(['switch', 'doctor', 'open', 'copy', 'remove', 'clean'])
</script>

<template>
  <a-menu>
    <a-menu-item v-if="showSwitch" key="switch" @click="$emit('switch', project)">
      <SwapOutlined /> {{ t('globalList.switchToHere') }}
    </a-menu-item>
    <a-menu-divider v-if="showSwitch" />
    <a-menu-item key="doctor" @click="$emit('doctor', project)">
      <MedicineBoxOutlined /> {{ t('globalList.fixLinks') }}
    </a-menu-item>
    <a-menu-item
      v-if="!showExistsGuard || project.exists"
      key="open" @click="$emit('open', project)"
    >
      <FolderOpenOutlined /> {{ t('common.openInExplorer') }}
    </a-menu-item>
    <a-menu-item key="copy" @click="$emit('copy', project)">
      <CopyOutlined /> {{ t('common.copy') }}
    </a-menu-item>
    <a-menu-divider />
    <a-menu-item key="remove" danger @click="$emit('remove', project)">
      <DeleteOutlined /> {{ t('globalList.removeProject') }}
    </a-menu-item>
    <a-menu-item key="clean" danger @click="$emit('clean', project)">
      <DeleteOutlined /> {{ t('globalList.removeClean') }}
    </a-menu-item>
  </a-menu>
</template>
