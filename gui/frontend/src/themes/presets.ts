/**
 * 主题色板：6 套内置预设（dracula 锚点 + 5 套 byte-stash 派生）+ 1 套自定义。
 *
 * 「好看」的来源（移植自 ax）：
 * - dracula 预设用 reference 现有 Dracula 硬编码值作精调中性色板（DRACULA_NEUTRAL），
 *   保证老用户零视觉跳变；primary = #BD93F9。
 * - blue 预设用精调 slate 中性色板（NEUTRAL），浅/深各一套；
 * - 其余预设只定义 primary + bgBase，由 deriveBaseLayer 派生带色相的协调中性色
 *   （背景/边框/输入框/浮层向白或黑渐变，文字保持中性 slate 保证可读）。
 *
 * byte-stash 扩展：在 ax 基础 12 token 之外补充状态色（success/warning/danger/info）+
 * 禁用文字 + 主色三态 + 阴影，按明暗固定提供，不参与预设色调（按钮/消息始终可辨）。
 *
 * 区域语义层（REGION_DEFS）独立于预设：任何预设下都可叠加区域微调，跨预设保留。
 */
import type { ThemeColors, PresetKey, BuiltinPreset, ResolvedMode, RegionKey } from './types'

/** 基础层 12 token → CSS 变量名（注入 :root 用；token.ts 的 BASE_BG/BASE_DIRECT 为权威映射，此处保留兼容导出） */
export const PALETTE_CSS_VARS: Record<string, string> = {
  primary: '--color-primary',
  primaryBg: '--color-primary-bg',
  colorHover: '--color-hover',
  bgBase: '--bg-base',
  bgSurface: '--bg-surface',
  bgInput: '--bg-input',
  bgElevated: '--bg-elevated',
  textPrimary: '--text-primary',
  textSecondary: '--text-secondary',
  textTertiary: '--text-tertiary',
  borderBase: '--border-base',
  borderLight: '--border-light',
}

/** 扩展层变量名（状态色 / 禁用文字 / 主色三态 / 阴影） */
export const EXT_CSS_VARS: Record<string, string> = {
  textDisabled: '--text-disabled',
  primaryHover: '--color-primary-hover',
  primaryActive: '--color-primary-active',
  primaryBorder: '--color-primary-border',
  success: '--color-success',
  warning: '--color-warning',
  danger: '--color-danger',
  info: '--color-info',
  successBg: '--success-bg',
  warningBg: '--warning-bg',
  dangerBg: '--danger-bg',
  infoBg: '--info-bg',
  shadowMd: '--shadow-md',
}

/** dracula 锚点中性色板（值取自 reference 原 variables.css，保证零视觉跳变） */
const DRACULA_NEUTRAL: Record<
  ResolvedMode,
  Omit<
    ThemeColors,
    | 'primary'
    | 'primaryBg'
    | 'colorHover'
    | 'textDisabled'
    | 'primaryHover'
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
> = {
  dark: {
    bgBase: '#0f172a',
    bgSurface: '#1e293b',
    bgInput: '#334155',
    bgElevated: '#3d4d63',
    textPrimary: '#f1f5f9',
    textSecondary: '#94a3b8',
    textTertiary: '#64748b',
    borderBase: '#3d4d63',
    borderLight: '#2d3e50',
  },
  light: {
    bgBase: '#ffffff',
    bgSurface: '#f8fafc',
    bgInput: '#f1f5f9',
    bgElevated: '#f1f5f9',
    textPrimary: '#0f172a',
    textSecondary: '#64748b',
    textTertiary: '#94a3b8',
    borderBase: '#e2e8f0',
    borderLight: '#f1f5f9',
  },
}

