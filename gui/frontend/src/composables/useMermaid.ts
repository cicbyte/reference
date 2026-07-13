/**
 * Mermaid rendering composable.
 *
 * Extracted from the duplicated blocks in WikiView.vue / LocalWikiView.vue.
 * Handles:
 *   - mermaid.initialize with theme that follows the theme store
 *   - re-init when the user toggles dark/light
 *   - renderMermaid(): scans rendered markdown for `language-mermaid` code
 *     blocks and replaces them with rendered SVGs, wiring click-to-enlarge.
 *
 * Usage:
 *   const { renderMermaid, modalOpen, modalSvg, zoom, panX, panY } = useMermaid()
 *   // after markdown v-html is patched, call await renderMermaid()
 *   // mount <MermaidModal v-model:open="modalOpen" :svg="modalSvg" ... />
 */
import { ref, watch } from 'vue'
import mermaid from 'mermaid'
import { useThemeStore } from '../stores/theme'

export function useMermaid(contentSelector = '.wc-md .language-mermaid') {
  const themeStore = useThemeStore()

  // modal state — pass these to <MermaidModal>
  const modalOpen = ref(false)
  const modalSvg = ref('')
  const zoom = ref(1)
  const panX = ref(0)
  const panY = ref(0)

  function applyTheme(isDark: boolean) {
    mermaid.initialize({
      startOnLoad: false,
      theme: isDark ? 'dark' : 'default',
      securityLevel: 'strict',
    })
  }

  applyTheme(themeStore.isDark)
  // re-init when the user toggles theme so diagrams stay legible
  watch(() => themeStore.isDark, (isDark) => applyTheme(isDark))

  function openModal(svg: string) {
    modalSvg.value = svg
    zoom.value = 1
    panX.value = 0
    panY.value = 0
    modalOpen.value = true
  }

  /** Render mermaid code blocks found in the DOM after a markdown patch. */
  async function renderMermaid() {
    const els = document.querySelectorAll(contentSelector)
    for (const el of Array.from(els)) {
      const code = el.textContent
      if (!code?.trim()) continue
      try {
        const id = 'mmd-' + Math.random().toString(36).slice(2, 9)
        const { svg } = await mermaid.render(id, code)
        const wrapper = (el as HTMLElement).closest('pre')
        if (wrapper) {
          const container = document.createElement('div')
          container.className = 'mermaid-rendered'
          container.innerHTML = svg
          container.style.cursor = 'zoom-in'
          container.addEventListener('click', () => openModal(svg))
          wrapper.replaceWith(container)
        }
      } catch {
        /* leave as a code block on parse error */
      }
    }
  }

  return { modalOpen, modalSvg, zoom, panX, panY, renderMermaid, openModal }
}
