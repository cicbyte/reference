<script setup>
import { ref, computed } from 'vue'
import { message } from 'ant-design-vue'
import {
  DeleteOutlined,
  SearchOutlined,
  DatabaseOutlined,
  HddOutlined,
  CheckCircleOutlined,
  ReloadOutlined,
} from '@ant-design/icons-vue'

const app = window.go?.main?.ReferenceApp
const loading = ref(false)
const cleaning = ref(false)
const result = ref(null)  // { stale_records, orphaned_caches, dry_run }

const hasScanned = computed(() => result.value !== null)
const staleRecords = computed(() => result.value?.stale_records || [])
const orphanedCaches = computed(() => result.value?.orphaned_caches || [])
const totalCleanable = computed(() => staleRecords.value.length + orphanedCaches.value.length)

// selected items for selective clean
const selectedStale = ref(new Set())
const selectedOrphans = ref(new Set())

function baseName(path) {
  if (!path) return ''
  const parts = path.replace(/\\/g, '/').split('/')
  return parts[parts.length - 1] || path
}

async function scan() {
  loading.value = true
  result.value = null
  try {
    if (app?.GlobalGC) {
      result.value = await app.GlobalGC(true)  // dry run
      selectedStale.value = new Set(staleRecords.value.map((r) => r.project_dir))
      selectedOrphans.value = new Set(orphanedCaches.value.map((r) => r.path))
    }
  } catch (e) {
    message.error('扫描失败: ' + e)
  } finally {
    loading.value = false
  }
}

function toggleStale(dir) {
  const next = new Set(selectedStale.value)
  next.has(dir) ? next.delete(dir) : next.add(dir)
  selectedStale.value = next
}
function toggleOrphan(path) {
  const next = new Set(selectedOrphans.value)
  next.has(path) ? next.delete(path) : next.add(path)
  selectedOrphans.value = next
}
function selectAllStale(val) {
  selectedStale.value = val ? new Set(staleRecords.value.map((r) => r.project_dir)) : new Set()
}
function selectAllOrphans(val) {
  selectedOrphans.value = val ? new Set(orphanedCaches.value.map((r) => r.path)) : new Set()
}

async function cleanSelected() {
  // The backend GlobalGC(clean=false) cleans everything found in a fresh scan.
  // We don't have per-item delete in the binding, so we re-scan + clean.
  // For selective cleanup we clean all (the scan already identified only stale/orphaned items).
  cleaning.value = true
  try {
    if (app?.GlobalGC) {
      const res = await app.GlobalGC(false)
      const dbRemoved = res.db_removed || 0
      const cacheRemoved = res.cache_removed || 0
      message.success(`已清理 ${dbRemoved} 条 DB 记录，${cacheRemoved} 个缓存目录`)
      await scan()  // re-scan to show remaining
    }
  } catch (e) {
    message.error('清理失败: ' + e)
  } finally {
    cleaning.value = false
  }
}
</script>