/** blue 预设的精调 slate 中性色板（浅/深各一套）—— 配色协调的根基 */
const NEUTRAL: Record<
  ResolvedMode,
  Omit<
    ThemeColors,
    | 'primary'
    | 'primaryBg'
    | 'colorHover'
    | 'textDisabled'
    | 'primaryHover'
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
> = {
  dark: {
    bgBase: '#0f172a',
    bgSurface: '#1e293b',
    bgInput: '#334155',
    bgElevated: '#273449',
    textPrimary: '#f1f5f9',
    textSecondary: '#94a3b8',
    textTertiary: '#64748b',
    borderBase: '#334155',
    borderLight: '#243049',
  },
  light: {
    bgBase: '#f8fafc',
    bgSurface: '#ffffff',
    bgInput: '#f1f5f9',
    bgElevated: '#f1f5f9',
    textPrimary: '#0f172a',
    textSecondary: '#475569',
    textTertiary: '#94a3b8',
    borderBase: '#cbd5e1',
    borderLight: '#e2e8f0',
  },
}

/** 6 套预设主色 + 底色（浅/深；dracula 用 reference 锚点值，其余与 ax 一致） */
export const PRESETS: Record<
  BuiltinPreset,
  { label: string; primary: Record<ResolvedMode, string>; bgBase: Record<ResolvedMode, string> }
> = {
  dracula: { label: 'Dracula', primary: { light: '#BD93F9', dark: '#BD93F9' }, bgBase: { light: '#ffffff', dark: '#0f172a' } },
  blue: { label: '经典蓝', primary: { light: '#2563eb', dark: '#60a5fa' }, bgBase: { light: '#f8fafc', dark: '#0f172a' } },
  emerald: { label: '翡翠绿', primary: { light: '#059669', dark: '#34d399' }, bgBase: { light: '#f1f6f3', dark: '#0a1612' } },
  violet: { label: '紫罗兰', primary: { light: '#7c3aed', dark: '#a78bfa' }, bgBase: { light: '#f5f2f8', dark: '#140f1d' } },
  amber: { label: '琥珀橙', primary: { light: '#d97706', dark: '#fbbf24' }, bgBase: { light: '#faf6f0', dark: '#1a140a' } },
  rose: { label: '玫瑰红', primary: { light: '#e11d48', dark: '#fb7185' }, bgBase: { light: '#f8f2f4', dark: '#1a0d11' } },
}

/** 预设列表（UI 用，固定顺序；dracula 首位为默认） */
export const PRESET_LIST: { key: BuiltinPreset; label: string }[] = [
  { key: 'dracula', label: 'Dracula' },
  { key: 'blue', label: '经典蓝' },
  { key: 'emerald', label: '翡翠绿' },
  { key: 'violet', label: '紫罗兰' },
  { key: 'amber', label: '琥珀橙' },
  { key: 'rose', label: '玫瑰红' },
]

/** 扩展层固定补充（按明暗；不参与预设色调） */
const SUPPLEMENT: Record<
  ResolvedMode,
  Omit<ThemeColors, keyof typeof NEUTRAL.dark | 'primary' | 'primaryBg' | 'colorHover'>
> = {
  dark: {
    textDisabled: '#475569',
    primaryHover: '', // 由 primary 派生（见 derivePrimaryStates）
    primaryActive: '',
    primaryBorder: '',
    success: '#22c55e',
    warning: '#f59e0b',
    danger: '#ef4444',
    info: '#06b6d4',
    successBg: 'rgba(34,197,94,0.14)',
    warningBg: 'rgba(245,158,11,0.14)',
    dangerBg: 'rgba(239,68,68,0.14)',
    infoBg: 'rgba(6,182,212,0.14)',
    shadowMd: '0 8px 24px rgba(0,0,0,0.4)',
  },
  light: {
    textDisabled: '#cbd5e1',
    primaryHover: '',
    primaryActive: '',
    primaryBorder: '',
    success: '#16a34a',
    warning: '#d97706',
    danger: '#dc2626',
    info: '#0891b2',
    successBg: 'rgba(22,163,74,0.10)',
    warningBg: 'rgba(217,119,6,0.10)',
    dangerBg: 'rgba(220,38,38,0.10)',
    infoBg: 'rgba(8,145,178,0.10)',
    shadowMd: '0 4px 12px rgba(15,23,42,0.08)',
  },
}

// ===== 色彩运算（移植自 ax，逐字对齐） =====

