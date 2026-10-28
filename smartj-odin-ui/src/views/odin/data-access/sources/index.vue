<template>
  <div class="source-manager">
    <!-- 左侧面板 -->
    <aside class="source-manager__aside">
      <div class="aside-header">
        <el-input
          v-model="searchText"
          :placeholder="t('common.search')"
          clearable
          size="small"
          :prefix-icon="Search"
        />
      </div>
      <div v-loading="dsStore.loading" class="aside-list">
        <div
          v-for="ds in filteredSources"
          :key="ds.id"
          class="source-item"
          :class="{ 'is-active': selectedId === ds.id }"
          @click="handleSelect(ds.id)"
        >
          <span class="status-dot" :class="statusClass(ds)" />
          <span class="source-type-icon">{{ typeIcon(ds.type) }}</span>
          <div class="source-info">
            <span class="source-name">{{ ds.name }}</span>
            <span class="source-host">{{ ds.host || ds.type }}</span>
          </div>
        </div>
        <el-empty v-if="!dsStore.loading && filteredSources.length === 0" :description="t('common.noData')" :image-size="40" />
      </div>
      <div class="aside-footer">
        <el-button v-permission="'odin:datasource:create'" type="primary" size="small" style="width: 100%" @click="handleAdd">
          + {{ t('odinDatasource.addTitle') }}
        </el-button>
      </div>
    </aside>

    <!-- 右侧详情 -->
    <main class="source-manager__main">
      <template v-if="currentSource">
        <div class="main-header">
          <h3 class="main-title">{{ currentSource.name }}</h3>
          <div class="main-actions">
            <el-button size="small" :loading="testing" @click="handleTest">
              {{ t('odinDatasource.testConn') }}
            </el-button>
            <el-button v-permission="'odin:datasource:sync'" size="small" type="warning" plain :loading="syncing" @click="handleSync">
              {{ t('odinDatasource.syncMeta') }}
            </el-button>
            <el-button v-permission="'odin:datasource:update'" size="small" @click="handleEdit(currentSource)">
              {{ t('common.edit') }}
            </el-button>
            <el-button v-permission="'odin:datasource:delete'" size="small" type="danger" plain @click="handleDelete(currentSource)">
              {{ t('common.delete') }}
            </el-button>
          </div>
        </div>

        <el-descriptions :column="3" border size="small" class="main-descriptions">
          <el-descriptions-item :label="t('odinDatasource.code')">{{ currentSource.code || '—' }}</el-descriptions-item>
          <el-descriptions-item :label="t('odinDatasource.name')">{{ currentSource.name }}</el-descriptions-item>
          <el-descriptions-item :label="t('odinDatasource.type')">{{ currentSource.type }}</el-descriptions-item>
          <el-descriptions-item :label="t('odinDatasource.host')">{{ currentSource.host || '—' }}</el-descriptions-item>
          <el-descriptions-item :label="t('odinDatasource.port')">{{ currentSource.port || '—' }}</el-descriptions-item>
          <el-descriptions-item :label="t('odinDatasource.dbName')">{{ currentSource.dbName || '—' }}</el-descriptions-item>
          <el-descriptions-item :label="t('odinDatasource.username')">{{ currentSource.username || '—' }}</el-descriptions-item>
          <el-descriptions-item :label="t('odinDatasource.status')">
            <el-tag :type="currentSource.status === 1 ? 'success' : 'danger'" size="small">
              {{ currentSource.status === 1 ? t('common.enabled') : t('common.disabled') }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item :label="t('odinDatasource.remark')" :span="2">{{ currentSource.remark || '—' }}</el-descriptions-item>
          <el-descriptions-item :label="t('odinDatasource.tags')" :span="3">{{ currentSource.tags || '—' }}</el-descriptions-item>
          <el-descriptions-item :label="t('odinDatasource.createTime')">{{ currentSource.createTime }}</el-descriptions-item>
          <el-descriptions-item :label="'Update Time'">{{ currentSource.updateTime || '—' }}</el-descriptions-item>
        </el-descriptions>

        <el-tabs v-model="activeTab" class="main-tabs">
          <el-tab-pane :label="t('odinDatasource.tables')" name="schema">
            <SchemaBrowser
              :datasource-id="currentSource.id"
              @select-table="handleTableSelect"
            />
          </el-tab-pane>
          <el-tab-pane :label="t('odinDatasource.tableMapping')" name="mapped">
            <TableMappingPanel :datasource-id="currentSource.id" />
          </el-tab-pane>
        </el-tabs>
      </template>
      <el-empty v-else :description="t('odinDatasource.selectSourceHint')" />
    </main>

    <!-- 新增/编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? t('odinDatasource.editTitle') : t('odinDatasource.addTitle')"
      width="640px"
      destroy-on-close
    >
      <el-form ref="editFormRef" :model="editForm" :rules="editRules" label-width="100px">
        <el-form-item :label="t('odinDatasource.code')" prop="code">
          <el-input
            v-model="editForm.code"
            :disabled="isEdit"
            :placeholder="isEdit ? editForm.code : t('common.pleaseInput') + t('odinDatasource.code')"
            :maxlength="64"
          />
        </el-form-item>
        <el-form-item :label="t('odinDatasource.name')" prop="name">
          <el-input v-model="editForm.name" :placeholder="t('common.pleaseInput') + t('odinDatasource.name')" :maxlength="128" />
        </el-form-item>
        <el-form-item :label="t('odinDatasource.type')" prop="type">
          <el-select v-model="editForm.type" :placeholder="t('common.pleaseSelect')" style="width: 100%">
            <el-option v-for="dt in DATASOURCE_TYPES" :key="dt" :label="typeLabel(dt)" :value="dt" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('odinDatasource.status')" prop="status">
          <el-switch v-model="editForm.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item :label="t('odinDatasource.host')" prop="host">
          <el-input v-model="editForm.host" placeholder="localhost" />
        </el-form-item>
        <el-form-item :label="t('odinDatasource.port')" prop="port">
          <el-input-number v-model="editForm.port" :min="0" :max="65535" controls-position="right" />
        </el-form-item>
        <el-form-item :label="t('odinDatasource.dbName')" prop="dbName">
          <el-input v-model="editForm.dbName" :placeholder="t('common.pleaseInput') + t('odinDatasource.dbName')" />
        </el-form-item>
        <el-form-item :label="t('odinDatasource.username')" prop="username">
          <el-input v-model="editForm.username" :placeholder="t('common.pleaseInput') + t('odinDatasource.username')" />
        </el-form-item>
        <el-form-item :label="t('odinDatasource.password')" prop="password">
          <el-input v-model="editForm.password" type="password" show-password :placeholder="isEdit ? t('odinDatasource.passwordTip') : t('common.pleaseInput')" />
        </el-form-item>
        <el-form-item :label="t('odinDatasource.schemaName')" prop="schemaName">
          <el-input v-model="editForm.schemaName" placeholder="public" />
        </el-form-item>
        <el-form-item :label="t('odinDatasource.remark')" prop="remark">
          <el-input v-model="editForm.remark" type="textarea" :rows="2" :maxlength="500" show-word-limit />
        </el-form-item>
        <el-form-item :label="t('odinDatasource.tags')" prop="tags">
          <el-input v-model="editForm.tags" :placeholder="t('odinDatasource.tagsTip')" :maxlength="255" />
        </el-form-item>
        <el-collapse class="advanced-collapse">
          <el-collapse-item :title="t('odinDatasource.advancedConfig')" name="advanced">
            <el-form-item :label="t('odinDatasource.jdbcUrl')" prop="jdbcUrl" label-width="130px">
              <el-input v-model="editForm.jdbcUrl" :placeholder="t('odinDatasource.jdbcUrlTip')" clearable />
            </el-form-item>
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item :label="t('odinDatasource.poolMaxSize')" prop="poolMaxSize" label-width="130px">
                  <el-input-number v-model="editForm.poolMaxSize" :min="1" :max="200" controls-position="right" style="width: 100%" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item :label="t('odinDatasource.poolMinIdle')" prop="poolMinIdle" label-width="130px">
                  <el-input-number v-model="editForm.poolMinIdle" :min="0" :max="100" controls-position="right" style="width: 100%" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item :label="t('odinDatasource.poolMaxLifeS')" prop="poolMaxLifeS" label-width="130px">
                  <el-input-number v-model="editForm.poolMaxLifeS" :min="0" :max="86400" controls-position="right" style="width: 100%" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item :label="t('odinDatasource.connTimeoutS')" prop="connTimeoutS" label-width="130px">
                  <el-input-number v-model="editForm.connTimeoutS" :min="1" :max="600" controls-position="right" style="width: 100%" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item :label="t('odinDatasource.queryTimeoutS')" prop="queryTimeoutS" label-width="130px">
                  <el-input-number v-model="editForm.queryTimeoutS" :min="1" :max="3600" controls-position="right" style="width: 100%" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-form-item :label="t('odinDatasource.options')" prop="options" label-width="130px">
              <el-input v-model="editForm.options" type="textarea" :rows="3" :placeholder="t('odinDatasource.optionsTip')" />
            </el-form-item>
          </el-collapse-item>
        </el-collapse>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ t('common.cancelBtn') }}</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">{{ t('common.confirmBtn') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import SchemaBrowser from '@/views/odin/components/datasource/SchemaBrowser.vue'
