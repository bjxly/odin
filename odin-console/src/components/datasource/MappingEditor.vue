<script setup lang="ts">
import { computed } from 'vue'
import {
  ElTable,
  ElTableColumn,
  ElSelect,
  ElOption,
  ElButton,
  ElTag,
  ElProgress,
  ElEmpty,
  ElTooltip,
  ElIcon,
} from 'element-plus'
import { Plus, Delete, Sort, Key } from '@element-plus/icons-vue'
import type {
  MappingColumnOption,
  MappingEditorRow,
  MappingPropertyOption,
} from '@/types/mapping'

const props = defineProps<{
  /** 编辑区行数据（由父级持有，接受建议 / 已保存映射统一进入此表） */
  rows: MappingEditorRow[]
  /** 当前本体类的属性选项 */
  properties: MappingPropertyOption[]
  /** 当前物理表的列选项 */
  columns: MappingColumnOption[]
  className?: string
  tableName?: string
}>()

const emit = defineEmits<{
  add: []
  remove: [key: string]
}>()

const asRow = (row: unknown): MappingEditorRow => row as MappingEditorRow

/** 已被占用的属性 / 列（用于禁用重复选项） */
const usedProps = computed(() => new Set(props.rows.filter((r) => r.propertyId).map((r) => r.propertyId)))
const usedCols = computed(() => new Set(props.rows.filter((r) => r.column).map((r) => r.column)))

/** 重复出现的属性 / 列（保存前需人工处理） */
function duplicates(values: string[]): Set<string> {
  const seen = new Set<string>()
  const dup = new Set<string>()
  for (const v of values) {
    if (!v) continue
    if (seen.has(v)) dup.add(v)
    else seen.add(v)
  }
  return dup
}

const dupProps = computed(() => duplicates(props.rows.map((r) => r.propertyId)))
const dupCols = computed(() => duplicates(props.rows.map((r) => r.column)))

const validRows = computed(() => props.rows.filter((r) => r.propertyId && r.column))
const incompleteRows = computed(() => props.rows.length - validRows.value.length)
const hasConflict = computed(() => dupProps.value.size > 0 || dupCols.value.size > 0)

function isPropDisabled(row: unknown, id: string): boolean {
  const r = asRow(row)
  return usedProps.value.has(id) && r.propertyId !== id
}

function isColDisabled(row: unknown, name: string): boolean {
  const r = asRow(row)
  return usedCols.value.has(name) && r.column !== name
}

function colTypeOf(name: string): string {
  return props.columns.find((c) => c.name === name)?.type ?? ''
}

/** 本体属性类型（range）回显 */
function propRangeOf(id: string): string {
  return props.properties.find((p) => p.id === id)?.range ?? ''
}

/** 行状态：有效 / 不完整 / 重复 */
function rowStatus(row: unknown): { type: 'success' | 'warning' | 'danger' | 'info'; label: string } {
  const r = asRow(row)
  if (dupProps.value.has(r.propertyId) || dupCols.value.has(r.column)) return { type: 'danger', label: '重复' }
  if (r.propertyId && r.column) return { type: 'success', label: '有效' }
  if (!r.propertyId && !r.column) return { type: 'info', label: '待选择' }
  return { type: 'warning', label: '不完整' }
}

function originTag(row: unknown): { type: 'primary' | 'success' | 'info'; label: string } {
  const o = asRow(row).origin
  if (o === 'suggest') return { type: 'success', label: '建议' }
  if (o === 'saved') return { type: 'primary', label: '已保存' }
  return { type: 'info', label: '手动' }
}

/**
 * 置信度统一以数值展示：
 * 已保存行 = 后端存储的 confidence；建议来源行 = 建议 confidence；手动行（或无存储值）= 100%
 * 兼容后端以百分数（0~100）返回的情况
 */
function confidenceOf(row: unknown): number {
  const c = asRow(row).confidence
  if (typeof c === 'number' && Number.isFinite(c)) return c > 1 ? c / 100 : c
  return 1
}

function confidenceColor(c: number): string {
  if (c >= 0.8) return '#38a89d'
  if (c >= 0.5) return '#e6a23c'
  return '#f56c6c'
}

/** 不完整 / 重复行加淡色底，便于快速定位 */
function rowClassName({ row }: { row: MappingEditorRow }): string {
  const st = rowStatus(row)
  if (st.label === '重复') return 'row-conflict'
  if (st.label === '有效') return ''
  return 'row-incomplete'
}
</script>

