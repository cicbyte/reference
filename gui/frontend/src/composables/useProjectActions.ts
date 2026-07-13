/**
 * Project action helpers (remove / doctor / copy path / open in explorer).
 *
 * Extracted from the near-identical blocks in ProjectRail.vue and
 * GlobalListView.vue. Both had copies of onRemove (Modal.confirm +
 * RemoveProject), onDoctor, onCopyPath.
 *
 * @param onRemoved optional callback invoked after a successful remove
 *                  (e.g. to reload the caller's local project list).
 */
import { ref } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { useProjectStore } from '../stores/project'

interface ProjectLike {
  dir: string
  name: string
  exists?: boolean
}

export function useProjectActions(onRemoved?: () => void | Promise<void>) {
  const app = window.go?.main?.ReferenceApp
  const project = useProjectStore()

  const diagnoseOpen = ref(false)
  const diagnoseDir = ref('')

  function onDoctor(p: ProjectLike) {
    diagnoseDir.value = p.dir
    diagnoseOpen.value = true
  }

  async function onOpenInExplorer(p: ProjectLike) {
    if (!app) return
    try {
      await app.OpenInExplorer(p.dir)
    } catch (e) {
      message.error('打开失败: ' + e)
    }
  }

  async function onCopyPath(p: ProjectLike) {
    if (!app) return
    try {
      await app.CopyPath(p.dir)
      message.success('路径已复制')
    } catch (e) {
      message.error('复制失败: ' + e)
    }
  }

  function onRemove(p: ProjectLike, clean: boolean) {
    const label = clean ? '移除并清除 .reference' : '移除项目'
    const content = clean
      ? `将删除 ${p.name} 的所有引用记录、.reference 目录及注入的 AI 配置文件。此操作不可撤销。`
      : `将删除 ${p.name} 的所有引用记录和链接，保留 .reference 目录。`
    Modal.confirm({
      title: `${label} — ${p.name}`,
      content,
      okText: label,
      okType: clean ? 'danger' : 'primary',
      cancelText: '取消',
      async onOk() {
        try {
          await app.RemoveProject(p.dir, clean)
          message.success(`${label}成功`)
          await project.loadProjects()
          if (onRemoved) await onRemoved()
        } catch (e) {
          message.error(`${label}失败: ` + e)
        }
      },
    })
  }

  return { diagnoseOpen, diagnoseDir, onDoctor, onOpenInExplorer, onCopyPath, onRemove }
}