/** hex(#rgb|#rrggbb) → [r,g,b]；非法兜底蓝 */
export function hexToRgb(hex: string): [number, number, number] {
  let h = hex.replace('#', '').trim()
  if (h.length === 3) h = h.split('').map((c) => c + c).join('')
  const n = parseInt(h, 16)
  if (Number.isNaN(n) || h.length !== 6) return [96, 165, 250]
  return [(n >> 16) & 255, (n >> 8) & 255, n & 255]
}

/** primary-bg：primary @ 15% alpha */
export function derivePrimaryBg(primary: string): string {
  const [r, g, b] = hexToRgb(primary)
  return `rgba(${r}, ${g}, ${b}, 0.15)`
}

/** primary-border：primary @ 40% alpha */
export function derivePrimaryBorder(primary: string): string {
  const [r, g, b] = hexToRgb(primary)
  return `rgba(${r}, ${g}, ${b}, 0.4)`
}

/** hover：dark = white@5%，light = black@4%（byte-stash 中性 hover 语义） */
export function deriveHover(mode: ResolvedMode): string {
  return mode === 'dark' ? 'rgba(255, 255, 255, 0.05)' : 'rgba(15, 23, 42, 0.04)'
}

/** dracula 专属 hover：主色 alpha（保持 reference 原 hover 为紫色主调的语义）
 *  dark=8%, light=12%（对齐 reference 原 --color-hover 值） */
function deriveDraculaHover(primary: string, mode: ResolvedMode): string {
  const [r, g, b] = hexToRgb(primary)
  return `rgba(${r}, ${g}, ${b}, ${mode === 'dark' ? 0.08 : 0.12})`
}

/** 主色 hover/active：hover 向白 12%（更亮），active 向黑 12%（更深） */
export function derivePrimaryStates(primary: string): { primaryHover: string; primaryActive: string } {
  return {
    primaryHover: mix(primary, 'white', 0.12),
    primaryActive: mix(primary, 'black', 0.12),
  }
}

/** hex 向白/黑混合 pct（0~1），返回 #rrggbb */
function mix(hex: string, target: 'white' | 'black', pct: number): string {
  const [r, g, b] = hexToRgb(hex)
  const t = target === 'white' ? 255 : 0
  const toHex = (n: number) => Math.round(n).toString(16).padStart(2, '0')
  return `#${toHex(r + (t - r) * pct)}${toHex(g + (t - g) * pct)}${toHex(b + (t - b) * pct)}`
}

/** 基础层派生：由「页面底色 + 明暗」派生协调中性色（保留主色 primary） */
export function deriveBaseLayer(base: string, mode: ResolvedMode, primary: string) {
  const common = { primary, primaryBg: derivePrimaryBg(primary), bgBase: base, colorHover: deriveHover(mode) }
  if (mode === 'dark') {
    return {
      ...common,
      bgSurface: mix(base, 'white', 0.05),
      bgInput: mix(base, 'white', 0.08),
      bgElevated: mix(base, 'white', 0.12),
      borderBase: mix(base, 'white', 0.22),
      borderLight: mix(base, 'white', 0.12),
      textPrimary: '#f1f5f9',
      textSecondary: '#94a3b8',
      textTertiary: '#64748b',
    }
  }
  return {
    ...common,
    bgSurface: mix(base, 'black', 0.03),
    bgInput: mix(base, 'black', 0.05),
    bgElevated: mix(base, 'black', 0.06),
    borderBase: mix(base, 'black', 0.18),
    borderLight: mix(base, 'black', 0.08),
    textPrimary: '#0f172a',
    textSecondary: '#475569',
    textTertiary: '#94a3b8',
  }
}

