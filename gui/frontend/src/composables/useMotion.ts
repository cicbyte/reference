/**
 * 全局动画入口（GSAP）
 *
 * 统一动画语言，全部参数可由用户在「设置 → 显示 → 动画」调节并持久化：
 *   - enabled：总开关
 *   - bounce：回弹 back.out(1.4) vs 平滑 power2.out
 *   - speed：整体时长缩放（快/正常/慢）
 *   - offset：入场位移幅度（无/轻微/明显）
 *   - stagger：卡片是否错峰入场
 *   - countUp：仪表盘数字是否滚动
 * 尊重 prefers-reduced-motion（系统级，优先于用户设置）。
 *
 * 暴露：
 * - vReveal / vRevealGroup：自定义指令（main.ts 全局注册为 v-reveal / v-reveal-group）
 * - revealOnChange：详情内容切换淡入（watch 依赖 → gsap fromTo 稳定 ref，绝不卸载子组件）
 * - countUp：数字从 0 滚到目标值
 * - useRouteTransition：路由切换过渡的 enter/leave
 * - useMotionPrefs / motionDisabled：偏好状态与禁用判断（设置页、空状态消费）
 *
 * 注意：入场终点用 clearProps:"all" 清除 inline 样式，把 transform/opacity 控制
 * 权交还给 CSS——否则 GSAP 留下的 inline transform 会压住元素已有的 :hover transform。
 *
 * 移植自 byte-stash（1:1），仅 localStorage key 改为 reference-motion。
 */
import gsap from 'gsap'
import { reactive, watch, nextTick, type WatchSource, type DirectiveBinding } from 'vue'

// ===== 动画偏好（持久化）=====
export type MotionSpeed = 'fast' | 'normal' | 'slow'
export type MotionOffset = 'none' | 'light' | 'strong'

export interface MotionPrefs {
  /** 动画总开关：关闭后所有过渡/入场瞬时完成 */
  enabled: boolean
  /** 回弹效果：true=back.out(1.4)，false=平滑 power2.out */
  bounce: boolean
  /** 整体时长缩放档位 */
  speed: MotionSpeed
  /** 入场位移幅度档位 */
  offset: MotionOffset
  /** 卡片是否错峰入场（关闭则同时出现） */
  stagger: boolean
  /** 仪表盘数字是否滚动 */
  countUp: boolean
}

const SPEED_FACTOR: Record<MotionSpeed, number> = { fast: 0.7, normal: 1, slow: 1.35 }
const OFFSET_PX: Record<MotionOffset, number> = { none: 0, light: 8, strong: 14 }

const MOTION_KEY = 'reference-motion'
const DEFAULT_PREFS: MotionPrefs = {
  enabled: true,
  bounce: true,
  speed: 'normal',
  offset: 'strong',
  stagger: true,
  countUp: true,
}

function loadPrefs(): MotionPrefs {
  try {
    return { ...DEFAULT_PREFS, ...JSON.parse(localStorage.getItem(MOTION_KEY) || '{}') }
  } catch {
    return { ...DEFAULT_PREFS }
  }
}

const motionPrefs = reactive<MotionPrefs>(loadPrefs())

/** 偏好变化 → 更新 gsap.defaults（影响后续所有继承默认 ease/duration 的 tween） */
function applyPrefs() {
  gsap.defaults({
    duration: 0.42 * SPEED_FACTOR[motionPrefs.speed],
    ease: motionPrefs.bounce ? 'back.out(1.4)' : 'power2.out',
  })
}
applyPrefs()

watch(motionPrefs, () => {
  localStorage.setItem(MOTION_KEY, JSON.stringify({ ...motionPrefs }))
  applyPrefs()
})

/** 供设置页消费的偏好（reactive，可直接 v-model 绑定） */
export function useMotionPrefs() {
  return { motionPrefs }
}

// ===== 工具 =====
/** 入场终点：复位后 clearProps，让 CSS hover/transform 等接管 */
const REVEAL_TO = { y: 0, autoAlpha: 1, clearProps: 'all' }

function prefersReduced(): boolean {
  return (
    typeof window !== 'undefined' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
  )
}

