<script setup lang="ts">
import { ref, computed, reactive, onMounted, watch, h, type VNode } from 'vue'
import {
  ElButton,
  ElInput,
  ElTag,
  ElDrawer,
  ElForm,
  ElFormItem,
  ElInputNumber,
  ElCheckbox,
  ElMessage,
  ElMessageBox,
  ElNotification,
  ElTooltip,
  ElEmpty,
  ElTable,
  ElTableColumn,
  ElBadge,
  ElSelect,
  ElOption,
} from 'element-plus'
import {
  Plus,
  Search,
  Coin,
  Edit,
  DataAnalysis,
  VideoPlay,
  Link,
  Refresh,
} from '@element-plus/icons-vue'
import SchemaBrowser from '@/components/datasource/SchemaBrowser.vue'
import MappingSuggest from '@/components/datasource/MappingSuggest.vue'
import MappingEditor from '@/components/datasource/MappingEditor.vue'
import { useDataSourceStore } from '@/stores/datasource'
import { useMappingStore } from '@/stores/mapping'
import { useOntologyStore } from '@/stores/ontology'
import { diffGroups, diffHeadline } from '@/utils/schemaDiff'
import type {
  MappingColumnOption,
  MappingEditorRow,
  MappingPropertyOption,
  PropertyMap,
} from '@/types/mapping'
import type { MappingVO } from '@/types/api'

const dsStore = useDataSourceStore()
const maps = useMappingStore()
const onto = useOntologyStore()

// 数据源列表（响应式派生自 datasource store）
const sources = computed(() => dsStore.sources)

// 获取当前源的 Schema（来自 datasourceStore.schemas）
const currentTables = computed(() => {
  if (!currentSource.value) return []
  return dsStore.getSchema(currentSource.value.id)?.tables ?? []
})

/** MappingSuggest 组件消费的建议行形状 */
interface SuggestItem {
  column: string
  columnType: string
  columnComment?: string
  propertyId: string | null
  propertyLabel?: string
  confidence: number
  reason: string
  accepted?: boolean
}

/** 映射三元组（物理导向）：数据源 → 物理表 → 目标本体类 */
const mapSourceId = ref('')
const mapTable = ref('')
const mapClassId = ref('')
const suggestLoading = ref(false)
const suggestions = ref<SuggestItem[]>([])
/** 用户是否手动选过目标类（选过后不再自动推荐） */
const classTouched = ref(false)

const mapSource = computed(() =>
  mapSourceId.value ? dsStore.sourceById(mapSourceId.value) ?? null : null,
)

/** 当前数据源的物理表（Schema 内省结果） */
const mapTables = computed(() =>
  mapSourceId.value ? dsStore.getSchema(mapSourceId.value)?.tables ?? [] : [],
)

const mapClassLabel = computed(
  () => onto.classes.find((c) => c.id === mapClassId.value)?.label || mapClassId.value || '—',
)

/** 属性 id（形如 Customer.code）→ 中文标签 */
function propertyLabelOf(propertyId: string | null): string | undefined {
  if (!propertyId) return undefined
  const [cid, pid] = propertyId.includes('.') ? propertyId.split('.') : ['', propertyId]
  const cls = onto.classes.find((c) => c.id === (cid || mapClassId.value))
  const prop = cls?.properties.find((p) => p.id === (pid || propertyId))
  return prop?.label || prop?.id
}

/** 调用后端建议接口，替代原先硬编码的 mock 建议 */
async function loadSuggestions() {
  const cls = onto.classes.find((c) => c.id === mapClassId.value)
  if (!mapSourceId.value || !cls?.numericId || !mapTable.value || onto.currentOntologyId == null) {
    suggestions.value = []
    return
  }
  suggestLoading.value = true
  try {
    const rows = await maps.suggestMappings({
      ontology_id: onto.currentOntologyId,
      class_id: cls.numericId,
      datasource_id: Number(mapSourceId.value),
      table_name: mapTable.value,
    })
    suggestions.value = (rows ?? []).map((s) => {
      const propertyId =
        s.suggested_property_name ??
        (s.suggested_property_id != null ? String(s.suggested_property_id) : null)
      return {
        column: s.column,
        columnType: s.column_type,
        columnComment: s.column_comment,
        propertyId,
        propertyLabel: s.suggested_property_label ?? propertyLabelOf(propertyId),
        confidence: s.confidence,
        reason: s.reason ?? '',
      }
    })
    syncSuggestAccepted()
  } finally {
    suggestLoading.value = false
  }
}

// ==================== 手动映射编辑器 ====================
/** 编辑区行：手动新增 / 接受建议 / 已保存映射统一在此编辑 */
const editorRows = ref<MappingEditorRow[]>([])
const savingMapping = ref(false)
const schemaLoading = ref(false)
let rowSeq = 0

function nextRowKey(): string {
  rowSeq += 1
  return `mrow_${rowSeq}_${Date.now().toString(36)}`
}

/** 数据源选项（datasource store 的 sources） */
const mapSourceOptions = computed(() =>
  dsStore.sources.map((s) => ({
    id: s.id,
    label: `${s.name}（${String(s.type).toUpperCase()}）`,
    type: String(s.type).toUpperCase(),
  })),
)

/** 可选物理表（来自当前数据源的 Schema 内省结果） */
const tableOptions = computed(() =>
  mapTables.value.map((t) => ({
    name: t.name,
    label: t.schema ? `${t.schema}.${t.name}` : t.name,
    columnCount: t.columns?.length ?? 0,
  })),
)

/** 归一化名称：去常见表前缀 / 分隔符 / 复数尾缀，便于松散比对 */
function normalizeName(raw: string): string {
  return raw
    .toLowerCase()
    .replace(/^(t|tb|tbl|tab)_/, '')
    .replace(/[_\-\s.]+/g, '')
    .replace(/ies$/, 'y')
    .replace(/(es|s)$/, '')
}

/** 名称重合度（0~1） */
function overlapScore(a: string, b: string): number {
  if (!a || !b) return 0
  if (a === b) return 1
  if (a.includes(b) || b.includes(a)) return 0.85
  const sa = new Set(a)
  const sb = new Set(b)
  let hit = 0
  for (const ch of sa) if (sb.has(ch)) hit += 1
  return hit / Math.max(sa.size, sb.size)
}

