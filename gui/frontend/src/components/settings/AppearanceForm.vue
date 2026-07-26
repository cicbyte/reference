<script setup lang="ts">
/**
 * 外观设置：明暗模式 + 6 套预设 + 自定义色板（基础 token + 状态色）+ 区域覆盖 + 导入导出
 *
 * 对齐 ax 三层架构（移植自 byte-stash）：
 * - 明暗模式：auto / light / dark（auto 跟随系统）
 * - 预设：buildPresetColors 派生（dracula 用 reference 锚点中性色，blue 精调 slate，其余由 bgBase 派生）
 * - 自定义：编辑基础 token + 状态色（深/浅各一套），primary 变化时自动重算派生项
 * - 区域覆盖：5 区域 × (背景/主文字/次文字)，跨预设保留
 * 改动即时生效（可视化拖动只更 CSS 变量；松开/JSON/导入时完整 apply + 持久化）。
 */
import { ref, computed, watch } from 'vue'
import { message } from 'ant-design-vue'
import { DownloadOutlined, ImportOutlined } from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'
import { useThemeStore } from '@/stores/theme'
import ThemePreviewCard from '@/components/settings/ThemePreviewCard.vue'
import { PRESET_LIST, REGION_DEFS, buildPresetColors, deriveBaseLayer } from '@/themes/presets'
import { applyColorsToDOM } from '@/themes/token'
import type {
  ThemeColors,
  BuiltinPreset,
  PresetKey,
  ResolvedMode,
  RegionKey,
  RegionProp,
  RegionOverrides,
  RegionOverrideKey,
  Mode,
} from '@/themes/types'

const { t } = useI18n()
const themeStore = useThemeStore()

// ===== 明暗模式 =====
function onSelectMode(m: Mode) {
  themeStore.setMode(m)
}

function onSelectPreset(key: BuiltinPreset | 'custom') {
  themeStore.setPreset(key as PresetKey)
}

// ===== 自定义色板编辑（基础 token + 状态色；深/浅各一套本地副本） =====
const editTab = ref<ResolvedMode>(themeStore.isDark ? 'dark' : 'light')
const editDark = ref<ThemeColors>({ ...themeStore.custom.dark })
const editLight = ref<ThemeColors>({ ...themeStore.custom.light })

const editColors = computed(() => (editTab.value === 'dark' ? editDark.value : editLight.value))

/** 即时视觉：拖动只更新 CSS 变量（轻量高频；不触发 antd/持久化） */
function applyLive() {
  applyColorsToDOM(editColors.value, editTab.value, themeStore.regions[editTab.value])
}
/** 完整应用：持久化两边 */
function applyCustom() {
  themeStore.setCustomColors('light', { ...editLight.value })
  themeStore.setCustomColors('dark', { ...editDark.value })
}

// 进入自定义时，用 store.custom 初始化本地副本（首次即创建）
watch(
  () => themeStore.preset,
  (p) => {
    if (p !== 'custom') return
    editDark.value = { ...themeStore.custom.dark }
    editLight.value = { ...themeStore.custom.light }
    editTab.value = themeStore.isDark ? 'dark' : 'light'
  },
  { immediate: true },
)

function colorVal(field: keyof ThemeColors): string {
  return (editColors.value as ThemeColors)[field] || '#000000'
}
function setEditColor(field: keyof ThemeColors, value: string) {
  ;(editColors.value as ThemeColors)[field] = value
}
function onColorInput(field: keyof ThemeColors, value: string) {
  setEditColor(field, value)
  applyLive()
}
function onColorCommit() {
  applyCustom()
}
function onHexCommit(field: keyof ThemeColors, value: string) {
  setEditColor(field, value)
  applyCustom()
}

