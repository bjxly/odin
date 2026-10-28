<script setup lang="ts">
import { ref, computed, reactive, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Edit, Delete, Plus } from '@element-plus/icons-vue'
import { useOdinOntologyStore } from '@/stores/odin/ontology'
import { useOdinDatasourceStore } from '@/stores/odin/datasource'
import { getOntProperties, deleteOntProperty, deleteOntRelation } from '@/api/odin/ontology'
import { getMappingPage, updateMapping, deleteMapping } from '@/api/odin/mapping'
import MappingEditor from '@/views/odin/components/datasource/MappingEditor.vue'
import type {
  OntClassVO,
  OntPropertyVO,
  OntRelationVO,
  OntRuleVO,
  MappingConfigVO,
  DatasourceVO,
  SchemaTableVO,
} from '@/types/odin'

const props = defineProps<{
  visible: boolean
  classId: string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  'edit-class': [cls: OntClassVO]
  'add-property': [classId: string]
  'edit-property': [prop: OntPropertyVO]
  'add-relation': []
  'refresh': []
}>()

const { t } = useI18n()
const store = useOdinOntologyStore()
const dsStore = useOdinDatasourceStore()

const drawerVisible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
})

const currentClass = computed<OntClassVO | null>(() => {
  if (!props.classId) return null
  return store.classes.find((c) => c.id === props.classId) ?? null
})

// ===== 属性 =====
const propList = ref<OntPropertyVO[]>([])
const propLoading = ref(false)

async function loadProperties() {
  if (!props.classId) { propList.value = []; return }
  propLoading.value = true
  try {
    propList.value = await getOntProperties(props.classId)
  } catch (e) {
    console.error('[ClassDrawer] loadProperties failed', e)
  } finally {
    propLoading.value = false
  }
}

// ===== 关系 =====
const relList = computed<OntRelationVO[]>(() => {
  if (!props.classId) return []
  return store.relations.filter(
    (r) => r.sourceClassId === props.classId || r.targetClassId === props.classId,
  )
})

// ===== 规则 =====
/** 解析规则 expression JSON（condition + action），非法/为空返回 null */
function parseRuleExpr(raw?: string): Record<string, any> | null {
  if (!raw) return null
  try {
    return JSON.parse(raw)
  } catch {
    return null
  }
}

/**
 * 仅展示与当前类相关的规则：
 * 1) 规则条件属性属于本类的属性（规则作用于本类实例）；
 * 2) 规则派生目标类（action.target_class）即本类编码（虚拟类由规则派生）。
 */
const ruleList = computed<OntRuleVO[]>(() => {
  if (!props.classId) return []
  const cls = currentClass.value
  if (!cls) return []
  const propCodes = new Set(propList.value.map((p) => p.code))
  return store.rules.filter((r) => {
    const expr = parseRuleExpr(r.expression)
    if (!expr) return false
    const prop = expr.condition?.property
    const target = expr.action?.target_class ?? expr.action?.targetClass
    return (prop && propCodes.has(String(prop))) || target === cls.code
  })
})

/** 规则条件的人读摘要（替代原始 JSON 串展示） */
function ruleCondSummary(r: OntRuleVO): string {
  const cond = parseRuleExpr(r.expression)?.condition
  if (!cond) return ''
  return `${cond.property ?? '?'} ${cond.operator ?? ''} ${cond.value ?? ''}`.trim()
}

// ===== 物理来源（本体类 ↔ 多数据源/多表/多字段映射，类中心视角）=====
interface SourceState {
  mapping: MappingConfigVO
  /** 解析出的 schema_table.id，供列选项与 AI 建议使用；未同步元数据时为空 */
  tableId: string
  savingMeta: boolean
}

const sourceStates = ref<SourceState[]>([])
const mappingLoading = ref(false)

