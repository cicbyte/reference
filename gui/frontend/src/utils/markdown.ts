/**
 * Safe Markdown rendering.
 *
 * `marked` (v5+) no longer sanitizes inline HTML, and wiki content comes from
 * semi-trusted remote git repos. Every `v-html` of markdown output MUST go
 * through `renderMarkdown()` here so DOMPurify strips event handlers / scripts.
 */
import { marked } from 'marked'
import DOMPurify from 'dompurify'

// Configure marked once. GFM + line breaks for readability.
marked.setOptions({ gfm: true, breaks: false })

// Allow mermaid code blocks to pass through (they are rendered later by our own
// mermaid renderer, not executed as HTML), but strip everything dangerous.
const ALLOWED_TAGS = [
  'p', 'br', 'hr', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
  'strong', 'em', 'del', 's', 'sub', 'sup', 'mark', 'abbr',
  'blockquote', 'code', 'pre', 'kbd', 'samp', 'var',
  'ul', 'ol', 'li', 'dl', 'dt', 'dd',
  'table', 'thead', 'tbody', 'tfoot', 'tr', 'th', 'td', 'caption', 'colgroup', 'col',
  'a', 'img', 'span', 'div', 'figure', 'figcaption',
  'details', 'summary',
]

const ALLOWED_ATTR = [
  'href', 'title', 'alt', 'src', 'width', 'height',
  'class', 'id',
  'colspan', 'rowspan', 'span',
  'target', 'rel',
  'open', 'start', 'reversed', 'type',
  'align', 'valign',
]

/** Render markdown to sanitized HTML (safe for v-html). */
export function renderMarkdown(src: string): string {
  if (!src) return ''
  const raw = (() => {
    try { return marked.parse(src, { async: false }) as string }
    catch { return String(src) }
  })()
  return DOMPurify.sanitize(raw, {
    ALLOWED_TAGS,
    ALLOWED_ATTR,
    ALLOW_DATA_ATTR: false,
  })
}