<template>
  <section class="mapping-editor">
    <!-- 头部 -->
    <header class="editor-head">
      <div class="head-main">
        <span class="head-icon"><ElIcon><Sort /></ElIcon></span>
        <div class="head-text">
          <h3>映射编辑区</h3>
          <p class="head-sub">
            <strong>{{ tableName || '未选择表' }}</strong>
            <span class="arrow">→</span>
            <strong>{{ className || '未选择类' }}</strong>
            <span class="head-tip">左为物理列、右为本体属性；可增删改，接受的建议也汇入此表统一编辑</span>
          </p>
        </div>
      </div>
      <div class="head-side">
        <div class="counters">
          <span class="counter">
            <b>{{ rows.length }}</b> 行
          </span>
          <span class="counter ok">
            <b>{{ validRows.length }}</b> 有效
          </span>
          <span v-if="incompleteRows" class="counter warn">
            <b>{{ incompleteRows }}</b> 待补全
          </span>
          <span v-if="hasConflict" class="counter err">
            <b>{{ dupProps.size + dupCols.size }}</b> 重复
          </span>
        </div>
        <ElButton type="primary" :icon="Plus" @click="emit('add')">添加映射</ElButton>
      </div>
    </header>

    <!-- 可编辑表格 -->
    <ElTable
      v-if="rows.length"
      :data="rows"
      row-key="key"
      border
      stripe
      size="small"
      max-height="360"
      :row-class-name="rowClassName"
      class="editor-table"
    >
      <ElTableColumn type="index" width="48" label="#" align="center" />

      <ElTableColumn label="物理列" min-width="230">
        <template #default="{ row }">
          <ElSelect
            :model-value="asRow(row).column || undefined"
            clearable
            filterable
            size="small"
            placeholder="选择物理列"
            style="width: 100%"
            @change="(v: any) => (asRow(row).column = typeof v === 'string' ? v : '')"
          >
            <ElOption
              v-for="c in columns"
              :key="c.name"
              :label="c.type ? `${c.name}（${c.type}）` : c.name"
              :value="c.name"
              :disabled="isColDisabled(row, c.name)"
            >
              <span class="opt-main">
                <ElIcon v-if="c.isPk" class="opt-key"><Key /></ElIcon>
                <code>{{ c.name }}</code>
                <em v-if="c.type" class="opt-type">{{ c.type }}</em>
              </span>
              <span class="opt-meta">
                <em v-if="c.comment" class="opt-comment">{{ c.comment }}</em>
                <em v-if="isColDisabled(row, c.name)" class="opt-used">已用</em>
              </span>
            </ElOption>
          </ElSelect>
          <div v-if="asRow(row).column" class="cell-hint">
            {{ colTypeOf(asRow(row).column) || '未知类型' }}
          </div>
        </template>
      </ElTableColumn>

      <ElTableColumn width="52" align="center">
        <template #default>
          <span class="link-arrow">→</span>
        </template>
      </ElTableColumn>

      <ElTableColumn label="本体属性" min-width="230">
        <template #default="{ row }">
          <ElSelect
            :model-value="asRow(row).propertyId || undefined"
            clearable
            filterable
            size="small"
            placeholder="选择本体属性"
            style="width: 100%"
            @change="(v: any) => (asRow(row).propertyId = typeof v === 'string' ? v : '')"
          >
            <ElOption
              v-for="p in properties"
              :key="p.id"
              :label="p.label === p.id ? p.id : `${p.label}（${p.id}）`"
              :value="p.id"
              :disabled="isPropDisabled(row, p.id)"
            >
              <span class="opt-main">
                <ElIcon v-if="p.isKey" class="opt-key"><Key /></ElIcon>
                {{ p.label }}
              </span>
              <span class="opt-meta">
                <code>{{ p.id }}</code>
                <em v-if="p.range" class="opt-type">{{ p.range }}</em>
                <em v-if="isPropDisabled(row, p.id)" class="opt-used">已用</em>
              </span>
            </ElOption>
          </ElSelect>
          <div v-if="asRow(row).propertyId" class="cell-hint">
            {{ propRangeOf(asRow(row).propertyId) || '未定义类型' }}
          </div>
        </template>
      </ElTableColumn>

      <ElTableColumn label="置信度" width="132" align="center">
        <template #default="{ row }">
          <div class="conf-cell">
            <ElProgress
              :percentage="Math.round(confidenceOf(row) * 100)"
              :color="confidenceColor(confidenceOf(row))"
              :stroke-width="6"
              :show-text="false"
              class="conf-bar"
            />
            <span class="conf-text" :style="{ color: confidenceColor(confidenceOf(row)) }">
              {{ Math.round(confidenceOf(row) * 100) }}%
            </span>
          </div>
        </template>
      </ElTableColumn>

      <ElTableColumn label="来源" width="86" align="center">
        <template #default="{ row }">
          <ElTag :type="originTag(row).type" size="small" effect="plain">
            {{ originTag(row).label }}
          </ElTag>
        </template>
      </ElTableColumn>

      <ElTableColumn label="状态" width="86" align="center">
        <template #default="{ row }">
          <ElTag :type="rowStatus(row).type" size="small">{{ rowStatus(row).label }}</ElTag>
        </template>
      </ElTableColumn>

      <ElTableColumn label="操作" width="70" align="center">
        <template #default="{ row }">
          <ElTooltip content="删除该映射行" placement="top">
            <ElButton
              type="danger"
              size="small"
              plain
              circle
              :icon="Delete"
              @click="emit('remove', asRow(row).key)"
            />
          </ElTooltip>
        </template>
      </ElTableColumn>
    </ElTable>

    <!-- 空态 -->
    <div v-else class="editor-empty">
      <ElEmpty :image-size="88" description="编辑区暂无映射行">
        <div class="empty-actions">
          <ElButton type="primary" :icon="Plus" @click="emit('add')">添加映射</ElButton>
          <span class="empty-tip">或在下方「映射建议」中点击接受，快速填充</span>
        </div>
      </ElEmpty>
    </div>

    <!-- 属性名回显（无匹配属性定义时提示） -->
    <p v-if="!properties.length" class="editor-note">
      当前本体类没有可用属性，请先到「本体图谱」为该类定义属性。
    </p>
    <p v-else-if="!columns.length" class="editor-note">
      未获取到表 <code>{{ tableName }}</code> 的列信息，请确认数据源已连接并已内省 Schema。
    </p>
  </section>
