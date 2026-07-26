/**
 * 主题 store（对齐 ax 三层架构，移植自 byte-stash）
 *
 * 状态：preset（dracula 锚点 + 5 套 byte-stash + custom）/ mode（auto·light·dark）/
 *       custom（自定义浅深色板）/ regions（区域覆盖表，浅深各一套，跨预设保留）
 * apply()：注入基础层 + 扩展层 + 区域覆盖 CSS 变量 + antd ConfigProvider。
 *
 * 「好看」来自 presets.ts 的派生：dracula 用 reference 锚点中性色，blue 精调 slate，
 * 其余由 bgBase 派生协调中性色。偏好纯前端持久化（localStorage），不碰 config——避免死字段。
 *
 * 双 API：新三层 API（preset/mode/custom/regions/palette...）+ 旧兼容 API
 * （isDark/currentTheme/setTheme/toggleTheme/initializeTheme），reference 现有 5 个消费方零改动。
 *
 * 迁移：reference 旧 localStorage 是 plain string 'dark'/'light'，normalizeState 三分支处理。
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  ThemeColors,
  PresetKey,
  BuiltinPreset,
  Mode,
  ResolvedMode,
  RegionKey,
  RegionProp,
  RegionOverrideKey,
  RegionOverrides,
} from '@/themes/types'
import {
  buildPresetColors,
  defaultCustom,
  defaultRegions,
  deriveBaseLayer,
  deriveHover,
  derivePrimaryBg,
  derivePrimaryBorder,
  derivePrimaryStates,
} from '@/themes/presets'
import { applyColorsToDOM, buildAntdToken } from '@/themes/token'

const STORAGE_KEY = 'reference-theme'
const THEME_STATE_VERSION = 4

interface ThemeState {
  version: number
  preset: PresetKey
  mode: Mode
  custom: { light: ThemeColors; dark: ThemeColors }
  regions: { light: RegionOverrides; dark: RegionOverrides }
}

function isPresetKey(v: unknown): v is PresetKey {
  return (
    v === 'dracula' ||
    v === 'blue' ||
    v === 'emerald' ||
    v === 'violet' ||
    v === 'amber' ||
    v === 'rose' ||
    v === 'custom'
  )
}
function isMode(v: unknown): v is Mode {
  return v === 'auto' || v === 'light' || v === 'dark'
}
function isValidRegions(v: unknown): v is { light: RegionOverrides; dark: RegionOverrides } {
  if (!v || typeof v !== 'object') return false
  const o = v as Record<string, unknown>
  return (['light', 'dark'] as const).every((m) => {
    const mo = o[m]
    if (!mo || typeof mo !== 'object') return false
    return Object.values(mo).every((val) => typeof val === 'string')
  })
}

export const useThemeStore = defineStore('theme', () => {
  // ===== state =====
  const preset = ref<PresetKey>('dracula')
  const mode = ref<Mode>('auto')
  const custom = ref<{ light: ThemeColors; dark: ThemeColors }>(defaultCustom())
  const regions = ref<{ light: RegionOverrides; dark: RegionOverrides }>({ light: {}, dark: {} })
  /** 旧 API 兼容：系统明暗（detectSystemTheme 填充） */
  const systemTheme = ref<ResolvedMode>('dark')
  let mediaListener: ((e: MediaQueryListEvent) => void) | null = null

  // ===== getters =====
  function resolveMode(m: Mode = mode.value): ResolvedMode {
    if (m !== 'auto') return m
    return window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }
  const resolvedMode = computed<ResolvedMode>(() => resolveMode())
  const isDark = computed(() => resolvedMode.value === 'dark')
  /** 旧 API 兼容：currentTheme 返回 'dark'/'light'（= resolvedMode） */
  const currentTheme = computed<ResolvedMode>(() => resolvedMode.value)

  /** 当前生效色板：custom → custom[mode]；预设 → buildPresetColors。展开以触发深层响应 */
  function currentColors(m: ResolvedMode = resolvedMode.value): ThemeColors {
    const src =
      preset.value === 'custom' ? custom.value[m] : buildPresetColors(preset.value as BuiltinPreset, m)
    return { ...src }
  }
  const palette = computed<ThemeColors>(() => currentColors())

  /** antd ConfigProvider token */
  const antdTheme = computed(() => buildAntdToken(palette.value, resolvedMode.value))

  // ===== actions =====
  function apply() {
    applyColorsToDOM(palette.value, resolvedMode.value, regions.value[resolvedMode.value])
    // highlight.js 主题跟随纯 CSS（derivation.css 基于 [data-theme] 切换），无需 JS
  }

  function setPreset(k: PresetKey) {
    preset.value = k
    apply()
    persist()
  }
  function setMode(m: Mode) {
    mode.value = m
    apply()
    persist()
  }
  /** 自定义：整体替换某变体色板（编辑器松开/JSON 应用时调） */
  function setCustomColors(m: ResolvedMode, colors: ThemeColors) {
    custom.value[m] = { ...colors }
    apply()
    persist()
  }
  /** 自定义：改单个 token（拖动即时 apply；primary 变化时重算派生项） */
  function setCustomToken(m: ResolvedMode, key: keyof ThemeColors, val: string) {
    const c = custom.value[m]
    c[key] = val
    if (key === 'primary') {
      c.primaryBg = derivePrimaryBg(val)
      c.primaryBorder = derivePrimaryBorder(val)
      const st = derivePrimaryStates(val)
      c.primaryHover = st.primaryHover
      c.primaryActive = st.primaryActive
    }
    c.colorHover = deriveHover(m)
    apply()
  }
  /** 自定义：基于该变体的页面底色自动派生协调色板（保留主色） */
  function autoFit(m: ResolvedMode) {
    const cur = custom.value[m]
    custom.value[m] = { ...cur, ...deriveBaseLayer(cur.bgBase, m, cur.primary) } as ThemeColors
    apply()
    persist()
  }
  function resetVariant(m: ResolvedMode) {
    custom.value[m] = buildPresetColors('dracula', m)
    apply()
    persist()
  }
  /** 区域覆盖：改某区域某属性 */
  function setRegionToken(m: ResolvedMode, key: RegionKey, prop: RegionProp, val: string) {
    regions.value[m][`${key}.${prop}` as RegionOverrideKey] = val
    apply()
    persist()
  }
  function resetRegions(m: ResolvedMode) {
    regions.value[m] = {}
    apply()
    persist()
  }
  /** 区域某属性当前生效值：覆盖值 ?? 当前色板派生默认（供 UI 显示） */
  function regionValue(m: ResolvedMode, key: RegionKey, prop: RegionProp): string {
    const k = `${key}.${prop}` as RegionOverrideKey
    return regions.value[m][k] ?? defaultRegions(currentColors(m))[key][prop]
  }

  function persist() {
    const state: ThemeState = {
      version: THEME_STATE_VERSION,
      preset: preset.value,
      mode: mode.value,
      custom: custom.value,
      regions: regions.value,
    }
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state))
  }

  function exportState(): ThemeState {
    return {
      version: THEME_STATE_VERSION,
      preset: preset.value,
      mode: mode.value,
      custom: { light: { ...custom.value.light }, dark: { ...custom.value.dark } },
      regions: { light: { ...regions.value.light }, dark: { ...regions.value.dark } },
    }
  }

  /** 应用一份原始主题（来自编辑器 / 导入文件）。失败返回错误，store 不变 */
  function applyRawState(raw: unknown): { ok: true } | { ok: false; error: string } {
    let parsed: unknown
    try {
      parsed = typeof raw === 'string' ? JSON.parse(raw) : raw
    } catch (e) {
      return { ok: false, error: 'JSON 语法错误：' + String(e) }
    }
    if (!parsed || typeof parsed !== 'object' || !Object.prototype.hasOwnProperty.call(parsed, 'preset')) {
      return { ok: false, error: '不是有效的主题 JSON（缺少 preset 字段）' }
    }
    const s = normalizeState(parsed)
    preset.value = s.preset
    mode.value = s.mode
    custom.value = s.custom
    regions.value = s.regions
    persist()
    apply()
    return { ok: true }
  }

  function freshState(): ThemeState {
    return {
      version: THEME_STATE_VERSION,
      preset: 'dracula',
      mode: 'auto',
      custom: defaultCustom(),
      regions: { light: {}, dark: {} },
    }
  }

  /** 任意 parsed → 合法 ThemeState（枚举校验 + 缺字段兜底）。loadMigrate 与 applyRawState 共用 */
  function normalizeState(parsed: unknown): ThemeState {
    // reference 旧 plain string 'dark'/'light'（localStorage 原值未 JSON 包装）
    if (typeof parsed === 'string') {
      if (parsed === 'dark' || parsed === 'light') {
        return {
          version: THEME_STATE_VERSION,
          preset: 'dracula',
          mode: parsed,
          custom: defaultCustom(),
          regions: { light: {}, dark: {} },
        }
      }
      return freshState()
    }
    // 新结构 {preset, mode, custom, regions}
    if (parsed && typeof parsed === 'object' && Object.prototype.hasOwnProperty.call(parsed, 'preset')) {
      const s = parsed as Partial<ThemeState>
      const presetKey = isPresetKey(s.preset) ? s.preset : 'dracula'
      return {
        version: THEME_STATE_VERSION,
        preset: presetKey,
        mode: isMode(s.mode) ? s.mode : 'auto',
        custom: s.custom && s.custom.light && s.custom.dark ? s.custom : defaultCustom(),
        regions: isValidRegions(s.regions) ? s.regions : { light: {}, dark: {} },
      }
    }
    return freshState()
  }

  /** 读 + 迁移 localStorage（plain string / 新结构 → v4） */
  function loadAndMigrate(): ThemeState {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return freshState()
    let parsed: unknown
    try {
      parsed = JSON.parse(raw)
    } catch {
      parsed = raw // plain string（reference 旧值 'dark'/'light' 无 JSON 包装）
    }
    return normalizeState(parsed)
  }

  /** 旧 API 兼容：探测系统明暗 */
  function detectSystemTheme() {
    systemTheme.value =
      window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }

  /**
   * 初始化：FOUC 内联脚本已在 CSS 解析前注入首帧变量；此处从 localStorage 权威迁移并 apply。幂等。
   */
  function initialize() {
    detectSystemTheme()
    const s = loadAndMigrate()
    preset.value = s.preset
    mode.value = s.mode
    custom.value = s.custom
    regions.value = isValidRegions(s.regions) ? s.regions : { light: {}, dark: {} }
    persist()
    apply()

    if (window.matchMedia) {
      const mq = window.matchMedia('(prefers-color-scheme: dark)')
      if (mediaListener) mq.removeEventListener('change', mediaListener)
      mediaListener = () => {
        detectSystemTheme()
        if (mode.value === 'auto') apply()
      }
      mq.addEventListener('change', mediaListener)
    }
  }

  // ===== 旧 API 兼容（reference 现有消费方零改动） =====
  /** 旧 API：setTheme('dark'|'light') → 内部转 setMode */
  function setTheme(t: string) {
    if (t === 'dark' || t === 'light') setMode(t)
  }
  /** 旧 API：切换明暗 */
  function toggleTheme() {
    setMode(resolvedMode.value === 'dark' ? 'light' : 'dark')
  }
  /** 旧 API：转发到 initialize（App.vue onMounted 调用点不变） */
  function initializeTheme() {
    initialize()
  }

  return {
    // state
    preset,
    mode,
    custom,
    regions,
    systemTheme,
    // getters
    resolvedMode,
    isDark,
    currentTheme,
    palette,
    antdTheme,
    // actions（新三层）
    apply,
    setPreset,
    setMode,
    setCustomColors,
    setCustomToken,
    autoFit,
    resetVariant,
    setRegionToken,
    resetRegions,
    regionValue,
    persist,
    exportState,
    applyRawState,
    initialize,
    // 旧 API 兼容
    setTheme,
    toggleTheme,
    detectSystemTheme,
    initializeTheme,
  }
})