import TableMappingPanel from '@/views/odin/components/datasource/TableMappingPanel.vue'
import { useOdinDatasourceStore } from '@/stores/odin/datasource'
import {
  getDatasourceById,
  createDatasource,
  updateDatasource,
  deleteDatasource,
  testDatasource,
  syncDatasource,
} from '@/api/odin/datasource'
import type { DatasourceVO, DatasourceDTO, DatasourceType, SchemaTableVO } from '@/types/odin'

const { t } = useI18n()
const route = useRoute()
const dsStore = useOdinDatasourceStore()

const DATASOURCE_TYPES: DatasourceType[] = [
  'mysql', 'postgresql', 'dameng', 'kingbase', 'gaussdb', 'oracle', 'sqlserver', 'sqlite', 'duckdb_file',
]

const DEFAULT_PORTS: Partial<Record<string, number>> = {
  mysql: 3306, postgresql: 5432, dameng: 5236, kingbase: 54321,
  gaussdb: 5432, oracle: 1521, sqlserver: 1433, sqlite: 0, duckdb_file: 0,
}

function typeLabel(type: string) {
  const key = `odinDatasource.typeLabels.${type}`
  const label = t(key)
  return label === key ? type : label
}

function typeIcon(type: string) {
  const icons: Record<string, string> = {
    mysql: '🐬', postgresql: '🐘', dameng: '🏔️', kingbase: '👑',
    gaussdb: '🌐', oracle: '🔴', sqlserver: '🪟', sqlite: '📦', duckdb_file: '🦆',
  }
  return icons[type] ?? '💾'
}

