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
import { useI18n } from 'vue-i18n'
import { useProjectStore } from '@/stores/project'

interface ProjectLike {
  dir: string
  name: string
  exists?: boolean
}

export function useProjectActions(onRemoved?: () => void | Promise<void>) {
  const { t } = useI18n()
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
      message.error(t('common.openFailed') + ': ' + e)
    }
  }

  async function onCopyPath(p: ProjectLike) {
    if (!app) return
    try {
      await app.CopyPath(p.dir)
      message.success(t('common.copied'))
    } catch (e) {
      message.error(t('common.copyFailed') + ': ' + e)
    }
  }

  function onRemove(p: ProjectLike, clean: boolean) {
    const label = clean ? t('globalList.removeClean') : t('globalList.removeProject')
    const content = clean
      ? t('globalList.removeCleanContent', { name: p.name })
      : t('globalList.removeContent', { name: p.name })
    Modal.confirm({
      title: t('globalList.removeTitle', { label, name: p.name }),
      content,
      okText: label,
      okType: clean ? 'danger' : 'primary',
      cancelText: t('common.cancelText'),
      async onOk() {
        try {
          await app.RemoveProject(p.dir, clean)
          message.success(t('globalList.removeSuccess', { label }))
          await project.loadProjects()
          if (onRemoved) await onRemoved()
        } catch (e) {
          message.error(t('globalList.removeFailed', { label }) + ': ' + e)
        }
      },
    })
  }

  return { diagnoseOpen, diagnoseDir, onDoctor, onOpenInExplorer, onCopyPath, onRemove }
}
