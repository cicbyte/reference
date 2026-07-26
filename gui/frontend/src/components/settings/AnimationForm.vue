<script setup lang="ts">
/**
 * 动画设置（显示分组 · 动画子分类）
 * 偏好状态与持久化都在 useMotion（localStorage reference-motion），这里绑定其
 * reactive motionPrefs——改动即时生效（gsap.defaults 下发）并写入。
 * 速度/位移用分段控件，其余用开关；尊重系统「减少动态效果」。
 */
import { computed } from 'vue'
import { ThunderboltOutlined } from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'
import { useMotionPrefs, type MotionSpeed, type MotionOffset } from '@/composables/useMotion'

const { t } = useI18n()
const { motionPrefs } = useMotionPrefs()

const speedOptions = computed(() => [
  { label: t('animation.speedFast'), value: 'fast' },
  { label: t('animation.speedNormal'), value: 'normal' },
  { label: t('animation.speedSlow'), value: 'slow' },
])
const offsetOptions = computed(() => [
  { label: t('animation.offsetNone'), value: 'none' },
  { label: t('animation.offsetLight'), value: 'light' },
  { label: t('animation.offsetStrong'), value: 'strong' },
])

// a-segmented 的 value 是 string|number，用 computed 收窄回枚举类型
const speedSel = computed({
  get: () => motionPrefs.speed,
  set: (v: string | number) => {
    motionPrefs.speed = v as MotionSpeed
  },
})
const offsetSel = computed({
  get: () => motionPrefs.offset,
  set: (v: string | number) => {
    motionPrefs.offset = v as MotionOffset
  },
})
</script>

<template>
  <div class="settings-form animation-form">
    <div class="form-header">
      <div class="form-title"><ThunderboltOutlined /> {{ t('settings.subTabs.animation') }}</div>
      <div class="form-desc">{{ t('settings.subTabs.animationDesc') }}</div>
    </div>

    <a-divider orientation="left" plain>{{ t('animation.sectionBasic') }}</a-divider>
    <div class="setting-row">
      <div class="row-label">
        <div class="row-title">{{ t('animation.enable') }}</div>
        <div class="row-help">{{ t('animation.enableDesc') }}</div>
      </div>
      <div class="row-control">
        <a-switch v-model:checked="motionPrefs.enabled" />
      </div>
    </div>
    <div class="setting-row">
      <div class="row-label">
        <div class="row-title">{{ t('animation.bounce') }}</div>
        <div class="row-help">{{ t('animation.bounceDesc') }}</div>
      </div>
      <div class="row-control">
        <a-switch v-model:checked="motionPrefs.bounce" :disabled="!motionPrefs.enabled" />
      </div>
    </div>

    <a-divider orientation="left" plain>{{ t('animation.sectionRhythm') }}</a-divider>
    <div class="setting-row">
      <div class="row-label">
        <div class="row-title">{{ t('animation.speed') }}</div>
        <div class="row-help">{{ t('animation.speedDesc') }}</div>
      </div>
      <div class="row-control">
        <a-segmented v-model:value="speedSel" :options="speedOptions" :disabled="!motionPrefs.enabled" />
      </div>
    </div>
    <div class="setting-row">
      <div class="row-label">
        <div class="row-title">{{ t('animation.offset') }}</div>
        <div class="row-help">{{ t('animation.offsetDesc') }}</div>
      </div>
      <div class="row-control">
        <a-segmented v-model:value="offsetSel" :options="offsetOptions" :disabled="!motionPrefs.enabled" />
      </div>
    </div>

    <a-divider orientation="left" plain>{{ t('animation.sectionDetail') }}</a-divider>
    <div class="setting-row">
      <div class="row-label">
        <div class="row-title">{{ t('animation.stagger') }}</div>
        <div class="row-help">{{ t('animation.staggerDesc') }}</div>
      </div>
      <div class="row-control">
        <a-switch v-model:checked="motionPrefs.stagger" :disabled="!motionPrefs.enabled" />
      </div>
    </div>
    <div class="setting-row">
      <div class="row-label">
        <div class="row-title">{{ t('animation.countUp') }}</div>
        <div class="row-help">{{ t('animation.countUpDesc') }}</div>
      </div>
      <div class="row-control">
        <a-switch v-model:checked="motionPrefs.countUp" :disabled="!motionPrefs.enabled" />
      </div>
    </div>

    <p class="anim-note">{{ t('animation.note') }}</p>
  </div>
</template>

<style scoped>
@import './settings-shared.css';

.anim-note {
  font-size: 11px;
  color: var(--color-text-tertiary);
  margin-top: var(--spacing-md);
  line-height: 1.6;
}
</style>