function statusClass(ds: DatasourceVO) {
  if (ds.status === 1) return 'dot-ok'
  return 'dot-down'
}

// ===== Left panel =====
const searchText = ref('')
const selectedId = ref('')
const activeTab = ref('schema')
const currentSource = computed(() => dsStore.currentSource)

const filteredSources = computed(() => {
  const q = searchText.value.toLowerCase()
  if (!q) return dsStore.sources
  return dsStore.sources.filter(
    (s) => s.name.toLowerCase().includes(q) || s.host?.toLowerCase().includes(q) || s.type.toLowerCase().includes(q),
  )
})

async function handleSelect(id: string) {
  selectedId.value = id
  await dsStore.selectSource(id)
}

function handleTableSelect(_table: SchemaTableVO) {
  // Could open detail or highlight
}

// ===== Test / Sync =====
const testing = ref(false)
const syncing = ref(false)

async function handleTest() {
  if (!currentSource.value) return
  testing.value = true
  try {
    const res = await testDatasource(currentSource.value.id)
    if (res.success) {
      ElMessage.success(t('odinDatasource.testSuccess', { ms: res.latencyMs ?? 0, ver: res.dbVersion ?? '-' }))
    } else {
      ElMessage.error(t('odinDatasource.testFail', { msg: res.errorMessage ?? '' }))
    }
  } finally {
    testing.value = false
  }
}

async function handleSync() {
  if (!currentSource.value) return
  syncing.value = true
  try {
    const count = await syncDatasource(currentSource.value.id)
    ElMessage.success(t('odinDatasource.syncDone', { count }))
  } finally {
    syncing.value = false
  }
}

// ===== Dialog =====
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const editFormRef = ref<FormInstance>()
const editForm = reactive<DatasourceDTO>({
  id: undefined,
  code: '',
  name: '',
  type: 'mysql',
  host: '',
  port: 3306,
  dbName: '',
  schemaName: '',
  username: '',
  password: '',
  jdbcUrl: '',
  options: '',
  tags: '',
  poolMaxSize: undefined,
  poolMinIdle: undefined,
  poolMaxLifeS: undefined,
  connTimeoutS: undefined,
  queryTimeoutS: undefined,
  status: 1,
  remark: '',
})

const editRules = computed<FormRules>(() => ({
  code: [
    { required: true, message: t('common.pleaseInput') + t('odinDatasource.code'), trigger: 'blur' },
    { max: 64, message: `${t('odinDatasource.code')} ≤ 64`, trigger: 'blur' },
  ],
  name: [{ required: true, message: t('common.pleaseInput') + t('odinDatasource.name'), trigger: 'blur' }],
  type: [{ required: true, message: t('common.pleaseSelect') + t('odinDatasource.type'), trigger: 'change' }],
  options: [
    {
      validator: (_rule, value, callback) => {
        const v = value == null ? '' : String(value).trim()
        if (!v) {
          callback()
          return
        }
        try {
          const parsed = JSON.parse(v)
          if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
            callback()
          } else {
            callback(new Error(t('odinDatasource.optionsInvalid')))
          }
        } catch {
          callback(new Error(t('odinDatasource.optionsInvalid')))
        }
      },
      trigger: 'blur',
    },
  ],
}))