/** 动画是否被禁用：用户关闭总开关 或 系统开启了「减少动态效果」 */
export function motionDisabled(): boolean {
  return !motionPrefs.enabled || prefersReduced()
}

/** 当前入场位移 px（由偏好 offset 决定；none=0 即纯淡入） */
function revealOffset(): number {
  return OFFSET_PX[motionPrefs.offset]
}

/** 当前时长缩放因子 */
function speedFactor(): number {
  return SPEED_FACTOR[motionPrefs.speed]
}

/** v-reveal：元素入场（淡入 + 上移 + 回弹/平滑）。binding.value = delay（秒） */
export const vReveal = {
  mounted(el: HTMLElement, binding: DirectiveBinding<number | undefined>) {
    if (motionDisabled()) {
      gsap.set(el, { autoAlpha: 1, y: 0 })
      return
    }
    gsap.fromTo(el, { y: revealOffset(), autoAlpha: 0 }, {
      ...REVEAL_TO,
      delay: binding.value ?? 0,
    })
  },
  unmounted(el: HTMLElement) {
    gsap.killTweensOf(el)
  },
}

/** v-reveal-group：容器的直接子元素依次入场。binding.value = stagger（秒，默认 0.06） */
export const vRevealGroup = {
  mounted(el: HTMLElement, binding: DirectiveBinding<number | undefined>) {
    const targets = el.children
    if (!targets.length) return
    if (motionDisabled()) {
      gsap.set(targets, { autoAlpha: 1, y: 0 })
      return
    }
    // stagger 偏好关 → 0（同时出现）
    const stagger = motionPrefs.stagger ? binding.value ?? 0.06 : 0
    gsap.fromTo(targets, { y: revealOffset(), autoAlpha: 0 }, { ...REVEAL_TO, stagger })
  },
  unmounted(el: HTMLElement) {
    gsap.killTweensOf(el.children)
  },
}

/**
 * 详情内容切换淡入：监听依赖（如 selected.name）变化，对新内容做 fromTo。
 * 关键：只动外层容器的 opacity+y，不卸载任何子组件（保持实例，避免重型组件重建）。
 */
export function revealOnChange(
  getEl: () => Element | null | undefined,
  dep: WatchSource,
): void {
  watch(dep, () => {
    nextTick(() => {
      const el = getEl()
      if (!el) return
      if (motionDisabled()) {
        gsap.set(el, { autoAlpha: 1, y: 0 })
        return
      }
      gsap.fromTo(
        el,
        { y: revealOffset(), autoAlpha: 0 },
        { y: 0, autoAlpha: 1, duration: 0.34 * speedFactor(), clearProps: 'all' },
      )
    })
  })
}

/** 数字从 0 滚到目标值（受 countUp 偏好与 speed 影响；ease 固定 power2.out） */
export function countUp(
  el: HTMLElement,
  to: number,
  formatter: (n: number) => string = (n) => String(Math.round(n)),
): void {
  if (motionDisabled() || !motionPrefs.countUp) {
    el.textContent = formatter(to)
    return
  }
  // 先归零，避免数据就绪瞬间 DOM 显示真值、再被 countUp 跳回 0 的闪烁
  el.textContent = formatter(0)
  const obj = { v: 0 }
  gsap.to(obj, {
    v: to,
    duration: 0.8 * speedFactor(),
    ease: 'power2.out',
    onUpdate: () => {
      el.textContent = formatter(obj.v)
    },
  })
}

/** 路由切换过渡：enter 弹入，leave 快速淡出。用于 <Transition :css="false"> */
export function useRouteTransition() {
  const onEnter = (el: Element, done: () => void) => {
    if (motionDisabled()) {
      gsap.set(el, { autoAlpha: 1, y: 0 })
      done()
      return
    }
    gsap.fromTo(el, { y: revealOffset(), autoAlpha: 0 }, { ...REVEAL_TO, onComplete: done })
  }
  const onLeave = (el: Element, done: () => void) => {
    if (motionDisabled()) {
      done()
      return
    }
    gsap.to(el, {
      autoAlpha: 0,
      y: -8,
      duration: 0.15 * speedFactor(),
      ease: 'power2.in',
      onComplete: done,
    })
  }
  return { onEnter, onLeave }
}
