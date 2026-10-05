<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  ElDialog,
  ElTable,
  ElTableColumn,
  ElTag,
  ElButton,
  ElPopconfirm,
  ElEmpty,
  ElIcon,
  ElMessage,
  ElDescriptions,
  ElDescriptionsItem,
} from 'element-plus'
import {
  Refresh,
  View,
  Download,
  Delete,
} from '@element-plus/icons-vue'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
}>()

// Mock 版本历史数据
const versions = ref([
  {
    id: 'v3',
    version: 3,
    createdAt: '2024-01-15 14:30:00',
    createdBy: 'admin',
    summary: '添加 CRM 达梦库客户映射',
    changes: {
      added: 2,
      modified: 1,
      deleted: 0,
    },
    status: 'current',
    classMaps: [
      { classId: 'Customer', sourceId: 'crm_dm', table: 'KHXX' },
    ],
  },
  {
    id: 'v2',
    version: 2,
    createdAt: '2024-01-14 10:15:00',
    createdBy: 'admin',
    summary: '完善 ERP 主库产品映射',
    changes: {
      added: 1,
      modified: 2,
      deleted: 0,
    },
    status: 'history',
    classMaps: [
      { classId: 'Product', sourceId: 'erp_pg', table: 'products' },
    ],
  },
  {
    id: 'v1',
    version: 1,
    createdAt: '2024-01-13 09:00:00',
    createdBy: 'admin',
    summary: '初始映射配置',
    changes: {
      added: 3,
      modified: 0,
      deleted: 0,
    },
    status: 'history',
    classMaps: [
      { classId: 'Customer', sourceId: 'erp_pg', table: 'customers' },
      { classId: 'Order', sourceId: 'erp_pg', table: 'orders' },
    ],
  },
])

// 选中的版本
const selectedVersion = ref<string | null>(null)

// 当前版本
const currentVersion = computed(() => versions.value.find((v) => v.status === 'current'))

// 选中的版本详情
const selectedVersionDetail = computed(() => {
  if (!selectedVersion.value) return null
  return versions.value.find((v) => v.id === selectedVersion.value)
})

// 查看版本详情
const viewVersion = (versionId: string) => {
  selectedVersion.value = versionId
}

// 回滚到指定版本
const rollbackToVersion = (versionId: string) => {
  const version = versions.value.find((v) => v.id === versionId)
  if (!version) return

  // 更新状态
  versions.value.forEach((v) => {
    v.status = v.id === versionId ? 'current' : 'history'
  })

  ElMessage.success(`已回滚到版本 ${version.version}`)
  selectedVersion.value = null
}

// 删除版本
const deleteVersion = (versionId: string) => {
  versions.value = versions.value.filter((v) => v.id !== versionId)
  if (selectedVersion.value === versionId) {
    selectedVersion.value = null
  }
  ElMessage.success('版本已删除')
}

// 导出版本
const exportVersion = (versionId: string) => {
  const version = versions.value.find((v) => v.id === versionId)
  if (!version) return

  const content = JSON.stringify(version, null, 2)
  const blob = new Blob([content], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `mapping-${versionId}.json`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)

  ElMessage.success('版本导出成功')
}

// 关闭对话框
const handleClose = () => {
  emit('update:open', false)
  selectedVersion.value = null
}
</script>

