<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  ElTree,
  ElInput,
  ElTable,
  ElTableColumn,
  ElTag,
  ElIcon,
  ElEmpty,
  ElButton,
  ElTooltip,
  ElBadge,
} from 'element-plus'
import {
  Grid,
  Key,
  Search,
  Refresh,
  Folder,
} from '@element-plus/icons-vue'

interface Column {
  name: string
  type: string
  nullable?: boolean
  comment?: string
  isPk?: boolean
  isFk?: boolean
}

interface Table {
  name: string
  schema?: string
  comment?: string
  columns: Column[]
}

const props = defineProps<{
  tables: Table[]
  sourceName: string
}>()

const emit = defineEmits<{
  refresh: []
}>()

const searchQuery = ref('')
const selectedTable = ref<Table | null>(null)
const expandedKeys = ref<string[]>([])

// 过滤表
const filteredTables = computed(() => {
  if (!searchQuery.value) return props.tables
  const query = searchQuery.value.toLowerCase()
  return props.tables.filter(
    (t) =>
      t.name.toLowerCase().includes(query) ||
      t.comment?.toLowerCase().includes(query) ||
      t.columns.some(
        (c) =>
          c.name.toLowerCase().includes(query) ||
          c.comment?.toLowerCase().includes(query),
      ),
  )
})

// 树形数据
const treeData = computed(() => {
  const schemaMap = new Map<string, Table[]>()

  for (const table of filteredTables.value) {
    const schema = table.schema || 'default'
    if (!schemaMap.has(schema)) {
      schemaMap.set(schema, [])
    }
    schemaMap.get(schema)!.push(table)
  }

  return Array.from(schemaMap.entries()).map(([schema, tables]) => ({
    id: `schema:${schema}`,
    label: schema,
    isSchema: true,
    children: tables.map((t) => ({
      id: `table:${t.name}`,
      label: t.name,
      comment: t.comment,
      columnCount: t.columns.length,
      pkCount: t.columns.filter((c) => c.isPk).length,
      isTable: true,
      table: t,
    })),
  }))
})

// 选中的列
const selectedColumns = computed(() => selectedTable.value?.columns ?? [])

// 统计信息
const stats = computed(() => {
  const totalTables = props.tables.length
  const totalColumns = props.tables.reduce((sum, t) => sum + t.columns.length, 0)
  const pkColumns = props.tables.reduce(
    (sum, t) => sum + t.columns.filter((c) => c.isPk).length,
    0,
  )
  return { totalTables, totalColumns, pkColumns }
})

// 处理节点点击
const handleNodeClick = (data: any) => {
  if (data.isTable && data.table) {
    selectedTable.value = data.table
  }
}

// 过滤节点方法
const filterNode = (value: string, data: any) => {
  if (!value) return true
  const query = value.toLowerCase()
  return (
    data.label?.toLowerCase().includes(query) ||
    data.comment?.toLowerCase().includes(query)
  )
}

// 类型颜色
const getTypeColor = (type: string) => {
  const t = type.toLowerCase()
  if (t.includes('int') || t.includes('serial')) return '#409eff'
  if (t.includes('varchar') || t.includes('text') || t.includes('char')) return '#67c23a'
  if (t.includes('timestamp') || t.includes('date') || t.includes('time')) return '#e6a23c'
  if (t.includes('decimal') || t.includes('numeric') || t.includes('float')) return '#f56c6c'
  if (t.includes('bool')) return '#909399'
  if (t.includes('json') || t.includes('jsonb')) return '#b37feb'
  return '#909399'
}

// 展开第一个 schema
watch(
  () => props.tables,
  () => {
    if (treeData.value.length > 0) {
      expandedKeys.value = [treeData.value[0].id]
    }
    selectedTable.value = null
  },
  { immediate: true },
)
</script>

