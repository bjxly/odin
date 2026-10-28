<template>
  <div class="schema-browser">
    <div class="schema-browser__toolbar">
      <el-input
        v-model="filterText"
        :placeholder="t('odinDatasource.tables')"
        clearable
        size="small"
        :prefix-icon="Search"
        style="width: 200px"
      />
      <el-button size="small" :icon="Refresh" :loading="loading" @click="refresh">
        {{ t('common.refresh') }}
      </el-button>
    </div>

    <el-tree
      ref="treeRef"
      v-loading="loading"
      :data="treeData"
      :props="treeProps"
      :filter-node-method="filterNode"
      :load="loadNode"
      lazy
      node-key="id"
      highlight-current
      :expand-on-click-node="false"
      @node-click="handleNodeClick"
    >
      <template #default="{ node, data }">
        <span class="tree-node" :class="getNodeDiffClass(data)">
          <el-icon v-if="data.nodeType === 'table'" class="node-icon"><Grid /></el-icon>
          <el-icon v-else class="node-icon"><Document /></el-icon>
          <span class="node-label">{{ node.label }}</span>
          <el-tag v-if="data.nodeType === 'table'" size="small" type="info" class="node-badge">
            {{ data.columnCount ?? '?' }}
          </el-tag>
          <el-tag v-else-if="data.dataType" size="small" effect="plain" class="node-type">
            {{ data.dataType }}
          </el-tag>
          <el-tag v-if="data.isPrimaryKey" size="small" type="warning" effect="plain" class="node-pk">
            PK
          </el-tag>
        </span>
      </template>
    </el-tree>

    <el-empty v-if="!loading && treeData.length === 0" :description="t('common.noData')" :image-size="60" />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Search, Refresh, Grid, Document } from '@element-plus/icons-vue'
import { useOdinDatasourceStore } from '@/stores/odin/datasource'
import type { SchemaTableVO, SchemaColumnVO } from '@/types/odin'

const props = defineProps<{
  datasourceId: string
}>()

const emit = defineEmits<{
  (e: 'select-table', table: SchemaTableVO): void
  (e: 'select-column', column: SchemaColumnVO, table: SchemaTableVO): void
}>()

const { t } = useI18n()
const dsStore = useOdinDatasourceStore()

const treeRef = ref<InstanceType<typeof import('element-plus')['ElTree']>>()
const filterText = ref('')
const loading = ref(false)

interface TreeNodeData {
  id: string
  label: string
  nodeType: 'table' | 'column'
  columnCount?: number
  dataType?: string
  isPrimaryKey?: boolean
  isLeaf: boolean
  raw?: SchemaTableVO | SchemaColumnVO
  parentTable?: SchemaTableVO
  diffStatus?: 'added' | 'removed' | 'modified'
}

const treeProps = {
  label: 'label',
  children: 'children',
  isLeaf: 'isLeaf',
}

const treeData = ref<TreeNodeData[]>([])

/** 根据 store 中缓存的 diff 信息标注节点 */
function getNodeDiffClass(data: TreeNodeData) {
  if (data.diffStatus === 'added') return 'diff-added'
  if (data.diffStatus === 'removed') return 'diff-removed'
  if (data.diffStatus === 'modified') return 'diff-modified'
  return ''
}

async function loadTables() {
  if (!props.datasourceId) return
  loading.value = true
  try {
    const tables = await dsStore.loadSchema(props.datasourceId)
    treeData.value = tables.map((t) => ({
      id: t.id,
      label: t.tableName,
      nodeType: 'table' as const,
      columnCount: undefined,
      isLeaf: false,
      raw: t,
    }))
  } finally {
    loading.value = false
  }
}

async function loadNode(node: any, resolve: (data: TreeNodeData[]) => void) {
  if (node.level === 0) {
    resolve([])
    return
  }
  const data = node.data as TreeNodeData
  if (data.nodeType === 'table' && data.raw) {
    const tableId = (data.raw as SchemaTableVO).id
    const columns = await dsStore.loadColumns(tableId)
    data.columnCount = columns.length
    const children: TreeNodeData[] = columns.map((col) => ({
      id: col.id,
      label: col.columnName,
      nodeType: 'column' as const,
      dataType: col.dataType,
      isPrimaryKey: col.isPrimaryKey === 1,
      isLeaf: true,
      raw: col,
      parentTable: data.raw as SchemaTableVO,
    }))
    resolve(children)
  } else {
    resolve([])
  }
}

function handleNodeClick(data: TreeNodeData) {
  if (data.nodeType === 'table' && data.raw) {
    emit('select-table', data.raw as SchemaTableVO)
  } else if (data.nodeType === 'column' && data.raw) {
    emit('select-column', data.raw as SchemaColumnVO, data.parentTable!)
  }
}

function filterNode(value: string, data: TreeNodeData) {
  if (!value) return true
  return data.label.toLowerCase().includes(value.toLowerCase())
}

function refresh() {
  dsStore.schemas.delete(props.datasourceId)
  loadTables()
}

watch(filterText, (val) => {
  treeRef.value?.filter(val)
})

watch(
  () => props.datasourceId,
  () => {
    treeData.value = []
    loadTables()
  },
)

onMounted(loadTables)
</script>

<style scoped>
.schema-browser {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.schema-browser__toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.tree-node {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: var(--sm-font-size-sm, 13px);
  padding: 2px 0;
}

.node-icon {
  font-size: 14px;
  color: var(--sm-text-secondary);
}

.node-label {
  color: var(--sm-text-primary);
}

.node-badge {
  margin-left: 4px;
}

.node-type {
  margin-left: 4px;
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: 11px;
}

.node-pk {
  margin-left: 4px;
}

/* Diff highlighting */
.diff-added {
  background: color-mix(in srgb, var(--el-color-success) 12%, transparent);
  border-radius: 4px;
  padding: 2px 6px;
}

.diff-removed {
  background: color-mix(in srgb, var(--el-color-danger) 10%, transparent);
  text-decoration: line-through;
  border-radius: 4px;
  padding: 2px 6px;
}

.diff-modified {
  background: color-mix(in srgb, var(--el-color-warning) 12%, transparent);
  border-radius: 4px;
  padding: 2px 6px;
}
</style>
