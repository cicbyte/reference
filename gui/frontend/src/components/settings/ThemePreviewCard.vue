<script setup lang="ts">
/**
 * 主题预览卡片（预设版）
 *
 * 接收预设 key，用 buildPresetColors(preset, 'dark') 派生完整色值渲染预览。
 */
import { CheckOutlined } from '@ant-design/icons-vue'
import { computed } from 'vue'
import { buildPresetColors } from '@/themes/presets'
import type { BuiltinPreset } from '@/themes/types'

const props = defineProps<{
  preset: BuiltinPreset
  label: string
  isActive: boolean
}>()

defineEmits<{
  (e: 'select', key: BuiltinPreset): void
}>()

const c = computed(() => buildPresetColors(props.preset, 'dark'))
</script>

<template>
  <div
    class="theme-preview-card"
    :class="{ active: isActive }"
    :style="{
      background: c.bgSurface,
      borderColor: isActive ? c.primary : c.borderBase,
    }"
    @click="$emit('select', preset)"
  >
    <div class="swatch-bar" :style="{ background: c.primary }" />
    <div class="preview-body">
      <div class="preview-title" :style="{ color: c.textPrimary }">{{ label }}</div>
      <div class="preview-lines">
        <div class="line w70" :style="{ background: c.textTertiary }" />
        <div class="line w50" :style="{ background: c.textDisabled }" />
        <div class="line w60" :style="{ background: c.primaryBg }" />
      </div>
      <div class="semantic-dots">
        <span class="dot" :style="{ background: c.success }" />
        <span class="dot" :style="{ background: c.warning }" />
        <span class="dot" :style="{ background: c.danger }" />
        <span class="dot" :style="{ background: c.info }" />
      </div>
    </div>
    <CheckOutlined v-if="isActive" class="check-mark" :style="{ color: c.primary }" />
  </div>
</template>

<style scoped>
.theme-preview-card {
  position: relative;
  display: flex;
  gap: 8px;
  padding: 10px;
  width: 140px;
  height: 90px;
  border: 1px solid;
  border-radius: 8px;
  cursor: pointer;
  transition: transform 150ms ease;
  overflow: hidden;
}
.theme-preview-card:hover {
  transform: translateY(-1px);
}
.theme-preview-card.active {
  box-shadow: 0 0 0 2px rgba(255, 255, 255, 0.1);
}
.swatch-bar {
  width: 4px;
  border-radius: 2px;
  flex-shrink: 0;
}
.preview-body {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
  min-width: 0;
}
.preview-title {
  font-size: 12px;
  font-weight: 600;
}
.preview-lines {
  display: flex;
  flex-direction: column;
  gap: 3px;
  margin-top: 2px;
}
.line {
  height: 4px;
  border-radius: 2px;
}
.line.w70 {
  width: 70%;
}
.line.w50 {
  width: 50%;
}
.line.w60 {
  width: 60%;
}
.semantic-dots {
  display: flex;
  gap: 4px;
  margin-top: auto;
}
.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}
.check-mark {
  position: absolute;
  right: 6px;
  bottom: 4px;
  font-size: 14px;
}
</style>