function resetVariant() {
  editDark.value = buildPresetColors('dracula', 'dark')
  editLight.value = buildPresetColors('dracula', 'light')
  applyCustom()
  // 区域覆盖是 custom 的子功能，重置时一并清空（两个变体 + 本地草稿），避免残留
  themeStore.resetRegions('light')
  themeStore.resetRegions('dark')
  regionDraft.value = {}
  message.success(t('settings.display.theme.resetDone'))
}
/** 基于当前编辑变体的页面底色自动派生协调色板（保留主色 + 状态色扩展） */
function onAutoFit() {
  const m = editTab.value
  const cur = { ...editColors.value }
  const next = { ...cur, ...deriveBaseLayer(cur.bgBase, m, cur.primary) } as ThemeColors
  if (m === 'dark') editDark.value = next
  else editLight.value = next
  themeStore.setCustomColors(m, next)
  message.success(t('settings.display.theme.autoFitDone'))
}

// ===== 区域覆盖 =====
const REGION_PROPS: { prop: RegionProp; label: string }[] = [
  { prop: 'bg', label: t('settings.display.theme.regionBg') },
  { prop: 'text', label: t('settings.display.theme.regionText') },
  { prop: 'textSub', label: t('settings.display.theme.regionTextSub') },
]
// 区域覆盖本地草稿：拖动时只更 CSS 变量（轻量），松开才写 store。
// card 区域接入 antdTheme，高频写 store 会触发 antd 全量重渲染导致卡顿，故分离 live/commit。
const regionDraft = ref<RegionOverrides>({ ...themeStore.regions[editTab.value] })
watch([() => themeStore.preset, editTab], () => {
  regionDraft.value = { ...themeStore.regions[editTab.value] }
})
function editTabColors(): ThemeColors {
  return themeStore.preset === 'custom'
    ? { ...themeStore.custom[editTab.value] }
    : buildPresetColors(themeStore.preset as BuiltinPreset, editTab.value)
}
function regionVal(key: RegionKey, prop: RegionProp): string {
  const k = `${key}.${prop}` as RegionOverrideKey
  return regionDraft.value[k] ?? themeStore.regionValue(editTab.value, key, prop)
}
function applyRegionLive() {
  applyColorsToDOM(editTabColors(), editTab.value, regionDraft.value)
}
function onRegionInput(key: RegionKey, prop: RegionProp, val: string) {
  ;(regionDraft.value as Record<string, string>)[`${key}.${prop}`] = val
  applyRegionLive()
}
function onRegionCommit(key: RegionKey, prop: RegionProp, val: string) {
  themeStore.setRegionToken(editTab.value, key, prop, val)
}

// ===== 字段组（可视化；含派生项展示，编辑派生项会被下次 primary 变化重算覆盖） =====
const TOKEN_GROUPS: { label: string; fields: { key: keyof ThemeColors; label: string }[] }[] = [
  { label: t('settings.display.theme.groupPrimary'), fields: [{ key: 'primary', label: t('settings.display.theme.tPrimary') }] },
  {
    label: t('settings.display.theme.groupBg'),
    fields: [
      { key: 'bgBase', label: t('settings.display.theme.tBgBase') },
      { key: 'bgSurface', label: t('settings.display.theme.tBgSurface') },
      { key: 'bgInput', label: t('settings.display.theme.tBgInput') },
      { key: 'bgElevated', label: t('settings.display.theme.tBgElevated') },
    ],
  },
  {
    label: t('settings.display.theme.groupText'),
    fields: [
      { key: 'textPrimary', label: t('settings.display.theme.tTextPrimary') },
      { key: 'textSecondary', label: t('settings.display.theme.tTextSecondary') },
      { key: 'textTertiary', label: t('settings.display.theme.tTextTertiary') },
      { key: 'textDisabled', label: t('settings.display.theme.tTextDisabled') },
    ],
  },
  {
    label: t('settings.display.theme.groupBorder'),
    fields: [
      { key: 'borderBase', label: t('settings.display.theme.tBorderBase') },
      { key: 'borderLight', label: t('settings.display.theme.tBorderLight') },
    ],
  },
  {
    label: t('settings.display.theme.groupPrimaryStates'),
    fields: [
      { key: 'primaryHover', label: t('settings.display.theme.tPrimaryHover') },
      { key: 'primaryActive', label: t('settings.display.theme.tPrimaryActive') },
      { key: 'primaryBorder', label: t('settings.display.theme.tPrimaryBorder') },
    ],
  },
  {
    label: t('settings.display.theme.groupStatus'),
    fields: [
      { key: 'success', label: t('settings.display.theme.tSuccess') },
      { key: 'warning', label: t('settings.display.theme.tWarning') },
      { key: 'danger', label: t('settings.display.theme.tDanger') },
      { key: 'info', label: t('settings.display.theme.tInfo') },
    ],
  },
]

