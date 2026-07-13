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
      <SwapOutlined /> 切换到此项目
    </a-menu-item>
    <a-menu-divider v-if="showSwitch" />
    <a-menu-item key="doctor" @click="$emit('doctor', project)">
      <MedicineBoxOutlined /> 修复断裂链接
    </a-menu-item>
    <a-menu-item
      v-if="!showExistsGuard || project.exists"
      key="open" @click="$emit('open', project)"
    >
      <FolderOpenOutlined /> 在文件管理器中打开
    </a-menu-item>
    <a-menu-item key="copy" @click="$emit('copy', project)">
      <CopyOutlined /> 复制路径
    </a-menu-item>
    <a-menu-divider />
    <a-menu-item key="remove" danger @click="$emit('remove', project)">
      <DeleteOutlined /> 移除项目
    </a-menu-item>
    <a-menu-item key="clean" danger @click="$emit('clean', project)">
      <DeleteOutlined /> 移除并清除 .reference
    </a-menu-item>
  </a-menu>
</template>