async function loadMappings() {
  if (!props.classId) { sourceStates.value = []; return }
  mappingLoading.value = true
  try {
    const res = await getMappingPage({ current: 1, size: 50, classId: props.classId })
    const list = res?.records ?? []
    // 按数据源去重加载 schema，解析每张表对应的元数据ID
    const dsIds = [...new Set(list.map((m) => m.datasourceId).filter(Boolean))]
    const schemaCache = new Map<string, SchemaTableVO[]>()
    for (const dsId of dsIds) {
      schemaCache.set(dsId, await dsStore.loadSchema(dsId))
    }
    sourceStates.value = list.map((mapping) => {
      const tables = schemaCache.get(mapping.datasourceId) ?? []
      const found = tables.find((tb) => tb.tableName === mapping.tableName)
      return { mapping, tableId: found?.id ?? '', savingMeta: false }
    })
  } catch (e) {
    console.error('[ClassDrawer] loadMappings failed', e)
    sourceStates.value = []
  } finally {
    mappingLoading.value = false
  }
}

/** 保存映射级元数据（优先级 / 过滤条件 / 状态） */
async function saveMeta(state: SourceState) {
  const m = state.mapping
  state.savingMeta = true
  try {
    await updateMapping({
      id: m.id,
      ontologyId: m.ontologyId,
      classId: m.classId,
      datasourceId: m.datasourceId,
      tableName: m.tableName,
      schemaName: m.schemaName,
      columnMappings: m.columnMappings,
      filterCondition: m.filterCondition,
      priority: m.priority,
      status: m.status,
      remark: m.remark,
    })
    ElMessage.success(t('common.updateSuccess'))
    emit('refresh')
  } catch (e) {
    console.error('[ClassDrawer] saveMeta failed', e)
  } finally {
    state.savingMeta = false
  }
}

/** 删除整条物理来源映射 */
async function handleDeleteSource(state: SourceState) {
  await ElMessageBox.confirm(
    t('odinOntology.drawerSourceDeleteConfirm'),
    t('common.tip'),
    { type: 'warning' },
  )
  await deleteMapping(state.mapping.id)
  ElMessage.success(t('common.deleteSuccess'))
  loadMappings()
  emit('refresh')
}

/** 列映射保存后刷新 */
function onSourceSaved() {
  loadMappings()
  emit('refresh')
}

// ===== 添加物理来源（新建 class ↔ 表 映射）=====
const addDialogVisible = ref(false)
const addForm = reactive({ datasourceId: '', tableId: '', tableName: '', schemaName: '' })
const datasourceOptions = ref<DatasourceVO[]>([])
const tableOptions = ref<SchemaTableVO[]>([])

async function openAddSource() {
  addForm.datasourceId = ''
  addForm.tableId = ''
  addForm.tableName = ''
  addForm.schemaName = ''
  tableOptions.value = []
  if (!datasourceOptions.value.length) {
    await dsStore.loadSources()
    datasourceOptions.value = dsStore.sources
  }
  addDialogVisible.value = true
}

async function onAddDsChange(dsId: string) {
  addForm.tableId = ''
  addForm.tableName = ''
  addForm.schemaName = ''
  tableOptions.value = dsId ? await dsStore.loadSchema(dsId) : []
}

function onAddTableChange(tableId: string) {
  const tb = tableOptions.value.find((x) => x.id === tableId)
  addForm.tableName = tb?.tableName ?? ''
  addForm.schemaName = tb?.schemaName ?? ''
}

function onAddSourceSaved() {
  addDialogVisible.value = false
  loadMappings()
  emit('refresh')
}

watch(() => props.classId, () => {
  if (props.classId && props.visible) {
    loadProperties()
    loadMappings()
  }
}, { immediate: true })

watch(() => props.visible, (v) => {
  if (v && props.classId) {
    loadProperties()
    loadMappings()
  }
})

// ===== 操作 =====
async function handleDeleteProp(row: OntPropertyVO) {
  await ElMessageBox.confirm(t('odinOntology.propertyConfirmDelete'), t('common.tip'), { type: 'warning' })
  await deleteOntProperty(row.id)
  ElMessage.success(t('common.deleteSuccess'))
  loadProperties()
  emit('refresh')
}

