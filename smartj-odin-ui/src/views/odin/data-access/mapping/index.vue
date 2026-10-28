<template>
  <div class="mapping-dashboard">
    <!-- 顶部覆盖率区 -->
    <div class="dashboard-header">
      <OntologySwitcher v-model="activeOntologyId" />
      <div class="coverage-section">
        <el-progress
          type="dashboard"
          :percentage="classCoverage"
          :width="80"
          :color="coverageColor"
        >
          <template #default="{ percentage }">
            <span class="coverage-value">{{ percentage }}%</span>
          </template>
        </el-progress>
        <div class="coverage-stats">
          <span>{{ t('odinMapping.mappedClasses') }}: {{ mappedClassCount }}/{{ totalClassCount }}</span>
          <span>{{ t('odinMapping.mappedProps') }}: {{ mappedPropCount }}/{{ totalPropCount }}</span>
        </div>
      </div>
    </div>

    <!-- 只读分析 Tabs -->
    <el-tabs v-model="activeView" class="dashboard-tabs">
      <!-- 映射明细（只读） -->
      <el-tab-pane :label="t('odinMapping.mappingDetail')" name="detail">
        <div v-loading="mappingStore.loading" class="tab-content">
          <el-table :data="mappingStore.mappings" border size="small">
            <el-table-column prop="ontologyName" :label="t('odinMapping.ontology')" min-width="120" align="left" show-overflow-tooltip />
            <el-table-column prop="className" :label="t('odinMapping.ontClass')" min-width="120" align="left" show-overflow-tooltip />
            <el-table-column prop="datasourceName" :label="t('odinMapping.datasource')" min-width="120" align="left" show-overflow-tooltip />
            <el-table-column :label="t('odinMapping.tableName')" min-width="150" align="left">
              <template #default="{ row }">
                {{ row.schemaName ? row.schemaName + '.' : '' }}{{ row.tableName }}
              </template>
            </el-table-column>
            <el-table-column prop="priority" :label="t('odinMapping.priority')" width="80" align="center" />
            <el-table-column :label="t('odinMapping.columnMappings')" width="90" align="center">
              <template #default="{ row }">{{ columnCount(row) }}</template>
            </el-table-column>
            <el-table-column :label="t('odinMapping.status')" width="80" align="center">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
                  {{ row.status === 1 ? t('common.enabled') : t('common.disabled') }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="t('common.operation')" width="90" align="center" fixed="right">
              <template #default="{ row }">
                <el-button
                  v-permission="'odin:mapping:validate'"
                  link
                  type="primary"
                  size="small"
                  :loading="validatingId === row.id"
                  @click="handleValidate(row)"
                >
                  {{ t('odinMapping.validate') }}
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!mappingStore.loading && !mappingStore.mappings.length" :description="t('common.noData')" :image-size="50" />
        </div>
      </el-tab-pane>

      <!-- 未映射类清单 -->
      <el-tab-pane name="unmappedClass">
        <template #label>
          {{ t('odinMapping.unmappedClasses') }}
          <el-badge v-if="unmappedClasses.length" :value="unmappedClasses.length" class="tab-badge" />
        </template>
        <div v-loading="classLoading" class="tab-content">
          <el-table :data="unmappedClasses" border size="small">
            <el-table-column prop="name" :label="t('odinMapping.ontClass')" min-width="180" align="left" />
            <el-table-column prop="description" :label="t('common.remark')" min-width="220" align="left" show-overflow-tooltip />
            <el-table-column :label="t('common.operation')" width="130" align="center" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="goClass(row.id)">
                  {{ t('odinMapping.jumpToClass') }}
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!classLoading && !unmappedClasses.length" :description="t('common.noData')" :image-size="50" />
        </div>
      </el-tab-pane>

      <!-- 未映射表清单 -->
      <el-tab-pane name="unmappedTable">
        <template #label>
          {{ t('odinMapping.unmappedTables') }}
          <el-badge v-if="unmappedTables.length" :value="unmappedTables.length" class="tab-badge" />
        </template>
        <div v-loading="tablesLoading" class="tab-content">
          <el-table :data="unmappedTables" border size="small">
            <el-table-column prop="datasourceName" :label="t('odinMapping.datasource')" min-width="160" align="left" />
            <el-table-column :label="t('odinMapping.tableName')" min-width="200" align="left">
              <template #default="{ row }">
                {{ row.schemaName ? row.schemaName + '.' : '' }}{{ row.tableName }}
              </template>
            </el-table-column>
            <el-table-column :label="t('common.operation')" width="130" align="center" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="goSource(row.datasourceId)">
                  {{ t('odinMapping.jumpToSource') }}
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!tablesLoading && !unmappedTables.length" :description="t('common.noData')" :image-size="50" />
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import OntologySwitcher from '@/views/odin/components/OntologySwitcher.vue'
import { useOdinMappingStore } from '@/stores/odin/mapping'
import { useOdinDatasourceStore } from '@/stores/odin/datasource'
import { useOdinWorkspaceStore } from '@/stores/odin/workspace'
import { getOntClassTree } from '@/api/odin/ontology'
import { parseColumnMappings } from '@/utils/mapping'
import type { OntClassVO } from '@/types/odin/ontology'
import type { SchemaTableVO } from '@/types/odin'
import type { MappingConfigVO } from '@/types/odin/mapping'

const { t } = useI18n()
const router = useRouter()
const mappingStore = useOdinMappingStore()
const dsStore = useOdinDatasourceStore()
const workspaceStore = useOdinWorkspaceStore()

const activeOntologyId = ref('')
const activeView = ref('detail')

// ===== 类清单（当前本体全部类）=====
const classOptions = ref<OntClassVO[]>([])
const classLoading = ref(false)

// ===== 全部物理表（跨数据源，用于未映射表分析）=====
interface DashTable {
  datasourceId: string
  datasourceName: string
  tableId: string
  schemaName: string
  tableName: string
}
const allTables = ref<DashTable[]>([])
const tablesLoading = ref(false)

function flattenClasses(tree: OntClassVO[]): OntClassVO[] {
  const out: OntClassVO[] = []
  for (const node of tree) {
    out.push(node)
    if (node.children?.length) out.push(...flattenClasses(node.children))
  }
  return out
}

// ===== 覆盖率 =====
const mappedClassIds = computed(() => new Set(mappingStore.mappings.map((m) => m.classId)))
const mappedClassCount = computed(
  () => classOptions.value.filter((c) => mappedClassIds.value.has(c.id)).length,
)
const totalClassCount = computed(() => classOptions.value.length)
const mappedPropCount = computed(() =>
  mappingStore.mappings.reduce((s, m) => s + parseColumnMappings(m.columnMappings).length, 0),
)
const totalPropCount = computed(() =>
  classOptions.value.reduce((s, c) => s + (c.propertyCount ?? 0), 0),
)
const classCoverage = computed(() =>
  totalClassCount.value ? Math.round((mappedClassCount.value / totalClassCount.value) * 100) : 0,
)
const coverageColor = computed(() => {
  const rate = classCoverage.value
  if (rate >= 70) return 'var(--el-color-success)'
  if (rate >= 40) return 'var(--el-color-warning)'
  return 'var(--el-color-danger)'
})

// ===== 未映射类 =====
const unmappedClasses = computed(() =>
  classOptions.value.filter((c) => !mappedClassIds.value.has(c.id)),
)

// ===== 未映射表 =====
const mappedTableKeys = computed(
  () => new Set(mappingStore.mappings.map((m) => `${m.datasourceId}|${m.schemaName ?? ''}|${m.tableName}`)),
)
const unmappedTables = computed(() =>
  allTables.value.filter((tb) => !mappedTableKeys.value.has(`${tb.datasourceId}|${tb.schemaName}|${tb.tableName}`)),
)

function columnCount(m: MappingConfigVO): number {
  return parseColumnMappings(m.columnMappings).length
}

// ===== 校验（弹出后端返回的真实结果）=====
const validatingId = ref('')
async function handleValidate(m: MappingConfigVO) {
  validatingId.value = m.id
  try {
    const result = await mappingStore.validateMappings(m.id)
    ElMessageBox.alert(result || t('odinMapping.validateSuccess'), t('odinMapping.validateResult'), {
      type: 'info',
      confirmButtonText: t('common.confirmBtn'),
    })
  } finally {
    validatingId.value = ''
  }
}

// ===== 跳转 =====
function goClass(classId: string) {
  router.push({ path: '/odin/ontology', query: { classId } })
}
function goSource(datasourceId: string) {
  router.push({ path: '/odin/data-access/sources', query: { datasourceId } })
}

// ===== 加载 =====
async function loadClasses(id: string) {
  classLoading.value = true
  try {
    classOptions.value = flattenClasses(await getOntClassTree(id))
  } catch {
    classOptions.value = []
  } finally {
    classLoading.value = false
  }
}

async function loadAllTables() {
  if (!dsStore.sources.length) {
    allTables.value = []
    return
  }
  tablesLoading.value = true
  try {
    const results = await Promise.all(
      dsStore.sources.map(async (ds) => {
        const tables = await dsStore.loadSchema(ds.id)
        return tables.map((tb: SchemaTableVO) => ({
          datasourceId: ds.id,
          datasourceName: ds.name,
          tableId: tb.id,
          schemaName: tb.schemaName ?? '',
          tableName: tb.tableName,
        }))
      }),
    )
    allTables.value = results.flat()
  } finally {
    tablesLoading.value = false
  }
}

watch(activeOntologyId, async (id) => {
  if (!id) return
  workspaceStore.setActive(id)
  await Promise.all([loadClasses(id), mappingStore.loadMappings(id)])
})

onMounted(async () => {
  if (!workspaceStore.ontologies.length) {
    await workspaceStore.loadOntologies()
  }
  activeOntologyId.value = workspaceStore.activeOntologyId
  await dsStore.loadSources()
  await loadAllTables()
})
</script>

<style scoped>
.mapping-dashboard {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.dashboard-header {
  display: flex;
  align-items: center;
  gap: 24px;
  padding: 16px 20px;
  background: var(--sm-bg-card);
  border-bottom: 1px solid var(--sm-border-light);
}

.coverage-section {
  display: flex;
  align-items: center;
  gap: 16px;
}

.coverage-value {
  font-size: 14px;
  font-weight: 600;
  color: var(--sm-text-primary);
}

.coverage-stats {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: var(--sm-font-size-sm, 13px);
  color: var(--sm-text-secondary);
}

.dashboard-tabs {
  flex: 1;
  overflow: hidden;
  padding: 0 20px;
}

.dashboard-tabs :deep(.el-tabs__content) {
  height: calc(100% - 40px);
  overflow-y: auto;
}

.tab-content {
  padding: 16px 0;
}

.tab-badge {
  margin-left: 6px;
  vertical-align: middle;
}

.tab-badge :deep(.el-badge__content) {
  border: none;
}
</style>