<template>
  <div class="schema-browser">
    <!-- 头部统计 -->
    <div class="browser-header">
      <div class="stats">
        <div class="stat-item">
          <span class="stat-value">{{ stats.totalTables }}</span>
          <span class="stat-label">张表</span>
        </div>
        <div class="stat-item">
          <span class="stat-value">{{ stats.totalColumns }}</span>
          <span class="stat-label">个列</span>
        </div>
        <div class="stat-item">
          <span class="stat-value">{{ stats.pkColumns }}</span>
          <span class="stat-label">主键</span>
        </div>
      </div>
      <ElButton :icon="Refresh" size="small" @click="emit('refresh')">刷新</ElButton>
    </div>

    <!-- 搜索框 -->
    <div class="search-box">
      <ElInput
        v-model="searchQuery"
        placeholder="搜索表名、列名、注释..."
        clearable
        :prefix-icon="Search"
      />
    </div>

    <!-- 主内容区 -->
    <div class="browser-content">
      <!-- 左侧表树 -->
      <div class="table-tree">
        <ElTree
          :data="treeData"
          :props="{ children: 'children', label: 'label' }"
          node-key="id"
          :default-expanded-keys="expandedKeys"
          :filter-node-method="filterNode"
          highlight-current
          @node-click="handleNodeClick"
        >
          <template #default="{ data }">
            <div class="tree-node">
              <template v-if="data.isSchema">
                <ElIcon class="node-icon"><Folder /></ElIcon>
                <span class="node-label">{{ data.label }}</span>
                <ElBadge :value="data.children?.length" :max="99" class="node-badge" />
              </template>
              <template v-else>
                <ElIcon class="node-icon"><Grid /></ElIcon>
                <span class="node-label">{{ data.label }}</span>
                <span class="node-meta" v-if="data.comment">{{ data.comment }}</span>
              </template>
            </div>
          </template>
        </ElTree>
      </div>

      <!-- 右侧列详情 -->
      <div class="column-detail">
        <template v-if="selectedTable">
          <div class="table-header">
            <h3>
              <ElIcon><Grid /></ElIcon>
              {{ selectedTable.name }}
            </h3>
            <span class="table-comment" v-if="selectedTable.comment">
              {{ selectedTable.comment }}
            </span>
          </div>

          <ElTable :data="selectedColumns" border stripe size="small" max-height="400">
            <ElTableColumn type="index" width="50" label="#" align="center" />
            <ElTableColumn prop="name" label="列名" min-width="120">
              <template #default="{ row }">
                <div class="column-name">
                  <ElIcon v-if="row.isPk" class="pk-icon"><Key /></ElIcon>
                  <code>{{ row.name }}</code>
                </div>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="type" label="类型" width="120">
              <template #default="{ row }">
                <ElTag :color="getTypeColor(row.type)" effect="dark" size="small">
                  {{ row.type }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="nullable" label="可空" width="70" align="center">
              <template #default="{ row }">
                <ElTag :type="row.nullable ? 'success' : 'danger'" size="small">
                  {{ row.nullable ? '是' : '否' }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="comment" label="注释" min-width="150">
              <template #default="{ row }">
                <span class="column-comment">{{ row.comment || '—' }}</span>
              </template>
            </ElTableColumn>
            <ElTableColumn label="操作" width="100" align="center">
              <template #default>
                <ElTooltip content="映射到本体属性" placement="top">
                  <ElButton size="small" type="primary" link>映射</ElButton>
                </ElTooltip>
              </template>
            </ElTableColumn>
          </ElTable>
        </template>

        <ElEmpty v-else description="请从左侧选择一张表" :image-size="120">
          <template #image>
            <div class="empty-icon">📊</div>
          </template>
        </ElEmpty>
      </div>
    </div>
  </div>
</template>

<style scoped>
.schema-browser {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
}

.browser-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: linear-gradient(135deg, #2b6cb0 0%, #3182ce 100%);
  color: white;
}

.stats {
  display: flex;
  gap: 24px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  line-height: 1;
}

.stat-label {
  font-size: 12px;
  opacity: 0.9;
  margin-top: 4px;
}

.search-box {
  padding: 12px 16px;
  border-bottom: 1px solid #ebeef5;
}

.browser-content {
  flex: 1;
  display: flex;
  min-height: 0;
  overflow: hidden;
}

.table-tree {
  width: 280px;
  border-right: 1px solid #ebeef5;
  overflow-y: auto;
  padding: 8px;
}

.tree-node {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
  flex: 1;
  min-width: 0;
}

.node-icon {
  font-size: 16px;
  color: #909399;
  flex-shrink: 0;
}

.node-label {
  font-size: 14px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-meta {
  font-size: 12px;
  color: #909399;
  margin-left: auto;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100px;
}

.node-badge {
  margin-left: auto;
}

.column-detail {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
}

.table-header {
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 2px solid #409eff;
}

.table-header h3 {
  margin: 0 0 8px;
  font-size: 18px;
  color: #303133;
  display: flex;
  align-items: center;
  gap: 8px;
}

.table-comment {
  font-size: 14px;
  color: #909399;
  font-style: italic;
}

.column-name {
  display: flex;
  align-items: center;
  gap: 6px;
}

.pk-icon {
  color: #e6a23c;
  font-size: 14px;
}

.column-comment {
  color: #606266;
  font-size: 13px;
}

.empty-icon {
  font-size: 64px;
  opacity: 0.5;
}

/* 自定义树样式 */
:deep(.el-tree-node__content) {
  height: 36px;
}

:deep(.el-tree-node.is-current > .el-tree-node__content) {
  background-color: #ecf5ff;
  border-radius: 4px;
}
</style>
