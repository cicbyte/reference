<script setup>
import { computed } from 'vue'
import {
  FolderOpenOutlined, ReloadOutlined, FolderOutlined,
  DeleteOutlined, EnvironmentOutlined, BranchesOutlined,
  ArrowRightOutlined, ThunderboltOutlined,
} from '@ant-design/icons-vue'
import { message as AMessage } from 'ant-design-vue'
import { useRepoStore } from '../../stores/repo'
import { relTime } from '../../composables/useRelativeTime'

const props = defineProps({
  open: { type: Boolean, default: false },
})
const emit = defineEmits(['update:open'])

const repo = useRepoStore()

function close() {
  emit('update:open', false)
}

async function switchTo(dir) {
  if (!dir) return
  try {
    const ok = await repo.switchTo(dir)
    if (ok) {
      AMessage.success(`已切换到 ${repo.name}`)
      close()
    }
  } catch (e) {
    AMessage.error('切换失败: ' + String(e?.message || e))
  }
}

async function onPick() {
  const dir = await repo.pickFolder()
  if (dir) await switchTo(dir)
}

async function onRemove(p, e) {
  e.stopPropagation()
  await repo.removeRecent(p)
  AMessage.success('已从最近列表移除')
}

async function onManualEnter(e) {
  const v = e.target.value?.trim()
  if (v) await switchTo(v)
}
</script>

<template>
  <a-modal
    :open="open"
    @update:open="emit('update:open', $event)"
    :footer="null"
    :mask-closable="true"
    destroy-on-close
    width="520px"
    wrap-class-name="repo-modal"
    class="repo-modal-root"
  >
    <!-- custom header strip -->
    <template #title>
      <div class="rm-head">
        <div class="rm-head-icon"><FolderOpenOutlined /></div>
        <div class="rm-head-text">
          <span class="rm-head-title">切换仓库</span>
          <span class="rm-head-sub">选择最近仓库，或浏览本地目录</span>
        </div>
      </div>
    </template>

    <div class="rm-body">
      <!-- current repo -->
      <div class="rm-current" :class="{ empty: !repo.path }">
        <div class="rm-cur-badge"><EnvironmentOutlined /></div>
        <div class="rm-cur-info">
          <div class="rm-cur-line">
            <span class="rm-cur-name">{{ repo.name || '未选择仓库' }}</span>
            <span v-if="repo.branch" class="rm-cur-branch"><BranchesOutlined /> {{ repo.branch }}</span>
          </div>
          <div v-if="repo.path" class="rm-cur-path" :title="repo.path">{{ repo.path }}</div>
          <div v-else class="rm-cur-hint">从下方选择或输入一个仓库路径</div>
        </div>
      </div>

      <!-- quick switch input -->
      <a-input
        class="rm-input"
        placeholder="粘贴仓库路径，回车快速切换"
        allow-clear
        @press-enter="onManualEnter"
      >
        <template #prefix><ThunderboltOutlined /></template>
      </a-input>

      <!-- recent header -->
      <div class="rm-list-head">
        <span class="rm-list-label">最近打开</span>
        <span class="rm-list-count">{{ repo.recent.length }}</span>
      </div>

      <!-- recent list -->
      <div class="rm-list">
        <div
          v-for="r in repo.recent"
          :key="r.path"
          class="rm-item"
          :class="{ active: r.path === repo.path }"
          @click="switchTo(r.path)"
        >
          <div class="rm-item-icon"><FolderOutlined /></div>
          <div class="rm-item-body">
            <div class="rm-item-line">
              <span class="rm-item-name">{{ r.name }}</span>
              <span class="rm-item-time">{{ relTime(r.lastOpened) }}</span>
            </div>
            <div class="rm-item-path" :title="r.path">{{ r.path }}</div>
          </div>
          <ArrowRightOutlined class="rm-item-arrow" />
          <button class="rm-item-del" title="从列表移除" @click="onRemove(r.path, $event)">
            <DeleteOutlined />
          </button>
        </div>

        <div v-if="!repo.recent.length" class="rm-empty">
          <div class="rm-empty-icon"><FolderOpenOutlined /></div>
          <span>暂无最近打开的仓库</span>
          <span class="rm-empty-tip">点击下方「浏览文件夹」开始</span>
        </div>
      </div>
    </div>

    <!-- footer -->
    <div class="rm-foot">
      <a-button class="rm-foot-ghost" @click="repo.detectCurrent()" title="重新检测当前目录">
        <template #icon><ReloadOutlined /></template>重新检测
      </a-button>
      <div class="rm-foot-spacer"></div>
      <a-button @click="close">关闭</a-button>
      <a-button type="primary" @click="onPick">
        <template #icon><FolderOpenOutlined /></template>浏览文件夹
      </a-button>
    </div>
  </a-modal>