async function handleDeleteRelation(row: OntRelationVO) {
  await ElMessageBox.confirm(t('odinOntology.relationConfirmDelete'), t('common.tip'), { type: 'warning' })
  await deleteOntRelation(row.id)
  ElMessage.success(t('common.deleteSuccess'))
  emit('refresh')
}

async function handleDeleteClass() {
  if (!currentClass.value) return
  await ElMessageBox.confirm(t('odinOntology.classConfirmDelete'), t('common.tip'), { type: 'warning' })
  await store.deleteClass(currentClass.value.id)
  ElMessage.success(t('common.deleteSuccess'))
  drawerVisible.value = false
  emit('refresh')
}

function parentClassName(): string {
  if (!currentClass.value?.parentClassId) return '—'
  const parent = store.classes.find((c) => c.id === currentClass.value!.parentClassId)
  return parent?.name ?? currentClass.value.parentClassId
}

function targetClassName(id: string): string {
  const cls = store.classes.find((c) => c.id === id)
  return cls?.name ?? id
}
</script>

<template>
  <el-drawer
    v-model="drawerVisible"
    :title="currentClass ? `${currentClass.name}（${currentClass.code}）` : t('odinOntology.drawerTitle')"
    size="760px"
    destroy-on-close
  >
    <template v-if="currentClass">
      <el-collapse model-value="basic">
        <!-- 基本信息 -->
        <el-collapse-item :title="t('odinOntology.drawerBasicInfo')" name="basic">
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item :label="t('odinOntology.className')">{{ currentClass.name }}</el-descriptions-item>
            <el-descriptions-item :label="t('odinOntology.classCode')">
              <code>{{ currentClass.code }}</code>
            </el-descriptions-item>
            <el-descriptions-item :label="t('odinOntology.description')">{{ currentClass.description || '—' }}</el-descriptions-item>
            <el-descriptions-item :label="t('odinOntology.parentClass')">{{ parentClassName() }}</el-descriptions-item>
            <el-descriptions-item :label="t('odinOntology.createTime')">{{ currentClass.createTime || '—' }}</el-descriptions-item>
          </el-descriptions>
        </el-collapse-item>

        <!-- 属性列表 -->
        <el-collapse-item :title="`${t('odinOntology.drawerProperties')}（${propList.length}）`" name="props">
          <div class="block-toolbar">
            <el-button size="small" type="primary" :icon="Plus" @click="emit('add-property', currentClass!.id)">
              {{ t('common.add') }}
            </el-button>
          </div>
          <el-table v-loading="propLoading" :data="propList" size="small" border max-height="280">
            <el-table-column prop="name" :label="t('odinOntology.drawerPropName')" min-width="120" align="left" show-overflow-tooltip />
            <el-table-column prop="code" :label="t('odinOntology.propCode')" min-width="110" align="left" show-overflow-tooltip />
            <el-table-column prop="dataType" :label="t('odinOntology.drawerPropType')" width="90" align="center" />
            <el-table-column prop="required" :label="t('odinOntology.drawerPropRequired')" width="70" align="center">
              <template #default="{ row }">
                <el-tag v-if="row.required === 1" type="danger" size="small">Y</el-tag>
                <span v-else>—</span>
              </template>
            </el-table-column>
            <el-table-column :label="t('common.operation')" width="120" align="center" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="emit('edit-property', row)">
                  {{ t('common.edit') }}
                </el-button>
                <el-button link type="danger" size="small" @click="handleDeleteProp(row)">
                  {{ t('common.delete') }}
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!propLoading && !propList.length" :description="t('odinOntology.drawerNoProperties')" :image-size="40" />
        </el-collapse-item>

        <!-- 关系列表 -->
        <el-collapse-item :title="`${t('odinOntology.drawerRelations')}（${relList.length}）`" name="rels">
          <div class="block-toolbar">
            <el-button size="small" type="primary" :icon="Plus" @click="emit('add-relation')">
              {{ t('common.add') }}
            </el-button>
          </div>
          <el-table :data="relList" size="small" border max-height="220">
            <el-table-column prop="name" :label="t('odinOntology.drawerRelName')" min-width="120" align="left" show-overflow-tooltip />
            <el-table-column :label="t('odinOntology.drawerRelTarget')" min-width="120" align="left" show-overflow-tooltip>
              <template #default="{ row }">
                {{ row.sourceClassId === classId ? targetClassName(row.targetClassId) : targetClassName(row.sourceClassId) }}
              </template>
            </el-table-column>
            <el-table-column prop="cardinality" :label="t('odinOntology.drawerRelCardinality')" width="90" align="center" />
            <el-table-column :label="t('common.operation')" width="80" align="center" fixed="right">
              <template #default="{ row }">
                <el-button link type="danger" size="small" @click="handleDeleteRelation(row)">
                  {{ t('common.delete') }}
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!relList.length" :description="t('odinOntology.drawerNoRelations')" :image-size="40" />
        </el-collapse-item>

        <!-- 物理来源（本体类 ↔ 多数据源/多表/多字段映射） -->
        <el-collapse-item :title="`${t('odinOntology.drawerSources')}（${sourceStates.length}）`" name="sources">
          <div v-loading="mappingLoading">
            <div class="block-toolbar">
              <el-button size="small" type="primary" :icon="Plus" @click="openAddSource">
                {{ t('odinOntology.drawerAddSource') }}
              </el-button>
              <span class="block-hint-inline">{{ t('odinOntology.drawerSourceEditHint') }}</span>
            </div>

            <div v-for="state in sourceStates" :key="state.mapping.id" class="source-card">
              <div class="source-card-header">
                <strong>{{ state.mapping.datasourceName || state.mapping.datasourceId }}</strong>
                <el-tag size="small" type="info">
                  {{ state.mapping.schemaName ? `${state.mapping.schemaName}.` : '' }}{{ state.mapping.tableName }}
                </el-tag>
                <el-button class="source-delete" link type="danger" size="small" @click="handleDeleteSource(state)">
                  {{ t('common.delete') }}
                </el-button>
              </div>

              <!-- 映射级元数据：优先级 / 过滤条件 / 状态 -->
              <div class="source-meta">
                <div class="meta-item">
                  <span class="meta-label">{{ t('odinMapping.priority') }}</span>
                  <el-input-number
                    v-model="state.mapping.priority"
                    size="small"
                    :min="0"
                    :max="9999"
                    controls-position="right"
                    style="width: 110px"
                  />
                </div>
                <div class="meta-item meta-grow">
                  <span class="meta-label">{{ t('odinMapping.filterCondition') }}</span>
                  <el-input
                    v-model="state.mapping.filterCondition"
                    size="small"
                    :placeholder="t('odinMapping.filterTip')"
                    clearable
                  />
                </div>
                <div class="meta-item">
                  <span class="meta-label">{{ t('odinMapping.status') }}</span>
                  <el-switch v-model="state.mapping.status" :active-value="1" :inactive-value="0" size="small" />
                </div>
                <el-button size="small" type="primary" :loading="state.savingMeta" @click="saveMeta(state)">
                  {{ t('common.save') }}
                </el-button>
              </div>

              <!-- 列映射编辑器（属性 ↔ 列，规范契约） -->
              <MappingEditor
                mode="edit"
                :class-id="classId"
                :mapping="state.mapping"
                :table-id="state.tableId"
                @saved="onSourceSaved"
              />
            </div>

            <el-empty v-if="!mappingLoading && !sourceStates.length" :description="t('odinOntology.drawerNoSources')" :image-size="40" />
          </div>
        </el-collapse-item>

        <!-- 关联规则（仅展示与本类相关的规则，点击图中规则节点可单独查看） -->
        <el-collapse-item :title="`${t('odinOntology.drawerRules')}（${ruleList.length}）`" name="rules">
          <ul v-if="ruleList.length" class="rule-list">
            <li v-for="r in ruleList" :key="r.id">
              <el-tag size="small" type="success">{{ r.ruleType }}</el-tag>
              <span class="rule-name">{{ r.name }}</span>
              <span v-if="ruleCondSummary(r)" class="rule-cond">{{ ruleCondSummary(r) }}</span>
              <code v-else-if="r.expression" class="rule-expr">{{ r.expression }}</code>
            </li>
          </ul>
          <el-empty v-else :description="t('odinOntology.drawerNoRules')" :image-size="40" />
        </el-collapse-item>
      </el-collapse>
    </template>

    <!-- 底部操作 -->
    <template #footer>
      <div class="drawer-footer">
        <el-button :icon="Edit" @click="currentClass && emit('edit-class', currentClass)">
          {{ t('odinOntology.drawerEditClass') }}
        </el-button>
        <el-button type="danger" :icon="Delete" @click="handleDeleteClass">
          {{ t('odinOntology.drawerDeleteClass') }}
        </el-button>
      </div>
    </template>
  </el-drawer>

  <!-- 添加物理来源：级联选 数据源 → 表，再编辑列映射并新建映射 -->
  <el-dialog
    v-model="addDialogVisible"
    :title="t('odinOntology.drawerAddSource')"
    width="760px"
    append-to-body
    destroy-on-close
  >
    <div class="add-source-cascade">
      <el-select
        v-model="addForm.datasourceId"
        :placeholder="t('odinMapping.datasource')"
        filterable
        style="width: 240px"
        @change="onAddDsChange"
      >
        <el-option v-for="ds in datasourceOptions" :key="ds.id" :label="ds.name" :value="ds.id" />
      </el-select>
      <el-select
        v-model="addForm.tableId"
        :placeholder="t('odinMapping.tableName')"
        filterable
        style="width: 260px"
        :disabled="!addForm.datasourceId"
        @change="onAddTableChange"
      >
        <el-option
          v-for="tb in tableOptions"
          :key="tb.id"
          :label="tb.schemaName ? `${tb.schemaName}.${tb.tableName}` : tb.tableName"
          :value="tb.id"
        />
      </el-select>
    </div>

    <MappingEditor
      v-if="addForm.tableId && currentClass"
      :key="addForm.tableId"
      mode="create"
      :class-id="classId"
      :ontology-id="currentClass.ontologyId"
      :datasource-id="addForm.datasourceId"
      :table-name="addForm.tableName"
      :schema-name="addForm.schemaName"
      :table-id="addForm.tableId"
      @saved="onAddSourceSaved"
    />
    <el-empty v-else :description="t('odinMapping.selectTableFirst')" :image-size="50" />

    <template #footer>
      <el-button @click="addDialogVisible = false">{{ t('common.cancelBtn') }}</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.block-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.block-hint-inline {
  font-size: 12px;
  color: var(--sm-text-secondary);
  line-height: 1.5;
}

.source-card {
  border: 1px solid var(--sm-border-light);
  border-radius: var(--sm-radius-md);
  padding: 10px 12px;
  margin-bottom: 12px;
  background: var(--sm-bg-card);
}

.source-card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  font-size: 13px;
}

.source-delete {
  margin-left: auto;
}

.source-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  padding-bottom: 10px;
  margin-bottom: 10px;
  border-bottom: 1px dashed var(--sm-border-light);
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.meta-grow {
  flex: 1;
  min-width: 200px;
}

.meta-label {
  font-size: 12px;
  color: var(--sm-text-secondary);
  white-space: nowrap;
}

.add-source-cascade {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.rule-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.rule-list li {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--sm-text-regular);
}

.rule-name {
  font-weight: 500;
  white-space: nowrap;
}

.rule-cond {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  font-size: 12px;
  color: var(--sm-text-secondary);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rule-expr {
  font-size: 11px;
  color: var(--sm-text-secondary);
  background: var(--sm-bg-page);
  border-radius: 3px;
  padding: 1px 6px;
  font-family: 'JetBrains Mono', Consolas, monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.drawer-footer {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}
</style>
