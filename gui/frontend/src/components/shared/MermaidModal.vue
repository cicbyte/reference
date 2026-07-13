<script setup>
/**
 * Mermaid diagram enlarge modal with zoom / pan / SVG+PNG export.
 *
 * Extracted verbatim from WikiView.vue / LocalWikiView.vue (they were
 * character-for-character duplicates). The parent owns `open` and `svg`
 * (the SVG string produced by mermaid.render); this component handles all
 * interaction (wheel-zoom, drag-pan, download).
 */
import {
  ZoomInOutlined, ZoomOutOutlined, DownloadOutlined, FileImageOutlined,
} from '@ant-design/icons-vue'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  open: { type: Boolean, default: false },
  svg: { type: String, default: '' },
})
const emit = defineEmits(['update:open'])

const { t } = useI18n()

const zoom = defineModel('zoom', { type: Number, default: 1 })
const panX = defineModel('panX', { type: Number, default: 0 })
const panY = defineModel('panY', { type: Number, default: 0 })

function zoomIn() { zoom.value = Math.min(zoom.value * 1.25, 5) }
function zoomOut() { zoom.value = Math.max(zoom.value / 1.25, 0.2) }
function zoomReset() { zoom.value = 1; panX.value = 0; panY.value = 0 }

function onWheel(e) {
  e.preventDefault()
  if (e.deltaY < 0) zoomIn()
  else zoomOut()
}

let dragging = false
let dragStart = { x: 0, y: 0 }
function onDragStart(e) {
  dragging = true
  dragStart = { x: e.clientX - panX.value, y: e.clientY - panY.value }
}
function onDragMove(e) {
  if (!dragging) return
  panX.value = e.clientX - dragStart.x
  panY.value = e.clientY - dragStart.y
}
function onDragEnd() { dragging = false }

function downloadSvg() {
  if (!props.svg) return
  const blob = new Blob([props.svg], { type: 'image/svg+xml' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'mermaid-diagram.svg'
  a.click()
  URL.revokeObjectURL(url)
}

function downloadPng() {
  if (!props.svg) return
  const svgBlob = new Blob([props.svg], { type: 'image/svg+xml' })
  const url = URL.createObjectURL(svgBlob)
  const img = new Image()
  img.onload = () => {
    const canvas = document.createElement('canvas')
    const scale = 2
    canvas.width = img.naturalWidth * scale
    canvas.height = img.naturalHeight * scale
    const ctx = canvas.getContext('2d')
    ctx.fillStyle = '#ffffff'
    ctx.fillRect(0, 0, canvas.width, canvas.height)
    ctx.scale(scale, scale)
    ctx.drawImage(img, 0, 0)
    URL.revokeObjectURL(url)
    canvas.toBlob((blob) => {
      if (!blob) return
      const pngUrl = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = pngUrl
      a.download = 'mermaid-diagram.png'
      a.click()
      URL.revokeObjectURL(pngUrl)
    })
  }
  img.src = url
}
</script>

<template>
  <a-modal
    :open="open"
    @update:open="emit('update:open', $event)"
    :footer="null"
    :title="null"
    width="80%"
    centered
    destroy-on-close
  >
    <div
      class="mermaid-modal-viewport"
      @wheel.prevent="onWheel"
      @mousedown="onDragStart"
      @mousemove="onDragMove"
      @mouseup="onDragEnd"
      @mouseleave="onDragEnd"
    >
      <div
        class="mermaid-modal-body"
        :style="{ transform: `translate(${panX}px, ${panY}px) scale(${zoom})` }"
        v-html="svg"
      ></div>
    </div>
    <div class="mermaid-modal-actions">
      <a-button size="small" @click="zoomOut"><ZoomOutOutlined /></a-button>
      <span class="zoom-level">{{ Math.round(zoom * 100) }}%</span>
      <a-button size="small" @click="zoomIn"><ZoomInOutlined /></a-button>
      <a-button size="small" @click="zoomReset">{{ t('mermaidModal.reset') }}</a-button>
      <span class="zoom-spacer"></span>
      <a-button @click="downloadSvg">
        <template #icon><DownloadOutlined /></template>
        SVG
      </a-button>
      <a-button @click="downloadPng">
        <template #icon><FileImageOutlined /></template>
        PNG
      </a-button>
    </div>
  </a-modal>
</template>

<style scoped>
.mermaid-modal-viewport {
  height: calc(80vh - 100px); overflow: hidden; cursor: grab;
  display: flex; align-items: center; justify-content: center;
  background: var(--color-background); border-radius: var(--radius-md);
}
.mermaid-modal-viewport:active { cursor: grabbing; }
.mermaid-modal-body {
  text-align: center; padding: 16px;
  transition: transform 0.1s ease;
  transform-origin: center center;
}
.mermaid-modal-body svg { max-width: 100%; height: auto; }
.mermaid-modal-actions {
  display: flex; align-items: center; gap: 8px;
  padding: 12px 0 0; border-top: 1px solid var(--color-border-light);
}
.zoom-level { font-size: 12px; color: var(--color-text-secondary); min-width: 42px; text-align: center; }
.zoom-spacer { flex: 1; }
</style>
