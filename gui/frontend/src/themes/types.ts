/**
 * 主题类型定义（对齐 ax 三层架构，移植自 byte-stash）
 *
 * 三层：
 * 1. 基础层 ThemeColors：ax 12 token（primary、bg 系列、text 系列、border 系列、colorHover）
 *    + byte-stash 扩展（textDisabled / primaryHover·Active·Border / 状态色 / shadowMd）
 * 2. 区域语义层 RegionOverrides：5 区域 × (bg/主文字/次文字)，独立于预设，跨预设保留
 * 3. 预设/自定义：6 套内置预设（dracula 锚点 + 5 套 byte-stash）+ 1 套自定义（用户可编辑）
 *
 * 区域色不再塞进 ThemeColors，改由区域语义层独立覆盖，默认在 :root 用 var() 回指
 * 基础层——切预设自动跟随，无需 realign。
 *
 * reference 专属：相比 byte-stash 多一套 `dracula` 预设，作为老用户零视觉跳变的锚点。
 */

/** 明暗三态（含 auto） */
export type AppearanceMode = 'light' | 'dark' | 'auto'
/** 解析后的明暗（auto 已消化） */
export type Appearance = 'light' | 'dark'

/** ax 别名（移植自 ax，便于派生算法逐字对齐） */
export type Mode = AppearanceMode
export type ResolvedMode = Appearance

/** 7 套预设 key（dracula 锚点 + 5 套 byte-stash + 自定义） */
export type PresetKey = 'dracula' | 'blue' | 'emerald' | 'violet' | 'amber' | 'rose' | 'custom'
/** 内置预设 key（不含 custom） */
export type BuiltinPreset = Exclude<PresetKey, 'custom'>

// ===== 区域语义层 =====
/** 5 个 UI 区域（byte-stash 命名：header/sider/footer/content/card） */
export type RegionKey = 'header' | 'sider' | 'footer' | 'content' | 'card'
/** 区域属性：背景 / 主文字 / 次要文字 */
export type RegionProp = 'bg' | 'text' | 'textSub'
/** 覆盖表键，如 "header.bg" */
export type RegionOverrideKey = `${RegionKey}.${RegionProp}`
/** 覆盖表：只存用户改过的项，未改的由 :root 默认 var() 回指基础层 */
export type RegionOverrides = Partial<Record<RegionOverrideKey, string>>

/**
 * 完整色值表（单明暗态）。
 * 基础层（ax 12，前 12 个）参与预设派生与 custom 编辑；
 * 扩展层（byte-stash 业务依赖：禁用文字/主色三态/状态色/阴影）按明暗固定补充，
 * 不参与预设色调，保证按钮/消息/标签始终可辨。
 */
export interface ThemeColors {
  // ===== 基础层（ax 12 token，可派生可编辑） =====
  primary: string
  primaryBg: string
  colorHover: string
  bgBase: string
  bgSurface: string
  bgInput: string
  bgElevated: string
  textPrimary: string
  textSecondary: string
  textTertiary: string
  borderBase: string // ax colorBorder
  borderLight: string // ax colorBorderLight
  // ===== 扩展层（byte-stash 业务依赖，固定补充） =====
  textDisabled: string
  primaryHover: string
  primaryActive: string
  primaryBorder: string
  success: string
  warning: string
  danger: string
  info: string
  successBg: string
  warningBg: string
  dangerBg: string
  infoBg: string
  shadowMd: string
}

/** 用户可编辑的基础层 token（派生项 primaryBg/colorHover 不开放） */
export type EditableToken = Exclude<
  keyof ThemeColors,
  | 'primaryBg'
  | 'colorHover'
  | 'textDisabled'
  | 'primaryActive'
  | 'primaryBorder'
  | 'success'
  | 'warning'
  | 'danger'
  | 'info'
  | 'successBg'
  | 'warningBg'
  | 'dangerBg'
  | 'infoBg'
  | 'shadowMd'
>
