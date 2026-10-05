<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  ElCard,
  ElButton,
  ElTag,
  ElRow,
  ElCol,
  ElInput,
  ElEmpty,
  ElMessage,
  ElDialog,
  ElUpload,
  ElIcon,
} from 'element-plus'
import type { UploadFile, UploadInstance } from 'element-plus'
import { Search, UploadFilled } from '@element-plus/icons-vue'
import { getDrivers, uploadDriver, type DriverInfo } from '@/api/datasource'

// 分类元数据
const categoryMeta: Record<string, { label: string; icon: string }> = {
  rdbms: { label: '关系型数据库', icon: '💾' },
  nosql: { label: '非关系型数据库', icon: '📦' },
  file: { label: '文件', icon: '📁' },
}

const drivers = ref<DriverInfo[]>([])
const loading = ref(false)

// 搜索和过滤
const searchQuery = ref('')
const categoryFilter = ref<string | null>(null)
const typeFilter = ref<'native' | 'jdbc' | null>(null)

const categories = computed(() => {
  const seen = new Set<string>()
  for (const d of drivers.value) {
    if (d.category) seen.add(d.category)
  }
  return [...seen].map((id) => ({
    id,
    label: categoryMeta[id]?.label ?? id,
    icon: categoryMeta[id]?.icon ?? '🔌',
  }))
})

const filteredDrivers = computed(() => {
  let result = [...drivers.value]

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(
      (d) =>
        d.display_name.toLowerCase().includes(query) ||
        d.name.toLowerCase().includes(query) ||
        d.description?.toLowerCase().includes(query) ||
        d.driver_lib?.toLowerCase().includes(query),
    )
  }

  if (categoryFilter.value) {
    result = result.filter((d) => d.category === categoryFilter.value)
  }

  if (typeFilter.value) {
    result = result.filter((d) => d.driver_type === typeFilter.value)
  }

  return result
})

const stats = computed(() => {
  const total = drivers.value.length
  const ready = drivers.value.filter((d) => d.status === 'ready').length
  const needDriver = drivers.value.filter((d) => d.status === 'need_driver').length
  return { total, ready, needDriver }
})

const driverTypeTag = (type: string) =>
  type === 'native'
    ? { label: 'Go 原生驱动', color: '#67c23a' }
    : { label: 'JDBC Agent', color: '#e6a23c' }

const statusTag = (status: string) =>
  status === 'ready'
    ? { label: '已就绪', type: 'success' as const }
    : status === 'need_driver'
      ? { label: '需上传驱动', type: 'warning' as const }
      : { label: status, type: 'info' as const }

// --- 上传驱动对话框 ---
const uploadDialogVisible = ref(false)
const uploadTarget = ref<DriverInfo | null>(null)
const uploading = ref(false)
const uploadRawFile = ref<File | null>(null)
const uploadRef = ref<UploadInstance>()

function openUploadDialog(driver: DriverInfo) {
  uploadTarget.value = driver
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

function beforeUploadCheck(file: File) {
  const isJar = file.name.toLowerCase().endsWith('.jar')
  if (!isJar) {
    ElMessage.error('仅支持上传 .jar 格式的驱动文件')
    return false
  }
  return true
}

async function submitUpload() {
  if (!uploadTarget.value) return
  if (!uploadRawFile.value) {
    ElMessage.warning('请先选择要上传的 JAR 驱动文件')
    return
  }
  if (!beforeUploadCheck(uploadRawFile.value)) return

  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', uploadRawFile.value)
    await uploadDriver(uploadTarget.value.name, formData)
    ElMessage.success('驱动上传成功')
    uploadDialogVisible.value = false
    uploadRef.value?.clearFiles()
    uploadRawFile.value = null
    await loadDrivers()
  } catch {
    // 错误提示已由请求拦截器统一处理
  } finally {
    uploading.value = false
  }
}

async function loadDrivers() {
  loading.value = true
  try {
    drivers.value = await getDrivers()
  } catch (e) {
    console.error('[ConnectorManager] 加载连接器列表失败', e)
    drivers.value = []
  } finally {
    loading.value = false
  }
}

onMounted(loadDrivers)
</script>