/** 表名 × 类名相似度（0~1），仅用于推荐排序 */
function classSimilarity(table: string, cls: { id: string; label: string }): number {
  if (!table) return 0
  const t = normalizeName(table)
  return Math.max(overlapScore(t, normalizeName(cls.id)), overlapScore(t, normalizeName(cls.label)))
}

/** 目标本体类候选：按表名与类名相似度排序推荐，仍允许任选 */
const targetClassOptions = computed(() =>
  onto.classes
    .filter((c) => !c.abstract && c.numericId != null)
    .map((c) => ({
      id: c.id,
      name: c.id,
      label: c.label || c.id,
      propCount: c.properties.length,
      score: classSimilarity(mapTable.value, c),
    }))
    .sort((a, b) => b.score - a.score || a.label.localeCompare(b.label)),
)

/** 相似度百分比文案 */
function scoreText(score: number): string {
  return `${Math.round(score * 100)}%`
}

/** 当前目标类的属性 → 编辑器下拉选项 */
const editorPropertyOptions = computed<MappingPropertyOption[]>(() =>
  (onto.classes.find((c) => c.id === mapClassId.value)?.properties ?? []).map((p) => ({
    id: p.id,
    label: p.label || p.id,
    range: p.range,
    isKey: !!p.isKey,
    sensitive: !!p.sensitive,
  })),
)

/** 当前物理表的列 → 编辑器下拉选项 */
const editorColumnOptions = computed<MappingColumnOption[]>(() => {
  if (!mapSourceId.value || !mapTable.value) return []
  const t = mapTables.value.find((x) => x.name === mapTable.value)
  return (t?.columns ?? []).map((c) => ({
    name: c.name,
    type: c.type,
    comment: c.comment,
    isPk: !!c.isPk,
  }))
})

const validEditorCount = computed(
  () => editorRows.value.filter((r) => r.propertyId && r.column).length,
)

/** 新增一行空映射（属性 / 列待选） */
function addEditorRow(partial?: Partial<MappingEditorRow>): MappingEditorRow {
  const row: MappingEditorRow = {
    key: nextRowKey(),
    propertyId: partial?.propertyId ?? '',
    column: partial?.column ?? '',
    confidence: partial?.confidence ?? null,
    origin: partial?.origin ?? 'manual',
  }
  editorRows.value = [...editorRows.value, row]
  return row
}

function onEditorAdd() {
  if (!editorPropertyOptions.value.length) {
    ElMessage.warning('当前本体类没有可用属性，请先在本体图谱中定义')
  }
  addEditorRow()
}

function onEditorRemove(key: string) {
  editorRows.value = editorRows.value.filter((r) => r.key !== key)
}

/** 写入 / 更新一行：同一属性或同一列视为同一条映射，保持 1:1 */
function upsertEditorRow(
  propertyId: string,
  column: string,
  confidence?: number | null,
  origin: MappingEditorRow['origin'] = 'suggest',
) {
  const list = [...editorRows.value]
  let i = list.findIndex((r) => r.propertyId === propertyId)
  if (i < 0) i = list.findIndex((r) => r.column === column)
  if (i >= 0) {
    list[i] = {
      ...list[i],
      propertyId,
      column,
      confidence: confidence ?? list[i].confidence ?? null,
      origin,
    }
  } else {
    list.push({ key: nextRowKey(), propertyId, column, confidence: confidence ?? null, origin })
  }
  editorRows.value = list
}

/** 用后端已保存的映射初始化编辑区（保留存储的置信度） */
function loadEditorFromSaved() {
  const saved = currentClassMap()?.properties ?? []
  editorRows.value = saved.map((p) => ({
    key: nextRowKey(),
    propertyId: p.propertyId,
    column: p.column,
    confidence: typeof p.confidence === 'number' ? p.confidence : null,
    origin: 'saved' as const,
  }))
}

/** 建议区中已进入编辑区的列标记为“已接受” */
function syncSuggestAccepted() {
  const cols = new Set(editorRows.value.filter((r) => r.column).map((r) => r.column))
  for (const s of suggestions.value) {
    if (cols.has(s.column)) s.accepted = true
  }
}

// 编辑区变动时双向同步建议区状态：列已在编辑区 → 已接受；已移除 → 恢复待处理
watch(
  editorRows,
  () => {
    const cols = new Set(editorRows.value.map((r) => r.column).filter(Boolean))
    for (const s of suggestions.value) {
      if (cols.has(s.column)) s.accepted = true
      else if (s.accepted === true) s.accepted = undefined
    }
  },
  { deep: true },
)

/** 三元组变化：重载该 (数据源, 物理表, 目标类) 的已保存映射并拉取建议 */
async function onMappingTargetChange() {
  suggestions.value = []
  loadEditorFromSaved()
  await loadSuggestions()
}

/** 默认目标类：该表已有映射 > 名称高度相似 > 该源其它已映射类 > 首个候选 */
function pickDefaultClass(table: string): string {
  const usable = (id: string) =>
    onto.classes.some((c) => c.id === id && !c.abstract && c.numericId != null)
  const saved = mapSourceId.value ? maps.mapsOfSource(mapSourceId.value) : []
  const ranked = targetClassOptions.value
  const has = (id: string) => ranked.some((o) => o.id === id)
  const onTable = saved.find((m) => m.table === table && usable(m.classId) && has(m.classId))
  if (onTable) return onTable.classId
  const best = ranked[0]
  if (best && best.score >= 0.5) return best.id
  const anySaved = saved.find((m) => usable(m.classId) && has(m.classId))
  if (anySaved) return anySaved.classId
  return best?.id ?? ''
}

/** 确保 Schema 就绪，按当前数据源重选表与目标类，再载入映射与建议 */
async function reloadForSource() {
  suggestions.value = []
  editorRows.value = []
  classTouched.value = false
  const sid = mapSourceId.value
  if (!sid) {
    mapTable.value = ''
    mapClassId.value = ''
    return
  }
  schemaLoading.value = true
  try {
    if (!dsStore.getSchema(sid)) await dsStore.fetchSchema(Number(sid))
  } finally {
    schemaLoading.value = false
  }
  const saved = maps.mapsOfSource(sid)
  const savedTable = saved.find((m) => mapTables.value.some((t) => t.name === m.table))?.table
  mapTable.value = savedTable ?? mapTables.value[0]?.name ?? ''
  mapClassId.value = pickDefaultClass(mapTable.value)
  await onMappingTargetChange()
}