/** 由预设 + 明暗组装完整 ThemeColors（基础派生 + 扩展补充 + 主色三态） */
export function buildPresetColors(preset: BuiltinPreset, mode: ResolvedMode): ThemeColors {
  const def = PRESETS[preset]
  const primary = def.primary[mode]
  // dracula 用 reference 锚点中性色；blue 沿用精调 slate；其余由带调性底色派生
  const base =
    preset === 'dracula'
      ? DRACULA_NEUTRAL[mode]
      : preset === 'blue'
        ? NEUTRAL[mode]
        : deriveBaseLayer(def.bgBase[mode], mode, primary)
  const sup = SUPPLEMENT[mode]
  const states = derivePrimaryStates(primary)
  return {
    ...base,
    primary,
    primaryBg: derivePrimaryBg(primary),
    // dracula 保留主色 alpha hover 语义；其余预设用 byte-stash 中性 hover
    colorHover: preset === 'dracula' ? deriveDraculaHover(primary, mode) : deriveHover(mode),
    textDisabled: sup.textDisabled,
    primaryHover: states.primaryHover,
    primaryActive: states.primaryActive,
    primaryBorder: derivePrimaryBorder(primary),
    success: sup.success,
    warning: sup.warning,
    danger: sup.danger,
    info: sup.info,
    successBg: sup.successBg,
    warningBg: sup.warningBg,
    dangerBg: sup.dangerBg,
    infoBg: sup.infoBg,
    shadowMd: sup.shadowMd,
  }
}

/** 自定义默认色板：dracula 预设浅/深副本（切到自定义无视觉跳变） */
export function defaultCustom(): { light: ThemeColors; dark: ThemeColors } {
  return { light: buildPresetColors('dracula', 'light'), dark: buildPresetColors('dracula', 'dark') }
}

/** 用户可编辑的基础层 token（派生项与扩展层不开放） */
export const EDITABLE_TOKENS: { group: string; key: keyof ThemeColors; label: string }[] = [
  { group: '主色', key: 'primary', label: '主题色' },
  { group: '背景', key: 'bgBase', label: '页面底色' },
  { group: '背景', key: 'bgSurface', label: '容器底色' },
  { group: '背景', key: 'bgInput', label: '输入框背景' },
  { group: '背景', key: 'bgElevated', label: '浮层背景' },
  { group: '文字', key: 'textPrimary', label: '主文字' },
  { group: '文字', key: 'textSecondary', label: '次要文字' },
  { group: '文字', key: 'textTertiary', label: '弱化文字' },
  { group: '边框', key: 'borderBase', label: '主边框' },
  { group: '边框', key: 'borderLight', label: '弱边框' },
]

// ===== 区域语义层（byte-stash 命名：header/sider/footer/content/card） =====

export interface RegionEntry {
  bg: string
  text: string
  textSub: string
}
export type RegionPalette = Record<RegionKey, RegionEntry>

/** 区域 key → UI 中文名 + 3 个 CSS 变量名（设置页渲染 + apply 注入共用） */
export const REGION_DEFS: {
  key: RegionKey
  label: string
  vars: { bg: string; text: string; textSub: string }
}[] = [
  { key: 'header', label: '顶部栏', vars: { bg: '--bg-header', text: '--text-header', textSub: '--text-header-sub' } },
  { key: 'sider', label: '侧边栏', vars: { bg: '--bg-sider', text: '--text-sider', textSub: '--text-sider-sub' } },
  { key: 'footer', label: '底部栏', vars: { bg: '--bg-footer', text: '--text-footer', textSub: '--text-footer-sub' } },
  { key: 'content', label: '内容区', vars: { bg: '--bg-content', text: '--text-content', textSub: '--text-content-sub' } },
  { key: 'card', label: '卡片/面板', vars: { bg: '--bg-card', text: '--text-card', textSub: '--text-card-sub' } },
]

/** 由某 ThemeColors 派生默认区域配色：chrome 类（顶/侧/底/卡片）用 bgSurface，内容区用 bgBase */
export function defaultRegions(c: ThemeColors): RegionPalette {
  const chrome: RegionEntry = { bg: c.bgSurface, text: c.textPrimary, textSub: c.textSecondary }
  return {
    header: { ...chrome },
    sider: { ...chrome },
    footer: { ...chrome },
    content: { bg: c.bgBase, text: c.textPrimary, textSub: c.textSecondary },
    card: { ...chrome },
  }
}
