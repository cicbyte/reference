<script setup>
import { ref } from 'vue'
import { message } from 'ant-design-vue'
import { useI18n } from 'vue-i18n'
import { useProjectStore } from '@/stores/project'

const { t } = useI18n()

const props = defineProps({
  open: { type: Boolean, default: false },
})
const emit = defineEmits(['update:open', 'added'])

const project = useProjectStore()
const form = ref({ target: '', local: false, name: '', branch: '' })
const loading = ref(false)

function close() {
  emit('update:open', false)
  form.value = { target: '', local: false, name: '', branch: '' }
}

async function handleSubmit() {
  if (!project.hasProject) {
    message.warning(t('addRepo.errNoProject'))
    return
  }
  if (!form.value.target) {
    message.warning(t('addRepo.errEmptyInput'))
    return
  }
  loading.value = true
  try {
    if (window.go?.main?.ReferenceApp) {
      await window.go.main.ReferenceApp.AddRepo(
        form.value.target,
        form.value.local,
        form.value.name,
        form.value.branch,
      )
      message.success(t('repos.added'))
      emit('added')
      close()
    }
  } catch (e) {
    message.error(t('repos.addFailed') + ': ' + e)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <a-modal
    :open="open"
    :title="t('addRepo.addToProject', { name: project.currentName || t('addRepo.currentProject') })"
    :footer="null"
    :mask-closable="true"
    destroy-on-close
    width="520px"
    @update:open="emit('update:open', $event)"
  >
    <a-form layout="vertical" :model="form" @finish="handleSubmit">
      <a-form-item :label="t('addRepo.repoType')">
        <a-radio-group v-model:value="form.local">
          <a-radio :value="false">{{ t('addRepo.remoteRepo') }}</a-radio>
          <a-radio :value="true">{{ t('addRepo.localRepo') }}</a-radio>
        </a-radio-group>
      </a-form-item>

      <a-form-item :label="form.local ? t('addRepo.targetLocal') : t('addRepo.targetRemote')" required>
        <a-input
          v-model:value="form.target"
          :placeholder="form.local ? t('addRepo.targetLocalPlaceholder') : t('addRepo.targetRemotePlaceholder')"
          allow-clear
        />
      </a-form-item>

      <a-form-item :label="t('addRepo.customName')">
        <a-input v-model:value="form.name" :placeholder="t('addRepo.customNamePlaceholder')" allow-clear />
      </a-form-item>

      <a-form-item :label="t('addRepo.branchOptional')" v-if="!form.local">
        <a-input v-model:value="form.branch" :placeholder="t('addRepo.branchPlaceholder')" allow-clear />
      </a-form-item>

      <a-form-item>
        <a-space>
          <a-button type="primary" html-type="submit" :loading="loading">{{ t('addRepo.submit') }}</a-button>
          <a-button @click="close">{{ t('common.cancel') }}</a-button>
        </a-space>
      </a-form-item>
    </a-form>
  </a-modal>
</template>