/** 打开映射抽屉：以卡片数据源为默认值初始化三元组 */
async function prepareMappingDrawer(source: any) {
  mapSourceId.value = String(source?.id ?? dsStore.sources[0]?.id ?? '')
  await reloadForSource()
}

/** 切换数据源：重新载入表与目标类 */
async function onMappingSourceChange() {
  await reloadForSource()
}

/** 切换物理表：未手动指定类时按相似度重新推荐目标类 */
async function onMappingTableChange() {
  if (!classTouched.value) mapClassId.value = pickDefaultClass(mapTable.value)
  await onMappingTargetChange()
}

/** 切换目标本体类 */
async function onMappingClassChange() {
  classTouched.value = true
  await onMappingTargetChange()
}

// 状态管理
const searchQuery = ref('')
const drawerType = ref<'schema' | 'edit' | 'mapping' | 'mappings' | null>(null)
const drawerVisible = ref(false)
const currentSource = ref<any>(null)

// 获取当前源的已保存映射（打开“查看映射”抽屉时拉取）
const sourceMappings = ref<MappingVO[]>([])
const mappingsLoading = ref(false)

async function loadSourceMappings(source: any) {
  mappingsLoading.value = true
  try {
    sourceMappings.value = await maps.queryMappings({
      datasource_id: Number(source.id),
      ...(onto.currentOntologyId != null ? { ontology_id: onto.currentOntologyId } : {}),
    })
  } finally {
    mappingsLoading.value = false
  }
}

const mappingsSummary = computed(() => ({
  classes: sourceMappings.value.length,
  props: sourceMappings.value.reduce((n, m) => n + (m.property_mappings?.length ?? 0), 0),
}))

/** 校验状态 → 标签颜色与文案 */
function validationTag(status?: string): { type: 'success' | 'warning' | 'danger' | 'info'; label: string } {
  switch (status) {
    case 'column_deleted':
      return { type: 'danger', label: '列已删除' }
    case 'column_changed':
      return { type: 'warning', label: '列已变更' }
    case 'validated':
      return { type: 'success', label: '已验证' }
    default:
      return { type: 'info', label: status || '未校验' }
  }
}

// 过滤后的数据源
const filteredSources = computed(() => {
  if (!searchQuery.value) return [...sources.value]
  const query = searchQuery.value.toLowerCase()
  return sources.value.filter(
    (s: any) =>
      s.name.toLowerCase().includes(query) ||
      s.id.toLowerCase().includes(query) ||
      s.type.toLowerCase().includes(query) ||
      s.tags?.some((t: string) => t.toLowerCase().includes(query)),
  )
})

// 类型图标和颜色 - 科技蓝主题
const typeConfig: Record<string, { icon: string; color: string; bg: string }> = {
  postgresql: { icon: '🐘', color: '#3182ce', bg: '#ebf8ff' },
  mysql: { icon: '🐬', color: '#2b6cb0', bg: '#ebf8ff' },
  dameng: { icon: '🏛️', color: '#2c5282', bg: '#ebf8ff' },
  kingbase: { icon: '👑', color: '#2a4365', bg: '#ebf8ff' },
  mongodb: { icon: '🍃', color: '#3182ce', bg: '#ebf8ff' },
  redis: { icon: '🔴', color: '#2b6cb0', bg: '#ebf8ff' },
  csv: { icon: '📄', color: '#3182ce', bg: '#ebf8ff' },
  clickhouse: { icon: '⚡', color: '#2c5282', bg: '#ebf8ff' },
}

const getTypeConfig = (type: string) => {
  return typeConfig[type] || { icon: '💾', color: '#909399', bg: '#f5f7fa' }
}

// 打开抽屉
const openDrawer = (type: 'schema' | 'edit' | 'mapping' | 'mappings', source: any) => {
  currentSource.value = source
  drawerType.value = type
  drawerVisible.value = true
  if (type === 'schema') void loadSchemaFor(source)
  if (type === 'mapping') void prepareMappingDrawer(source)
  if (type === 'mappings') void loadSourceMappings(source)
}

/** 打开 Schema 抽屉即强制拉取（后端在缓存为空时会自动内省，首次打开即有数据） */
async function loadSchemaFor(source: any) {
  schemaLoading.value = true
  try {
    const res = await dsStore.fetchSchema(Number(source.id))
    if (!res) ElMessage.error('Schema 加载失败，请先测试连接后重试')
  } finally {
    schemaLoading.value = false
  }
}

// 关闭抽屉
const closeDrawer = () => {
  const wasSchema = drawerType.value === 'schema'
  drawerVisible.value = false
  drawerType.value = null
  // Schema 抽屉关闭后重新拉取列表，刷新卡片上的表数量
  if (wasSchema) void dsStore.fetchDatasources()
}

// 测试连接（调用后端测试接口）
const handleTest = async (source: any) => {
  const r = await dsStore.testDatasource(Number(source.id))
  if (r.status === 'ok') ElMessage.success(`连接正常 · ${r.latencyMs} ms`)
  else if (r.status === 'warn') ElMessage.warning(r.testMessage || '连接不稳定')
  else ElMessage.error(r.testMessage || '连接失败')
}

// 编辑表单
const editForm = reactive({
  name: '',
  host: '',
  port: 0,
  database: '',
  username: '',
  password: '',
  defaultSchema: '',
  readOnly: true,
  description: '',
})

// 打开编辑抽屉
const openEditDrawer = (source: any) => {
  editForm.name = source.name
  editForm.host = source.host || ''
  editForm.port = source.port || 0
  editForm.database = source.database || ''
  editForm.username = source.user || ''
  // 密码出于安全考虑不回显，留空表示不修改
  editForm.password = ''
  editForm.defaultSchema = source.defaultSchema || ''
  editForm.readOnly = true
  editForm.description = source.description || ''
  openDrawer('edit', source)
}