<template>
  <div class="connector-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1>连接器管理</h1>
        <p class="subtitle">
          查看系统支持的数据库连接类型与驱动方式，Go 原生驱动开箱即用，JDBC 驱动可上传代理驱动文件
        </p>
      </div>
      <div class="header-stats">
        <div class="stat-item">
          <span class="stat-value">{{ stats.total }}</span>
          <span class="stat-label">连接器总数</span>
        </div>
        <div class="stat-item">
          <span class="stat-value ready">{{ stats.ready }}</span>
          <span class="stat-label">已就绪</span>
        </div>
        <div class="stat-item">
          <span class="stat-value need-driver">{{ stats.needDriver }}</span>
          <span class="stat-label">需上传驱动</span>
        </div>
      </div>
    </div>

    <!-- 过滤器 -->
    <div class="filter-section">
      <ElInput
        v-model="searchQuery"
        placeholder="搜索连接器名称、驱动库..."
        :prefix-icon="Search"
        clearable
        class="search-input"
      />

      <div class="filter-tags">
        <ElTag
          :type="categoryFilter === null ? 'primary' : 'info'"
          effect="plain"
          class="filter-tag"
          @click="categoryFilter = null"
        >
          全部分类
        </ElTag>
        <ElTag
          v-for="cat in categories"
          :key="cat.id"
          :type="categoryFilter === cat.id ? 'primary' : 'info'"
          effect="plain"
          class="filter-tag"
          @click="categoryFilter = categoryFilter === cat.id ? null : cat.id"
        >
          {{ cat.icon }} {{ cat.label }}
        </ElTag>
      </div>

      <div class="filter-tags">
        <ElTag
          :type="typeFilter === null ? 'primary' : 'info'"
          effect="plain"
          class="filter-tag"
          @click="typeFilter = null"
        >
          全部驱动方式
        </ElTag>
        <ElTag
          :type="typeFilter === 'native' ? 'primary' : 'info'"
          effect="plain"
          class="filter-tag"
          @click="typeFilter = typeFilter === 'native' ? null : 'native'"
        >
          Go 原生驱动
        </ElTag>
        <ElTag
          :type="typeFilter === 'jdbc' ? 'primary' : 'info'"
          effect="plain"
          class="filter-tag"
          @click="typeFilter = typeFilter === 'jdbc' ? null : 'jdbc'"
        >
          JDBC Agent
        </ElTag>
      </div>
    </div>

    <!-- 连接器列表 -->
    <div v-loading="loading" class="connectors-grid">
      <ElEmpty
        v-if="!loading && filteredDrivers.length === 0"
        description="没有找到匹配的连接器"
      />

      <ElRow :gutter="20">
        <ElCol v-for="driver in filteredDrivers" :key="driver.name" :span="8">
          <ElCard class="connector-card" shadow="hover">
            <div class="card-header">
              <div class="connector-icon">
                {{ categoryMeta[driver.category]?.icon ?? '🔌' }}
              </div>
              <div class="connector-info">
                <h3>{{ driver.display_name }}</h3>
                <div class="connector-meta">
                  <ElTag size="small" effect="plain" type="info">
                    {{ categoryMeta[driver.category]?.label ?? driver.category }}
                  </ElTag>
                  <ElTag
                    size="small"
                    :style="{
                      background: driverTypeTag(driver.driver_type).color + '1a',
                      color: driverTypeTag(driver.driver_type).color,
                      borderColor: driverTypeTag(driver.driver_type).color + '66',
                    }"
                  >
                    {{ driverTypeTag(driver.driver_type).label }}
                  </ElTag>
                  <ElTag size="small" effect="plain" :type="statusTag(driver.status).type">
                    {{ statusTag(driver.status).label }}
                  </ElTag>
                </div>
              </div>
            </div>

            <p class="connector-desc">{{ driver.description }}</p>

            <div class="connector-details">
              <div class="detail-row">
                <span class="label">驱动库</span>
                <code class="value">{{ driver.driver_lib || '—' }}</code>
              </div>
              <div class="detail-row">
                <span class="label">标识</span>
                <code class="value">{{ driver.name }}</code>
              </div>
            </div>

            <div class="card-actions">
              <span v-if="driver.driver_type === 'native'" class="no-config-hint">
                Go 原生驱动，无需配置
              </span>
              <template v-else>
                <ElButton
                  v-if="driver.status === 'need_driver'"
                  type="primary"
                  size="small"
                  :icon="UploadFilled"
                  @click="openUploadDialog(driver)"
                >
                  上传驱动
                </ElButton>
                <ElButton
                  v-else
                  size="small"
                  type="warning"
                  plain
                  :icon="UploadFilled"
                  @click="openUploadDialog(driver)"
                >
                  更新驱动
                </ElButton>
              </template>
            </div>
          </ElCard>
        </ElCol>
      </ElRow>
    </div>

    <!-- 上传驱动对话框 -->
    <ElDialog
      v-model="uploadDialogVisible"
      :title="`上传驱动 - ${uploadTarget?.display_name ?? ''}`"
      width="480px"
      destroy-on-close
    >
      <div class="upload-dialog-body">
        <p class="upload-hint">
          {{ uploadTarget?.description }}
        </p>
        <p class="upload-lib">
          所需驱动文件：<code>{{ uploadTarget?.driver_lib || 'JDBC 驱动 JAR' }}</code>
        </p>
        <ElUpload
          ref="uploadRef"
          drag
          accept=".jar"
          :limit="1"
          :auto-upload="false"
          :on-change="handleFileChange"
          :on-remove="handleFileRemove"
          :on-exceed="() => ElMessage.warning('只能上传一个文件，请先移除已选文件')"
          class="driver-upload"
        >
          <ElIcon class="el-icon--upload"><UploadFilled /></ElIcon>
          <div class="el-upload__text">
            将 .jar 驱动文件拖到此处，或<em>点击选择文件</em>
          </div>
        </ElUpload>
      </div>
      <template #footer>
        <ElButton @click="uploadDialogVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="uploading" @click="submitUpload">
          上传
        </ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<style scoped>
