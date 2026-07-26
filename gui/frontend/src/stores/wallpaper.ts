/**
 * 壁纸 store（移植自 byte-stash，适配 Wails base64 data URL）
 *
 * 分工：
 * - 壁纸【文件】由 Go 后端管（~/.cicbyte/apps/reference/wallpapers/，见 wallpaper.go）
 * - 壁纸【设置】（开关/选中/模糊/遮罩/面板透明度/面板毛玻璃）持久化在 localStorage
 *   `reference-wallpaper`，纯 UI 偏好，不碰 config——避免死字段。
 *
 * apply() 在 :root 注入两个变量驱动整窗面板毛玻璃半透明：
 * - `--wp-panel-alpha`：区域背景半透明度（derivation.css 用 color-mix 消费，0%=不透明）
 * - `--wp-blur`：模糊量（壁纸层 filter:blur 用；开启面板毛玻璃时 backdrop-filter 也用）
 * 并设 `<html data-wallpaper="on|off">` / `data-wp-glass="on|off">`。
 * 壁纸层/遮罩层在 AppLayout.vue 渲染，遮罩色 maskColor 跟随 theme.resolvedMode。
 *
 * 资源：Wails 无 Tauri asset: 对等协议，壁纸图走 Go 端 base64 data URL
 * （WallpaperDataURL），urlCache 缓存避免重复编码大图。
 */
import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import { useThemeStore } from '@/stores/theme'

const STORAGE_KEY = 'reference-wallpaper'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const app = (window as any)?.go?.main?.ReferenceApp

interface WallpaperState {
  version: 1
  enabled: boolean
  activeName: string | null
  blur: number // 0~40 px
  mask: number // 0~0.8（遮罩 alpha）
  panelAlpha: number // 0~0.7（面板背景 alpha）
  glass: boolean // 面板毛玻璃开关（backdrop-filter）
}

function clamp(v: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, v))
}

export const useWallpaperStore = defineStore('wallpaper', () => {
  const enabled = ref(false)
  const activeName = ref<string | null>(null)
  const blur = ref(0)
  const mask = ref(0.3)
  const panelAlpha = ref(0.4)
  const glass = ref(false)
  const list = ref<string[]>([])
  /** name → base64 data URL 缓存（避免重复编码大图） */
  const urlCache = ref<Record<string, string>>({})

  const theme = useThemeStore()

  /** 某壁纸的 data URL（缓存命中或空串；调 ensureUrl 异步填充） */
  function thumbUrl(name: string): string {
    return urlCache.value[name] ?? ''
  }

  /** 异步获取壁纸 data URL 并缓存（命中直接返回） */
  async function ensureUrl(name: string): Promise<string> {
    if (!name) return ''
    if (urlCache.value[name]) return urlCache.value[name]
    if (!app?.WallpaperDataURL) return ''
    try {
      const url = await app.WallpaperDataURL(name)
      if (url) urlCache.value[name] = url
      return url
    } catch {
      return ''
    }
  }

  /** 当前激活壁纸 URL（enabled && activeName && 已缓存时；AppLayout 据此渲染壁纸层） */
  const activeUrl = computed(() => {
    if (!enabled.value || !activeName.value) return null
    return urlCache.value[activeName.value] ?? null
  })
  /** 是否有可选壁纸——独立于 enabled，避免「enabled=false → activeUrl=null → hasWallpaper=false
   *  → 开关 disabled → 无法重新开启」的循环依赖。仅看是否有选中项（init 已校验文件存在），
   *  用于开关 / 滑块的 disabled 判断；渲染层用 activeUrl。 */
  const hasWallpaper = computed(() => !!activeName.value)

  /** 遮罩层背景色：深色压暗（黑）、浅色压亮（白），跟随当前明暗 */
  const maskColor = computed(() => {
    const rgb = theme.resolvedMode === 'dark' ? '0,0,0' : '255,255,255'
    return `rgba(${rgb}, ${mask.value})`
  })

  /** 注入 CSS 变量 + data-wallpaper/data-wp-glass；禁用或无激活项时清零 */
  function apply() {
    const root = document.documentElement
    const on = enabled.value && !!activeName.value
    if (on) {
      root.style.setProperty('--wp-blur', `${blur.value}px`)
      root.style.setProperty('--wp-panel-alpha', `${Math.round(panelAlpha.value * 100)}%`)
      root.setAttribute('data-wallpaper', 'on')
    } else {
      root.style.setProperty('--wp-blur', '0px')
      root.style.setProperty('--wp-panel-alpha', '0%')
      root.setAttribute('data-wallpaper', 'off')
    }
    root.setAttribute('data-wp-glass', on && glass.value ? 'on' : 'off')
  }

  function persist() {
    const s: WallpaperState = {
      version: 1,
      enabled: enabled.value,
      activeName: activeName.value,
      blur: blur.value,
      mask: mask.value,
      panelAlpha: panelAlpha.value,
      glass: glass.value,
    }
    localStorage.setItem(STORAGE_KEY, JSON.stringify(s))
  }

  function load() {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return
    try {
      const s = JSON.parse(raw) as Partial<WallpaperState>
      enabled.value = !!s.enabled
      activeName.value = s.activeName ?? null
      blur.value = clamp(Number(s.blur ?? 0), 0, 40)
      mask.value = clamp(Number(s.mask ?? 0.3), 0, 0.8)
      panelAlpha.value = clamp(Number(s.panelAlpha ?? 0.4), 0, 0.7)
      glass.value = !!s.glass
    } catch {
      /* 损坏：保持默认 */
    }
  }

  async function refreshList() {
    if (!app?.WallpaperList) {
      list.value = []
      return
    }
    try {
      list.value = await app.WallpaperList()
    } catch {
      list.value = []
    }
    // 批量预取缩略图 data URL（并行；urlCache 更新触发网格重渲染）
    await Promise.all(list.value.map((name) => ensureUrl(name)))
  }

  /** 初始化：load 设置 → 拉列表 → 校验激活项仍存在 → 预取激活 URL → apply */
  async function init() {
    load()
    await refreshList()
    if (activeName.value && !list.value.includes(activeName.value)) {
      activeName.value = null
    }
    if (activeName.value) {
      await ensureUrl(activeName.value)
    }
    apply()
  }

  // ===== setters =====
  function setEnabled(v: boolean) {
    enabled.value = v
  }
  async function setActive(name: string | null) {
    activeName.value = name
    if (name) {
      enabled.value = true
      await ensureUrl(name)
    }
  }
  function setBlur(v: number) {
    blur.value = clamp(v, 0, 40)
  }
  function setMask(v: number) {
    mask.value = clamp(v, 0, 0.8)
  }
  function setPanelAlpha(v: number) {
    panelAlpha.value = clamp(v, 0, 0.7)
  }
  function setGlass(v: boolean) {
    glass.value = v
  }
  function clear() {
    activeName.value = null
  }

  // 任一相关 state 变化 → apply；拖动 slider 时 persist 防抖
  let persistTimer: ReturnType<typeof setTimeout> | null = null
  watch([enabled, activeName, blur, mask, panelAlpha, glass], () => {
    apply()
    if (persistTimer) clearTimeout(persistTimer)
    persistTimer = setTimeout(() => persist(), 300)
  })

  return {
    enabled,
    activeName,
    blur,
    mask,
    panelAlpha,
    glass,
    list,
    activeUrl,
    hasWallpaper,
    maskColor,
    thumbUrl,
    ensureUrl,
    apply,
    persist,
    refreshList,
    init,
    setEnabled,
    setActive,
    setBlur,
    setMask,
    setPanelAlpha,
    setGlass,
    clear,
  }
})