// ==================== 刷新 Schema ====================
/** 通知中每类差异最多展示的明细条数，超出折叠为「…等 N 项」 */
const MAX_DIFF_ITEMS = 6

/**
 * 刷新 Schema（调用后端差量刷新接口 POST /datasources/:id/schema/refresh）。
 * 后端返回嵌套结构 `{ schema, diff }`：store 已解包 `schema` 写入缓存
 * （SchemaBrowser 通过 currentTables 自动重渲染）并把 `diff` 归一化为 SchemaDiffVO，
 * 这里据其计数与明细如实提示「X 新增 / Y 删除 / Z 变更」（表级 + 列级）。
 */
const handleRefreshSchema = async () => {
  if (!currentSource.value) return
  const id = Number(currentSource.value.id)
  schemaLoading.value = true
  try {
    ElMessage.info('正在刷新 Schema...')
    const res = await dsStore.refreshSchema(id)
    if (!res) {
      ElMessage.error('Schema 刷新失败，请先测试连接')
      return
    }
    const tableCount = res.table_count ?? res.tables?.length ?? 0
    const columnCount =
      res.column_count ??
      res.tables?.reduce((n, t) => n + (t.columns?.length ?? 0), 0) ??
      0
    const counts = res.diff?.counts
    if (!counts || counts.total === 0) {
      ElMessage.success(
        `Schema 刷新完成：${tableCount} 张表 / ${columnCount} 个字段，无结构变化`,
      )
      return
    }
    const groups = diffGroups(res.diff)
    ElNotification({
      title: `刷新完成：${tableCount} 张表 / ${columnCount} 个字段`,
      type: 'warning',
      duration: 8000,
      message: h('div', { style: 'font-size:12px;line-height:1.8;max-width:440px' }, [
        h(
          'div',
          { style: 'font-weight:600;color:#1a365d;margin-bottom:2px' },
          diffHeadline(res.diff),
        ),
        ...groups.map((g) => {
          const kids: VNode[] = [
            h(
              'span',
              { style: 'color:#2b6cb0;font-weight:600' },
              `${g.label} ${g.items.length}：`,
            ),
            h(
              'code',
              { style: "font-family:'JetBrains Mono',Consolas,monospace" },
              g.items.slice(0, MAX_DIFF_ITEMS).join('、'),
            ),
          ]
          if (g.items.length > MAX_DIFF_ITEMS) {
            kids.push(h('span', { style: 'color:#a0aec0' }, ` …等 ${g.items.length} 项`))
          }
          return h('div', { style: 'color:#4a5568;word-break:break-all' }, kids)
        }),
        h(
          'div',
          { style: 'color:#a0aec0;margin-top:4px' },
          '受影响列的映射校验状态已由后端同步更新',
        ),
      ]),
    })
  } finally {
    schemaLoading.value = false
  }
}

const savingEdit = ref(false)

// 保存编辑（走 datasource store 的更新接口）
const handleSaveEdit = async () => {
  if (!currentSource.value) return
  savingEdit.value = true
  try {
    await dsStore.updateDatasource(Number(currentSource.value.id), {
      name: editForm.name,
      host: editForm.host || undefined,
      port: editForm.port || undefined,
      database_name: editForm.database || undefined,
      username: editForm.username || undefined,
      password: editForm.password || undefined,
      schema_name: editForm.defaultSchema || undefined,
      description: editForm.description || undefined,
    })
    ElMessage.success('数据源配置已更新')
    closeDrawer()
  } catch {
    ElMessage.error('数据源配置更新失败')
  } finally {
    savingEdit.value = false
  }
}

/** 当前 (数据源, 物理表, 目标类) 三元组已存在的映射配置 */
function currentClassMap() {
  if (!mapSourceId.value) return undefined
  return maps
    .mapsOfSource(mapSourceId.value)
    .find((m) => m.classId === mapClassId.value && m.table === mapTable.value)
}

/** 将属性映射写回后端（存在则更新，否则创建） */
async function persistProperties(properties: PropertyMap[]) {
  const cls = onto.classes.find((c) => c.id === mapClassId.value)
  if (!mapSourceId.value || !cls?.numericId || !mapTable.value || onto.currentOntologyId == null) {
    return false
  }
  const table = mapTables.value.find((t) => t.name === mapTable.value)
  const existing = currentClassMap()
  const payload = {
    ontology_id: onto.currentOntologyId,
    class_id: cls.numericId,
    datasource_id: Number(mapSourceId.value),
    source_table: mapTable.value,
    source_schema: table?.schema,
    // 后端按 property_mappings 数组持久化（property_name / column_name）
    property_mappings: properties.map((p) => ({
      property_name: p.propertyId,
      column_name: p.column,
      ...(p.valueMap ? { value_map: p.valueMap } : {}),
    })),
  }
  try {
    if (existing) await maps.updateMapping(Number(existing.id), payload)
    else await maps.createMapping(payload)
    // 重新拉取映射，保证 classMaps / 卡片计数同步
    await maps.fetchMappings(
      onto.currentOntologyId != null ? { ontology_id: onto.currentOntologyId } : undefined,
    )
    return true
  } catch {
    return false
  }
}

/** 接受单条建议 → 汇入编辑区（不立即落库，统一由「保存映射」提交） */
const onAcceptSuggest = (item: SuggestItem) => {
  if (!item.propertyId) {
    item.accepted = false
    ElMessage.warning('该建议未匹配到本体属性，可在编辑区手动指定')
    return
  }
  item.accepted = true
  upsertEditorRow(item.propertyId, item.column, item.confidence, 'suggest')
  ElMessage.success(`已加入编辑区：${item.column} → ${item.propertyLabel || item.propertyId}`)
}

/** 忽略单条建议（含已接受 / 已保存行的撤销）→ 从编辑区移除该列（保留手动行） */
const onRejectSuggest = (item: SuggestItem) => {
  item.accepted = false
  const before = editorRows.value.length
  editorRows.value = editorRows.value.filter(
    (r) => !(r.origin !== 'manual' && r.column === item.column),
  )
  const removed = before - editorRows.value.length
  ElMessage.info(
    removed
      ? `已忽略建议，并从编辑区移除 ${removed} 行（${item.column}）`
      : `已忽略建议：${item.column}`,
  )
}