</template>

<style scoped>
/* ---- header ---- */
.rm-head { display: flex; align-items: center; gap: 12px; }
.rm-head-icon {
  display: flex; align-items: center; justify-content: center;
  width: 38px; height: 38px; border-radius: var(--radius-md);
  background: var(--color-primary-bg); color: var(--color-primary);
  font-size: 18px;
}
.rm-head-text { display: flex; flex-direction: column; gap: 1px; }
.rm-head-title { font-size: 16px; font-weight: 600; color: var(--color-text); line-height: 1.2; }
.rm-head-sub { font-size: 11px; color: var(--color-text-tertiary); }

/* ---- body ---- */
.rm-body { display: flex; flex-direction: column; gap: var(--spacing-md); padding-top: 4px; }

/* current repo card */
.rm-current {
  display: flex; align-items: center; gap: 12px;
  padding: 12px 14px;
  border-radius: var(--radius-md);
  background: linear-gradient(135deg, var(--color-primary-bg), transparent);
  border: 1px solid var(--color-primary);
  box-shadow: var(--shadow-xs);
}
.rm-current.empty {
  background: var(--color-surface);
  border: 1px dashed var(--color-border);
  box-shadow: none;
}
.rm-cur-badge {
  display: flex; align-items: center; justify-content: center;
  width: 34px; height: 34px; border-radius: var(--radius-sm);
  background: var(--color-primary); color: #fff; font-size: 16px;
  flex-shrink: 0;
}
.rm-current.empty .rm-cur-badge { background: var(--color-border); }
.rm-cur-info { flex: 1; min-width: 0; }
.rm-cur-line { display: flex; align-items: center; gap: 8px; }
.rm-cur-name { font-size: 15px; font-weight: 600; color: var(--color-text); }
.rm-cur-branch {
  display: inline-flex; align-items: center; gap: 4px;
  font-size: 11px; padding: 1px 8px; border-radius: 999px;
  background: var(--color-primary); color: #fff;
}
.rm-cur-path {
  margin-top: 3px; font-family: 'Cascadia Code', monospace; font-size: 11px;
  color: var(--color-text-tertiary);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.rm-cur-hint { margin-top: 3px; font-size: 12px; color: var(--color-text-tertiary); }

/* input */
.rm-input :deep(.ant-input-affix-wrapper) { border-radius: var(--radius-md); }

/* list header */
.rm-list-head { display: flex; align-items: center; gap: 8px; margin-top: 2px; }
.rm-list-label { font-size: 12px; font-weight: 600; color: var(--color-text-secondary); text-transform: uppercase; letter-spacing: 0.04em; }
.rm-list-count {
  display: inline-flex; align-items: center; justify-content: center;
  min-width: 18px; height: 18px; padding: 0 5px;
  font-size: 11px; border-radius: 999px;
  background: var(--color-surface-raised); color: var(--color-text-secondary);
}

/* list */
.rm-list { display: flex; flex-direction: column; gap: 6px; max-height: 260px; overflow-y: auto; margin: 0 -2px; padding: 0 2px; }
.rm-list::-webkit-scrollbar { width: 6px; }
.rm-list::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 3px; }

