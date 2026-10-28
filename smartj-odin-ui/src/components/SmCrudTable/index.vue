<template>
  <div
    class="sm-crud-table"
    :style="tableWrapperStyle as any"
  >
    <!-- 工具栏 -->
    <div class="sm-crud-table__header">
      <div class="sm-crud-table__title">
        <slot name="title" />
      </div>
      <div class="sm-crud-table__header-actions">
        <el-button-group>
          <slot name="header-actions" />
          <!-- 列设置按钮 -->
          <ColumnSelector
            v-if="showColumnSettings && dataColumns.length > 0"
            :configs="columnConfigs"
            @toggle="colPref.toggleColumn"
            @update-order="colPref.updateOrder"
            @reset="colPref.resetToDefaults"
          />
        </el-button-group>
      </div>
    </div>

    <!-- 表格 -->
    <el-table
      ref="tableRef"
      v-loading="loading"
      :data="data"
      :border="border"
      :stripe="stripe"
      :height="height"
      :max-height="maxHeight"
      :row-key="rowKey"
      :default-sort="defaultSort"
      @selection-change="onSelectionChange"
      @sort-change="onSortChange"
      @row-click="(row: any, col: any, ev: any) => emit('row-click', row, col, ev)"
      @row-dblclick="(row: any, col: any, ev: any) => emit('row-dblclick', row, col, ev)"
    >
      <el-table-column
        v-if="showSelection"
        type="selection"
        width="50"
        align="center"
      />
      <el-table-column
        v-if="showIndex"
        type="index"
        label="#"
        width="60"
        align="center"
        fixed="left"
      />

      <template
        v-for="col in displayColumns"
        :key="col.prop || col.label"
      >
        <el-table-column
          :prop="col.prop"
          :label="col.label"
          :width="col.width"
          :min-width="col.minWidth"
          :sortable="col.sortable"
          :align="col.align || 'center'"
          :fixed="col.fixed"
          :class-name="col.className"
          :show-overflow-tooltip="col.showOverflowTooltip !== false"
        >
          <template
            v-if="col.slot"
            #default="scope"
          >
            <slot
              :name="col.slot"
              v-bind="scope"
            />
          </template>
          <template
            v-else-if="col.formatter"
            #default="scope"
          >
            <span>{{ col.formatter?.(scope.row, scope.column, scope.row[col.prop!], scope.$index) }}</span>
          </template>
        </el-table-column>
      </template>

      <el-table-column
        v-if="$slots.operate"
        :label="$t('common.operation')"
        :width="operateWidth"
        align="center"
        fixed="right"
      >
        <template #default="scope">
          <slot
            name="operate"
            v-bind="scope"
          />
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div
      v-if="showPagination"
      class="sm-crud-table__pagination"
    >
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="pageSizes"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        background
        @current-change="onPageChange"
        @size-change="onSizeChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import type { TableInstance } from 'element-plus'
import ColumnSelector from './ColumnSelector.vue'
import { useColumnPreference } from './useColumnPreference'
import type { TableColumn } from './types'

export type { TableColumn } from './types'

const props = withDefaults(
  defineProps<{
    data: any[]
    columns: TableColumn[]
    total?: number
    loading?: boolean
    border?: boolean
    stripe?: boolean
    rowKey?: string
    /** 固定高度（触发固定表头 + 内容滚动） */
    height?: number | string
    /** 最大高度 */
    maxHeight?: number | string
    /** 自适应高度（撑满父容器） */
    autoHeight?: boolean
    showSelection?: boolean
    showIndex?: boolean
    showPagination?: boolean
    showColumnSettings?: boolean
    /** 列配置持久化 key（不传则不持久化） */
    columnPreferenceKey?: string
    page?: number
    pageSize?: number
    pageSizes?: number[]
    operateWidth?: number | string
    /** 远程排序字段 */
    sortProp?: string
    /** 远程排序方向 */
    sortOrder?: 'ascending' | 'descending' | null
  }>(),
  {
    total: 0,
    loading: false,
    border: true,
    stripe: true,
    rowKey: 'id',
    autoHeight: false,
    showSelection: false,
    showIndex: false,
    showPagination: true,
    showColumnSettings: true,
    page: 1,
    pageSize: 10,
    pageSizes: () => [10, 20, 50, 100],
    operateWidth: 180,
    sortOrder: null,
  }
)