/** 全部接受：高置信度建议批量汇入编辑区 */
const onAcceptAllSuggest = () => {
  const rows = suggestions.value.filter((s) => s.propertyId && s.confidence >= 0.8)
  if (!rows.length) {
    ElMessage.warning('没有置信度 ≥80% 的建议可接受')
    return
  }
  for (const s of rows) {
    s.accepted = true
    upsertEditorRow(s.propertyId as string, s.column, s.confidence, 'suggest')
  }
  ElMessage.success(`已汇入 ${rows.length} 条高置信度映射，确认后点「保存映射」`)
}

/** 全部忽略：清除建议对应行（保留手动新增行） */
const onRejectAllSuggest = () => {
  const cols = new Set(suggestions.value.map((s) => s.column))
  for (const s of suggestions.value) s.accepted = false
  const before = editorRows.value.length
  editorRows.value = editorRows.value.filter(
    (r) => !(r.origin !== 'manual' && cols.has(r.column)),
  )
  ElMessage.info(`已忽略全部建议（编辑区移除 ${before - editorRows.value.length} 行）`)
}

/** 保存：编辑区中所有有效行（物理列 + 本体属性均已选）构建 property_mappings 提交 */
const onSaveMappings = async () => {
  if (!mapSourceId.value) {
    ElMessage.warning('请选择数据源')
    return
  }
  if (!mapTable.value) {
    ElMessage.warning('请选择物理表')
    return
  }
  const cls = onto.classes.find((c) => c.id === mapClassId.value)
  if (!cls?.numericId) {
    ElMessage.warning('请选择有效的目标本体类')
    return
  }
  if (onto.currentOntologyId == null) {
    ElMessage.warning('缺少本体上下文，请先选择本体')
    return
  }

  const valid = editorRows.value.filter((r) => r.propertyId && r.column)
  const skipped = editorRows.value.length - valid.length

  // 重复映射会相互覆盖，先阻断
  const dupProps = valid.map((r) => r.propertyId).filter((v, i, a) => a.indexOf(v) !== i)
  const dupCols = valid.map((r) => r.column).filter((v, i, a) => a.indexOf(v) !== i)
  if (dupProps.length || dupCols.length) {
    ElMessage.error(
      `存在重复映射：${[...new Set([...dupProps, ...dupCols])].join('、')}，请调整后再保存`,
    )
    return
  }

  const existing = currentClassMap()
  if (!valid.length) {
    if (!existing) {
      ElMessage.warning('请至少完成一条映射（物理列 + 本体属性）')
      return
    }
    try {
      await ElMessageBox.confirm(
        `将清空 ${mapTable.value} → ${cls.label} 的全部映射，是否继续？`,
        '确认清空映射',
        { type: 'warning', confirmButtonText: '清空并保存', cancelButtonText: '取消' },
      )
    } catch {
      return
    }
  }

  // 未改动的行保留原有 value_map，避免覆盖枚举翻译配置
  const savedProps = existing?.properties ?? []
  const properties: PropertyMap[] = valid.map((r) => {
    const prev = savedProps.find((p) => p.propertyId === r.propertyId && p.column === r.column)
    return { propertyId: r.propertyId, column: r.column, valueMap: prev?.valueMap }
  })

  savingMapping.value = true
  try {
    const ok = await persistProperties(properties)
    if (!ok) {
      ElMessage.error('映射保存失败')
      return
    }
    ElMessage.success(
      skipped
        ? `映射已保存（${valid.length} 条，跳过 ${skipped} 条不完整行）`
        : `映射已保存（${valid.length} 条）`,
    )
    closeDrawer()
  } finally {
    savingMapping.value = false
  }
}

onMounted(async () => {
  await dsStore.fetchDatasources()
  await maps.fetchMappings(
    onto.currentOntologyId != null ? { ontology_id: onto.currentOntologyId } : undefined,
  )
})
</script>