// ===== JSON 代码编辑 + 导入导出 =====
const editorMode = ref<'visual' | 'json'>('visual')
const jsonText = ref('')

function serializeTheme(): string {
  return JSON.stringify(themeStore.exportState(), null, 2)
}
function enterJsonMode() {
  jsonText.value = serializeTheme()
  editorMode.value = 'json'
}
function applyJson() {
  const r = themeStore.applyRawState(jsonText.value)
  if (r.ok) {
    editDark.value = { ...themeStore.custom.dark }
    editLight.value = { ...themeStore.custom.light }
    message.success(t('settings.display.theme.jsonApplied'))
  } else {
    message.error(r.error)
  }
}
function exportToFile() {
  const blob = new Blob([serializeTheme()], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'reference-theme.json'
  a.click()
  URL.revokeObjectURL(url)
}
function onImportFile(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = () => {
    const r = themeStore.applyRawState(reader.result)
    if (r.ok) {
      editDark.value = { ...themeStore.custom.dark }
      editLight.value = { ...themeStore.custom.light }
      message.success(t('settings.display.theme.imported'))
    } else {
      message.error(r.error)
    }
  }
  reader.readAsText(file)
  ;(e.target as HTMLInputElement).value = ''
}
</script>

<template>
  <div class="appearance-form">
    <!-- 明暗模式 -->
    <div class="setting-row">
      <div class="row-label">
        <div class="row-title">{{ t('settings.display.theme.mode') }}</div>
        <div class="row-help">{{ t('settings.display.theme.modeHelp') }}</div>
      </div>
      <a-radio-group :value="themeStore.mode" button-style="solid" @update:value="onSelectMode">
        <a-radio-button value="auto">{{ t('settings.display.theme.modeAuto') }}</a-radio-button>
        <a-radio-button value="dark">{{ t('settings.display.dark') }}</a-radio-button>
        <a-radio-button value="light">{{ t('settings.display.light') }}</a-radio-button>
      </a-radio-group>
    </div>

    <a-divider orientation="left" plain>{{ t('settings.display.theme.scheme') }}</a-divider>
    <div class="theme-grid">
      <ThemePreviewCard
        v-for="p in PRESET_LIST"
        :key="p.key"
        :preset="p.key"
        :label="p.label"
        :is-active="themeStore.preset === p.key"
        @select="onSelectPreset"
      />
      <!-- 自定义卡片（选中后下方展开编辑面板） -->
      <div
        class="custom-preview-card"
        :class="{ active: themeStore.preset === 'custom' }"
        :style="{
          background: themeStore.custom.dark.bgSurface,
          borderColor: themeStore.preset === 'custom'
            ? themeStore.custom.dark.primary
            : themeStore.custom.dark.borderBase,
        }"
        @click="onSelectPreset('custom')"
      >
        <div class="cpc-swatch" :style="{ background: themeStore.custom.dark.primary }" />
        <div class="cpc-body">
          <div class="cpc-name" :style="{ color: themeStore.custom.dark.textPrimary }">
            {{ t('settings.display.theme.custom') }}
          </div>
          <div class="cpc-lines">
            <div class="cpc-line w70" :style="{ background: themeStore.custom.dark.textTertiary }" />
            <div class="cpc-line w50" :style="{ background: themeStore.custom.dark.textDisabled }" />
            <div class="cpc-line w60" :style="{ background: themeStore.custom.dark.primaryBg }" />
          </div>
          <div class="cpc-dots">
            <span class="cpc-dot" :style="{ background: themeStore.custom.dark.success }" />
            <span class="cpc-dot" :style="{ background: themeStore.custom.dark.warning }" />
            <span class="cpc-dot" :style="{ background: themeStore.custom.dark.danger }" />
          </div>
        </div>
      </div>
    </div>

    <!-- 自定义编辑（选中自定义时内联展开，改动即时生效） -->
    <div v-if="themeStore.preset === 'custom'" class="custom-editor">
      <div class="editor-head">
        <span class="editor-title">{{ t('settings.display.theme.customTitle') }}</span>
        <div class="head-actions">
          <a-button size="small" @click="exportToFile">
            <DownloadOutlined /> {{ t('settings.display.theme.export') }}
          </a-button>
          <a-button size="small">
            <ImportOutlined /> {{ t('settings.display.theme.import') }}
            <input class="import-input" type="file" accept="application/json" @change="onImportFile" />
          </a-button>
          <div class="mode-tabs">
            <button class="mode-tab" :class="{ active: editorMode === 'visual' }" @click="editorMode = 'visual'">
              {{ t('settings.display.theme.visual') }}
            </button>
            <button class="mode-tab" :class="{ active: editorMode === 'json' }" @click="enterJsonMode">
              {{ t('settings.display.theme.json') }}
            </button>
          </div>
        </div>
      </div>

      <!-- 可视化编辑 -->
      <template v-if="editorMode === 'visual'">
        <div class="visual-subhead">
          <div class="editor-tabs">
            <button class="editor-tab" :class="{ active: editTab === 'dark' }" @click="editTab = 'dark'">
              {{ t('settings.display.theme.variantDark') }}
            </button>
            <button class="editor-tab" :class="{ active: editTab === 'light' }" @click="editTab = 'light'">
              {{ t('settings.display.theme.variantLight') }}
            </button>
          </div>
          <span class="editor-hint">{{ t('settings.display.theme.liveHint') }}</span>
          <div class="subhead-actions">
            <a-button size="small" @click="onAutoFit">{{ t('settings.display.theme.autoFit') }}</a-button>
            <a-button size="small" @click="resetVariant">{{ t('settings.display.theme.reset') }}</a-button>
          </div>
        </div>

        <div class="editor-body">
          <div v-for="group in TOKEN_GROUPS" :key="group.label" class="color-group">
            <div class="color-group-label">{{ group.label }}</div>
            <div v-for="f in group.fields" :key="f.key" class="color-row">
              <input
                type="color"
                :value="colorVal(f.key)"
                class="color-picker"
                @input="onColorInput(f.key, ($event.target as HTMLInputElement).value)"
                @change="onColorCommit"
              />
              <span class="color-label">{{ f.label }}</span>
              <input
                type="text"
                :value="colorVal(f.key)"
                class="color-hex"
                @change="onHexCommit(f.key, ($event.target as HTMLInputElement).value)"
              />
            </div>
          </div>

          <!-- 区域覆盖 -->
          <div class="color-group">
            <div class="color-group-label">{{ t('settings.display.theme.regionOverride') }}</div>
            <div v-for="reg in REGION_DEFS" :key="reg.key" class="region-block">
              <div class="region-label">{{ reg.label }}</div>
              <div v-for="rp in REGION_PROPS" :key="rp.prop" class="color-row">
                <input
                  type="color"
                  :value="regionVal(reg.key, rp.prop) || '#000000'"
                  class="color-picker"
                  @input="onRegionInput(reg.key, rp.prop, ($event.target as HTMLInputElement).value)"
                  @change="onRegionCommit(reg.key, rp.prop, ($event.target as HTMLInputElement).value)"
                />
                <span class="color-label">{{ rp.label }}</span>
                <input type="text" :value="regionVal(reg.key, rp.prop) || ''" class="color-hex" disabled />
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- JSON 代码编辑（用 textarea 替代 Monaco） -->
      <template v-else>
        <div class="json-wrap">
          <a-textarea v-model:value="jsonText" :rows="18" class="json-textarea" />
          <div class="json-footer">
            <a-button size="small" @click="enterJsonMode">{{ t('settings.display.theme.refresh') }}</a-button>
            <a-button size="small" type="primary" @click="applyJson">
              {{ t('settings.display.theme.apply') }}
            </a-button>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
