import type { TableColumnCtx } from 'element-plus'

/**
 * 表格列定义（增强版）
 */
export interface TableColumn {
  /** 字段名 */
  prop?: string
  /** 列标题 */
  label: string
  /** 列宽 */
  width?: number | string
  /** 最小列宽 */
  minWidth?: number | string
  /** 排序 */
  sortable?: boolean | 'custom'
  /** 对齐方式 */
  align?: 'left' | 'center' | 'right'
  /** 自定义渲染插槽名 */
  slot?: string
  /** 格式化函数 */
  formatter?: (row: any, column: TableColumnCtx<any>, cellValue: any, index: number) => string
  /** 超出省略 tooltip */
  showOverflowTooltip?: boolean
  /** 默认是否可见（默认 true） */
  visible?: boolean
  /** 列固定 */
  fixed?: boolean | 'left' | 'right'
  /** 列的 className */
  className?: string
}

/**
 * 列配置持久化结构
 */
export interface ColumnPreference {
  prop: string
  visible: boolean
  order: number
  width?: number
}

/**
 * 分页参数
 */
export interface PaginationState {
  current: number
  size: number
}

/**
 * 排序参数
 */
export interface SortState {
  prop: string
  order: 'ascending' | 'descending' | null
}