<template>
  <div class="data-access-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1>数据接入</h1>
        <p class="subtitle">管理和配置您的数据源</p>
      </div>
      <div class="header-right">
        <ElInput
          v-model="searchQuery"
          placeholder="搜索数据源..."
          clearable
          :prefix-icon="Search"
          class="search-input"
        />
        <ElButton type="primary" :icon="Plus">
          新建数据源
        </ElButton>
      </div>
    </div>

    <!-- 数据源卡片网格 -->
    <div v-loading="dsStore.loading" class="cards-container">
      <div v-if="filteredSources.length === 0" class="empty-state">
        <ElEmpty description="没有找到匹配的数据源" />
      </div>
      
      <div v-else class="cards-grid">
        <div
          v-for="source in filteredSources"
          :key="source.id"
          class="source-card"
        >
          <!-- 卡片头部 -->
          <div class="card-header">
            <div class="type-badge" :style="{ background: getTypeConfig(source.type).bg }">
              <span class="type-icon">{{ getTypeConfig(source.type).icon }}</span>
            </div>
            <div class="header-info">
              <h3 class="source-name">{{ source.name }}</h3>
              <div class="source-meta">
                <ElTag
                  :type="source.status === 'connected' ? 'success' : source.status === 'testing' ? 'warning' : 'info'"
                  size="small"
                  effect="plain"
                >
                  {{ source.status === 'connected' ? '已连接' : source.status === 'testing' ? '测试中' : '未连接' }}
                </ElTag>
                <ElTag type="info" size="small">
                  {{ source.type.toUpperCase() }}
                </ElTag>
              </div>
            </div>
            <ElTooltip content="测试连接" placement="top">
              <ElButton
                class="test-btn"
                :icon="VideoPlay"
                size="small"
                circle
                @click.stop="handleTest(source)"
              />
            </ElTooltip>
          </div>

          <!-- 卡片内容 -->
          <div class="card-body">
            <div class="info-row">
              <span class="label">主机</span>
              <span class="value">{{ source.host || '—' }}:{{ source.port || '—' }}</span>
            </div>
            <div class="info-row">
              <span class="label">数据库</span>
              <span class="value">{{ source.database || '—' }}</span>
            </div>
            <div class="info-row">
              <span class="label">表数量</span>
              <span class="value">{{ source.tableCount || 0 }} 张</span>
            </div>
            <div class="info-row">
              <span class="label">已映射类</span>
              <span class="value">
                <ElBadge
                  :value="source.mappedClassCount || 0"
                  :type="source.mappedClassCount > 0 ? 'primary' : 'info'"
                  class="mapping-badge"
                />
              </span>
            </div>
            <div class="info-row" v-if="source.description">
              <span class="label">描述</span>
              <span class="value desc">{{ source.description }}</span>
            </div>
          </div>

          <!-- 标签 -->
          <div class="card-tags" v-if="source.tags?.length">
            <ElTag v-for="tag in source.tags" :key="tag" size="small" effect="plain" type="info">
              {{ tag }}
            </ElTag>
          </div>

          <!-- 操作按钮 -->
          <div class="card-actions">
            <ElButton
              type="primary"
              :icon="Coin"
              size="small"
              @click="openDrawer('schema', source)"
            >
              Schema
            </ElButton>
            <ElButton
              :icon="Edit"
              size="small"
              @click="openEditDrawer(source)"
            >
              编辑
            </ElButton>
            <ElButton
              type="success"
              :icon="DataAnalysis"
              size="small"
              @click="openDrawer('mapping', source)"
            >
              映射配置
            </ElButton>
            <ElButton
              v-if="source.mappedClassCount > 0"
              type="warning"
              :icon="Link"
              size="small"
              @click="openDrawer('mappings', source)"
            >
              查看映射
            </ElButton>
          </div>
        </div>
      </div>
    </div>

    <!-- Schema 浏览抽屉 -->
    <ElDrawer
      v-model="drawerVisible"
      :title="drawerType === 'schema' ? `Schema 浏览 - ${currentSource?.name}` : ''"
      size="80%"
      direction="rtl"
      :before-close="closeDrawer"
      v-if="drawerType === 'schema'"
    >
      <div v-loading="schemaLoading" class="schema-drawer-body">
        <SchemaBrowser
          v-if="currentTables.length"
          :tables="currentTables"
          :source-name="currentSource?.name || ''"
          @refresh="handleRefreshSchema"
        />
        <ElEmpty
          v-else-if="!schemaLoading"
          description="未获取到表结构，请先测试连接后重试"
        >
          <ElButton type="primary" :icon="Refresh" @click="handleRefreshSchema">
            刷新 Schema
          </ElButton>
        </ElEmpty>
      </div>
    </ElDrawer>

    <!-- 编辑抽屉 -->
    <ElDrawer
      v-model="drawerVisible"
      :title="`编辑数据源 - ${currentSource?.name}`"
      size="480px"
      direction="rtl"
      :before-close="closeDrawer"
      v-if="drawerType === 'edit'"
    >
      <ElForm label-width="100px" class="edit-form">
        <ElFormItem label="显示名称">
          <ElInput v-model="editForm.name" placeholder="数据源名称" />
        </ElFormItem>
        <ElFormItem label="主机">
          <ElInput v-model="editForm.host" placeholder="IP 或域名" />
        </ElFormItem>
        <ElFormItem label="端口">
          <ElInputNumber v-model="editForm.port" :min="0" :max="65535" style="width: 100%" />
        </ElFormItem>
        <ElFormItem label="数据库">
          <ElInput v-model="editForm.database" placeholder="数据库名称" />
        </ElFormItem>
        <ElFormItem label="用户名">
          <ElInput v-model="editForm.username" placeholder="数据库用户名" />
        </ElFormItem>
        <ElFormItem label="密码">
          <ElInput
            v-model="editForm.password"
            type="password"
            show-password
            placeholder="留空则不修改"
          />
        </ElFormItem>
        <ElFormItem label="默认 Schema">
          <ElInput v-model="editForm.defaultSchema" placeholder="如 public、CRM" />
        </ElFormItem>
        <ElFormItem label="只读接入">
          <ElCheckbox v-model="editForm.readOnly">只读模式</ElCheckbox>
        </ElFormItem>
        <ElFormItem label="描述">
          <ElInput v-model="editForm.description" type="textarea" :rows="3" placeholder="数据源描述" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <div class="drawer-footer">
          <ElButton @click="closeDrawer">取消</ElButton>
          <ElButton type="primary" :loading="savingEdit" @click="handleSaveEdit">保存</ElButton>
        </div>
      </template>
    </ElDrawer>

    <!-- 映射配置抽屉 -->
    <ElDrawer
      v-model="drawerVisible"
      :title="`映射配置 - ${mapSource?.name || currentSource?.name || ''}`"
      size="90%"
      direction="rtl"
      :before-close="closeDrawer"
      v-if="drawerType === 'mapping'"
    >
      <div v-loading="schemaLoading" class="mapping-drawer-body">
        <!-- 物理导向三元组：数据源 → 物理表 → 目标本体类 -->
        <div class="mapping-toolbar">
          <span class="tb-label">数据源</span>
          <ElSelect
            v-model="mapSourceId"
            filterable
            placeholder="选择数据源"
            class="tb-select"
            @change="onMappingSourceChange"
          >
            <ElOption v-for="s in mapSourceOptions" :key="s.id" :label="s.label" :value="s.id" />
          </ElSelect>

          <span class="tb-sep">→</span>

          <span class="tb-label">物理表</span>
          <ElSelect
            v-model="mapTable"
            filterable
            placeholder="选择物理表"
            class="tb-select"
            :loading="schemaLoading"
            @change="onMappingTableChange"
          >
            <ElOption v-for="t in tableOptions" :key="t.name" :label="t.label" :value="t.name">
              <span class="tb-opt-main">
                <code>{{ t.name }}</code>
              </span>
              <span class="tb-opt-meta">{{ t.columnCount }} 列</span>
            </ElOption>
          </ElSelect>

          <span class="tb-sep">→</span>

          <span class="tb-label">目标本体类</span>
          <ElSelect
            v-model="mapClassId"
            filterable
            placeholder="选择目标本体类"
            class="tb-select"
            @change="onMappingClassChange"
          >
            <ElOption
              v-for="c in targetClassOptions"
              :key="c.id"
              :label="`${c.label}（${c.name}）`"
              :value="c.id"
            >
              <span class="tb-opt-main">{{ c.label }}</span>
              <span class="tb-opt-meta">
                <code>{{ c.name }}</code>
                <em>{{ c.propCount }} 属性</em>
                <em v-if="c.score >= 0.5" class="tb-opt-rec">推荐 {{ scoreText(c.score) }}</em>
              </span>
            </ElOption>
          </ElSelect>

          <ElTooltip content="重新载入该三元组的已保存映射与映射建议" placement="top">
            <ElButton :icon="Refresh" :loading="suggestLoading" @click="onMappingTargetChange">
              重新载入
            </ElButton>
          </ElTooltip>

          <span v-if="!mapSourceId" class="tb-hint">请先选择数据源</span>
          <span v-else-if="!tableOptions.length" class="tb-hint">
            该数据源暂无表结构，请先测试连接或在 Schema 抽屉中刷新
          </span>
          <span v-else-if="!targetClassOptions.length" class="tb-hint">
            当前本体没有可用类，请先到「本体图谱」创建
          </span>
        </div>

        <!-- 映射编辑器：物理列 → 本体属性 -->
        <MappingEditor
          :rows="editorRows"
          :properties="editorPropertyOptions"
          :columns="editorColumnOptions"
          :class-name="mapClassLabel"
          :table-name="mapTable"
          @add="onEditorAdd"
          @remove="onEditorRemove"
        />

        <!-- 映射建议（快速填充手段） -->
        <MappingSuggest
          v-loading="suggestLoading"
          :suggestions="suggestions"
          :class-name="mapClassLabel"
          :table-name="mapTable"
          :source-name="mapSource?.name || currentSource?.name || ''"
          @accept="onAcceptSuggest"
          @reject="onRejectSuggest"
          @accept-all="onAcceptAllSuggest"
          @reject-all="onRejectAllSuggest"
        />
      </div>
      <template #footer>
        <div class="drawer-footer">
          <span class="footer-summary">
            <b>{{ mapTable || '—' }}</b>
            <span class="fs-arrow">→</span>
            <b>{{ mapClassLabel }}</b>
            · 编辑区 {{ editorRows.length }} 行 · 有效
            <b>{{ validEditorCount }}</b> 条
          </span>
          <ElButton @click="closeDrawer">取消</ElButton>
          <ElButton type="primary" :loading="savingMapping" @click="onSaveMappings">
            保存映射
          </ElButton>
        </div>
      </template>
    </ElDrawer>

    <!-- 查看映射抽屉 -->
    <ElDrawer
      v-model="drawerVisible"
      :title="`已映射本体类 - ${currentSource?.name}`"
      size="560px"
      direction="rtl"
      :before-close="closeDrawer"
      v-if="drawerType === 'mappings'"
    >
      <div class="mappings-content" v-loading="mappingsLoading">
        <div class="mappings-summary">
          <div class="summary-item">
            <span class="summary-value">{{ mappingsSummary.classes }}</span>
            <span class="summary-label">个本体类</span>
          </div>
          <div class="summary-item">
            <span class="summary-value">{{ mappingsSummary.props }}</span>
            <span class="summary-label">个属性映射</span>
          </div>
        </div>

        <ElEmpty v-if="!mappingsLoading && sourceMappings.length === 0" description="该数据源暂无已保存的映射" />

        <ElTable v-else :data="sourceMappings" border stripe row-key="id">
          <ElTableColumn type="expand">
            <template #default="{ row }">
              <div class="prop-map-detail">
                <ElTable :data="row.property_mappings ?? []" size="small" border>
                  <ElTableColumn label="本体属性" min-width="150">
                    <template #default="{ row: pm }">
                      <span>{{ pm.property_name || pm.property_id || '—' }}</span>
                    </template>
                  </ElTableColumn>
                  <ElTableColumn label="物理列" min-width="150">
                    <template #default="{ row: pm }">
                      <code>{{ pm.column_name || '—' }}</code>
                      <ElTag v-if="pm.column_data_type" size="small" type="info" class="pm-type">
                        {{ pm.column_data_type }}
                      </ElTag>
                    </template>
                  </ElTableColumn>
                  <ElTableColumn label="校验状态" width="120" align="center">
                    <template #default="{ row: pm }">
                      <ElTag :type="validationTag(pm.validation_status).type" size="small">
                        {{ validationTag(pm.validation_status).label }}
                      </ElTag>
                    </template>
                  </ElTableColumn>
                </ElTable>
              </div>
            </template>
          </ElTableColumn>
          <ElTableColumn label="本体类" width="140">
            <template #default="{ row }">
              <ElTag type="primary" effect="plain">{{ row.class_name || row.class_id }}</ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="source_table" label="物理表" min-width="140">
            <template #default="{ row }">
              <code>{{ row.source_table }}</code>
            </template>
          </ElTableColumn>
          <ElTableColumn label="属性数" width="90" align="center">
            <template #default="{ row }">
              {{ row.property_mappings?.length ?? 0 }}
            </template>
          </ElTableColumn>
        </ElTable>
      </div>
      <template #footer>
        <div class="drawer-footer">
          <ElButton @click="closeDrawer">关闭</ElButton>
          <ElButton type="primary" @click="closeDrawer(); openDrawer('mapping', currentSource)">
            配置新映射
          </ElButton>
        </div>
      </template>
    </ElDrawer>
  </div>