@import './settings-shared.css';

.appearance-form {
  padding: 8px 4px;
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

/* 明暗模式行 */
.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 10px 0;
}
.row-label {
  flex: 1;
  min-width: 0;
}
.row-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text);
}
.row-help {
  font-size: 12px;
  color: var(--color-text-tertiary);
  margin-top: 2px;
}

.theme-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

/* 自定义卡片（对齐 ThemePreviewCard 尺寸） */
.custom-preview-card {
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
.custom-preview-card:hover {
  transform: translateY(-1px);
}
.custom-preview-card.active {
  box-shadow: 0 0 0 2px rgba(255, 255, 255, 0.1);
}
.cpc-swatch {
  width: 4px;
  border-radius: 2px;
  flex-shrink: 0;
}
.cpc-body {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
  min-width: 0;
}
.cpc-name {
  font-size: 12px;
  font-weight: 600;
}
.cpc-lines {
  display: flex;
  flex-direction: column;
  gap: 3px;
  margin-top: 2px;
}
.cpc-line {
  height: 4px;
  border-radius: 2px;
}
.cpc-line.w70 {
  width: 70%;
}
.cpc-line.w50 {
  width: 50%;
}
.cpc-line.w60 {
  width: 60%;
}
.cpc-dots {
  display: flex;
  gap: 4px;
  margin-top: auto;
}
.cpc-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

/* ===== 自定义编辑面板 ===== */
.custom-editor {
  margin-top: 16px;
  padding: 16px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-light);
  border-radius: 8px;
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}
.editor-head {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
.editor-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}
.head-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}
.mode-tabs {
  display: flex;
  gap: 2px;
}
.mode-tab {
  padding: 4px 12px;
  border: 1px solid var(--border-base);
  border-radius: 4px;
  background: transparent;
  color: var(--text-tertiary);
  font-size: 12px;
  cursor: pointer;
}
.mode-tab.active {
  background: var(--color-primary);
  color: #fff;
  border-color: var(--color-primary);
}