.rm-item {
  position: relative;
  display: flex; align-items: center; gap: 12px;
  padding: 9px 12px;
  border: 1px solid var(--color-border-light); border-radius: var(--radius-md);
  cursor: pointer; background: var(--color-background);
  transition: all var(--transition-fast);
}
.rm-item:hover {
  border-color: var(--color-primary);
  background: var(--color-hover);
  transform: translateX(2px);
}
.rm-item.active {
  border-color: var(--color-primary);
  background: var(--color-primary-bg);
}
.rm-item-icon {
  display: flex; align-items: center; justify-content: center;
  width: 30px; height: 30px; border-radius: var(--radius-sm);
  background: var(--color-surface); color: var(--color-text-tertiary);
  font-size: 15px; flex-shrink: 0; transition: all var(--transition-fast);
}
.rm-item:hover .rm-item-icon, .rm-item.active .rm-item-icon {
  background: var(--color-primary-bg); color: var(--color-primary);
}
.rm-item-body { flex: 1; min-width: 0; }
.rm-item-line { display: flex; align-items: center; gap: 8px; }
.rm-item-name { font-size: 13px; font-weight: 500; color: var(--color-text); }
.rm-item-time { font-size: 11px; color: var(--color-text-tertiary); margin-left: auto; }
.rm-item-path {
  font-family: 'Cascadia Code', monospace; font-size: 11px; color: var(--color-text-tertiary);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.rm-item-arrow {
  font-size: 11px; color: var(--color-primary);
  opacity: 0; transition: all var(--transition-fast); flex-shrink: 0;
}
.rm-item:hover .rm-item-arrow { opacity: 0.7; }
.rm-item-del {
  display: flex; align-items: center; justify-content: center;
  width: 24px; height: 24px; border: none; border-radius: var(--radius-sm);
  background: transparent; color: var(--color-text-tertiary); cursor: pointer;
  opacity: 0; transition: all var(--transition-fast); flex-shrink: 0;
}
.rm-item:hover .rm-item-del { opacity: 0.5; }
.rm-item-del:hover { opacity: 1 !important; background: var(--color-error-bg); color: var(--color-error); }

/* empty */
.rm-empty { display: flex; flex-direction: column; align-items: center; gap: 6px; padding: 32px 0; }
.rm-empty-icon { font-size: 30px; color: var(--color-text-tertiary); opacity: 0.35; }
.rm-empty span:nth-child(2) { font-size: 13px; color: var(--color-text-secondary); }
.rm-empty-tip { font-size: 11px; color: var(--color-text-tertiary); }

/* footer */
.rm-foot { display: flex; align-items: center; gap: var(--spacing-sm); margin-top: var(--spacing-md); padding-top: var(--spacing-md); border-top: 1px solid var(--color-border-light); }
.rm-foot-spacer { flex: 1; }
</style>

<!-- non-scoped: tune the ant-modal shell to match the design system -->
<style>
.repo-modal .ant-modal {
  border-radius: var(--radius-lg) !important;
  overflow: hidden;
  box-shadow: var(--shadow-lg) !important;
}
.repo-modal .ant-modal-content {
  border-radius: var(--radius-lg) !important;
  padding: var(--spacing-lg) var(--spacing-lg) var(--spacing-md) !important;
  background: var(--color-background) !important;
}
.repo-modal .ant-modal-header { margin-bottom: var(--spacing-md) !important; padding: 0 !important; background: transparent !important; }
.repo-modal .ant-modal-close { top: 16px !important; inset-inline-end: 16px !important; }
.repo-modal .ant-btn-primary {
  background: var(--color-primary) !important;
  box-shadow: var(--shadow-primary) !important;
}
.repo-modal .ant-btn-primary:hover { background: var(--color-primary-hover) !important; }
</style>