</template>

<style scoped>
.mapping-editor {
  background: #fff;
  border: 1px solid #bee3f8;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 12px rgba(49, 130, 206, 0.08);
}

/* 头部 */
.editor-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  padding: 14px 20px;
  background: linear-gradient(135deg, #f0f7ff 0%, #e6f2fd 100%);
  border-bottom: 1px solid #bee3f8;
}

.head-main {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.head-icon {
  width: 38px;
  height: 38px;
  flex-shrink: 0;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 18px;
  background: linear-gradient(135deg, #3182ce 0%, #2b6cb0 100%);
  box-shadow: 0 4px 10px rgba(43, 108, 176, 0.28);
}

.head-text h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #1a365d;
  letter-spacing: 0.2px;
}

.head-sub {
  margin: 4px 0 0;
  font-size: 12px;
  color: #4a6fa5;
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.head-sub strong {
  color: #2b6cb0;
  font-weight: 600;
}

.arrow {
  color: #90cdf4;
  font-weight: 700;
}

.head-tip {
  padding-left: 8px;
  margin-left: 4px;
  border-left: 1px dashed #bee3f8;
  color: #7b9cc4;
}

.head-side {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-shrink: 0;
}

.counters {
  display: flex;
  gap: 10px;
  font-size: 12px;
  color: #4a6fa5;
}

.counter b {
  font-size: 16px;
  font-weight: 700;
  color: #2b6cb0;
  margin-right: 2px;
}

.counter.ok b {
  color: #2c9c8f;
}

.counter.warn b {
  color: #d69e2e;
}

.counter.err b {
  color: #e53e3e;
}

/* 表格 */
.editor-table {
  margin: 0;
}

.editor-table :deep(.el-table__cell) {
  vertical-align: middle;
}

.editor-table :deep(.row-incomplete) {
  background: #fffdf5;
}

.editor-table :deep(.row-conflict) {
  background: #fff5f5;
}

.link-arrow {
  color: #3182ce;
  font-weight: 700;
  font-size: 15px;
}

.cell-hint {
  margin-top: 2px;
  font-size: 11px;
  color: #90a4bd;
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
}

/* 下拉选项 */
.opt-main {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-weight: 500;
}

.opt-key {
  color: #e6a23c;
  font-size: 12px;
}

.opt-meta {
  float: right;
  margin-left: 16px;
  font-size: 11px;
  color: #a0aec0;
  display: inline-flex;
  gap: 6px;
  align-items: center;
}

.opt-meta code {
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
}

.opt-meta em {
  font-style: normal;
}

.opt-comment {
  max-width: 140px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.opt-type {
  font-style: normal;
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  color: #63b3ed;
}

.opt-used {
  color: #e6a23c;
}

/* 置信度 */
.conf-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  justify-content: center;
}

.conf-bar {
  width: 56px;
}

.conf-text {
  font-size: 12px;
  font-weight: 700;
}

/* 空态与提示 */
.editor-empty {
  padding: 8px 0 4px;
}

.empty-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.empty-tip {
  font-size: 12px;
  color: #90a4bd;
}

.editor-note {
  margin: 0;
  padding: 10px 20px;
  font-size: 12px;
  color: #b7791f;
  background: #fffaf0;
  border-top: 1px solid #feebc8;
}

.editor-note code {
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  color: #975a16;
}
</style>