.visual-subhead {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
.editor-tabs {
  display: flex;
  gap: 2px;
}
.editor-tab {
  padding: 4px 12px;
  border: 1px solid var(--border-base);
  border-radius: 4px;
  background: transparent;
  color: var(--text-tertiary);
  font-size: 12px;
  cursor: pointer;
}
.editor-tab.active {
  background: var(--color-primary);
  color: #fff;
  border-color: var(--color-primary);
}
.editor-hint {
  font-size: 11px;
  color: var(--text-tertiary);
}
.subhead-actions {
  display: flex;
  gap: 8px;
  margin-left: auto;
}

.editor-body {
  display: flex;
  flex-direction: column;
  gap: 20px;
  flex: 1;
  overflow-y: auto;
  min-height: 0;
  padding-right: 4px;
}
.color-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.color-group-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.color-row {
  display: flex;
  align-items: center;
  gap: 10px;
}
.color-picker {
  width: 32px;
  height: 32px;
  border: 1px solid var(--border-base);
  border-radius: 4px;
  cursor: pointer;
  background: none;
  padding: 2px;
}
.color-picker::-webkit-color-swatch-wrapper {
  padding: 0;
}
.color-picker::-webkit-color-swatch {
  border: none;
  border-radius: 3px;
}
.color-label {
  flex: 1;
  font-size: 13px;
  color: var(--text-secondary);
}
.color-hex {
  width: 110px;
  border: 1px solid var(--border-light);
  border-radius: 3px;
  padding: 3px 6px;
  font-size: 12px;
  font-family: 'JetBrains Mono', monospace;
  background: var(--bg-input);
  color: var(--text-secondary);
}
.color-hex:focus {
  outline: none;
  border-color: var(--color-primary);
}
.color-hex:disabled {
  opacity: 0.6;
}

.region-block {
  padding: 8px 0 0;
  border-top: 1px dashed var(--border-light);
}
.region-block:first-of-type {
  border-top: none;
}
.region-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  margin-bottom: 4px;
}

.json-wrap {
  display: flex;
  flex-direction: column;
  gap: 12px;
  flex: 1;
  min-height: 0;
}
.json-textarea {
  flex: 1;
  min-height: 240px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
}
.json-footer {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.import-input {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
}
</style>
