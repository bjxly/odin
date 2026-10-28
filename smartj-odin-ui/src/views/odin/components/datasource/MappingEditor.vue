<template>
  <div class="mapping-editor">
    <!-- AI 建议工具条（需 classId + tableId） -->
    <div v-if="showSuggest && tableId" class="mapping-editor__toolbar">
      <el-button size="small" :loading="suggestLoading" @click="toggleSuggest">
        {{ t('odinMapping.suggest') }}
      </el-button>
      <span class="mapping-editor__hint">{{ t('odinMapping.suggestTip') }}</span>
    </div>
    <MappingSuggest
      v-if="suggestVisible"
      class="mapping-editor__suggest"
      :suggestions="suggestions"
      @select="handleApplySuggest"
    />

    <el-table :data="rows" border size="small" style="width: 100%">
      <el-table-column :label="t('odinMapping.propertyCode')" min-width="170" align="left">
        <template #default="{ row }">
          <el-select
            v-model="row.propertyCode"
            :placeholder="t('common.pleaseSelect')"
            size="small"
            filterable
            allow-create
            default-first-option
            style="width: 100%"
            @change="(v: string) => onPropertyChange(row, v)"
          >
            <el-option v-for="p in propertyOptions" :key="p.code" :label="`${p.name} (${p.code})`" :value="p.code" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column :label="t('odinMapping.columnName')" min-width="150" align="left">
        <template #default="{ row }">
          <el-select
            v-model="row.columnName"
            :placeholder="t('common.pleaseSelect')"
            size="small"
            filterable
            allow-create
            default-first-option
            style="width: 100%"
          >
            <el-option v-for="c in columnOptions" :key="c" :label="c" :value="c" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column :label="t('odinMapping.transform')" width="130" align="left">
        <template #default="{ row }">
          <el-select v-model="row.transform" :placeholder="'none'" size="small" clearable style="width: 100%">
            <el-option v-for="tf in TRANSFORMS" :key="tf" :label="tf" :value="tf" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column :label="t('odinMapping.valueMap')" min-width="150" align="left">
        <template #default="{ row }">
          <el-input v-model="row.valueMap" size="small" placeholder='{"M":"Male"}' />
        </template>
      </el-table-column>
      <el-table-column
        v-if="hasConfidence"
        :label="t('odinMapping.confidence')"
        width="90"
        align="center"
      >
        <template #default="{ row }">
          <span v-if="row.confidence != null">{{ row.confidence }}</span>
          <span v-else>—</span>
        </template>
      </el-table-column>
      <el-table-column :label="t('common.operation')" width="70" align="center" fixed="right">
        <template #default="{ $index }">
          <el-button link type="danger" size="small" @click="removeRow($index)">
            {{ t('common.delete') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-empty v-if="!rows.length" :description="t('common.noData')" :image-size="40" />

    <div class="mapping-editor__footer">
      <el-button size="small" :icon="Plus" @click="addRow">{{ t('common.add') }}</el-button>
      <el-button size="small" type="primary" :loading="saving" @click="handleSave">
        {{ t('common.save') }}
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import MappingSuggest from './MappingSuggest.vue'
import { getOntProperties } from '@/api/odin/ontology'
import { createMapping, updateMapping, suggestColumnMappings } from '@/api/odin/mapping'
import { parseColumnMappings, serializeColumnMappings } from '@/utils/mapping'
import { useOdinDatasourceStore } from '@/stores/odin/datasource'
import type { OntPropertyVO } from '@/types/odin/ontology'
import type { MappingConfigVO, MappingSuggestVO } from '@/types/odin/mapping'

/** 编辑态行：valueMap 以字符串承载便于输入，保存时由 serialize 归一 */
interface EditRow {
  propertyCode: string
  propertyName?: string
  columnName: string
  transform: string
  valueMap: string
  confidence?: number
  validationStatus?: string
}

const props = withDefaults(
  defineProps<{
    /** 目标本体类ID */
    classId: string
    /** 编辑模式：create 新建映射行 / edit 更新已有映射 */
    mode?: 'create' | 'edit'
    /** edit 模式下的既有映射 */
    mapping?: MappingConfigVO | null
    /** create 模式所需：本体ID */
    ontologyId?: string
    /** create 模式所需：数据源ID */
    datasourceId?: string
    /** create 模式所需：物理表名 */
    tableName?: string
    /** create 模式所需：模式名 */
    schemaName?: string
    /** 物理表元数据ID（用于加载列选项与 AI 建议） */
    tableId?: string
    /** 是否显示 AI 建议入口 */
    showSuggest?: boolean
  }>(),
  {
    mode: 'edit',
    mapping: null,
    showSuggest: true,
  },
)

const emit = defineEmits<{
  (e: 'saved', id: string): void
}>()

const { t } = useI18n()
const dsStore = useOdinDatasourceStore()

const TRANSFORMS = ['none', 'uppercase', 'lowercase', 'trim', 'split', 'json_extract']

const rows = ref<EditRow[]>([])
const propertyOptions = ref<OntPropertyVO[]>([])
const columnOptions = ref<string[]>([])
const saving = ref(false)
const suggestions = ref<MappingSuggestVO[]>([])
const suggestVisible = ref(false)
const suggestLoading = ref(false)

const hasConfidence = computed(() => rows.value.some((r) => r.confidence != null))

/** column_mappings JSON → 编辑行 */
function toEditRows(json?: string | null): EditRow[] {
  return parseColumnMappings(json).map((r) => ({
    propertyCode: r.propertyCode,
    propertyName: r.propertyName,
    columnName: r.columnName,
    transform: r.transform ?? 'none',
    valueMap:
      r.valueMap == null ? '' : typeof r.valueMap === 'string' ? r.valueMap : JSON.stringify(r.valueMap),
    confidence: r.confidence,
    validationStatus: r.validationStatus,
  }))
}

async function loadProperties() {
  if (!props.classId) {
    propertyOptions.value = []
    return
  }
  try {
    propertyOptions.value = await getOntProperties(props.classId)
  } catch {
    propertyOptions.value = []
  }
}

async function loadColumns() {
  if (!props.tableId) {
    columnOptions.value = []
    return
  }
  const cols = await dsStore.loadColumns(props.tableId)
  columnOptions.value = cols.map((c) => c.columnName)
}

/** 选择属性时同步展示名快照 */
function onPropertyChange(row: EditRow, code: string) {
  const p = propertyOptions.value.find((x) => x.code === code)
  row.propertyName = p?.name
}

function addRow() {
  rows.value.push({ propertyCode: '', columnName: '', transform: 'none', valueMap: '' })
}

function removeRow(index: number) {
  rows.value.splice(index, 1)
}

async function toggleSuggest() {
  if (suggestVisible.value) {
    suggestVisible.value = false
    return
  }
  if (!props.tableId || !props.classId) return
  suggestLoading.value = true
  try {
    suggestions.value = await suggestColumnMappings(props.classId, props.tableId)
    suggestVisible.value = true
  } catch {
    suggestions.value = []
  } finally {
    suggestLoading.value = false
  }
}

/** 应用 AI 建议：按 propertyCode 合并进当前行 */
function handleApplySuggest(items: MappingSuggestVO[]) {
  for (const s of items) {
    const idx = rows.value.findIndex((r) => r.propertyCode === s.propertyCode)
    if (idx >= 0) {
      rows.value[idx].columnName = s.columnName
      rows.value[idx].confidence = s.confidence
      rows.value[idx].propertyName = s.propertyName
    } else {
      rows.value.push({
        propertyCode: s.propertyCode,
        propertyName: s.propertyName,
        columnName: s.columnName,
        transform: 'none',
        valueMap: '',
        confidence: s.confidence,
        validationStatus: 'suggested',
      })
    }
  }
  suggestVisible.value = false
}

async function handleSave() {
  saving.value = true
  try {
    const columnMappings = serializeColumnMappings(rows.value)
    if (props.mode === 'create') {
      const ontologyId = props.ontologyId ?? props.mapping?.ontologyId
      const datasourceId = props.datasourceId ?? props.mapping?.datasourceId
      const tableName = props.tableName ?? props.mapping?.tableName
      const schemaName = props.schemaName ?? props.mapping?.schemaName
      if (!ontologyId || !props.classId || !datasourceId || !tableName) {
        ElMessage.warning(t('odinMapping.selectTableFirst'))
        return
      }
      const id = await createMapping({
        ontologyId,
        classId: props.classId,
        datasourceId,
        tableName,
        schemaName,
        columnMappings,
        priority: 10,
        status: 1,
      })
      ElMessage.success(t('common.createSuccess'))
      emit('saved', id)
    } else {
      const m = props.mapping
      if (!m) {
        ElMessage.warning(t('odinMapping.selectOntologyFirst'))
        return
      }
      await updateMapping({
        id: m.id,
        ontologyId: m.ontologyId,
        classId: m.classId,
        datasourceId: m.datasourceId,
        tableName: m.tableName,
        schemaName: m.schemaName,
        columnMappings,
        filterCondition: m.filterCondition,
        priority: m.priority,
        status: m.status,
        remark: m.remark,
      })
      ElMessage.success(t('common.updateSuccess'))
      emit('saved', m.id)
    }
  } finally {
    saving.value = false
  }
}

function reload() {
  rows.value = toEditRows(props.mapping?.columnMappings)
}

watch(() => props.classId, () => loadProperties())
watch(() => props.tableId, () => loadColumns())
// 仅当映射身份或列映射内容变化时重载行；避免父级改 priority/filter/status 触发重载丢弃未保存编辑
watch(() => [props.mapping?.id, props.mapping?.columnMappings], () => reload())

onMounted(async () => {
  await Promise.all([loadProperties(), loadColumns()])
  reload()
})
</script>

<style scoped>
.mapping-editor {
  width: 100%;
}

.mapping-editor__toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.mapping-editor__hint {
  font-size: 12px;
  color: var(--sm-text-secondary);
}

.mapping-editor__suggest {
  margin-bottom: 12px;
}

.mapping-editor__footer {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
}
</style>
