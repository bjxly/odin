<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  ElDialog,
  ElTabs,
  ElTabPane,
  ElButton,
  ElSelect,
  ElOption,
  ElInput,
  ElCheckbox,
  ElAlert,
  ElIcon,
  ElMessage,
} from 'element-plus'
import {
  Download,
  Upload,
  Document,
  CopyDocument,
} from '@element-plus/icons-vue'
import { useOntologyStore } from '@/stores/ontology'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
}>()

const store = useOntologyStore()

const activeTab = ref<'export' | 'import'>('export')

// 导出配置
const exportFormat = ref<'json' | 'jsonld' | 'owl' | 'yaml'>('json')
const exportOptions = ref({
  includeClasses: true,
  includeRelations: true,
  includeRules: true,
  includeLayout: false,
  includeMappings: false,
})

// 导入配置
const importFormat = ref<'json' | 'jsonld' | 'owl' | 'yaml'>('json')
const importFile = ref<File | null>(null)
const importContent = ref('')
const importMode = ref<'merge' | 'replace'>('merge')

// 文件上传 ref（供模板中 click 触发）
const fileInput = ref<HTMLInputElement | null>(null)

// 导出预览
const exportPreview = computed(() => {
  const ontology = store.currentOntology
  if (!ontology) return ''

  const data: any = {
    version: '1.0',
    id: ontology.id,
    name: ontology.name,
    description: ontology.description,
    exportedAt: new Date().toISOString(),
  }

  if (exportOptions.value.includeClasses) {
    data.classes = store.classes.map((c) => ({
      id: c.id,
      label: c.label,
      description: c.description,
      domain: c.domain,
      synonyms: c.synonyms,
      properties: c.properties,
    }))
  }

  if (exportOptions.value.includeRelations) {
    data.relations = store.relations
  }

  if (exportOptions.value.includeRules) {
    data.rules = store.rules
  }

  if (exportOptions.value.includeLayout) {
    data.layout = store.layout
  }

  if (exportFormat.value === 'json') {
    return JSON.stringify(data, null, 2)
  } else if (exportFormat.value === 'jsonld') {
    return JSON.stringify(convertToJsonLD(data), null, 2)
  } else if (exportFormat.value === 'owl') {
    return convertToOWL(data)
  } else {
    return convertToYAML(data)
  }
})

// 转换为 JSON-LD
function convertToJsonLD(data: any): any {
  return {
    '@context': {
      '@vocab': 'http://www.w3.org/2002/07/owl#',
      'rdf': 'http://www.w3.org/1999/02/22-rdf-syntax-ns#',
      'rdfs': 'http://www.w3.org/2000/01/rdf-schema#',
      'skos': 'http://www.w3.org/2004/02/skos/core#',
    },
    '@type': 'Ontology',
    '@id': data.id,
    'label': data.name,
    'description': data.description,
    'classes': data.classes?.map((c: any) => ({
      '@type': 'Class',
      '@id': c.id,
      'label': c.label,
      'description': c.description,
      'properties': c.properties?.map((p: any) => ({
        '@type': 'Property',
        '@id': p.id,
        'label': p.label,
        'range': p.range,
      })),
    })),
  }
}

// 转换为 OWL
function convertToOWL(data: any): string {
  let owl = `<?xml version="1.0" encoding="UTF-8"?>
<rdf:RDF
  xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
  xmlns:rdfs="http://www.w3.org/2000/01/rdf-schema#"
  xmlns:owl="http://www.w3.org/2002/07/owl#"
  xmlns:skos="http://www.w3.org/2004/02/skos/core#">
  
  <owl:Ontology rdf:about="${data.id}">
    <rdfs:label>${data.name}</rdfs:label>
    <rdfs:comment>${data.description || ''}</rdfs:comment>
  </owl:Ontology>
`

  // 添加类
  if (data.classes) {
    for (const cls of data.classes) {
      owl += `
  <owl:Class rdf:about="${cls.id}">
    <rdfs:label>${cls.label}</rdfs:label>
    <rdfs:comment>${cls.description || ''}</rdfs:comment>
  </owl:Class>`
    }
  }

  // 添加属性
  if (data.classes) {
    for (const cls of data.classes) {
      if (cls.properties) {
        for (const prop of cls.properties) {
          owl += `
  <owl:DatatypeProperty rdf:about="${prop.id}">
    <rdfs:label>${prop.label}</rdfs:label>
    <rdfs:domain rdf:resource="${cls.id}"/>
    <rdfs:range rdf:resource="http://www.w3.org/2001/XMLSchema#${prop.range || 'string'}"/>
  </owl:DatatypeProperty>`
        }
      }
    }
  }

  owl += '\n</rdf:RDF>'
  return owl
}