watch(
  () => editForm.type,
  () => {
    const knownDefaults = Object.values(DEFAULT_PORTS) as number[]
    if (editForm.port == null || knownDefaults.includes(editForm.port)) {
      editForm.port = DEFAULT_PORTS[editForm.type] ?? 3306
    }
  },
)

function resetEditForm() {
  editForm.id = undefined
  editForm.code = ''
  editForm.name = ''
  editForm.type = 'mysql'
  editForm.host = ''
  editForm.port = 3306
  editForm.dbName = ''
  editForm.schemaName = ''
  editForm.username = ''
  editForm.password = ''
  editForm.jdbcUrl = ''
  editForm.options = ''
  editForm.tags = ''
  editForm.poolMaxSize = undefined
  editForm.poolMinIdle = undefined
  editForm.poolMaxLifeS = undefined
  editForm.connTimeoutS = undefined
  editForm.queryTimeoutS = undefined
  editForm.status = 1
  editForm.remark = ''
}

function handleAdd() {
  isEdit.value = false
  resetEditForm()
  dialogVisible.value = true
}

async function handleEdit(row: DatasourceVO) {
  isEdit.value = true
  const detail = await getDatasourceById(row.id)
  Object.assign(editForm, {
    id: detail.id,
    code: detail.code,
    name: detail.name,
    type: detail.type,
    host: detail.host ?? '',
    port: detail.port ?? DEFAULT_PORTS[detail.type] ?? 3306,
    dbName: detail.dbName ?? '',
    schemaName: detail.schemaName ?? '',
    username: detail.username ?? '',
    password: '',
    jdbcUrl: detail.jdbcUrl ?? '',
    options: detail.options ?? '',
    tags: detail.tags ?? '',
    poolMaxSize: detail.poolMaxSize,
    poolMinIdle: detail.poolMinIdle,
    poolMaxLifeS: detail.poolMaxLifeS,
    connTimeoutS: detail.connTimeoutS,
    queryTimeoutS: detail.queryTimeoutS,
    status: detail.status,
    remark: detail.remark ?? '',
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await editFormRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  try {
    if (isEdit.value) {
      await updateDatasource(editForm)
      ElMessage.success(t('common.updateSuccess'))
    } else {
      await createDatasource(editForm)
      ElMessage.success(t('common.createSuccess'))
    }
    dialogVisible.value = false
    await dsStore.loadSources()
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(row: DatasourceVO) {
  await ElMessageBox.confirm(t('odinDatasource.confirmDelete'), t('common.tip'), { type: 'warning' })
  await deleteDatasource(row.id)
  ElMessage.success(t('common.deleteSuccess'))
  selectedId.value = ''
  dsStore.currentSource = null
  await dsStore.loadSources()
}

onMounted(async () => {
  await dsStore.loadSources()
  // 支持从映射看板深链预选数据源（?datasourceId=xxx）
  const dsId = route.query.datasourceId
  if (typeof dsId === 'string' && dsId) {
    await handleSelect(dsId)
  }
})
</script>

<style scoped>
.source-manager {
  display: flex;
  height: 100%;
  gap: 0;
  overflow: hidden;
}

.source-manager__aside {
  width: 280px;
  min-width: 280px;
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--sm-border-light);
  background: var(--sm-bg-card);
}

.aside-header {
  padding: 12px;
  border-bottom: 1px solid var(--sm-border-lighter);
}

.aside-list {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
}

.source-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  cursor: pointer;
  transition: background 0.15s;
}

.source-item:hover {
  background: var(--sm-bg-page);
}

.source-item.is-active {
  background: color-mix(in srgb, var(--sm-color-primary) 8%, transparent);
  border-right: 3px solid var(--sm-color-primary);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.dot-ok { background: var(--el-color-success); }
.dot-warn { background: var(--el-color-warning); }
.dot-down { background: var(--el-color-danger); }

.source-type-icon {
  font-size: 16px;
  flex-shrink: 0;
}

.source-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.source-name {
  font-size: var(--sm-font-size-sm, 13px);
  font-weight: 500;
  color: var(--sm-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.source-host {
  font-size: 11px;
  color: var(--sm-text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.aside-footer {
  padding: 12px;
  border-top: 1px solid var(--sm-border-lighter);
}

.source-manager__main {
  flex: 1;
  overflow-y: auto;
  padding: 20px 24px;
}

.main-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.main-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--sm-text-primary);
}

.main-actions {
  display: flex;
  gap: 8px;
}

.main-descriptions {
  margin-bottom: 20px;
}

.main-tabs {
  margin-top: 8px;
}

.advanced-collapse {
  margin: 0 0 8px 30px;
  border-top: none;
}
</style>