</template>

<style scoped>
.data-access-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: linear-gradient(180deg, #f0f7ff 0%, #e8f4fd 100%);
  overflow: hidden;
}

/* 页面头部 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24px 32px;
  background: linear-gradient(135deg, #2b6cb0 0%, #3182ce 100%);
  border-bottom: none;
}

.header-left h1 {
  margin: 0 0 4px;
  font-size: 24px;
  font-weight: 600;
  color: white;
}

.subtitle {
  margin: 0;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.85);
}

.header-right {
  display: flex;
  gap: 12px;
  align-items: center;
}

.search-input {
  width: 280px;
}

.search-input :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  box-shadow: none;
}

.search-input :deep(.el-input__wrapper:hover),
.search-input :deep(.el-input__wrapper.is-focus) {
  background: rgba(255, 255, 255, 0.3);
  border-color: rgba(255, 255, 255, 0.5);
}

.search-input :deep(.el-input__inner) {
  color: white;
}

.search-input :deep(.el-input__inner::placeholder) {
  color: rgba(255, 255, 255, 0.6);
}

.search-input :deep(.el-input__prefix .el-icon) {
  color: rgba(255, 255, 255, 0.8);
}

/* 卡片容器 */
.cards-container {
  flex: 1;
  overflow-y: auto;
  padding: 24px 32px;
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 400px;
}