const emit = defineEmits<{
  (e: 'update:page', val: number): void
  (e: 'update:pageSize', val: number): void
  (e: 'update:sortProp', val: string): void
  (e: 'update:sortOrder', val: 'ascending' | 'descending' | null): void
  (e: 'page-change', page: number): void
  (e: 'size-change', size: number): void
  (e: 'selection-change', rows: any[]): void
  (e: 'sort-change', sort: { prop: string; order: string }): void
  (e: 'row-click', row: any, column: any, event: Event): void
  (e: 'row-dblclick', row: any, column: any, event: Event): void
}>()

const tableRef = ref<TableInstance>()
const currentPage = ref(props.page)
const pageSize = ref(props.pageSize)
// 过滤出数据列（排除无 prop 且无 slot 的纯展示列，但保留有 formatter 的列）
const dataColumns = computed(() => props.columns.filter((c) => c.prop || c.slot || c.formatter))

// 列偏好管理
const colPref = useColumnPreference(props.showColumnSettings ? props.columnPreferenceKey : undefined, () => dataColumns.value)

// 实际渲染的列（按用户偏好过滤和排序）
const displayColumns = computed(() => {
  if (!props.showColumnSettings) return dataColumns.value
  return colPref.visibleColumns.value
})

// 列配置（传给 ColumnSelector）
const columnConfigs = computed(() => colPref.columnConfigs.value)

// 远程排序默认值
const defaultSort = computed(() => {
  if (props.sortProp && props.sortOrder) {
    return { prop: props.sortProp, order: props.sortOrder }
  }
  return undefined
})

// 表格容器样式
const tableWrapperStyle = computed(() => {
  if (props.autoHeight) {
    return { display: 'flex', flexDirection: 'column', height: '100%', overflow: 'hidden' }
  }
  return {}
})

// ========== 分页 ==========
watch(
  () => props.page,
  (v) => {
    currentPage.value = v
  }
)
watch(
  () => props.pageSize,
  (v) => {
    pageSize.value = v
  }
)

function onPageChange(page: number) {
  emit('update:page', page)
  emit('page-change', page)
}

function onSizeChange(size: number) {
  emit('update:pageSize', size)
  emit('size-change', size)
}

// ========== 多选 ==========
function onSelectionChange(rows: any[]) {
  emit('selection-change', rows)
}

// ========== 排序 ==========
function onSortChange({ prop, order }: { prop: string; order: string | null }) {
  emit('update:sortProp', prop || '')
  emit('update:sortOrder', order as 'ascending' | 'descending' | null)
  emit('sort-change', { prop: prop || '', order: order || '' })
}

// ========== 暴露方法 ==========
defineExpose({
  tableRef,
  /** 清空选中行 */
  clearSelection() {
    tableRef.value?.clearSelection()
  },
  /** 手动设置排序 */
  sort(prop: string, order: 'ascending' | 'descending') {
    tableRef.value?.sort(prop, order)
  },
  /** 清空排序 */
  clearSort() {
    tableRef.value?.clearSort()
  },
})
</script>

<style scoped>
.sm-crud-table {
  width: 100%;
}
.sm-crud-table__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--sm-bg-card);
  border-radius: var(--sm-radius-md) var(--sm-radius-md) 0 0;
  margin-bottom: 0;
}
.sm-crud-table__title {
  font-size: 16px;
  font-weight: 600;
  color: var(--sm-text-primary);
}
.sm-crud-table__header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
.sm-crud-table__pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