<template>
  <ElDialog
    :model-value="open"
    @update:model-value="handleClose"
    title="映射版本管理"
    width="900px"
    destroy-on-close
  >
    <div class="version-container">
      <!-- 版本列表 -->
      <div class="version-list">
        <div class="list-header">
          <h3>版本历史</h3>
          <ElTag type="success" effect="plain">
            当前版本: v{{ currentVersion?.version }}
          </ElTag>
        </div>

        <ElEmpty v-if="versions.length === 0" description="暂无版本历史" />

        <div v-else class="versions">
          <div
            v-for="version in versions"
            :key="version.id"
            class="version-item"
            :class="{
              active: selectedVersion === version.id,
              current: version.status === 'current',
            }"
            @click="viewVersion(version.id)"
          >
            <div class="version-header">
              <div class="version-info">
                <span class="version-number">v{{ version.version }}</span>
                <ElTag
                  :type="version.status === 'current' ? 'success' : 'info'"
                  size="small"
                  effect="plain"
                >
                  {{ version.status === 'current' ? '当前' : '历史' }}
                </ElTag>
              </div>
              <span class="version-time">{{ version.createdAt }}</span>
            </div>

            <p class="version-summary">{{ version.summary }}</p>

            <div class="version-changes">
              <span class="change added">+{{ version.changes.added }}</span>
              <span class="change modified">~{{ version.changes.modified }}</span>
              <span class="change deleted">-{{ version.changes.deleted }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 版本详情 -->
      <div class="version-detail">
        <template v-if="selectedVersionDetail">
          <div class="detail-header">
            <h3>版本详情 - v{{ selectedVersionDetail.version }}</h3>
            <div class="detail-actions">
              <ElButton
                v-if="selectedVersionDetail.status !== 'current'"
                type="primary"
                size="small"
                :icon="Refresh"
                @click="rollbackToVersion(selectedVersionDetail.id)"
              >
                回滚到此版本
              </ElButton>
              <ElButton
                size="small"
                :icon="Download"
                @click="exportVersion(selectedVersionDetail.id)"
              >
                导出
              </ElButton>
              <ElPopconfirm
                v-if="selectedVersionDetail.status !== 'current'"
                title="确定删除此版本？"
                @confirm="deleteVersion(selectedVersionDetail.id)"
              >
                <template #reference>
                  <ElButton type="danger" size="small" :icon="Delete" plain>
                    删除
                  </ElButton>
                </template>
              </ElPopconfirm>
            </div>
          </div>

          <ElDescriptions :column="2" border size="small">
            <ElDescriptionsItem label="版本号">
              v{{ selectedVersionDetail.version }}
            </ElDescriptionsItem>
            <ElDescriptionsItem label="状态">
              <ElTag
                :type="selectedVersionDetail.status === 'current' ? 'success' : 'info'"
                size="small"
              >
                {{ selectedVersionDetail.status === 'current' ? '当前版本' : '历史版本' }}
              </ElTag>
            </ElDescriptionsItem>
            <ElDescriptionsItem label="创建时间">
              {{ selectedVersionDetail.createdAt }}
            </ElDescriptionsItem>
            <ElDescriptionsItem label="创建者">
              {{ selectedVersionDetail.createdBy }}
            </ElDescriptionsItem>
            <ElDescriptionsItem label="变更摘要" :span="2">
              {{ selectedVersionDetail.summary }}
            </ElDescriptionsItem>
          </ElDescriptions>

          <div class="changes-section">
            <h4>变更统计</h4>
            <div class="changes-stats">
              <div class="stat-item">
                <span class="stat-value added">+{{ selectedVersionDetail.changes.added }}</span>
                <span class="stat-label">新增</span>
              </div>
              <div class="stat-item">
                <span class="stat-value modified">~{{ selectedVersionDetail.changes.modified }}</span>
                <span class="stat-label">修改</span>
              </div>
              <div class="stat-item">
                <span class="stat-value deleted">-{{ selectedVersionDetail.changes.deleted }}</span>
                <span class="stat-label">删除</span>
              </div>
            </div>
          </div>

          <div class="class-maps-section">
            <h4>涉及的映射</h4>
            <ElTable :data="selectedVersionDetail.classMaps" border size="small">
              <ElTableColumn prop="classId" label="本体类" width="120">
                <template #default="{ row }">
                  <ElTag type="primary" effect="plain">{{ row.classId }}</ElTag>
                </template>
              </ElTableColumn>
              <ElTableColumn prop="sourceId" label="数据源" width="120" />
              <ElTableColumn prop="table" label="物理表" />
            </ElTable>
          </div>
        </template>

        <div v-else class="no-selection">
          <ElIcon :size="64"><View /></ElIcon>
          <p>选择一个版本查看详情</p>
        </div>
      </div>
    </div>
  </ElDialog>
</template>

<style scoped>
.version-container {
  display: flex;
  gap: 24px;
  min-height: 500px;
}

.version-list {
  width: 320px;
  display: flex;
  flex-direction: column;
}

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.list-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #1a365d;
}

.versions {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.version-item {
  padding: 16px;
  background: white;
  border: 2px solid #e4e7ed;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.version-item:hover {
  border-color: #3182ce;
}

.version-item.active {
  border-color: #3182ce;
  background: #f0f7ff;
}

.version-item.current {
  border-color: #67c23a;
  background: #f0fff4;
}

.version-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.version-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.version-number {
  font-size: 16px;
  font-weight: 700;
  color: #1a365d;
}

.version-time {
  font-size: 12px;
  color: #909399;
}

.version-summary {
  font-size: 13px;
  color: #606266;
  margin: 0 0 10px;
}

.version-changes {
  display: flex;
  gap: 12px;
}

.change {
  font-size: 13px;
  font-weight: 600;
  font-family: 'Monaco', 'Menlo', monospace;
}

.change.added {
  color: #67c23a;
}

.change.modified {
  color: #e6a23c;
}

.change.deleted {
  color: #f56c6c;
}

.version-detail {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.detail-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1a365d;
}

.detail-actions {
  display: flex;
  gap: 8px;
}

.changes-section {
  margin-top: 20px;
}

.changes-section h4 {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 600;
  color: #2d3748;
}

.changes-stats {
  display: flex;
  gap: 24px;
  padding: 16px;
  background: #f8fafc;
  border-radius: 8px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  font-family: 'Monaco', 'Menlo', monospace;
}

.stat-value.added {
  color: #67c23a;
}

.stat-value.modified {
  color: #e6a23c;
}

.stat-value.deleted {
  color: #f56c6c;
}

.stat-label {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.class-maps-section {
  margin-top: 20px;
}

.class-maps-section h4 {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 600;
  color: #2d3748;
}

.no-selection {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #909399;
}

.no-selection p {
  margin: 12px 0 0;
  font-size: 14px;
}
</style>