/* 卡片网格 */
.cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: 20px;
}

/* 数据源卡片 */
.source-card {
  background: white;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 2px 12px rgba(49, 130, 206, 0.08);
  border: 1px solid #bee3f8;
  transition: all 0.3s ease;
  display: flex;
  flex-direction: column;
}

.source-card:hover {
  box-shadow: 0 8px 24px rgba(49, 130, 206, 0.15);
  transform: translateY(-2px);
  border-color: #3182ce;
}

/* 卡片头部 */
.card-header {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 20px;
}

.type-badge {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: linear-gradient(135deg, #ebf8ff 0%, #bee3f8 100%);
  border: 1px solid #90cdf4;
}

.type-icon {
  font-size: 28px;
}

.header-info {
  flex: 1;
  min-width: 0;
}

.source-name {
  margin: 0 0 8px;
  font-size: 18px;
  font-weight: 600;
  color: #1a365d;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.source-meta {
  display: flex;
  gap: 8px;
  align-items: center;
}

.test-btn {
  opacity: 0;
  transition: opacity 0.2s;
}

.source-card:hover .test-btn {
  opacity: 1;
}

/* 卡片内容 */
.card-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 16px;
  background: linear-gradient(135deg, #f0f7ff 0%, #ebf8ff 100%);
  border-radius: 12px;
  margin-bottom: 16px;
  border: 1px solid #bee3f8;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  font-size: 13px;
}

.label {
  color: #4a6fa5;
  font-weight: 500;
  flex-shrink: 0;
  min-width: 60px;
}

.value {
  color: #1a365d;
  font-weight: 600;
  text-align: right;
  word-break: break-all;
}

.value.desc {
  font-weight: 400;
  color: #4a6fa5;
  font-size: 12px;
  line-height: 1.4;
}

/* 标签 */
.card-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 16px;
}

.card-tags :deep(.el-tag) {
  background: #ebf8ff;
  border-color: #bee3f8;
  color: #2b6cb0;
}

/* 操作按钮 */
.card-actions {
  display: flex;
  gap: 8px;
  padding-top: 16px;
  border-top: 1px solid #bee3f8;
}

.card-actions .el-button {
  flex: 1;
}

.card-actions :deep(.el-button--primary) {
  background: linear-gradient(135deg, #3182ce 0%, #2b6cb0 100%);
  border-color: #2b6cb0;
}

.card-actions :deep(.el-button--primary:hover) {
  background: linear-gradient(135deg, #2b6cb0 0%, #2c5282 100%);
}

.card-actions :deep(.el-button--success) {
  background: linear-gradient(135deg, #38a89d 0%, #2c9c8f 100%);
  border-color: #2c9c8f;
}

.card-actions :deep(.el-button--success:hover) {
  background: linear-gradient(135deg, #2c9c8f 0%, #228b7d 100%);
}

/* 抽屉样式 */
.edit-form {
  padding: 20px;
}

.schema-drawer-body {
  height: 100%;
}

/* 映射抽屉：目标工具条 + 编辑区 + 建议区 */
.mapping-drawer-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 16px 20px 24px;
}

.mapping-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  padding: 14px 18px;
  background: #fff;
  border: 1px solid #bee3f8;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(49, 130, 206, 0.08);
}

.tb-label {
  font-size: 13px;
  font-weight: 600;
  color: #4a6fa5;
  flex-shrink: 0;
}

.tb-select {
  width: 240px;
}

.tb-sep {
  color: #90cdf4;
  font-weight: 700;
  font-size: 15px;
}

.tb-opt-main {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-weight: 500;
}

.tb-opt-main code {
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
}

.tb-opt-meta {
  float: right;
  margin-left: 16px;
  font-size: 11px;
  color: #a0aec0;
  display: inline-flex;
  gap: 6px;
  align-items: center;
}

.tb-opt-meta code {
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
}

.tb-opt-meta em {
  font-style: normal;
}

.tb-opt-rec {
  color: #2c9c8f;
  font-weight: 600;
}

.fs-arrow {
  color: #90cdf4;
  font-weight: 700;
  margin: 0 2px;
}

.tb-hint {
  font-size: 12px;
  color: #b7791f;
  background: #fffaf0;
  border: 1px solid #feebc8;
  padding: 4px 10px;
  border-radius: 6px;
}

.drawer-footer {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 12px;
  padding: 16px 20px;
}

.footer-summary {
  margin-right: auto;
  font-size: 13px;
  color: #4a6fa5;
}

.footer-summary b {
  font-size: 15px;
  font-weight: 700;
  color: #2b6cb0;
  margin: 0 2px;
}

:deep(.el-drawer__header) {
  background: linear-gradient(135deg, #2b6cb0 0%, #3182ce 100%);
  color: white;
  margin-bottom: 0;
  padding: 20px 24px;
}

:deep(.el-drawer__title) {
  color: white;
  font-weight: 600;
}

:deep(.el-drawer__close-btn) {
  color: white;
}

:deep(.el-drawer__close-btn:hover) {
  color: rgba(255, 255, 255, 0.8);
}

/* 查看映射样式 */
.mappings-content {
  padding: 20px;
}

.prop-map-detail {
  padding: 8px 16px 12px 48px;
  background: #f7fbff;
}

.pm-type {
  margin-left: 8px;
}

.mappings-summary {
  display: flex;
  gap: 24px;
  margin-bottom: 24px;
  padding: 20px;
  background: linear-gradient(135deg, #f0f7ff 0%, #ebf8ff 100%);
  border-radius: 12px;
  border: 1px solid #bee3f8;
}

.summary-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.summary-value {
  font-size: 32px;
  font-weight: 700;
  color: #2b6cb0;
  line-height: 1;
}

.summary-label {
  font-size: 14px;
  color: #4a6fa5;
  margin-top: 4px;
}

.mapping-badge :deep(.el-badge__content) {
  font-size: 14px;
  padding: 4px 8px;
}
</style>
