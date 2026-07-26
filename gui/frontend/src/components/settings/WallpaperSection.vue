<script setup>
/**
 * 壁纸设置（显示分组 · 壁纸子分类，对齐 byte-stash WallpaperSection，适配 Wails）
 *
 * - 上传图片（PickImageFile → WallpaperUpload → 复制到 wallpapers/）→ 自动设为当前
 * - 壁纸库缩略图网格：点选切换、hover 删除（缩略图走 base64 data URL 缓存）
 * - 模糊 / 遮罩 / 面板透明度 三档滑块（store apply 注入 --wp-blur / --wp-panel-alpha）
 * - 面板毛玻璃开关（store apply 注入 data-wp-glass，全局 CSS 触发 backdrop-filter）
 * - 清除壁纸（恢复纯色面板）
 */
import { ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { PictureOutlined, PlusOutlined, DeleteOutlined } from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'
import { useWallpaperStore } from '@/stores/wallpaper'

const { t } = useI18n()
const wallpaper = useWallpaperStore()
const uploading = ref(false)

const app = window.go?.main?.ReferenceApp

async function onUpload() {
  uploading.value = true
  try {
    const src = await app?.PickImageFile?.()
    if (!src) return
    const filename = await app.WallpaperUpload(src)
    await wallpaper.refreshList()
    await wallpaper.setActive(filename)
    message.success(t('wallpaper.uploaded'))
  } catch (e) {
    message.error(typeof e === 'string' ? e : t('wallpaper.uploadFailed'))
  } finally {
    uploading.value = false
  }
}

function onPick(filename) {
  wallpaper.setActive(filename)
}

async function onDelete(filename) {
  try {
    if (wallpaper.activeName === filename) wallpaper.clear()
    await app?.WallpaperDelete?.(filename)
    await wallpaper.refreshList()
    message.success(t('wallpaper.deleted'))
  } catch (e) {
    message.error(typeof e === 'string' ? e : t('wallpaper.deleteFailed'))
  }
}

onMounted(async () => {
  // store 已由 AppLayout init，但设置页可能先于首帧打开——确保 list 就绪
  if (!wallpaper.list.length) await wallpaper.init()
})
</script>

<template>
  <div class="wallpaper-section">
    <div class="form-header">
      <div class="form-title"><PictureOutlined /> {{ t('settings.subTabs.wallpaper') }}</div>
      <div class="form-desc">{{ t('settings.subTabs.wallpaperDesc') }}</div>
    </div>

    <div class="wp-enable-row">
      <a-switch
        :checked="wallpaper.enabled"
        :disabled="!wallpaper.hasWallpaper"
        size="small"
        @change="wallpaper.setEnabled"
      />
      <span class="wp-enable-label">{{ t('wallpaper.enable') }}</span>
      <span class="wp-hint">{{ t('wallpaper.enableHint') }}</span>
    </div>

    <a-alert type="info" show-icon class="wp-tip">
      <template #message>{{ t('wallpaper.tipTitle') }}</template>
      <template #description>
        <ul class="wp-tip-list">
          <li>{{ t('wallpaper.tipRatio') }}</li>
          <li>{{ t('wallpaper.tipFormat') }}</li>
          <li>{{ t('wallpaper.tipEffect') }}</li>
          <li>{{ t('wallpaper.tipStorage') }}</li>
          <li v-html="t('wallpaper.tipCustom')"></li>
        </ul>
      </template>
    </a-alert>

    <div class="wp-grid">
      <div class="wp-tile wp-upload" :class="{ busy: uploading }" @click="onUpload">
        <PlusOutlined />
        <span>{{ uploading ? t('wallpaper.uploading') : t('wallpaper.upload') }}</span>
      </div>
      <div
        v-for="f in wallpaper.list"
        :key="f"
        class="wp-tile wp-thumb"
        :class="{ active: wallpaper.activeName === f }"
        :style="wallpaper.thumbUrl(f) ? { backgroundImage: `url('${wallpaper.thumbUrl(f)}')` } : {}"
        :title="f"
        @click="onPick(f)"
      >
        <button class="wp-del" :title="t('common.delete')" @click.stop="onDelete(f)">
          <DeleteOutlined />
        </button>
      </div>
    </div>

    <div class="wp-controls">
      <div class="wp-row">
        <span class="wp-label">{{ t('wallpaper.blur') }}</span>
        <a-slider
          :value="wallpaper.blur"
          :min="0"
          :max="40"
          :step="1"
          :disabled="!wallpaper.hasWallpaper || !wallpaper.enabled"
          class="wp-slider"
          @change="(v) => wallpaper.setBlur(v)"
        />
        <span class="wp-val">{{ wallpaper.blur }}px</span>
      </div>
      <div class="wp-row">
        <span class="wp-label">{{ t('wallpaper.mask') }}</span>
        <a-slider
          :value="Math.round(wallpaper.mask * 100)"
          :min="0"
          :max="80"
          :step="1"
          :disabled="!wallpaper.hasWallpaper || !wallpaper.enabled"
          class="wp-slider"
          @change="(v) => wallpaper.setMask(v / 100)"
        />
        <span class="wp-val">{{ Math.round(wallpaper.mask * 100) }}%</span>
      </div>
      <div class="wp-row">
        <span class="wp-label">{{ t('wallpaper.panelAlpha') }}</span>
        <a-slider
          :value="Math.round(wallpaper.panelAlpha * 100)"
          :min="0"
          :max="70"
          :step="1"
          :disabled="!wallpaper.hasWallpaper || !wallpaper.enabled"
          class="wp-slider"
          @change="(v) => wallpaper.setPanelAlpha(v / 100)"
        />
        <span class="wp-val">{{ Math.round(wallpaper.panelAlpha * 100) }}%</span>
      </div>
      <div class="wp-row wp-row-switch">
        <span class="wp-label">{{ t('wallpaper.glass') }}</span>
        <a-switch
          :checked="wallpaper.glass"
          :disabled="!wallpaper.hasWallpaper || !wallpaper.enabled"
          size="small"
          @change="wallpaper.setGlass"
        />
        <span class="wp-hint">{{ t('wallpaper.glassHint') }}</span>
      </div>
    </div>

    <div class="wp-actions">
      <span class="wp-hint">{{ t('wallpaper.clearHint') }}</span>
      <a-button v-if="wallpaper.hasWallpaper" size="small" @click="wallpaper.clear()">
        {{ t('wallpaper.clear') }}
      </a-button>
    </div>
  </div>
</template>

<style scoped>
@import './settings-shared.css';

.wallpaper-section {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.wp-enable-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}
.wp-enable-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-secondary);
}
.wp-hint {
  font-size: 11px;
  color: var(--color-text-tertiary);
  line-height: 1.6;
}
.wp-tip {
  margin-bottom: 16px;
}
.wp-tip-list {
  margin: 0;
  padding-left: 18px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
  color: var(--color-text-secondary);
}
.wp-tip-list :deep(code) {
  font-family: 'Cascadia Code', monospace;
  background: var(--color-surface-raised);
  padding: 0 4px;
  border-radius: 3px;
  font-size: 11px;
}
.wp-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
.wp-tile {
  width: 96px;
  height: 60px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  overflow: hidden;
  position: relative;
  cursor: pointer;
  background-size: cover;
  background-position: center;
  transition: all var(--transition-fast);
}
.wp-tile:hover {
  transform: translateY(-1px);
  border-color: var(--color-primary-border);
}
.wp-tile.active {
  box-shadow: 0 0 0 2px var(--color-primary);
}
.wp-upload {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  color: var(--color-text-tertiary);
  font-size: 12px;
  border-style: dashed;
}
.wp-upload.busy {
  opacity: 0.6;
  cursor: wait;
}
.wp-del {
  position: absolute;
  top: 2px;
  right: 2px;
  width: 20px;
  height: 20px;
  border: none;
  border-radius: var(--radius-xs);
  background: rgba(0, 0, 0, 0.55);
  color: #fff;
  cursor: pointer;
  font-size: 11px;
  display: none;
  align-items: center;
  justify-content: center;
}
.wp-thumb:hover .wp-del {
  display: flex;
}
.wp-controls {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.wp-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.wp-row-switch {
  gap: 8px;
}
.wp-label {
  width: 84px;
  font-size: 13px;
  color: var(--color-text-secondary);
  flex-shrink: 0;
}
.wp-slider {
  flex: 1;
  min-width: 0;
}
.wp-val {
  width: 44px;
  text-align: right;
  font-size: 12px;
  color: var(--color-text-tertiary);
  font-family: 'Cascadia Code', monospace;
  flex-shrink: 0;
}
.wp-actions {
  margin-top: 12px;
  display: flex;
  align-items: center;
  gap: 12px;
}
</style>
