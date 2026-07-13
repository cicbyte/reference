// Side-effect module: registers languages and exposes hljs on window.
// MUST be imported separately (and before) the line-numbers plugin, since
// static `import` statements are hoisted and evaluate before the body runs.
import hljs from 'highlight.js/lib/core'
import hljsGo from 'highlight.js/lib/languages/go'
import hljsJs from 'highlight.js/lib/languages/javascript'
import hljsTs from 'highlight.js/lib/languages/typescript'
import hljsXml from 'highlight.js/lib/languages/xml'
import hljsCss from 'highlight.js/lib/languages/css'
import hljsJson from 'highlight.js/lib/languages/json'
import hljsYaml from 'highlight.js/lib/languages/yaml'
import hljsMd from 'highlight.js/lib/languages/markdown'
import hljsPy from 'highlight.js/lib/languages/python'
import hljsBash from 'highlight.js/lib/languages/bash'
import hljsSql from 'highlight.js/lib/languages/sql'
import hljsRust from 'highlight.js/lib/languages/rust'
import 'highlight.js/styles/github-dark.css'

const langs = {
  go: hljsGo, javascript: hljsJs, typescript: hljsTs, xml: hljsXml,
  css: hljsCss, json: hljsJson, yaml: hljsYaml, markdown: hljsMd,
  python: hljsPy, bash: hljsBash, sql: hljsSql, rust: hljsRust,
}
Object.entries(langs).forEach(([name, mod]) => hljs.registerLanguage(name, mod))

// Expose so the line-numbers plugin (a window-aware IIFE) can augment it.
window.hljs = hljs

export default hljs

export function hljsLangForName(name) {
  const ext = (name.split('.').pop() || '').toLowerCase()
  const map = {
    go: 'go', js: 'javascript', ts: 'typescript', html: 'xml', vue: 'xml',
    css: 'css', json: 'json', yml: 'yaml', yaml: 'yaml', md: 'markdown',
    py: 'python', sh: 'bash', sql: 'sql', rs: 'rust',
  }
  return map[ext] || ''
}
