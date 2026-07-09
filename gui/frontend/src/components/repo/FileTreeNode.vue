<script setup>
/**
 * Recursive file-tree node renderer — VSCode-explorer style.
 * Directories first, file-type icon colors, chevron rotation, hover/active.
 */
import {
  FolderFilled, FileOutlined, FileTextOutlined, FileMarkdownOutlined,
  CodeOutlined, FileZipOutlined, DownOutlined, RightOutlined,
} from '@ant-design/icons-vue'

const props = defineProps({
  nodes: { type: Array, default: () => [] },
  depth: { type: Number, default: 0 },
  tree: { type: Object, default: () => ({}) },
  expanded: { type: Set, default: () => new Set() },
  selected: { default: '' },
  loadingDirs: { type: Set, default: () => new Set() },
})
const emit = defineEmits(['toggle', 'open'])

function isSelected(path) {
  return props.selected === path
}

function onNodeClick(node) {
  emit('open', node)
  if (node.isDir) emit('toggle', node)
}

// pick a file-type icon + color by extension (VSCode-ish)
function fileIcon(name) {
  if (!name) return { icon: FileOutlined, color: 'var(--color-text-tertiary)' }
  const ext = (name.split('.').pop() || '').toLowerCase()
  const map = {
    go: { icon: CodeOutlined, color: '#00ADD8' },
    js: { icon: CodeOutlined, color: '#F7DF1E' },
    ts: { icon: CodeOutlined, color: '#3178C6' },
    vue: { icon: CodeOutlined, color: '#42B883' },
    py: { icon: CodeOutlined, color: '#3776AB' },
    rs: { icon: CodeOutlined, color: '#DEA584' },
    md: { icon: FileMarkdownOutlined, color: '#6C6C6C' },
    json: { icon: CodeOutlined, color: '#CBC24A' },
    yml: { icon: FileTextOutlined, color: '#CB171E' },
    yaml: { icon: FileTextOutlined, color: '#CB171E' },
    toml: { icon: FileTextOutlined, color: '#9C4221' },
    html: { icon: CodeOutlined, color: '#E34F26' },
    css: { icon: CodeOutlined, color: '#1572B6' },
    zip: { icon: FileZipOutlined, color: '#7B68EE' },
    gz: { icon: FileZipOutlined, color: '#7B68EE' },
    png: { icon: FileOutlined, color: '#A8A8A8' },
    jpg: { icon: FileOutlined, color: '#A8A8A8' },
  }
  return map[ext] || { icon: FileOutlined, color: 'var(--color-text-tertiary)' }
}
</script>

<template>
  <template v-for="node in nodes" :key="node.path">
    <div
      class="ftn-node"
      :class="{ selected: isSelected(node.path), 'is-dir': node.isDir }"
      :style="{ paddingLeft: 6 + depth * 14 + 'px' }"
      @click="onNodeClick(node)"
    >
      <span class="ftn-chevron" v-if="node.isDir">
        <a-spin v-if="loadingDirs.has(node.path)" size="small" />
        <DownOutlined v-else-if="expanded.has(node.path)" />
        <RightOutlined v-else />
      </span>
      <span class="ftn-chevron-spacer" v-else></span>

      <component
        :is="node.isDir ? FolderFilled : fileIcon(node.name).icon"
        class="ftn-icon"
        :style="!node.isDir ? { color: fileIcon(node.name).color } : {}"
      />
      <span class="ftn-name" :title="node.name">{{ node.name }}</span>
    </div>

    <FileTreeNode
      v-if="node.isDir && expanded.has(node.path)"
      :nodes="tree[node.path] || []"
      :depth="depth + 1"
      :tree="tree"
      :expanded="expanded"
      :selected="selected"
      :loading-dirs="loadingDirs"
      @toggle="(n) => emit('toggle', n)"
      @open="(n) => emit('open', n)"
    />
  </template>
</template>

<style scoped>
.ftn-node {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 3px 8px;
  cursor: pointer;
  font-size: 13px;
  line-height: 1.5;
  color: var(--color-text);
  white-space: nowrap;
  user-select: none;
  border-radius: var(--radius-xs);
  transition: background var(--transition-fast);
}
.ftn-node:hover { background: var(--color-hover); }
.ftn-node.selected { background: var(--color-primary-bg); }
.ftn-node.selected .ftn-name { color: var(--color-primary); font-weight: 500; }

.ftn-chevron {
  width: 16px;
  height: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: var(--color-text-tertiary);
  font-size: 9px;
}
.ftn-chevron-spacer { width: 16px; flex-shrink: 0; }
.ftn-icon { font-size: 15px; flex-shrink: 0; color: var(--color-warning); }
.ftn-name { overflow: hidden; text-overflow: ellipsis; flex: 1; }
</style>
