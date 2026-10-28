<template>
  <div class="example-page">
    <h2>SmCrudTable 示例</h2>
    <p class="example-desc">展示增强后的 CRUD 表格组件：列显隐选择、拖拽排序、固定表头、远程排序 v-model、持久化配置。</p>

    <SmCrudTable
      ref="crudTableRef"
      v-model:page="page"
      v-model:page-size="pageSize"
      v-model:sort-prop="sortProp"
      v-model:sort-order="sortOrder"
      :data="pagedData"
      :columns="columns"
      :total="mockData.length"
      :max-height="400"
      :show-selection="true"
      :show-index="true"
      :show-column-settings="true"
      column-preference-key="example-task-table"
      @selection-change="onSelectionChange"
      @page-change="onPageChange"
      @sort-change="onSortChange"
      @row-click="onRowClick"
    >
      <template #title>任务管理</template>
      <template #header-actions>
        <el-button type="primary">
          <el-icon><Plus /></el-icon>
          新增
        </el-button>
        <el-button
          type="danger"
          :disabled="selectedRows.length === 0"
        >
          批量删除 ({{ selectedRows.length }})
        </el-button>
        <el-button @click="resetColumnPrefs">清除列配置缓存</el-button>
      </template>

      <template #status="{ row }">
        <SmDictTag
          dict-type="task_status"
          :value="row.status"
        />
      </template>

      <template #type="{ row }">
        <SmDictTag
          dict-type="task_type"
          :value="row.type"
        />
      </template>

      <template #operate="{ row }">
        <el-button
          link
          type="primary"
          size="small"
          @click="handleEdit(row)"
        >
          <el-icon><Edit /></el-icon>
          编辑
        </el-button>
        <el-button
          link
          type="danger"
          size="small"
          @click="handleDelete(row)"
        >
          <el-icon><Delete /></el-icon>
          删除
        </el-button>
      </template>
    </SmCrudTable>

    <!-- 排序状态展示 -->
    <div class="example-state">
      <p>
        <strong>当前排序：</strong>
        {{ sortProp ? `${sortProp} ${sortOrder}` : '无' }}
      </p>
      <p>
        <strong>选中行数：</strong>
        {{ selectedRows.length }}
      </p>
      <p><el-text type="info">提示：点击右上角「列设置」按钮可调整列的显隐和顺序，配置自动保存至 localStorage</el-text></p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Plus, Edit, Delete } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import SmCrudTable from '@/components/SmCrudTable/index.vue'
import SmDictTag from '@/components/SmDictTag/index.vue'
import type { TableColumn } from '@/components/SmCrudTable/types'

const page = ref(1)
const pageSize = ref(10)
const selectedRows = ref<any[]>([])
const sortProp = ref('')
const sortOrder = ref<'ascending' | 'descending' | null>(null)
const crudTableRef = ref()

const columns: TableColumn[] = [
  { prop: 'id', label: 'ID', width: 80 },
  { prop: 'name', label: '任务名称', minWidth: 160 },
  { prop: 'status', label: '状态', width: 120, slot: 'status' },
  { prop: 'type', label: '类型', width: 120, slot: 'type' },
  { prop: 'assignee', label: '负责人', width: 100, visible: false },
  { prop: 'createTime', label: '创建时间', width: 180, sortable: 'custom' },
  { prop: 'updateTime', label: '更新时间', width: 180, sortable: 'custom' },
]

const mockData = Array.from({ length: 25 }, (_, i) => {
  const statuses = [0, 0, 1, 2, 2, 3, 4]
  const types = ['shell', 'http', 'python', 'sql']
  const assignees = ['张三', '李四', '王五', '赵六']
  const baseDay = 10 + Math.floor(i / 3)
  return {
    id: i + 1,
    name: ['数据备份任务', '健康检查', '日报生成', '日志清理', 'ETL 流程', '文件同步', '缓存预热', '数据库巡检', '安全扫描', '性能监控'][i % 10] || `任务 ${i + 1}`,
    status: statuses[i % statuses.length],
    type: types[i % types.length],
    assignee: assignees[i % assignees.length],
    createTime: `2025-01-${String(baseDay).padStart(2, '0')} ${String(8 + (i % 10)).padStart(2, '0')}:00:00`,
    updateTime: `2025-03-${String(baseDay + 1).padStart(2, '0')} ${String(10 + (i % 8)).padStart(2, '0')}:30:00`,
  }
})

const pagedData = computed(() => {
  let data = [...mockData]
  // 远程排序模拟
  if (sortProp.value && sortOrder.value) {
    data.sort((a: any, b: any) => {
      const va = a[sortProp.value] ?? ''
      const vb = b[sortProp.value] ?? ''
      const dir = sortOrder.value === 'ascending' ? 1 : -1
      return String(va).localeCompare(String(vb)) * dir
    })
  }
  const start = (page.value - 1) * pageSize.value
  return data.slice(start, start + pageSize.value)
})

function onSelectionChange(rows: any[]) {
  selectedRows.value = rows
}

function onPageChange(p: number) {
  page.value = p
}

function onSortChange(_sort: { prop: string; order: string }) {
  // 由 v-model 自动处理，这里做额外日志
}

function onRowClick(_row: any) {
  // 行点击事件示例
}

function handleEdit(row: any) {
  ElMessage.info(`编辑任务: ${row.name}`)
}

function handleDelete(row: any) {
  ElMessageBox.confirm(`确认删除任务「${row.name}」?`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })
    .then(() => {
      ElMessage.success(`已删除: ${row.name}`)
    })
    .catch(() => {})
}

function resetColumnPrefs() {
  localStorage.removeItem('sm-table-pref:example-task-table')
  ElMessage.success('列配置缓存已清除，刷新页面后将恢复默认')
}
</script>

<style scoped>
.example-page {
  padding: 20px;
}
.example-desc {
  color: var(--el-text-color-secondary);
  margin-bottom: 16px;
}
.example-state {
  margin-top: 16px;
  padding: 12px 16px;
  background: var(--sm-bg-secondary);
  border-radius: var(--sm-radius-md);
  font-size: var(--sm-font-size-base);
  color: var(--sm-text-secondary);
  line-height: 1.8;
}
</style>
