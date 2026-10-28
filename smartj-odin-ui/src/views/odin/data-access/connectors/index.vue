<template>
  <div class="connector-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1>{{ t('odinConnector.title') }}</h1>
        <p class="subtitle">{{ t('odinConnector.subtitle') }}</p>
      </div>
      <div class="header-stats">
        <div class="stat-item">
          <span class="stat-value">{{ stats.total }}</span>
          <span class="stat-label">{{ t('odinConnector.stats.total') }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-value ready">{{ stats.ready }}</span>
          <span class="stat-label">{{ t('odinConnector.stats.ready') }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-value need-driver">{{ stats.needDriver }}</span>
          <span class="stat-label">{{ t('odinConnector.stats.needDriver') }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-value disabled">{{ stats.disabled }}</span>
          <span class="stat-label">{{ t('odinConnector.stats.disabled') }}</span>
        </div>
      </div>
    </div>

    <!-- 过滤器 -->
    <div class="filter-section">
      <div class="filter-row">
        <el-input
          v-model="searchQuery"
          :placeholder="t('odinConnector.filter.searchPlaceholder')"
          :prefix-icon="Search"
          clearable
          class="search-input"
        />
        <el-button
          v-permission="'odin:datasource:create'"
          type="primary"
          :icon="Plus"
          @click="handleAdd"
        >
          {{ t('common.add') }}
        </el-button>
      </div>

      <div class="filter-tags">
        <el-tag
          :type="categoryFilter === null ? 'primary' : 'info'"
          effect="plain"
          class="filter-tag"
          @click="categoryFilter = null"
        >
          {{ t('odinConnector.filter.allCategories') }}
        </el-tag>
        <el-tag
          v-for="cat in categories"
          :key="cat.id"
          :type="categoryFilter === cat.id ? 'primary' : 'info'"
          effect="plain"
          class="filter-tag"
          @click="categoryFilter = categoryFilter === cat.id ? null : cat.id"
        >
          {{ cat.icon }} {{ t(`odinConnector.category.${cat.id}`) }}
        </el-tag>
      </div>

      <div class="filter-tags">
        <el-tag
          :type="sourceFilter === null ? 'primary' : 'info'"
          effect="plain"
          class="filter-tag"
          @click="sourceFilter = null"
        >
          {{ t('odinConnector.filter.allSources') }}
        </el-tag>
        <el-tag
          :type="sourceFilter === 'builtin' ? 'primary' : 'info'"
          effect="plain"
          class="filter-tag"
          @click="sourceFilter = sourceFilter === 'builtin' ? null : 'builtin'"
        >
          {{ t('odinConnector.filter.builtin') }}
        </el-tag>
        <el-tag
          :type="sourceFilter === 'custom' ? 'primary' : 'info'"
          effect="plain"
          class="filter-tag"
          @click="sourceFilter = sourceFilter === 'custom' ? null : 'custom'"
        >
          {{ t('odinConnector.filter.custom') }}
        </el-tag>
      </div>

      <div class="filter-tags">
        <el-tag
          :type="statusFilter === null ? 'primary' : 'info'"
          effect="plain"
          class="filter-tag"
          @click="statusFilter = null"
        >
          {{ t('odinConnector.filter.allStatuses') }}
        </el-tag>
        <el-tag
          :type="statusFilter === 'ready' ? 'primary' : 'info'"
          effect="plain"
          class="filter-tag"
          @click="statusFilter = statusFilter === 'ready' ? null : 'ready'"
        >
          {{ t('odinConnector.filter.ready') }}
        </el-tag>
        <el-tag
          :type="statusFilter === 'need_driver' ? 'primary' : 'info'"
          effect="plain"
          class="filter-tag"
          @click="statusFilter = statusFilter === 'need_driver' ? null : 'need_driver'"
        >
          {{ t('odinConnector.filter.needDriver') }}
        </el-tag>
        <el-tag
          :type="statusFilter === 'disabled' ? 'primary' : 'info'"
          effect="plain"
          class="filter-tag"
          @click="statusFilter = statusFilter === 'disabled' ? null : 'disabled'"
        >
          {{ t('odinConnector.filter.disabled') }}
        </el-tag>
      </div>
    </div>

    <!-- 连接器卡片网格 -->
    <div v-loading="loading" class="connectors-grid">
      <el-empty
        v-if="!loading && filteredConnectors.length === 0"
        :description="t('odinConnector.empty')"
      />

      <el-row :gutter="20">
        <el-col
          v-for="c in filteredConnectors"
          :key="c.id"
          :xs="24"
          :sm="12"
          :md="8"
          :lg="8"
          :xl="6"
        >
          <el-card class="connector-card" shadow="hover" :body-style="{ padding: '16px' }">
            <div class="card-header">
              <div class="connector-icon">{{ c.icon }}</div>
              <div class="connector-info">
                <h3>{{ c.displayName }}</h3>
                <div class="connector-meta">
                  <el-tag size="small" effect="plain" type="info">
                    {{ t(`odinConnector.category.${c.category}`) }}
                  </el-tag>
                  <el-tag size="small" :style="versionTagStyle">
                    JDBC · {{ driverVersion(c.driverLib) }}
                  </el-tag>
                  <el-tag size="small" effect="plain" :type="statusTagType(c.status)">
                    {{ statusLabel(c.status) }}
                  </el-tag>
                  <el-tag v-if="c.builtin" size="small" effect="plain" type="success">
                    {{ t('odinConnector.builtin') }}
                  </el-tag>
                </div>
              </div>
            </div>

            <p class="connector-desc">{{ c.description || '—' }}</p>

            <div class="connector-details">
              <div class="detail-row">
                <span class="label">{{ t('odinConnector.card.driverLib') }}</span>
                <code class="value">{{ c.driverLib || '—' }}</code>
              </div>
              <div class="detail-row">
                <span class="label">{{ t('odinConnector.card.code') }}</span>
                <code class="value">{{ c.code }}</code>
              </div>
              <div v-if="c.defaultPort" class="detail-row">
                <span class="label">{{ t('odinConnector.card.defaultPort') }}</span>
                <code class="value">{{ c.defaultPort }}</code>
              </div>
            </div>

            <div class="card-actions">
              <el-button
                v-if="c.status === 'need_driver'"
                type="primary"
                size="small"
                :icon="UploadFilled"
                @click="openUploadDialog(c)"
              >
                {{ t('odinConnector.card.uploadDriver') }}
              </el-button>
              <el-button
                v-else
                size="small"
                type="warning"
                plain
                :icon="UploadFilled"
                @click="openUploadDialog(c)"
              >
                {{ t('odinConnector.card.updateDriver') }}
              </el-button>
              <el-button
                v-if="!c.builtin"
                v-permission="'odin:datasource:update'"
                link
                type="primary"
                size="small"
                @click="handleEdit(c)"
              >
                {{ t('common.edit') }}
              </el-button>
              <el-button
                v-if="!c.builtin"
                v-permission="'odin:datasource:delete'"
                link
                type="danger"
                size="small"
                @click="handleDelete(c)"
              >
                {{ t('common.delete') }}
              </el-button>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 新增/编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? t('odinConnector.editTitle') : t('odinConnector.addTitle')"
      width="560px"
      destroy-on-close
    >
      <el-form ref="editFormRef" :model="editForm" :rules="editRules" label-width="110px">
        <el-form-item :label="t('odinConnector.name')" prop="name">
          <el-input v-model="editForm.name" :placeholder="t('common.pleaseInput') + t('odinConnector.name')" :maxlength="128" />
        </el-form-item>
        <el-form-item :label="t('odinConnector.code')" prop="code">
          <el-input v-model="editForm.code" :placeholder="t('common.pleaseInput') + t('odinConnector.code')" :maxlength="64" />
        </el-form-item>
        <el-form-item :label="t('odinConnector.dbType')" prop="dbType">
          <el-select v-model="editForm.dbType" :placeholder="t('common.pleaseSelect')" style="width: 100%">
            <el-option v-for="dt in DB_TYPES" :key="dt.value" :label="dt.label" :value="dt.value" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('odinConnector.driverClass')" prop="driverClass">
          <el-input v-model="editForm.driverClass" placeholder="com.mysql.cj.jdbc.Driver" :maxlength="256" />
        </el-form-item>
        <el-form-item :label="t('odinConnector.urlPattern')" prop="urlPattern">
          <el-input v-model="editForm.urlPattern" placeholder="jdbc:mysql://{host}:{port}/{db}" :maxlength="500" />
        </el-form-item>
        <el-form-item :label="t('odinConnector.defaultPort')" prop="defaultPort">
          <el-input-number v-model="editForm.defaultPort" :min="0" :max="65535" controls-position="right" />
        </el-form-item>
        <el-form-item :label="t('odinConnector.description')" prop="description">
          <el-input v-model="editForm.description" type="textarea" :rows="2" :maxlength="500" show-word-limit />
        </el-form-item>
        <el-form-item :label="t('odinConnector.status')" prop="status">
          <el-radio-group v-model="editForm.status">
            <el-radio :value="1">{{ t('common.enabled') }}</el-radio>
            <el-radio :value="0">{{ t('common.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ t('common.cancelBtn') }}</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">{{ t('common.confirmBtn') }}</el-button>
      </template>
    </el-dialog>

    <!-- 上传驱动对话框 -->
    <el-dialog
      v-model="uploadDialogVisible"
      :title="t('odinConnector.upload.title', { name: uploadTarget?.displayName ?? '' })"
      width="480px"
      destroy-on-close
    >
      <div class="upload-dialog-body">
        <p class="upload-hint">{{ uploadTarget?.description }}</p>
        <p class="upload-lib">
          {{ t('odinConnector.upload.needFile') }}
          <code>{{ uploadTarget?.driverLib || 'JDBC Driver JAR' }}</code>
        </p>
        <el-upload
          ref="uploadRef"
          drag
          accept=".jar"
          :limit="1"
          :auto-upload="false"
          :on-change="handleFileChange"
          :on-remove="handleFileRemove"
          :on-exceed="() => ElMessage.warning(t('odinConnector.upload.onlyOne'))"
          class="driver-upload"
        >
          <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
          <div class="el-upload__text">
            {{ t('odinConnector.upload.drop') }}<em>{{ t('odinConnector.upload.select') }}</em>
          </div>
        </el-upload>
      </div>
      <template #footer>
        <el-button @click="uploadDialogVisible = false">{{ t('common.cancelBtn') }}</el-button>
        <el-button type="primary" :loading="uploading" @click="submitUpload">
          {{ t('odinConnector.upload.submit') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules, UploadInstance, UploadFile } from 'element-plus'
import { Search, UploadFilled, Plus } from '@element-plus/icons-vue'
import {
  getConnectorList,
  getDrivers,
  createConnector,
  updateConnector,
  deleteConnectors,
  testConnector,
  uploadDriver,
} from '@/api/odin/connector'
import type { ConnectorVO, ConnectorDTO, DriverInfo } from '@/types/odin/connector'

const { t } = useI18n()

// ===== 数据库类型元数据 =====
const DB_TYPES = [
  { label: 'MySQL', value: 'mysql' },
  { label: 'PostgreSQL', value: 'postgresql' },
  { label: 'SQL Server', value: 'sqlserver' },
  { label: '达梦 DM8', value: 'dameng' },
  { label: '人大金仓 KingbaseES', value: 'kingbase' },
  { label: 'openGauss / GaussDB', value: 'opengauss' },
  { label: 'Oracle', value: 'oracle' },
  { label: 'SQLite', value: 'sqlite' },
  { label: 'DuckDB', value: 'duckdb' },
  { label: 'GBase 8s', value: 'gbase' },
]

/** 分类元数据：dbType → category / icon */
const CATEGORY_META: Record<string, { category: string; icon: string }> = {
  mysql: { category: 'rdbms', icon: '🐬' },
  postgresql: { category: 'rdbms', icon: '🐘' },
  sqlserver: { category: 'rdbms', icon: '🪟' },
  dameng: { category: 'rdbms', icon: '💾' },
  kingbase: { category: 'rdbms', icon: '👑' },
  opengauss: { category: 'rdbms', icon: '🌐' },
  oracle: { category: 'rdbms', icon: '🔴' },
  sqlite: { category: 'rdbms', icon: '📦' },
  duckdb: { category: 'rdbms', icon: '🦆' },
  gbase: { category: 'rdbms', icon: '🅶' },
  mongodb: { category: 'nosql', icon: '🍃' },
  redis: { category: 'nosql', icon: '🧠' },
  elasticsearch: { category: 'nosql', icon: '🔍' },
  csv: { category: 'file', icon: '📄' },
  excel: { category: 'file', icon: '📊' },
  json: { category: 'file', icon: '📋' },
}

/** 连接器展示模型（融合后端 ConnectorVO + 原型 DriverInfo 字段） */
interface ConnectorDisplay {
  id: string
  code: string
  displayName: string
  dbType: string
  category: string
  icon: string
  driverType: 'native' | 'jdbc'
  driverLib: string
  driverClass: string
  urlPattern: string
  defaultPort?: number
  description: string
  status: 'ready' | 'need_driver' | 'disabled'
  builtin: boolean
  createTime?: string
  raw?: ConnectorVO
}

/** 将后端 DriverInfo（内置 JDBC 驱动）归一化为展示模型 */
function normalizeDriver(d: DriverInfo): ConnectorDisplay {
  const meta = CATEGORY_META[d.dbType] ?? { category: 'other', icon: '🔌' }
  const version = d.version ?? ''
  return {
    id: d.id,
    code: d.dbType,
    displayName: d.driverName,
    dbType: d.dbType,
    category: meta.category,
    icon: meta.icon,
    driverType: 'jdbc',
    driverLib: version ? `${d.description ?? d.driverName} ${version}` : (d.description ?? d.driverName),
    driverClass: d.driverClass,
    urlPattern: d.urlPattern,
    defaultPort: d.defaultPort,
    description: d.description ?? '',
    // 框架内置驱动均随包提供，统一按已就绪处理
    status: 'ready',
    builtin: d.builtin ?? true,
  }
}

// ===== 状态 =====
const loading = ref(false)
const connectors = ref<ConnectorDisplay[]>([])
const searchQuery = ref('')
const categoryFilter = ref<string | null>(null)
const sourceFilter = ref<'builtin' | 'custom' | null>(null)
const statusFilter = ref<'ready' | 'need_driver' | 'disabled' | null>(null)

// ===== 派生 =====
const categories = computed(() => {
  const seen = new Set<string>()
  for (const c of connectors.value) {
    if (c.category) seen.add(c.category)
  }
  const order = ['rdbms', 'nosql', 'file', 'other']
  return [...seen]
    .sort((a, b) => order.indexOf(a) - order.indexOf(b))
    .map((id) => ({
      id,
      icon: CATEGORY_META[Object.keys(CATEGORY_META).find((k) => CATEGORY_META[k].category === id) ?? '']?.icon ?? '🔌',
    }))
})

const filteredConnectors = computed(() => {
  let result = [...connectors.value]
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    result = result.filter(
      (c) =>
        c.displayName.toLowerCase().includes(q) ||
        c.code.toLowerCase().includes(q) ||
        c.description?.toLowerCase().includes(q) ||
        c.driverLib?.toLowerCase().includes(q) ||
        c.driverClass?.toLowerCase().includes(q),
    )
  }
  if (categoryFilter.value) result = result.filter((c) => c.category === categoryFilter.value)
  if (sourceFilter.value) result = result.filter((c) => (c.builtin ? 'builtin' : 'custom') === sourceFilter.value)
  if (statusFilter.value) result = result.filter((c) => c.status === statusFilter.value)
  return result
})

const stats = computed(() => {
  const list = connectors.value
  return {
    total: list.length,
    ready: list.filter((c) => c.status === 'ready').length,
    needDriver: list.filter((c) => c.status === 'need_driver').length,
    disabled: list.filter((c) => c.status === 'disabled').length,
  }
})

/** 版本标签样式（统一 JDBC 主题色） */
const versionTagStyle = {
  background: 'color-mix(in srgb, var(--sm-color-primary) 12%, transparent)',
  color: 'var(--sm-color-primary)',
  borderColor: 'color-mix(in srgb, var(--sm-color-primary) 40%, transparent)',
}

/** 从驱动库字符串（形如 "mysql-connector-j 9.7.0"）解析版本号 */
function driverVersion(lib: string): string {
  if (!lib) return '—'
  const parts = lib.trim().split(/\s+/)
  return parts.length > 1 ? parts[parts.length - 1] : lib
}

function statusTagType(status: string): 'success' | 'warning' | 'info' | 'danger' {
  if (status === 'ready') return 'success'
  if (status === 'need_driver') return 'warning'
  if (status === 'disabled') return 'danger'
  return 'info'
}

function statusLabel(status: string): string {
  if (status === 'ready') return t('odinConnector.filter.ready')
  if (status === 'need_driver') return t('odinConnector.filter.needDriver')
  if (status === 'disabled') return t('odinConnector.filter.disabled')
  return status
}

/** 将后端 ConnectorVO 归一化为展示模型（框架统一使用 JDBC 驱动） */
function normalizeConnector(vo: ConnectorVO): ConnectorDisplay {
  const meta = CATEGORY_META[vo.dbType] ?? { category: 'other', icon: '🔌' }
  const status: ConnectorDisplay['status'] = vo.status === 0 ? 'disabled' : 'ready'
  return {
    id: vo.id,
    code: vo.code || vo.dbType,
    displayName: vo.name,
    dbType: vo.dbType,
    category: meta.category,
    icon: meta.icon,
    driverType: 'jdbc',
    driverLib: vo.driverClass?.split('.').pop() ?? '',
    driverClass: vo.driverClass,
    urlPattern: vo.urlPattern,
    defaultPort: vo.defaultPort,
    description: vo.description ?? '',
    status,
    builtin: false,
    createTime: vo.createTime,
    raw: vo,
  }
}

async function loadData() {
  loading.value = true
  try {
    // 内置 JDBC 驱动来自后端 drivers()，自定义连接器来自 list()，二者并行加载
    const [drivers, list] = await Promise.all([
      getDrivers().catch((e) => {
        console.error('[connectors] load drivers failed', e)
        return [] as DriverInfo[]
      }),
      getConnectorList().catch((e) => {
        console.error('[connectors] load list failed', e)
        return [] as ConnectorVO[]
      }),
    ])
    const builtins = (drivers ?? []).map(normalizeDriver)
    const customs = (list ?? []).map(normalizeConnector)
    // 去重：自定义连接器若与内置驱动 code 相同，则覆盖内置项
    const customCodes = new Set(customs.map((c) => c.code))
    const mergedBuiltins = builtins.filter((b) => !customCodes.has(b.code))
    connectors.value = [...customs, ...mergedBuiltins]
  } finally {
    loading.value = false
  }
}

// ===== CRUD =====
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const editFormRef = ref<FormInstance>()
const editForm = reactive<ConnectorDTO>({
  id: undefined,
  name: '',
  code: '',
  driverClass: '',
  urlPattern: '',
  dbType: 'mysql',
  defaultPort: undefined,
  description: '',
  status: 1,
})

const editRules = computed<FormRules>(() => ({
  name: [{ required: true, message: t('common.pleaseInput') + t('odinConnector.name'), trigger: 'blur' }],
  code: [{ required: true, message: t('common.pleaseInput') + t('odinConnector.code'), trigger: 'blur' }],
  dbType: [{ required: true, message: t('common.pleaseSelect') + t('odinConnector.dbType'), trigger: 'change' }],
  driverClass: [{ required: true, message: t('common.pleaseInput') + t('odinConnector.driverClass'), trigger: 'blur' }],
  urlPattern: [{ required: true, message: t('common.pleaseInput') + t('odinConnector.urlPattern'), trigger: 'blur' }],
}))

function resetEditForm() {
  editForm.id = undefined
  editForm.name = ''
  editForm.code = ''
  editForm.driverClass = ''
  editForm.urlPattern = ''
  editForm.dbType = 'mysql'
  editForm.defaultPort = undefined
  editForm.description = ''
  editForm.status = 1
}

function handleAdd() {
  isEdit.value = false
  resetEditForm()
  dialogVisible.value = true
}

function handleEdit(c: ConnectorDisplay) {
  isEdit.value = true
  Object.assign(editForm, {
    id: c.raw?.id ?? c.id,
    name: c.displayName,
    code: c.code,
    driverClass: c.driverClass,
    urlPattern: c.urlPattern,
    dbType: c.dbType,
    defaultPort: c.defaultPort,
    description: c.description,
    status: c.status === 'disabled' ? 0 : 1,
  })
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await editFormRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  try {
    if (isEdit.value) {
      await updateConnector(editForm)
      ElMessage.success(t('common.updateSuccess'))
    } else {
      await createConnector(editForm)
      ElMessage.success(t('common.createSuccess'))
    }
    dialogVisible.value = false
    loadData()
  } finally {
    submitLoading.value = false
  }
}

async function handleDelete(c: ConnectorDisplay) {
  await ElMessageBox.confirm(t('odinConnector.confirmDelete'), t('common.tip'), { type: 'warning' })
  await deleteConnectors([c.raw?.id ?? c.id])
  ElMessage.success(t('common.deleteSuccess'))
  loadData()
}

// ===== 测试连接（保留，可由详情/后续入口调用） =====
const testingId = ref('')
async function handleTest(c: ConnectorDisplay) {
  if (testingId.value) return
  testingId.value = c.id
  try {
    await testConnector(c.raw?.id ?? c.id)
    ElMessage.success(t('odinConnector.testSuccess'))
  } catch (e) {
    ElMessage.error(t('odinConnector.testFail', { msg: (e as Error).message || '' }))
  } finally {
    testingId.value = ''
  }
}
// 显式引用以避免未使用告警（测试入口后续可挂到详情按钮）
void handleTest

// ===== 上传驱动 =====
const uploadDialogVisible = ref(false)
const uploadTarget = ref<ConnectorDisplay | null>(null)
const uploading = ref(false)
const uploadRawFile = ref<File | null>(null)
const uploadRef = ref<UploadInstance>()

function openUploadDialog(c: ConnectorDisplay) {
  uploadTarget.value = c
  uploadRawFile.value = null
  uploadDialogVisible.value = true
  uploadRef.value?.clearFiles()
}

function handleFileChange(file: UploadFile) {
  uploadRawFile.value = file.raw ?? null
}

function handleFileRemove() {
  uploadRawFile.value = null
}

async function submitUpload() {
  if (!uploadTarget.value) return
  if (!uploadRawFile.value) {
    ElMessage.warning(t('odinConnector.upload.selectFirst'))
    return
  }
  const isJar = uploadRawFile.value.name.toLowerCase().endsWith('.jar')
  if (!isJar) {
    ElMessage.error(t('odinConnector.upload.onlyJar'))
    return
  }
  uploading.value = true
  try {
    await uploadDriver(uploadRawFile.value)
    ElMessage.success(t('odinConnector.uploadSuccess'))
    uploadDialogVisible.value = false
    uploadRef.value?.clearFiles()
    uploadRawFile.value = null
    await loadData()
  } catch {
    ElMessage.error(t('odinConnector.uploadFail'))
  } finally {
    uploading.value = false
  }
}

onMounted(loadData)
</script>

<style scoped>
.connector-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--sm-bg-page);
  overflow-y: auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--sm-spacing-xxl);
  padding: var(--sm-spacing-xxl) 32px;
  background: linear-gradient(
    135deg,
    var(--sm-color-primary) 0%,
    color-mix(in srgb, var(--sm-color-primary) 78%, white) 100%
  );
  color: #fff;
}

.header-left h1 {
  margin: 0 0 4px;
  font-size: 24px;
  font-weight: 600;
}

.subtitle {
  margin: 0;
  font-size: var(--sm-font-size-base);
  opacity: 0.88;
  max-width: 640px;
  line-height: 1.5;
}

.header-stats {
  display: flex;
  gap: var(--sm-spacing-xxl);
  flex-shrink: 0;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  line-height: 1;
}

.stat-value.ready {
  color: color-mix(in srgb, var(--sm-color-success) 70%, white);
}

.stat-value.need-driver {
  color: color-mix(in srgb, var(--sm-color-warning) 70%, white);
}

.stat-value.disabled {
  color: color-mix(in srgb, var(--sm-color-danger) 70%, white);
}

.stat-label {
  font-size: var(--sm-font-size-xs);
  opacity: 0.85;
  margin-top: 4px;
}

.filter-section {
  padding: var(--sm-spacing-xl) 32px;
  background: var(--sm-bg-card);
  border-bottom: 1px solid var(--sm-border-light);
}

.filter-row {
  display: flex;
  align-items: center;
  gap: var(--sm-spacing-md);
  margin-bottom: var(--sm-spacing-lg);
}

.search-input {
  flex: 1;
  max-width: 420px;
}

.filter-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--sm-spacing-sm);
  margin-bottom: var(--sm-spacing-md);
}

.filter-tags:last-child {
  margin-bottom: 0;
}

.filter-tag {
  cursor: pointer;
  transition: transform 0.2s;
}

.filter-tag:hover {
  transform: scale(1.05);
}

.connectors-grid {
  flex: 1;
  padding: var(--sm-spacing-xxl) 32px;
}

.connector-card {
  margin-bottom: var(--sm-spacing-xl);
  border-radius: 12px;
  transition: transform 0.3s, box-shadow 0.3s;
  border: 1px solid var(--sm-border-lighter);
}

.connector-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 24px color-mix(in srgb, var(--sm-text-primary) 10%, transparent);
}

