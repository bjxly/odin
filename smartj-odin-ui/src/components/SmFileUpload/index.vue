<template>
  <div class="sm-file-upload">
    <el-upload
      ref="uploadRef"
      :action="action"
      :accept="accept"
      :limit="limit"
      :drag="drag"
      :list-type="listType"
      :file-list="modelValue"
      :before-upload="handleBeforeUpload"
      :on-success="handleSuccess"
      :on-error="handleError"
      :on-exceed="handleExceed"
      :on-remove="handleRemove"
      :on-change="handleChange"
    >
      <template v-if="drag">
        <el-icon class="sm-file-upload__icon"><i-ep-upload-filled /></el-icon>
        <div class="sm-file-upload__text">
          {{ $t('cmp.uploadDragText') }}
          <em>{{ $t('cmp.uploadClick') }}</em>
        </div>
      </template>
      <template v-else>
        <el-button type="primary">{{ $t('cmp.uploadClick') }}</el-button>
      </template>
      <template #tip>
        <div
          v-if="tipText"
          class="sm-file-upload__tip"
        >
          {{ tipText }}
        </div>
      </template>
    </el-upload>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import type { UploadFile, UploadFiles, UploadInstance, UploadRawFile } from 'element-plus'

const { t } = useI18n()

const props = withDefaults(
  defineProps<{
    /** Upload request URL */
    action: string
    /** Accepted file types, e.g. '.jpg,.png,.pdf' */
    accept?: string
    /** Maximum single file size in MB */
    maxSize?: number
    /** Maximum number of files allowed */
    limit?: number
    /** Enable drag-and-drop zone */
    drag?: boolean
    /** Upload list display type */
    listType?: 'text' | 'picture' | 'picture-card'
    /** Bound file list via v-model */
    modelValue?: UploadFiles
  }>(),
  {
    accept: '',
    maxSize: 0,
    limit: 0,
    drag: true,
    listType: 'text',
    modelValue: () => [],
  }
)

const emit = defineEmits<{
  'update:modelValue': [files: UploadFiles]
  success: [response: any, file: UploadFile, fileList: UploadFiles]
  error: [error: Error, file: UploadFile, fileList: UploadFiles]
}>()

const uploadRef = ref<UploadInstance>()

/** Build a human-readable tip from accept / maxSize / limit */
const tipText = computed(() => {
  const parts: string[] = []
  if (props.accept) {
    parts.push(t('cmp.uploadFormatTip', { format: props.accept }))
  }
  if (props.maxSize > 0) {
    parts.push(t('cmp.uploadMaxSize', { size: props.maxSize }))
  }
  if (props.limit > 0) {
    parts.push(t('cmp.uploadMaxCount', { count: props.limit }))
  }
  return parts.join('；')
})

/** Derive a user-friendly file-size string */
function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes}B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)}MB`
}

/**
 * Validate file before upload.
 * Checks file extension (when accept is set) and file size (when maxSize is set).
 */
function handleBeforeUpload(file: UploadRawFile): boolean {
  // -- file type validation --
  if (props.accept) {
    const allowed = props.accept.split(',').map((ext) => ext.trim().toLowerCase())
    const fileName = file.name.toLowerCase()
    const matched = allowed.some((ext) => {
      if (ext.startsWith('.')) {
        return fileName.endsWith(ext)
      }
      // MIME type match, e.g. "image/png"
      if (ext.includes('/')) {
        return file.type === ext || file.type.startsWith(ext.replace('*', ''))
      }
      return fileName.endsWith(`.${ext}`)
    })
    if (!matched) {
      ElMessage.warning(t('cmp.uploadInvalidFormat', { format: props.accept }))
      return false
    }
  }

  // -- file size validation --
  if (props.maxSize > 0 && file.size > props.maxSize * 1024 * 1024) {
    ElMessage.warning(t('cmp.uploadExceedSize', { max: props.maxSize, size: formatSize(file.size) }))
    return false
  }

  return true
}

function handleSuccess(response: any, file: UploadFile, fileList: UploadFiles) {
  emit('update:modelValue', fileList)
  emit('success', response, file, fileList)
}

function handleError(error: Error, file: UploadFile, fileList: UploadFiles) {
  emit('update:modelValue', fileList)
  emit('error', error, file, fileList)
}

function handleExceed() {
  if (props.limit > 0) {
    ElMessage.warning(t('cmp.uploadExceedCount', { count: props.limit }))
  }
}

function handleRemove(_file: UploadFile, fileList: UploadFiles) {
  emit('update:modelValue', fileList)
}

function handleChange(_file: UploadFile, fileList: UploadFiles) {
  emit('update:modelValue', fileList)
}

/** Expose internal el-upload instance for imperative calls */
defineExpose({ uploadRef })
</script>

<style scoped>
.sm-file-upload {
  width: 100%;
}

.sm-file-upload :deep(.el-upload-dragger) {
  padding: var(--sm-spacing-xl) var(--sm-spacing-xxl);
  border-color: var(--sm-border-base);
  border-radius: var(--sm-radius-lg);
  background-color: var(--sm-bg-card);
  transition:
    border-color 0.2s,
    background-color 0.2s;
}

.sm-file-upload :deep(.el-upload-dragger:hover) {
  border-color: var(--sm-color-primary);
  background-color: var(--sm-bg-hover);
}

.sm-file-upload__icon {
  font-size: 48px;
  color: var(--sm-text-placeholder);
  margin-bottom: var(--sm-spacing-sm);
}

.sm-file-upload__text {
  font-size: var(--sm-font-size-base);
  color: var(--sm-text-regular);
  line-height: 1.6;
}

.sm-file-upload__text em {
  font-style: normal;
  color: var(--sm-color-primary);
}

.sm-file-upload__tip {
  margin-top: var(--sm-spacing-sm);
  font-size: var(--sm-font-size-xs);
  color: var(--sm-text-secondary);
}
</style>
