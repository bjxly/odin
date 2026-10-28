<template>
  <div v-loading="loading" class="table-mapping-panel">
    <el-table :data="groups" border size="small" row-key="key">
      <el-table-column type="expand">
        <template #default="{ row }">
          <div class="group-detail">
            <div v-for="m in row.mappings" :key="m.id" class="mapping-item">
              <el-tag size="small" type="success">{{ m.className || m.classId }}</el-tag>
              <span class="mapping-meta">{{ t('odinMapping.priority') }}: {{ m.priority ?? '—' }}</span>
              <span v-if="m.filterCondition" class="mapping-meta">filter: {{ m.filterCondition }}</span>
              <el-tag size="small" :type="m.status === 1 ? 'success' : 'info'">
                {{ m.status === 1 ? t('common.enabled') : t('common.disabled') }}
              </el-tag>
              <span class="mapping-meta">{{ columnCount(m) }} {{ t('odinMapping.columnMappings') }}</span>
              <div class="mapping-item__ops">
                <el-button link type="primary" size="small" @click="openEditColumn(row, m)">
                  {{ t('common.edit') }}
                </el-button>
                <el-button link type="danger" size="small" @click="handleDelete(m)">
                  {{ t('common.delete') }}
                </el-button>
              </div>
            </div>
            <el-empty v-if="!row.mappings.length" :description="t('odinMapping.noClassMapped')" :image-size="40" />
          </div>
        </template>
      </el-table-column>
      <el-table-column :label="t('odinMapping.tableName')" min-width="180" align="left">
        <template #default="{ row }">
          {{ row.schemaName ? row.schemaName + '.' : '' }}{{ row.tableName }}
        </template>
      </el-table-column>
      <el-table-column :label="t('odinMapping.mappedToClasses')" min-width="240" align="left">
        <template #default="{ row }">
          <template v-if="row.mappings.length">
            <el-tag v-for="(c, i) in classTags(row)" :key="i" size="small" class="class-tag">{{ c }}</el-tag>
          </template>
          <span v-else class="unmapped">—</span>
        </template>
      </el-table-column>
      <el-table-column :label="t('odinMapping.mappedColumnCount')" width="110" align="center">
        <template #default="{ row }">
          {{ row.mappings.reduce((s: number, m: MappingConfigVO) => s + columnCount(m), 0) }}
        </template>
      </el-table-column>
      <el-table-column :label="t('common.operation')" width="150" align="center" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" :icon="Plus" @click="openAddClass(row)">
            {{ t('odinMapping.addMapping') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-empty v-if="!loading && !groups.length" :description="t('common.noData')" :image-size="50" />

    <!-- 编辑列映射（表 → 某个已映射类） -->
    <el-dialog
      v-model="editDialogVisible"
      :title="t('odinMapping.editTitle')"
      width="780px"
      append-to-body
      destroy-on-close
    >
      <MappingEditor
        v-if="editTarget"
        mode="edit"
        :class-id="editTarget.classId"
        :mapping="editTarget"
        :table-id="editTableId"
        @saved="onEditSaved"
      />
      <template #footer>
        <el-button @click="editDialogVisible = false">{{ t('common.cancelBtn') }}</el-button>
      </template>
    </el-dialog>

    <!-- 新增类映射（物理表 → 选本体类） -->
    <el-dialog
      v-model="addDialogVisible"
      :title="t('odinMapping.createMappingTitle')"
      width="780px"
      append-to-body
      destroy-on-close
    >
      <div class="add-cascade">
        <el-select
          v-model="addForm.ontologyId"
          :placeholder="t('odinMapping.ontology')"
          filterable
          style="width: 220px"
          @change="onAddOntologyChange"
        >
          <el-option v-for="o in wsStore.ontologies" :key="o.id" :label="o.name" :value="o.id" />
        </el-select>
        <el-select
          v-model="addForm.classId"
          :placeholder="t('odinMapping.ontClass')"
          filterable
          style="width: 220px"
          :disabled="!addForm.ontologyId"
        >
          <el-option v-for="c in classOptions" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>
        <el-tag size="small" type="info">
          {{ addCtx.schemaName ? addCtx.schemaName + '.' : '' }}{{ addCtx.tableName }}
        </el-tag>
      </div>

      <MappingEditor
        v-if="addForm.classId"
        :key="addForm.classId"
        mode="create"
        :class-id="addForm.classId"
        :ontology-id="addForm.ontologyId"
        :datasource-id="datasourceId"
        :table-name="addCtx.tableName"
        :schema-name="addCtx.schemaName"
        :table-id="addCtx.tableId"
        @saved="onAddSaved"
      />
      <el-empty v-else :description="t('odinMapping.selectClassFirst')" :image-size="50" />

      <template #footer>
        <el-button @click="addDialogVisible = false">{{ t('common.cancelBtn') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import MappingEditor from './MappingEditor.vue'
import { useOdinDatasourceStore } from '@/stores/odin/datasource'
import { useOdinMappingStore } from '@/stores/odin/mapping'
import { useOdinWorkspaceStore } from '@/stores/odin/workspace'
import { getOntClassTree } from '@/api/odin/ontology'
import { deleteMapping } from '@/api/odin/mapping'
import { parseColumnMappings } from '@/utils/mapping'
import type { SchemaTableVO, MappingConfigVO, OntClassVO } from '@/types/odin'

const props = defineProps<{ datasourceId: string }>()

const { t } = useI18n()
const dsStore = useOdinDatasourceStore()
const mappingStore = useOdinMappingStore()
const wsStore = useOdinWorkspaceStore()

const loading = ref(false)
const tables = ref<SchemaTableVO[]>([])
const mappings = ref<MappingConfigVO[]>([])

/** 表中心分组：每张物理表 + 映射到它的类映射集合 */
interface TableGroup {
  key: string
  tableId: string
  schemaName: string
  tableName: string
  mappings: MappingConfigVO[]
}

const groups = computed<TableGroup[]>(() => {
  const map = new Map<string, TableGroup>()
  for (const tb of tables.value) {
    const key = `${tb.schemaName ?? ''}.${tb.tableName}`
    map.set(key, {
      key,
      tableId: tb.id,
      schemaName: tb.schemaName ?? '',
      tableName: tb.tableName,
      mappings: [],
    })
  }
  for (const m of mappings.value) {
    const key = `${m.schemaName ?? ''}.${m.tableName}`
    let g = map.get(key)
    if (!g) {
      g = { key, tableId: '', schemaName: m.schemaName ?? '', tableName: m.tableName, mappings: [] }
      map.set(key, g)
    }
    g.mappings.push(m)
  }
  return [...map.values()].sort((a, b) => a.tableName.localeCompare(b.tableName))
})

async function load() {
  if (!props.datasourceId) {
    tables.value = []
    mappings.value = []
    return
  }
  loading.value = true
  try {
    const [tb, mp] = await Promise.all([
      dsStore.loadSchema(props.datasourceId),
      mappingStore.loadMappingsByDatasource(props.datasourceId),
    ])
    tables.value = tb
    mappings.value = mp
  } finally {
    loading.value = false
  }
}

watch(() => props.datasourceId, () => load())
onMounted(load)

function classTags(g: TableGroup): string[] {
  return g.mappings.map((m) => m.className || m.classId)
}

function columnCount(m: MappingConfigVO): number {
  return parseColumnMappings(m.columnMappings).length
}

// ===== 编辑列映射 =====
const editDialogVisible = ref(false)
const editTarget = ref<MappingConfigVO | null>(null)
const editTableId = ref('')

function openEditColumn(g: TableGroup, m: MappingConfigVO) {
  editTarget.value = m
  editTableId.value = g.tableId
  editDialogVisible.value = true
}

function onEditSaved() {
  editDialogVisible.value = false
  load()
}

// ===== 删除映射 =====
async function handleDelete(m: MappingConfigVO) {
  await ElMessageBox.confirm(t('odinMapping.confirmDeleteMapping'), t('common.tip'), { type: 'warning' })
  await deleteMapping(m.id)
  ElMessage.success(t('common.deleteSuccess'))
  load()
}

// ===== 新增类映射（表 → 选本体类）=====
const addDialogVisible = ref(false)
const addCtx = reactive({ tableId: '', tableName: '', schemaName: '' })
const addForm = reactive({ ontologyId: '', classId: '' })
const classOptions = ref<OntClassVO[]>([])

async function openAddClass(g: TableGroup) {
  addCtx.tableId = g.tableId
  addCtx.tableName = g.tableName
  addCtx.schemaName = g.schemaName
  addForm.ontologyId = ''
  addForm.classId = ''
  classOptions.value = []
  if (!wsStore.ontologies.length) await wsStore.loadOntologies()
  addDialogVisible.value = true
}

function flattenClasses(tree: OntClassVO[]): OntClassVO[] {
  const out: OntClassVO[] = []
  for (const n of tree) {
    out.push(n)
    if (n.children?.length) out.push(...flattenClasses(n.children))
  }
  return out
}

async function onAddOntologyChange(ontId: string) {
  addForm.classId = ''
  classOptions.value = []
  if (!ontId) return
  try {
    classOptions.value = flattenClasses(await getOntClassTree(ontId))
  } catch {
    classOptions.value = []
  }
}

function onAddSaved() {
  addDialogVisible.value = false
  load()
}
</script>

<style scoped>
.table-mapping-panel {
  width: 100%;
}

.group-detail {
  padding: 8px 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.mapping-item {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  padding: 6px 10px;
  border: 1px solid var(--sm-border-light);
  border-radius: var(--sm-radius-md);
  background: var(--sm-bg-card);
}

.mapping-meta {
  font-size: 12px;
  color: var(--sm-text-secondary);
}

.mapping-item__ops {
  margin-left: auto;
  display: flex;
  gap: 4px;
}

.class-tag {
  margin-right: 6px;
}

.unmapped {
  color: var(--sm-text-secondary);
}

.add-cascade {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
</style>