// 转换为 YAML
function convertToYAML(data: any): string {
  let yaml = `# 本体导出
version: "${data.version}"
id: ${data.id}
name: ${data.name}
description: ${data.description || ''}
exported_at: ${data.exportedAt}

`

  if (data.classes) {
    yaml += 'classes:\n'
    for (const cls of data.classes) {
      yaml += `  - id: ${cls.id}
    label: ${cls.label}
    description: ${cls.description || ''}
`
      if (cls.synonyms?.length) {
        yaml += `    synonyms: [${cls.synonyms.join(', ')}]
`
      }
      if (cls.properties?.length) {
        yaml += `    properties:
`
        for (const prop of cls.properties) {
          yaml += `      - id: ${prop.id}
        label: ${prop.label}
        range: ${prop.range}
`
        }
      }
    }
  }

  return yaml
}

// 导出文件
const handleExport = () => {
  const content = exportPreview.value
  const extensions: Record<string, string> = {
    json: 'json',
    jsonld: 'jsonld',
    owl: 'owl',
    yaml: 'yaml',
  }

  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `ontology-${store.activeOntologyId}.${extensions[exportFormat.value]}`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)

  ElMessage.success('本体导出成功')
}

// 复制到剪贴板
const handleCopy = async () => {
  try {
    await navigator.clipboard.writeText(exportPreview.value)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

// 处理文件上传
const handleFileUpload = (event: Event) => {
  const input = event.target as HTMLInputElement
  if (input.files?.length) {
    importFile.value = input.files[0]
    const reader = new FileReader()
    reader.onload = (e) => {
      importContent.value = e.target?.result as string
    }
    reader.readAsText(input.files[0])
  }
}

// 导入本体
const handleImport = () => {
  if (!importContent.value) {
    ElMessage.warning('请选择文件或粘贴内容')
    return
  }

  try {
    let data: any
    if (importFormat.value === 'json' || importFormat.value === 'jsonld') {
      data = JSON.parse(importContent.value)
    } else {
      // YAML 解析需要额外库，这里简化处理
      ElMessage.warning('YAML 导入暂不支持，请使用 JSON 格式')
      return
    }

    // 应用导入
    if (importMode.value === 'replace') {
      // 替换模式：清除现有数据
      store.resetToMock()
    }

    // 合并类
    if (data.classes) {
      for (const cls of data.classes) {
        if (!store.classExists(cls.id)) {
          store.addClass(cls)
        }
      }
    }

    // 合并关系
    if (data.relations) {
      for (const rel of data.relations) {
        store.addRelation(rel)
      }
    }

    ElMessage.success(`本体导入成功（${importMode.value === 'replace' ? '替换' : '合并'}模式）`)
    emit('update:open', false)
  } catch (e) {
    ElMessage.error(`导入失败: ${e instanceof Error ? e.message : '格式错误'}`)
  }
}

// 关闭对话框
const handleClose = () => {
  emit('update:open', false)
}
</script>

<template>
  <ElDialog
    :model-value="open"
    @update:model-value="handleClose"
    title="本体导入导出"
    width="800px"
    destroy-on-close
  >
    <ElTabs v-model="activeTab">
      <!-- 导出标签页 -->
      <ElTabPane label="导出" name="export">
        <div class="export-section">
          <div class="options-row">
            <div class="option-group">
              <label>导出格式</label>
              <ElSelect v-model="exportFormat" style="width: 200px">
                <ElOption label="JSON" value="json" />
                <ElOption label="JSON-LD" value="jsonld" />
                <ElOption label="OWL (RDF/XML)" value="owl" />
                <ElOption label="YAML" value="yaml" />
              </ElSelect>
            </div>

            <div class="option-group">
              <label>包含内容</label>
              <div class="checkboxes">
                <ElCheckbox v-model="exportOptions.includeClasses">类</ElCheckbox>
                <ElCheckbox v-model="exportOptions.includeRelations">关系</ElCheckbox>
                <ElCheckbox v-model="exportOptions.includeRules">规则</ElCheckbox>
                <ElCheckbox v-model="exportOptions.includeLayout">布局</ElCheckbox>
                <ElCheckbox v-model="exportOptions.includeMappings">映射</ElCheckbox>
              </div>
            </div>
          </div>

          <div class="preview-section">
            <div class="preview-header">
              <span>预览</span>
              <ElButton size="small" :icon="CopyDocument" @click="handleCopy">
                复制
              </ElButton>
            </div>
            <pre class="preview-content"><code>{{ exportPreview }}</code></pre>
          </div>

          <div class="actions">
            <ElButton type="primary" :icon="Download" @click="handleExport">
              导出文件
            </ElButton>
          </div>
        </div>
      </ElTabPane>

      <!-- 导入标签页 -->
      <ElTabPane label="导入" name="import">
        <div class="import-section">
          <ElAlert type="info" :closable="false" style="margin-bottom: 16px">
            <template #title>
              导入本体将合并或替换当前数据，请谨慎操作
            </template>
          </ElAlert>

          <div class="options-row">
            <div class="option-group">
              <label>导入格式</label>
              <ElSelect v-model="importFormat" style="width: 200px">
                <ElOption label="JSON" value="json" />
                <ElOption label="JSON-LD" value="jsonld" />
                <ElOption label="YAML" value="yaml" disabled />
              </ElSelect>
            </div>

            <div class="option-group">
              <label>导入模式</label>
              <ElSelect v-model="importMode" style="width: 200px">
                <ElOption label="合并（保留现有）" value="merge" />
                <ElOption label="替换（清除现有）" value="replace" />
              </ElSelect>
            </div>
          </div>

          <div class="upload-section">
            <div class="upload-area" @click="fileInput?.click()">
              <ElIcon :size="48"><Upload /></ElIcon>
              <p>点击或拖拽文件到此处</p>
              <span>支持 .json, .jsonld 格式</span>
              <input
                ref="fileInput"
                type="file"
                accept=".json,.jsonld"
                style="display: none"
                @change="handleFileUpload"
              />
            </div>

            <div v-if="importFile" class="file-info">
              <ElIcon><Document /></ElIcon>
              <span>{{ importFile.name }}</span>
              <span class="file-size">({{ (importFile.size / 1024).toFixed(1) }} KB)</span>
            </div>
          </div>

          <div class="content-section">
            <label>或粘贴内容</label>
            <ElInput
              v-model="importContent"
              type="textarea"
              :rows="10"
              placeholder="粘贴 JSON 或 JSON-LD 内容..."
            />
          </div>

          <div class="actions">
            <ElButton type="primary" :icon="Upload" @click="handleImport">
              导入本体
            </ElButton>
          </div>
        </div>
      </ElTabPane>
    </ElTabs>
  </ElDialog>
</template>

<style scoped>
.export-section,
.import-section {
  padding: 16px 0;
}

.options-row {
  display: flex;
  gap: 32px;
  margin-bottom: 20px;
}

.option-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.option-group label {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.checkboxes {
  display: flex;
  gap: 16px;
}

.preview-section {
  margin-bottom: 20px;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.preview-header span {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.preview-content {
  background: #1e293b;
  color: #e2e8f0;
  padding: 16px;
  border-radius: 8px;
  max-height: 300px;
  overflow: auto;
  font-size: 12px;
  line-height: 1.6;
}

.actions {
  display: flex;
  justify-content: flex-end;
}

.upload-section {
  margin-bottom: 20px;
}

.upload-area {
  border: 2px dashed #dcdfe6;
  border-radius: 12px;
  padding: 40px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s;
  color: #909399;
}

.upload-area:hover {
  border-color: #409eff;
  color: #409eff;
  background: #f5f7fa;
}

.upload-area p {
  margin: 12px 0 4px;
  font-size: 16px;
  font-weight: 500;
}

.upload-area span {
  font-size: 13px;
}

.file-info {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
  padding: 12px;
  background: #f0f9ff;
  border-radius: 8px;
  color: #2b6cb0;
}

.file-size {
  color: #909399;
}

.content-section {
  margin-bottom: 20px;
}

.content-section label {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}
</style>