.connector-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: linear-gradient(180deg, #f0f7ff 0%, #e8f4fd 100%);
  overflow-y: auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 24px;
  padding: 24px 32px;
  background: linear-gradient(135deg, #2b6cb0 0%, #3182ce 100%);
  color: white;
}

.header-left h1 {
  margin: 0 0 4px;
  font-size: 24px;
  font-weight: 600;
}

.subtitle {
  margin: 0;
  font-size: 14px;
  opacity: 0.85;
  max-width: 640px;
  line-height: 1.5;
}

.header-stats {
  display: flex;
  gap: 24px;
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
  color: #86efac;
}

.stat-value.need-driver {
  color: #fde68a;
}

.stat-label {
  font-size: 12px;
  opacity: 0.8;
  margin-top: 4px;
}

.filter-section {
  padding: 20px 32px;
  background: white;
  border-bottom: 1px solid #e4e7ed;
}

.search-input {
  width: 100%;
  max-width: 400px;
  margin-bottom: 16px;
}

.filter-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}

.filter-tags:last-child {
  margin-bottom: 0;
}

.filter-tag {
  cursor: pointer;
  transition: all 0.2s;
}

.filter-tag:hover {
  transform: scale(1.05);
}

.connectors-grid {
  flex: 1;
  padding: 24px 32px;
}

.connector-card {
  margin-bottom: 20px;
  border-radius: 12px;
  transition: all 0.3s;
}

.connector-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.connector-icon {
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, #ebf8ff 0%, #bee3f8 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  flex-shrink: 0;
}

.connector-info h3 {
  margin: 0 0 6px;
  font-size: 16px;
  font-weight: 600;
  color: #1a365d;
}

.connector-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.connector-desc {
  font-size: 13px;
  color: #718096;
  line-height: 1.5;
  margin: 0 0 12px;
  min-height: 40px;
}

.connector-details {
  background: #f8fafc;
  border-radius: 8px;
  padding: 10px;
  margin-bottom: 12px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  padding: 4px 0;
  gap: 8px;
}

.detail-row .label {
  color: #718096;
  flex-shrink: 0;
}

.detail-row .value {
  color: #2d3748;
  font-weight: 500;
  text-align: right;
  word-break: break-all;
}

.detail-row code {
  background: #edf2f7;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 11px;
}

.card-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid #e4e7ed;
}

.card-actions .el-button {
  flex: 1;
}

.no-config-hint {
  flex: 1;
  text-align: center;
  font-size: 12px;
  color: #909399;
}

.upload-dialog-body {
  padding: 4px 0;
}

.upload-hint {
  margin: 0 0 8px;
  font-size: 13px;
  color: #606266;
  line-height: 1.5;
}

.upload-lib {
  margin: 0 0 16px;
  font-size: 13px;
  color: #606266;
}

.upload-lib code {
  background: #f0f7ff;
  color: #2b6cb0;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}

.driver-upload {
  width: 100%;
}
</style>