.card-header {
  display: flex;
  align-items: center;
  gap: var(--sm-spacing-md);
  margin-bottom: var(--sm-spacing-md);
}

.connector-icon {
  width: 48px;
  height: 48px;
  background: linear-gradient(
    135deg,
    color-mix(in srgb, var(--sm-color-primary) 12%, transparent) 0%,
    color-mix(in srgb, var(--sm-color-primary) 26%, transparent) 100%
  );
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  flex-shrink: 0;
}

.connector-info {
  min-width: 0;
  flex: 1;
}

.connector-info h3 {
  margin: 0 0 6px;
  font-size: var(--sm-font-size-lg);
  font-weight: 600;
  color: var(--sm-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.connector-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.connector-desc {
  font-size: var(--sm-font-size-sm);
  color: var(--sm-text-secondary);
  line-height: 1.5;
  margin: 0 0 var(--sm-spacing-md);
  min-height: 40px;
}

.connector-details {
  background: var(--sm-bg-header);
  border-radius: var(--sm-radius-lg);
  padding: 10px;
  margin-bottom: var(--sm-spacing-md);
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: var(--sm-font-size-xs);
  padding: 4px 0;
  gap: var(--sm-spacing-sm);
}

.detail-row .label {
  color: var(--sm-text-secondary);
  flex-shrink: 0;
}

.detail-row .value {
  color: var(--sm-text-primary);
  font-weight: 500;
  text-align: right;
  word-break: break-all;
}

.detail-row code {
  background: color-mix(in srgb, var(--sm-text-secondary) 14%, transparent);
  padding: 2px 6px;
  border-radius: var(--sm-radius-md);
  font-size: 11px;
}

.card-actions {
  display: flex;
  align-items: center;
  gap: var(--sm-spacing-sm);
  padding-top: var(--sm-spacing-md);
  border-top: 1px solid var(--sm-border-lighter);
  flex-wrap: wrap;
}

.card-actions .el-button {
  flex: 0 0 auto;
}

.card-actions .el-button:not(.is-link) {
  flex: 1;
}

.upload-dialog-body {
  padding: 4px 0;
}

.upload-hint {
  margin: 0 0 var(--sm-spacing-sm);
  font-size: var(--sm-font-size-sm);
  color: var(--sm-text-regular);
  line-height: 1.5;
}

.upload-lib {
  margin: 0 0 var(--sm-spacing-lg);
  font-size: var(--sm-font-size-sm);
  color: var(--sm-text-regular);
}

.upload-lib code {
  background: color-mix(in srgb, var(--sm-color-primary) 12%, transparent);
  color: var(--sm-color-primary);
  padding: 2px 6px;
  border-radius: var(--sm-radius-md);
  font-size: var(--sm-font-size-xs);
}

.driver-upload {
  width: 100%;
}
</style>
