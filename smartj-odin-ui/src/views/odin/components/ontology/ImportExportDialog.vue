<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { UploadFilled, Download } from '@element-plus/icons-vue'
import { exportOntology, importOntology } from '@/api/odin/ontology'

const props = defineProps<{
  visible: boolean
  ontologyId: string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  imported: []
}>()

const { t } = useI18n()

const dialogVisible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
})

const activeTab = ref('export')

// ===== 导出 =====
const exportLoading = ref(false)
const exportJson = ref('')

async function handleExport() {
  if (!props.ontologyId) return
  exportLoading.value = true
  try {
    const data = await exportOntology(props.ontologyId)
    exportJson.value = typeof data === 'string' ? data : JSON.stringify(data, null, 2)
  } catch (e) {
    console.error('[ImportExport] export failed', e)
    ElMessage.error(t('odinOntology.exportLoading'))
  } finally {
    exportLoading.value = false
  }
}

function handleDownload() {
  const blob = new Blob([exportJson.value], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `ontology_${props.ontologyId}.json`
  a.click()
  URL.revokeObjectURL(url)
}

// ===== 导入 =====
const importLoading = ref(false)
const importPreview = ref('')
const importFile = ref<File | null>(null)

function handleFileChange(file: File) {
  importFile.value = file
  const reader = new FileReader()
  reader.onload = (e) => {
    const text = e.target?.result as string
    try {
      JSON.parse(text)
      importPreview.value = text
    } catch {
      importPreview.value = ''
      ElMessage.error(t('odinOntology.importInvalidJson'))
    }
  }
  reader.readAsText(file)
}

function handleDrop(e: DragEvent) {
  const file = e.dataTransfer?.files?.[0]
  if (file) handleFileChange(file)
}

function handleFileSelect(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) handleFileChange(file)
}

async function handleImport() {
  if (!importPreview.value || !props.ontologyId) return
  importLoading.value = true
  try {
    const data = JSON.parse(importPreview.value)
    await importOntology(props.ontologyId, data)
    ElMessage.success(t('odinOntology.importSuccess'))
    dialogVisible.value = false
    emit('imported')
  } catch (e) {
    console.error('[ImportExport] import failed', e)
    ElMessage.error(t('odinOntology.importFailed'))
  } finally {
    importLoading.value = false
  }
}

watch(() => props.visible, (v) => {
  if (v) {
    exportJson.value = ''
    importPreview.value = ''
    importFile.value = null
    activeTab.value = 'export'
    handleExport()
  }
})
</script>

<template>
  <el-dialog v-model="dialogVisible" :title="t('odinOntology.importExportTitle')" width="680px" destroy-on-close>
    <el-tabs v-model="activeTab">
      <!-- 导出 -->
      <el-tab-pane :label="t('odinOntology.exportTab')" name="export">
        <div v-loading="exportLoading">
          <p style="margin: 0 0 8px; font-size: 13px; color: var(--sm-text-secondary)">
            {{ t('odinOntology.exportPreview') }}
          </p>
          <pre class="json-preview">{{ exportJson || '...' }}</pre>
          <div style="margin-top: 12px; text-align: right">
            <el-button type="primary" :icon="Download" :disabled="!exportJson" @click="handleDownload">
              {{ t('odinOntology.exportDownload') }}
            </el-button>
          </div>
        </div>
      </el-tab-pane>

      <!-- 导入 -->
      <el-tab-pane :label="t('odinOntology.importTab')" name="import">
        <div
          class="upload-area"
          @drop.prevent="handleDrop"
          @dragover.prevent
        >
          <el-icon :size="40" color="var(--sm-text-placeholder)"><UploadFilled /></el-icon>
          <p>{{ t('odinOntology.importDragHint') }}</p>
          <input
            type="file"
            accept=".json"
            style="position: absolute; inset: 0; opacity: 0; cursor: pointer"
            @change="handleFileSelect"
          />
        </div>
        <div v-if="importPreview" style="margin-top: 12px">
          <p style="margin: 0 0 8px; font-size: 13px; color: var(--sm-text-secondary)">
            {{ t('odinOntology.importPreview') }}
            <el-tag v-if="importFile" size="small" type="info">{{ importFile.name }}</el-tag>
          </p>
          <pre class="json-preview">{{ importPreview.slice(0, 2000) }}{{ importPreview.length > 2000 ? '...' : '' }}</pre>
          <div style="margin-top: 12px; text-align: right">
            <el-button type="primary" :loading="importLoading" @click="handleImport">
              {{ t('odinOntology.importConfirm') }}
            </el-button>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
  </el-dialog>
</template>

<style scoped>
.json-preview {
  margin: 0;
  padding: 10px 12px;
  background: var(--sm-bg-page);
  border: 1px solid var(--sm-border-lighter);
  border-radius: var(--sm-radius-md);
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: 11px;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 320px;
  overflow: auto;
  color: var(--sm-text-regular);
}

.upload-area {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 32px;
  border: 2px dashed var(--sm-border-base);
  border-radius: var(--sm-radius-lg);
  background: var(--sm-bg-page);
  transition: border-color 0.2s;
  cursor: pointer;
}

.upload-area:hover {
  border-color: var(--sm-color-primary);
}

.upload-area p {
  margin: 0;
  font-size: 13px;
  color: var(--sm-text-secondary);
}
</style>
