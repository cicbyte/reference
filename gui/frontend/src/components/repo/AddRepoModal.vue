<script setup>
import { ref } from 'vue'
import { message } from 'ant-design-vue'
import { useProjectStore } from '../../stores/project'

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
    message.warning('请先选择一个项目')
    return
  }
  if (!form.value.target) {
    message.warning('请输入仓库地址或路径')
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
      message.success('仓库添加成功')
      emit('added')
      close()
    }
  } catch (e) {
    message.error('添加失败: ' + e)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <a-modal
    :open="open"
    :title="'添加仓库到 ' + (project.currentName || '当前项目')"
    :footer="null"
    :mask-closable="true"
    destroy-on-close
    width="520px"
    @update:open="emit('update:open', $event)"
  >
    <a-form layout="vertical" :model="form" @finish="handleSubmit">
      <a-form-item label="仓库类型">
        <a-radio-group v-model:value="form.local">
          <a-radio :value="false">远程仓库 (Git URL)</a-radio>
          <a-radio :value="true">本地仓库</a-radio>
        </a-radio-group>
      </a-form-item>

      <a-form-item :label="form.local ? '本地路径' : 'Git URL'" required>
        <a-input
          v-model:value="form.target"
          :placeholder="form.local ? '/path/to/repo' : 'https://github.com/owner/repo 或 owner/repo'"
          allow-clear
        />
      </a-form-item>

      <a-form-item label="自定义名称（可选）">
        <a-input v-model:value="form.name" placeholder="留空则自动识别" allow-clear />
      </a-form-item>

      <a-form-item label="指定分支（可选）" v-if="!form.local">
        <a-input v-model:value="form.branch" placeholder="留空则使用默认分支" allow-clear />
      </a-form-item>

      <a-form-item>
        <a-space>
          <a-button type="primary" html-type="submit" :loading="loading">添加</a-button>
          <a-button @click="close">取消</a-button>
        </a-space>
      </a-form-item>
    </a-form>
  </a-modal>
</template>
