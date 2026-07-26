/**
 * 主题注入（对齐 ax 三层架构 + byte-stash 变量层半透明）
 *
 * - applyColorsToDOM()：基础 bg 注入 `--bg-*-orig`（原始色），由 :root 的
 *   `--bg-*: color-mix(var(--bg-*-orig), transparent var(--wp-panel-alpha,0%))`
 *   统一派生半透明——一处定义，所有用 var(--bg-*) 的容器自动透出壁纸（全覆盖，
 *   不再依赖逐组件改 background）。非 bg 基础 + 扩展层 + 区域文字直接注入；
 *   区域 bg 也走 -orig（覆盖值 ?? 默认派生）。
 * - buildAntdToken()：生成 Ant Design ConfigProvider token，读同一份 ThemeColors。
 *
 * highlight.js 主题跟随不在此处——hljs 基于 <html data-theme> 属性选择器纯 CSS 切换，
 * 见 derivation.css 的 .hljs 规则。
 */
import { theme as antdTheme } from 'ant-design-vue'
import type { ThemeColors, Appearance, RegionOverrides, RegionOverrideKey } from './types'
import { EXT_CSS_VARS, REGION_DEFS, defaultRegions } from './presets'

/** 基础 bg → -orig（供 :root color-mix 半透明派生） */
const BASE_BG: Record<string, string> = {
  bgBase: '--bg-base-orig',
  bgSurface: '--bg-surface-orig',
  bgElevated: '--bg-elevated-orig',
  bgInput: '--bg-input-orig',
}
/** 基础非 bg → 直接注入（不需半透明） */
const BASE_DIRECT: Record<string, string> = {
  primary: '--color-primary',
  primaryBg: '--color-primary-bg',
  colorHover: '--color-hover',
  textPrimary: '--text-primary',
  textSecondary: '--text-secondary',
  textTertiary: '--text-tertiary',
  borderBase: '--border-base',
  borderLight: '--border-light',
}

/** 把一份 ThemeColors 注入 DOM：基础 bg→-orig + 非bg直接 + 扩展 + 区域 + 明暗属性 */
export function applyColorsToDOM(colors: ThemeColors, appearance: Appearance, regions?: RegionOverrides): void {
  const root = document.documentElement

  // 基础 bg → -orig（:root color-mix 消费，壁纸时 --wp-panel-alpha>0 半透明）
  for (const key in BASE_BG) {
    root.style.setProperty(BASE_BG[key], colors[key as keyof ThemeColors] ?? '')
  }
  // 基础非 bg + 扩展层 → 直接
  for (const key in BASE_DIRECT) {
    root.style.setProperty(BASE_DIRECT[key], colors[key as keyof ThemeColors] ?? '')
  }
  for (const key in EXT_CSS_VARS) {
    root.style.setProperty(EXT_CSS_VARS[key], colors[key as keyof ThemeColors] ?? '')
  }
  root.style.setProperty('--accent', colors.primary)

  // 区域语义层：bg → -orig（用户覆盖值 ?? 默认派生，统一走半透明）；文字 → 直接（覆盖 set / 未覆盖 remove）
  const defs = defaultRegions(colors)
  REGION_DEFS.forEach((d) => {
    const bgOverride = regions ? regions[`${d.key}.bg` as RegionOverrideKey] : undefined
    root.style.setProperty(`${d.vars.bg}-orig`, bgOverride ?? defs[d.key].bg)
    ;(['text', 'textSub'] as const).forEach((prop) => {
      const v = regions ? regions[`${d.key}.${prop}` as RegionOverrideKey] : undefined
      if (v != null) root.style.setProperty(d.vars[prop], v)
      else root.style.removeProperty(d.vars[prop])
    })
  })

  root.setAttribute('data-theme', appearance)
}

/** antd 轨：生成 ConfigProvider token，直接读 ThemeColors，不让 antd algorithm 自派生。
 *  cardBg 可选：Card/Modal 背景跟随 card 区域覆盖（未传则用 bgSurface）。 */
export function buildAntdToken(colors: ThemeColors, appearance: Appearance, cardBg?: string) {
  const isDark = appearance === 'dark'
  const cBg = cardBg ?? colors.bgSurface
  return {
    algorithm: isDark ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
    token: {
      colorPrimary: colors.primary,
      colorBgBase: colors.bgBase,
      colorBgContainer: colors.bgSurface,
      colorBgElevated: colors.bgElevated,
      colorBgSpotlight: colors.bgElevated,
      colorBgLayout: colors.bgBase,
      colorTextBase: colors.textPrimary,
      colorText: colors.textPrimary,
      colorTextSecondary: colors.textSecondary,
      colorTextTertiary: colors.textTertiary,
      colorTextQuaternary: colors.textDisabled,
      colorBorder: colors.borderBase,
      colorBorderSecondary: colors.borderLight,
      colorSuccess: colors.success,
      colorWarning: colors.warning,
      colorError: colors.danger,
      colorInfo: colors.info,
      borderRadius: 8,
      fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif',
      fontSize: 14,
    },
    components: {
      Layout: {
        headerBg: colors.bgSurface,
        siderBg: colors.bgSurface,
        bodyBg: colors.bgBase,
      },
      Menu: {
        itemSelectedBg: colors.primaryBg,
        itemSelectedColor: colors.primary,
      },
      Card: {
        colorBgContainer: cBg,
        borderRadiusLG: 8,
      },
      Modal: {
        contentBg: cBg,
        headerBg: cBg,
        titleColor: colors.textPrimary,
      },
      Dropdown: {
        colorBgElevated: colors.bgElevated,
      },
      Popover: {
        colorBgElevated: colors.bgElevated,
      },
      Input: {
        colorBgContainer: colors.bgInput,
      },
    },
  }
}
