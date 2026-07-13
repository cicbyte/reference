<script setup>
import { ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { InfoCircleOutlined, FolderOpenOutlined, CopyOutlined } from '@ant-design/icons-vue'
import AiSparkIcon from '../common/AiSparkIcon.vue'

const loading = ref(true)
const version = ref(null)
const paths = ref(null)

onMounted(async () => {
  try {
    if (window.go?.main?.ReferenceApp) {
      const app = window.go.main.ReferenceApp
      version.value = await app.GetVersionInfo()
      const cfg = await app.GetAppConfig()
      paths.value = cfg.paths
    }
  } catch (e) {
    console.error('Load about info failed:', e)
  } finally {
    loading.value = false
  }
})

async function copyText(text) {
  try {
    await window.go.main.ReferenceApp.CopyPath(text)
    message.success('已复制')
  } catch { message.error('复制失败') }
}

function openDir(dir) {
  window.go.main.ReferenceApp.OpenInExplorer(dir).catch((e) => message.error('打开失败: ' + e))
}
</script>

<template>
  <div class="settings-form">
    <div class="form-header">
      <div class="form-title"><InfoCircleOutlined /> 关于</div>
      <div class="form-desc">应用版本与系统路径信息。</div>
    </div>

    <a-spin :spinning="loading">
      <!-- version hero card -->
      <div class="hero-card">
        <div class="hero-logo"><AiSparkIcon :size="40" /></div>
        <div class="hero-info">
          <div class="hero-name">reference</div>
          <div class="hero-tagline">本地 Git 仓库引用管理器 · 为 AI 辅助编程而生</div>
          <div class="hero-version">
            <span class="ver-badge">v{{ version?.version || 'dev' }}</span>
            <span class="ver-meta mono" v-if="version?.commit">{{ String(version.commit).slice(0, 8) }}</span>
            <span class="ver-meta" v-if="version?.buildTime">{{ version.buildTime }}</span>
          </div>
        </div>
      </div>

      <!-- system paths grid -->
      <div class="setting-group">
        <div class="group-title">系统路径</div>
        <div class="path-grid">
          <div v-for="p in [
            { label: '配置文件', value: paths?.config, icon: '⚙' },
            { label: '数据库', value: paths?.db, icon: '🗄' },
            { label: '仓库缓存', value: paths?.repos, icon: '📦' },
            { label: '知识库', value: paths?.wiki, icon: '📚' },
            { label: '日志目录', value: paths?.logDir, icon: '📝' },
          ]" :key="p.label" class="path-cell">
            <div class="path-cell-head">
              <span class="path-icon">{{ p.icon }}</span>
              <span class="path-label">{{ p.label }}</span>
              <div class="path-actions" v-if="p.value">
                <a-button size="small" type="text" title="复制" @click="copyText(p.value)"><CopyOutlined /></a-button>
                <a-button size="small" type="text" title="在文件管理器中打开" @click="openDir(p.value)"><FolderOpenOutlined /></a-button>
              </div>
            </div>
            <div class="path-value mono" :title="p.value">{{ p.value || '—' }}</div>
          </div>
        </div>
      </div>

      <!-- links -->
      <div class="setting-group">
        <div class="group-title">资源</div>
        <div class="link-list">
          <a href="https://github.com/cicbyte/reference" target="_blank" class="link-item">
            <span class="link-ico">⌥</span>
            <div>
              <div class="link-title">GitHub 仓库</div>
              <div class="link-desc">源码、Issue 与发布</div>
            </div>
          </a>
        </div>
      </div>
    </a-spin>
  </div>
</template>

<style scoped>
@import './settings-shared.css';

/* hero card */
.hero-card {
  display: flex; align-items: center; gap: 16px;
  padding: 20px 24px; margin-bottom: var(--spacing-md);
  border-radius: var(--radius-md); border: 1px solid var(--color-border);
  background: linear-gradient(135deg, var(--color-surface), var(--color-surface-raised));
}
.hero-logo {
  display: flex; align-items: center; justify-content: center;
  width: 64px; height: 64px; flex-shrink: 0;
  border-radius: var(--radius-md); background: var(--color-primary-bg);
}
.hero-name { font-size: 20px; font-weight: 700; color: var(--color-text); }
.hero-tagline { font-size: 13px; color: var(--color-text-tertiary); margin-top: 2px; }
.hero-version { display: flex; align-items: center; gap: 10px; margin-top: 10px; }
.ver-badge {
  font-size: 12px; font-weight: 600; padding: 2px 8px; border-radius: 4px;
  background: var(--color-primary); color: #fff;
}
.ver-meta { font-size: 11px; color: var(--color-text-tertiary); }

/* path grid */
.path-grid {
  display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px;
}
.path-cell {
  padding: 10px 12px; border-radius: var(--radius-sm);
  background: var(--color-background); border: 1px solid var(--color-border-light);
}
.path-cell-head { display: flex; align-items: center; gap: 6px; margin-bottom: 4px; }
.path-icon { font-size: 13px; }
.path-label { font-size: 12px; font-weight: 600; color: var(--color-text-secondary); flex: 1; }
.path-actions { display: flex; gap: 2px; opacity: 0; transition: opacity var(--transition-fast); }
.path-cell:hover .path-actions { opacity: 0.7; }
.path-value {
  font-size: 11px; color: var(--color-text-tertiary);
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}

/* links */
.link-list { display: flex; flex-direction: column; gap: 8px; }
.link-item {
  display: flex; align-items: center; gap: 12px; padding: 10px 12px;
  border-radius: var(--radius-sm); background: var(--color-background);
  border: 1px solid var(--color-border-light); text-decoration: none;
  transition: all var(--transition-fast);
}
.link-item:hover { border-color: var(--color-primary); background: var(--color-primary-bg); }
.link-ico {
  display: flex; align-items: center; justify-content: center;
  width: 32px; height: 32px; border-radius: var(--radius-sm);
  background: var(--color-surface-raised); color: var(--color-primary); font-size: 16px;
}
.link-title { font-size: 13px; font-weight: 500; color: var(--color-text); }
.link-desc { font-size: 11px; color: var(--color-text-tertiary); }
</style>