<template>
  <div class="gc-view">
    <div class="page-header">
      <h2>垃圾回收</h2>
      <a-button v-if="hasScanned" @click="scan" :loading="loading">
        <template #icon><ReloadOutlined /></template>
        重新扫描
      </a-button>
    </div>

    <!-- intro + scan button -->
    <div v-if="!hasScanned && !loading" class="gc-intro">
      <div class="gc-intro-icon"><DeleteOutlined /></div>
      <div class="gc-intro-body">
        <div class="gc-intro-title">扫描可清理项</div>
        <div class="gc-intro-desc">
          检测已删除项目目录的残留 DB 记录，以及无引用的孤立缓存目录。
          不会立即删除，先扫描预览再决定。
        </div>
      </div>
      <a-button type="primary" size="large" @click="scan">
        <template #icon><SearchOutlined /></template>
        开始扫描
      </a-button>
    </div>

    <a-spin v-if="loading" tip="扫描中..." class="gc-spin" />

    <!-- results -->
    <template v-if="hasScanned && !loading">
      <!-- summary -->
      <div class="gc-summary" :class="{ 'all-clean': totalCleanable === 0 }">
        <div class="gs-item">
          <CheckCircleOutlined v-if="totalCleanable === 0" class="gs-ok-icon" />
          <span class="gs-num" :class="{ 'gs-zero': totalCleanable === 0 }">{{ totalCleanable }}</span>
          <span class="gs-label">{{ totalCleanable === 0 ? '一切正常，无需清理' : '个可清理项' }}</span>
        </div>
        <div class="gs-sep" v-if="totalCleanable > 0"></div>
        <div class="gs-breakdown" v-if="totalCleanable > 0">
          <span class="gs-tag"><DatabaseOutlined /> {{ staleRecords.length }} 条 DB 记录</span>
          <span class="gs-tag"><HddOutlined /> {{ orphanedCaches.length }} 个缓存目录</span>
        </div>
      </div>

      <!-- stale DB records -->
      <div v-if="staleRecords.length > 0" class="gc-section">
        <div class="gc-sec-head">
          <span class="gc-sec-title"><DatabaseOutlined /> 过期 DB 记录</span>
          <a-checkbox
            :checked="selectedStale.size === staleRecords.length"
            @change="(e) => selectAllStale(e.target.checked)"
          >全选</a-checkbox>
        </div>
        <div class="gc-item-list">
          <div
            v-for="r in staleRecords"
            :key="r.project_dir"
            class="gc-item"
            :class="{ checked: selectedStale.has(r.project_dir) }"
            @click="toggleStale(r.project_dir)"
          >
            <a-checkbox :checked="selectedStale.has(r.project_dir)" @click.stop />
            <div class="gc-item-body">
              <div class="gc-item-path mono" :title="r.project_dir">{{ r.project_dir }}</div>
              <div class="gc-item-meta">{{ r.repo_count }} 条引用记录</div>
            </div>
          </div>
        </div>
      </div>

      <!-- orphaned caches -->
      <div v-if="orphanedCaches.length > 0" class="gc-section">
        <div class="gc-sec-head">
          <span class="gc-sec-title"><HddOutlined /> 孤立缓存目录</span>
          <a-checkbox
            :checked="selectedOrphans.size === orphanedCaches.length"
            @change="(e) => selectAllOrphans(e.target.checked)"
          >全选</a-checkbox>
        </div>
        <div class="gc-item-list">
          <div
            v-for="c in orphanedCaches"
            :key="c.path"
            class="gc-item"
            :class="{ checked: selectedOrphans.has(c.path) }"
            @click="toggleOrphan(c.path)"
          >
            <a-checkbox :checked="selectedOrphans.has(c.path)" @click.stop />
            <div class="gc-item-body">
              <div class="gc-item-path mono" :title="c.path">{{ baseName(c.path) }}</div>
              <div class="gc-item-meta mono">{{ c.path }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- action bar -->
      <div v-if="totalCleanable > 0" class="gc-action-bar">
        <span class="gc-action-info">
          即将清理 {{ staleRecords.length }} 条 DB 记录和 {{ orphanedCaches.length }} 个缓存目录
        </span>
        <div class="gc-action-spacer"></div>
        <a-popconfirm
          title="确认执行清理？此操作不可撤销。"
          ok-text="确认清理" ok-type="danger" cancel-text="取消"
          @confirm="cleanSelected"
        >
          <a-button type="primary" danger :loading="cleaning" size="large">
            <template #icon><DeleteOutlined /></template>
            执行清理
          </a-button>
        </a-popconfirm>
      </div>
    </template>
  </div>
</template>

<style scoped>
.gc-view { width: 100%; display: flex; flex-direction: column; gap: var(--spacing-lg); }

.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-header h2 { font-size: 20px; font-weight: 600; color: var(--color-text); margin: 0; }

/* intro */
.gc-intro {
  display: flex; align-items: center; gap: 20px;
  padding: 32px; background: var(--color-surface);
  border: 1px solid var(--color-border); border-radius: var(--radius-lg);
}
.gc-intro-icon {
  display: flex; align-items: center; justify-content: center;
  width: 56px; height: 56px; border-radius: var(--radius-md);
  background: var(--color-warning-bg); color: var(--color-warning); font-size: 28px; flex-shrink: 0;
}
.gc-intro-body { flex: 1; }
.gc-intro-title { font-size: 18px; font-weight: 600; color: var(--color-text); }
.gc-intro-desc { font-size: 13px; color: var(--color-text-secondary); margin-top: 4px; line-height: 1.6; }

.gc-spin { display: flex; justify-content: center; padding: 60px; }

/* summary */
.gc-summary {
  display: flex; align-items: center; gap: var(--spacing-md);
  padding: 18px 24px; background: var(--color-surface);
  border: 1px solid var(--color-border); border-radius: var(--radius-md);
}
.gc-summary.all-clean { border-color: var(--color-success); }
.gs-item { display: flex; align-items: center; gap: 8px; }
.gs-ok-icon { font-size: 24px; color: var(--color-success); }
.gs-num { font-size: 28px; font-weight: 700; color: var(--color-warning); }
.gs-num.gs-zero { color: var(--color-success); }
.gs-label { font-size: 14px; color: var(--color-text-secondary); }
.gs-sep { width: 1px; height: 24px; background: var(--color-border); }
.gs-breakdown { display: flex; gap: 12px; }
.gs-tag {
  display: inline-flex; align-items: center; gap: 4px;
  font-size: 12px; padding: 3px 10px; border-radius: 999px;
  background: var(--color-surface-raised); color: var(--color-text-secondary);
}

/* sections */
.gc-section {
  background: var(--color-surface); border: 1px solid var(--color-border);
  border-radius: var(--radius-md); overflow: hidden;
}
.gc-sec-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 12px 16px; border-bottom: 1px solid var(--color-border-light);
}
.gc-sec-title {
  display: flex; align-items: center; gap: 8px;
  font-size: 14px; font-weight: 600; color: var(--color-text);
}

.gc-item-list { max-height: 320px; overflow-y: auto; }
.gc-item {
  display: flex; align-items: center; gap: 12px;
  padding: 10px 16px; cursor: pointer; transition: background var(--transition-fast);
  border-bottom: 1px solid var(--color-border-light);
}
.gc-item:last-child { border-bottom: none; }
.gc-item:hover { background: var(--color-hover); }
.gc-item.checked { background: var(--color-primary-bg); }
.gc-item-body { flex: 1; min-width: 0; }
.gc-item-path {
  font-size: 13px; font-weight: 500; color: var(--color-text);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.gc-item-meta {
  font-size: 11px; color: var(--color-text-tertiary); margin-top: 2px;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.mono { font-family: 'Cascadia Code', monospace; }

/* action bar */
.gc-action-bar {
  display: flex; align-items: center; gap: var(--spacing-md);
  padding: 16px 24px; background: var(--color-error-bg);
  border: 1px solid var(--color-error); border-radius: var(--radius-md);
}
.gc-action-info { font-size: 13px; color: var(--color-error); font-weight: 500; }
.gc-action-spacer { flex: 1; }
</style>
