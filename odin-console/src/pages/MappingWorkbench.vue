<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  ElTabs,
  ElTabPane,
  ElSelect,
  ElOption,
  ElTable,
  ElTableColumn,
  ElButton,
  ElTag,
  ElAlert,
  ElEmpty,
  ElInput,
  ElCheckbox,
  ElCheckboxGroup,
  ElMessage,
  ElPopconfirm,
  ElProgress,
  ElCard,
  ElDialog,
} from 'element-plus'
import { useOntologyStore } from '@/stores/ontology'
import { useMappingStore } from '@/stores/mapping'
import { useDataSourceStore } from '@/stores/datasource'
import ValueMapDialog from '@/components/mapping/ValueMapDialog.vue'
import type { PropertyMap, SuggestRow, ValidateResult } from '@/types/mapping'
import type { MappingVO } from '@/types/api'

const onto = useOntologyStore()
const maps = useMappingStore()
const ds = useDataSourceStore()

const tab = ref('by-class')
const busy = ref(false)

/** 数据源 id（后端数字主键的字符串形式）→ 可读名称 */
function sourceLabel(id: string) {
  const s = ds.sourceInfo(id)
  return s?.label || s?.code || id
}

/** ClassMap.id 即后端数字主键的字符串形式 */
function numericMapId(mapId: string): number | null {
  const n = Number(mapId)
  return Number.isFinite(n) && mapId !== '' ? n : null
}

/** 把 ClassMap 可编辑部分序列化为后端载荷（后端仅识别 property_mappings） */
function toPayload(m: { properties: PropertyMap[] }) {
  return {
    // 对齐后端 mappingPayload：property_name / column_name / value_map
    property_mappings: (m.properties ?? []).map((p) => ({
      property_name: p.propertyId,
      column_name: p.column,
      ...(p.valueMap ? { value_map: p.valueMap } : {}),
    })),
  }
}

// ---- 按类 ----
const selectedClassId = ref('')
const selectedMapId = ref<string | null>(null)

const classOptions = computed(() =>
  onto.classes.filter((c) => !c.abstract).map((c) => ({ id: c.id, label: `${c.label}（${c.id}）` })),
)

const currentClass = computed(() => onto.classes.find((c) => c.id === selectedClassId.value))

const classMapList = computed(() => maps.mapsOfClass(selectedClassId.value))

// 类列表异步到位后自动选中首个
watch(
  classOptions,
  (list) => {
    if (!list.length) {
      selectedClassId.value = ''
      return
    }
    if (!list.some((c) => c.id === selectedClassId.value)) {
      selectedClassId.value = list[0].id
    }
  },
  { immediate: true },
)

// 映射列表变化（切类 / 增删）后修正当前选中项
watch(
  classMapList,
  (list) => {
    if (!list.length) {
      selectedMapId.value = null
      return
    }
    if (!selectedMapId.value || !list.some((m) => m.id === selectedMapId.value)) {
      selectedMapId.value = list[0].id
    }
  },
  { immediate: true },
)

const currentMap = computed(() =>
  selectedMapId.value ? maps.findMap(selectedMapId.value) : undefined,
)

/** 覆盖率：以本体非抽象类为分母 */
const coverage = computed(() => {
  const total = onto.classes.filter((c) => !c.abstract).length
  const mapped = new Set(
    maps.classMaps
      .filter((m) => onto.classes.some((c) => c.id === m.classId && !c.abstract))
      .map((m) => m.classId),
  ).size
  return { total, mapped, percent: total ? Math.round((mapped / total) * 100) : 0 }
})

/** 属性行：本体属性 + 当前 map 的列映射 */
const propRows = computed(() => {
  const cls = currentClass.value
  if (!cls) return []
  const map = currentMap.value
  return cls.properties.map((p) => {
    const pm = map?.properties.find((x) => x.propertyId === p.id)
    return {
      propId: p.id,
      propLabel: p.label,
      range: p.range,
      isKey: !!p.isKey,
      sensitive: !!p.sensitive,
      column: pm?.column ?? '',
      valueMap: pm?.valueMap,
      mapped: !!pm,
    }
  })
})

const currentTableCols = computed(() => {
  if (!currentMap.value) return []
  const t = maps.tableOf(currentMap.value.sourceId, currentMap.value.table, currentMap.value.schema)
  return t?.columns ?? []
})

// ---- 跨数据源物理来源汇总（本体侧反向镜像）----
const classSourceRows = ref<MappingVO[]>([])
const classSourceLoading = ref(false)

/** 选中类后跨数据源拉取该类全部映射配置 */
async function loadClassSources() {
  const cls = onto.classes.find((c) => c.id === selectedClassId.value)
  if (!cls?.numericId) {
    classSourceRows.value = []
    return
  }
  classSourceLoading.value = true
  try {
    classSourceRows.value = await maps.queryMappings({ class_id: cls.numericId })
  } finally {
    classSourceLoading.value = false
  }
}

watch(
  () => [selectedClassId.value, onto.classes.length] as const,
  () => {
    void loadClassSources()
  },
  { immediate: true },
)

interface SourceGroupItem {
  property: string
  propertyLabel: string
  column: string
  columnType: string
  confidence: number | null
}

interface SourceGroup {
  key: string
  sourceName: string
  sourceType: string
  tableLabel: string
  items: SourceGroupItem[]
  propCount: number
  avgConfidence: number
}

/** 解析 property_mappings（数组优先，其次 JSON 字符串） */
function parsePropertyMappings(m: MappingVO): any[] {
  if (Array.isArray(m.property_mappings) && m.property_mappings.length) return m.property_mappings
  const raw = m.property_mappings_json || m.mappings_json
  try {
    const parsed = raw ? JSON.parse(raw) : []
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

function propertyLabelOf(pid: string): string {
  const cls = onto.classes.find((c) => c.id === selectedClassId.value)
  return cls?.properties.find((p) => p.id === pid)?.label || pid
}

/** 按 (datasource_name.source_table) 分组 */
const classSourceGroups = computed<SourceGroup[]>(() =>
  classSourceRows.value.map((m, i) => {
    const items = parsePropertyMappings(m).map((p: any) => {
      const pid = String(p.property_name ?? p.propertyId ?? p.property_id ?? '')
      const conf = typeof p.confidence === 'number' ? p.confidence : null
      return {
        property: pid,
        propertyLabel: propertyLabelOf(pid),
        column: String(p.column_name ?? p.column ?? ''),
        columnType: String(p.column_data_type ?? p.data_type ?? ''),
        confidence: conf != null ? (conf > 1 ? conf / 100 : conf) : null,
      } as SourceGroupItem
    })
    const sid = String(m.datasource_id)
    const confs = items.map((x) => x.confidence).filter((c): c is number => c != null)
    return {
      key: `${m.id ?? i}`,
      sourceName: m.datasource_name || sourceLabel(sid),
      sourceType: String(m.datasource_type || ds.sourceInfo(sid)?.type || '').toUpperCase(),
      tableLabel: `${m.source_schema ? `${m.source_schema}.` : ''}${m.source_table}`,
      items,
      propCount: items.length || (m.mapped_count ?? 0),
      avgConfidence: confs.length ? confs.reduce((a, b) => a + b, 0) / confs.length : 1,
    }
  }),
)

const classSourceSummary = computed(() => {
  const groups = classSourceGroups.value
  return {
    sources: new Set(groups.map((g) => g.sourceName)).size,
    tables: groups.length,
    props: new Set(groups.flatMap((g) => g.items.map((i) => i.property))).size,
    rows: groups.reduce((n, g) => n + g.items.length, 0),
  }
})

/** 属性 → 贡献它的物理源（一个属性可能来自多张表） */
const propSourceIndex = computed(() => {
  const acc = new Map<string, string[]>()
  for (const g of classSourceGroups.value) {
    const label = `${g.sourceName}.${g.tableLabel}`
    for (const it of g.items) {
      if (!it.property) continue
      const arr = acc.get(it.property)
      if (arr) {
        if (!arr.includes(label)) arr.push(label)
      } else {
        acc.set(it.property, [label])
      }
    }
  }
  return acc
})

function sourcesOfProp(pid: string): string[] {
  return propSourceIndex.value.get(pid) ?? []
}

function confText(c: number | null): string {
  return c == null ? '—' : `${Math.round(c * 100)}%`
}

function confColor(c: number | null): string {
  if (c == null) return '#94a3b8'
  if (c >= 0.8) return '#2c9c8f'
  if (c >= 0.5) return '#d69e2e'
  return '#e53e3e'
}

// value_map 对话框
const vmOpen = ref(false)
const vmPropId = ref('')
const vmPropLabel = ref('')
const vmValueMap = ref<Record<string, string>>({})

function openValueMap(propId: string, label: string, vm?: Record<string, string>) {
  if (!currentMap.value) return
  vmPropId.value = propId
  vmPropLabel.value = label
  vmValueMap.value = vm ? { ...vm } : {}
  vmOpen.value = true
}

/** 统一落库：局部修改 → updateMapping */
async function patchCurrentMap(
  patch: Partial<Pick<PropertyMapHolder, 'keyColumns' | 'properties' | 'filters' | 'primary'>>,
  okMsg?: string,
) {
  const m = currentMap.value
  if (!m) return
  const id = numericMapId(m.id)
  if (id == null) {
    ElMessage.error('映射缺少后端 ID，无法保存')
    return
  }
  busy.value = true
  try {
    await maps.updateMapping(id, toPayload({ ...m, ...patch }))
    if (okMsg) ElMessage.success(okMsg)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : String(e))
  } finally {
    busy.value = false
  }
}

interface PropertyMapHolder {
  keyColumns: string[]
  properties: PropertyMap[]
  filters?: string[]
  primary?: boolean
}

async function onSaveValueMap(vm: Record<string, string>) {
  const m = currentMap.value
  if (!m) return
  const next = m.properties.map((p) =>
    p.propertyId === vmPropId.value
      ? { ...p, valueMap: Object.keys(vm).length ? vm : undefined }
      : p,
  )
  await patchCurrentMap({ properties: next }, 'value_map 已更新')
}

async function onColumnChange(propId: string, column: string | undefined) {
  const m = currentMap.value
  if (!m) return
  const others = m.properties.filter((p) => p.propertyId !== propId)
  if (!column) {
    await patchCurrentMap({ properties: others })
    return
  }
  const existing = m.properties.find((p) => p.propertyId === propId)
  await patchCurrentMap({
    properties: [...others, { propertyId: propId, column, valueMap: existing?.valueMap }],
  })
}

async function onKeysChange(keys: Array<string | number | boolean>) {
  await patchCurrentMap({ keyColumns: keys.map(String) })
}

async function onFiltersInput(v: string) {
  await patchCurrentMap({
    filters: v
      .split(/[;；]/)
      .map((s) => s.trim())
      .filter(Boolean),
  })
}

async function removeCurrentMap() {
  const m = currentMap.value
  if (!m) return
  const id = numericMapId(m.id)
  if (id == null) {
    ElMessage.error('映射缺少后端 ID，无法删除')
    return
  }
  busy.value = true
  try {
    await maps.deleteMapping(id)
    validateResult.value = null
    ElMessage.success('映射已删除')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : String(e))
  } finally {
    busy.value = false
  }
}

// 新建映射
const newMapOpen = ref(false)
const newMapForm = reactive({
  sourceId: '',
  table: '',
})

const newMapTableOptions = computed(() => {
  const s = maps.schemaOf(newMapForm.sourceId)
  return s?.tables ?? []
})

function openNewMap() {
  newMapForm.sourceId = maps.schemas[0]?.sourceId ?? ds.sourceInfos[0]?.id ?? ''
  newMapForm.table = ''
  newMapOpen.value = true
}

async function createMap() {
  if (!newMapForm.table) {
    ElMessage.warning('请选择物理表')
    return
  }
  const cls = onto.classes.find((c) => c.id === selectedClassId.value)
  if (!cls?.numericId) {
    ElMessage.warning('请先选择有效的本体类')
    return
  }
  if (onto.currentOntologyId == null) {
    ElMessage.warning('缺少本体上下文')
    return
  }
  const t = maps.tableOf(newMapForm.sourceId, newMapForm.table)
  busy.value = true
  try {
    const res = await maps.createMapping({
      ontology_id: onto.currentOntologyId,
      class_id: cls.numericId,
      datasource_id: Number(newMapForm.sourceId),
      source_table: newMapForm.table,
      source_schema: t?.schema,
      property_mappings: [],
    })
    if (res?.id != null) selectedMapId.value = String(res.id)
    newMapOpen.value = false
    ElMessage.success('已创建映射，请配置属性列')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : String(e))
  } finally {
    busy.value = false
  }
}

// 校验
const validateResult = ref<ValidateResult | null>(null)
const validating = ref(false)

async function runValidate() {
  const m = currentMap.value
  if (!m) return
  const id = numericMapId(m.id)
  if (id == null) {
    ElMessage.error('映射缺少后端 ID，无法校验')
    return
  }
  validating.value = true
  try {
    const res = await maps.validateMapping(id)
    if (!res) {
      validateResult.value = {
        ok: false,
        issues: [{ level: 'error', code: 'validate_failed', message: '校验请求失败' }],
      }
      ElMessage.error('校验失败')
      return
    }
    const issues = (res.issues ?? []).map((i) => ({
      level: (i.level === 'error' ? 'error' : i.level === 'warning' ? 'warn' : 'ok') as
        | 'error'
        | 'warn'
        | 'ok',
      code: i.code,
      message: i.message,
    }))
    if (!issues.length) {
      issues.push({
        level: res.valid ? 'ok' : 'error',
        code: 'summary',
        message: res.valid ? '校验通过' : '校验未通过',
      })
    }
    validateResult.value = { ok: res.valid, issues }
    if (res.valid) ElMessage.success('校验通过')
    else ElMessage.warning('校验发现问题，见下方结果')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : String(e))
  } finally {
    validating.value = false
  }
}

// ---- 按源 ----
const srcSourceId = ref('')
const srcTable = ref('')
const srcClassId = ref('')
const srcKeys = ref<string[]>([])
const suggestRows = ref<SuggestRow[]>([])
const suggesting = ref(false)

const srcTableOptions = computed(() => maps.schemaOf(srcSourceId.value)?.tables ?? [])

// schemas 由 datasource store 异步派生，到位后自动选中首个源
watch(
  () => maps.schemas,
  (list) => {
    if (!list.length) return
    if (!list.some((s) => s.sourceId === srcSourceId.value)) {
      srcSourceId.value = list[0].sourceId
    }
  },
  { immediate: true },
)

// 选源后自动选中首个表
watch(
  srcTableOptions,
  (list) => {
    if (!list.length) {
      srcTable.value = ''
      return
    }
    if (!list.some((t) => t.name === srcTable.value)) {
      srcTable.value = list[0].name
    }
  },
  { immediate: true },
)

watch(
  () => [srcSourceId.value, srcTable.value],
  () => {
    const t = maps.tableOf(srcSourceId.value, srcTable.value)
    const pk = t?.columns.find((c) => c.isPk)?.name
    srcKeys.value = pk ? [pk] : []
    suggestRows.value = []
  },
  { immediate: true },
)

const srcCols = computed(() => maps.tableOf(srcSourceId.value, srcTable.value)?.columns ?? [])

async function runSuggest() {
  const cls = onto.classes.find((c) => c.id === srcClassId.value)
  if (!cls?.numericId || onto.currentOntologyId == null) {
    ElMessage.warning('请选择目标类')
    return
  }
  if (!srcSourceId.value || !srcTable.value) {
    ElMessage.warning('请选择数据源与物理表')
    return
  }
  suggesting.value = true
  try {
    const res = await maps.suggestMappings({
      ontology_id: onto.currentOntologyId,
      class_id: cls.numericId,
      datasource_id: Number(srcSourceId.value),
      table_name: srcTable.value,
    })
    suggestRows.value = (res ?? []).map((s) => ({
      column: s.column,
      type: s.column_type,
      comment: s.column_comment,
      propertyId:
        s.suggested_property_name ??
        (s.suggested_property_id != null ? String(s.suggested_property_id) : null),
      confidence: s.confidence,
      reason: s.reason ?? '',
    }))
    if (!suggestRows.value.length) ElMessage.warning('无建议')
    else ElMessage.success(`生成 ${suggestRows.value.filter((r) => r.propertyId).length} 条建议`)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : String(e))
  } finally {
    suggesting.value = false
  }
}

function acceptHighConfidence(threshold = 0.5) {
  for (const r of suggestRows.value) {
    if (r.confidence < threshold) r.propertyId = null
  }
  ElMessage.info(`已清空置信度 < ${threshold} 的建议`)
}

async function applySuggest() {
  if (!srcKeys.value.length) {
    ElMessage.warning('请先劾选业务键列')
    return
  }
  const cls = onto.classes.find((c) => c.id === srcClassId.value)
  if (!cls?.numericId || onto.currentOntologyId == null) {
    ElMessage.warning('请选择目标类')
    return
  }
  const t = maps.tableOf(srcSourceId.value, srcTable.value)
  const properties: PropertyMap[] = suggestRows.value
    .filter((r) => r.propertyId)
    .map((r) => ({ propertyId: r.propertyId!, column: r.column }))
  busy.value = true
  try {
    const existing = maps.classMaps.find(
      (m) =>
        m.classId === srcClassId.value &&
        m.sourceId === srcSourceId.value &&
        m.table === srcTable.value,
    )
    if (existing) {
      const id = numericMapId(existing.id)
      if (id == null) throw new Error('映射缺少后端 ID')
      await maps.updateMapping(id, toPayload({ properties }))
    } else {
      await maps.createMapping({
        ontology_id: onto.currentOntologyId,
        class_id: cls.numericId,
        datasource_id: Number(srcSourceId.value),
        source_table: srcTable.value,
        source_schema: t?.schema,
        property_mappings: properties.map((p) => ({
          property_name: p.propertyId,
          column_name: p.column,
          ...(p.valueMap ? { value_map: p.valueMap } : {}),
        })),
      })
    }
    ElMessage.success('已写入映射')
    // 切到按类视图
    selectedClassId.value = srcClassId.value
    tab.value = 'by-class'
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : String(e))
  } finally {
    busy.value = false
  }
}

const mappedCount = computed(() => suggestRows.value.filter((r) => r.propertyId).length)

// ---- 头部操作 ----
async function onSaveMappings() {
  maps.saveMappings()
  await maps.fetchMappings({ ontology_id: onto.currentOntologyId ?? undefined })
  ElMessage.success('映射已保存')
}

async function reloadMappings() {
  maps.dirty = false
  validateResult.value = null
  await maps.fetchMappings({ ontology_id: onto.currentOntologyId ?? undefined })
  ElMessage.info('已重新加载服务端映射')
}

onMounted(async () => {
  if (!ds.datasources.length) await ds.fetchDatasources()
  await maps.fetchMappings({ ontology_id: onto.currentOntologyId ?? undefined })
})

// 切换本体时重拉映射
watch(
  () => onto.currentOntologyId,
  async (id) => {
    if (id == null) return
    validateResult.value = null
    await maps.fetchMappings({ ontology_id: id })
  },
)
</script>

<template>
  <div class="mapping-page">
    <div class="page-head">
      <div>
        <h2>映射工作台</h2>
        <p class="sub">
          覆盖率 {{ coverage.mapped }}/{{ coverage.total }} 类 ·
          <ElProgress
            :percentage="coverage.percent"
            :stroke-width="6"
            style="width: 120px; display: inline-flex; vertical-align: middle"
          />
          <span v-if="maps.dirty" class="dirty">映射有未保存修改</span>
        </p>
      </div>
      <div class="head-actions">
        <ElButton size="small" type="primary" :disabled="!maps.dirty" @click="onSaveMappings">
          保存映射
        </ElButton>
        <ElPopconfirm title="放弃本地状态并重新加载服务端映射？" @confirm="reloadMappings">
          <template #reference>
            <ElButton size="small">重置映射</ElButton>
          </template>
        </ElPopconfirm>
      </div>
    </div>

    <ElTabs v-model="tab">
      <!-- ========== 按类对照 ========== -->
      <ElTabPane label="按类对照" name="by-class">
        <div class="toolbar-row">
          <span class="lbl">本体类</span>
          <ElSelect v-model="selectedClassId" filterable style="width: 220px" size="small">
            <ElOption v-for="c in classOptions" :key="c.id" :label="c.label" :value="c.id" />
          </ElSelect>

          <template v-if="classMapList.length">
            <span class="lbl">物理映射</span>
            <ElSelect v-model="selectedMapId" style="width: 280px" size="small">
              <ElOption
                v-for="m in classMapList"
                :key="m.id"
                :label="`${sourceLabel(m.sourceId)}.${m.schema ? m.schema + '.' : ''}${m.table}`"
                :value="m.id"
              />
            </ElSelect>
          </template>

          <ElButton size="small" type="primary" @click="openNewMap">新建映射</ElButton>
          <ElButton
            v-if="currentMap"
            size="small"
            type="danger"
            plain
            @click="removeCurrentMap"
          >
            删除当前映射
          </ElButton>
        </div>

        <!-- 跨数据源来源汇总：这个类由哪些物理源拼成 -->
        <section v-loading="classSourceLoading" class="src-summary">
          <div class="src-summary-head">
            <h3 class="sec inline">物理来源汇总</h3>
            <span class="src-summary-meta">
              {{ classSourceSummary.sources }} 个数据源 · {{ classSourceSummary.tables }} 张表 ·
              贡献 {{ classSourceSummary.props }} 个属性（{{ classSourceSummary.rows }} 条映射）
            </span>
            <ElButton
              size="small"
              text
              type="primary"
              :loading="classSourceLoading"
              @click="loadClassSources"
            >
              刷新
            </ElButton>
          </div>

          <ElEmpty
            v-if="!classSourceLoading && !classSourceGroups.length"
            description="该类暂无跨数据源映射"
            :image-size="56"
          />

          <div v-else class="src-groups">
            <article v-for="g in classSourceGroups" :key="g.key" class="src-group">
              <header class="src-group-head">
                <code class="src-group-title">{{ g.sourceName }}.{{ g.tableLabel }}</code>
                <ElTag v-if="g.sourceType" size="small" effect="plain" type="primary">
                  {{ g.sourceType }}
                </ElTag>
                <span class="src-group-count">{{ g.propCount }} 个属性</span>
                <span class="src-group-conf" :style="{ color: confColor(g.avgConfidence) }">
                  平均置信度 {{ confText(g.avgConfidence) }}
                </span>
              </header>
              <ul v-if="g.items.length" class="src-prop-list">
                <li v-for="(it, idx) in g.items" :key="`${g.key}_${idx}`">
                  <span class="sp-prop">
                    {{ it.propertyLabel }}
                    <code v-if="it.propertyLabel !== it.property">{{ it.property }}</code>
                  </span>
                  <span class="sp-arrow">←</span>
                  <code class="sp-col">{{ it.column || '—' }}</code>
                  <ElTag v-if="it.columnType" size="small" type="info" effect="plain">
                    {{ it.columnType }}
                  </ElTag>
                  <span class="sp-conf" :style="{ color: confColor(it.confidence) }">
                    {{ confText(it.confidence) }}
                  </span>
                </li>
              </ul>
              <p v-else class="src-group-empty">该配置尚未写入属性映射</p>
            </article>
          </div>
        </section>

        <ElEmpty
          v-if="!currentMap"
          description="该类还没有物理映射，点「新建映射」或到「按源浏览」生成建议"
        >
          <ElButton type="primary" size="small" @click="openNewMap">新建映射</ElButton>
        </ElEmpty>

        <template v-else>
          <ElCard shadow="never" class="map-card">
            <div class="map-meta">
              <ElTag type="success">{{ sourceLabel(currentMap.sourceId) }}</ElTag>
              <code>{{ currentMap.schema }}.{{ currentMap.table }}</code>
              <span v-if="currentMap.primary" class="meta-tag">主映射</span>
              <ElButton size="small" :loading="validating" @click="runValidate">校验映射</ElButton>
            </div>

            <div class="key-row">
              <span class="lbl">业务键 Subject</span>
              <ElCheckboxGroup :model-value="currentMap.keyColumns" @change="onKeysChange">
                <ElCheckbox v-for="c in currentTableCols" :key="c.name" :value="c.name">
                  {{ c.name }}
                </ElCheckbox>
              </ElCheckboxGroup>
            </div>

            <div class="filter-row">
              <span class="lbl">过滤条件</span>
              <ElInput
                size="small"
                style="max-width: 360px"
                :model-value="(currentMap.filters ?? []).join('; ')"
                placeholder="如 deleted_at IS NULL"
                @change="onFiltersInput"
              />
            </div>
          </ElCard>

          <h3 class="sec">属性对照</h3>
          <ElTable :data="propRows" size="small" border>
            <ElTableColumn prop="propLabel" label="本体属性" min-width="110">
              <template #default="{ row }">
                <span>{{ row.propLabel }}</span>
                <ElTag v-if="row.isKey" size="small" type="warning" style="margin-left: 4px">
                  键
                </ElTag>
                <ElTag v-if="row.sensitive" size="small" type="danger" style="margin-left: 4px">
                  敏
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="range" label="类型" width="80" />
            <ElTableColumn label="物理列" min-width="150">
              <template #default="{ row }">
                <ElSelect
                  :model-value="row.column || undefined"
                  clearable
                  filterable
                  size="small"
                  placeholder="选择列"
                  style="width: 100%"
                  @change="(v: string | undefined) => onColumnChange(row.propId, v)"
                >
                  <ElOption
                    v-for="c in currentTableCols"
                    :key="c.name"
                    :label="`${c.name} (${c.type})`"
                    :value="c.name"
                  />
                </ElSelect>
              </template>
            </ElTableColumn>
            <ElTableColumn label="value_map" min-width="120">
              <template #default="{ row }">
                <ElButton
                  v-if="row.range === 'enum' || row.valueMap"
                  size="small"
                  text
                  type="primary"
                  @click="openValueMap(row.propId, row.propLabel, row.valueMap)"
                >
                  {{
                    row.valueMap
                      ? `${Object.keys(row.valueMap).length} 组`
                      : '配置'
                  }}
                </ElButton>
                <span v-else class="muted">—</span>
              </template>
            </ElTableColumn>
            <ElTableColumn label="状态" width="80">
              <template #default="{ row }">
                <ElTag :type="row.mapped ? 'success' : 'info'" size="small">
                  {{ row.mapped ? '已映射' : '未映射' }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn label="来源" min-width="170">
              <template #default="{ row }">
                <template v-if="sourcesOfProp(row.propId).length">
                  <ElTag
                    v-for="s in sourcesOfProp(row.propId)"
                    :key="s"
                    size="small"
                    effect="plain"
                    type="success"
                    class="src-chip"
                  >
                    {{ s }}
                  </ElTag>
                </template>
                <span v-else class="muted">—</span>
              </template>
            </ElTableColumn>
          </ElTable>

          <template v-if="validateResult">
            <h3 class="sec">校验结果</h3>
            <div class="validate-list">
              <ElAlert
                v-for="(iss, i) in validateResult.issues"
                :key="i"
                :type="iss.level === 'error' ? 'error' : iss.level === 'warn' ? 'warning' : 'success'"
                :title="iss.message"
                :closable="false"
                style="margin-bottom: 6px"
              />
            </div>
          </template>
        </template>
      </ElTabPane>

      <!-- ========== 按源浏览 ========== -->
      <ElTabPane label="按源浏览 / 建议" name="by-source">
        <div class="toolbar-row wrap">
          <span class="lbl">数据源</span>
          <ElSelect v-model="srcSourceId" style="width: 140px" size="small">
            <ElOption v-for="s in maps.schemas" :key="s.sourceId" :label="sourceLabel(s.sourceId)" :value="s.sourceId" />
          </ElSelect>
          <span class="lbl">物理表</span>
          <ElSelect v-model="srcTable" filterable style="width: 180px" size="small">
            <ElOption
              v-for="t in srcTableOptions"
              :key="t.name"
              :label="t.schema ? `${t.schema}.${t.name}` : t.name"
              :value="t.name"
            />
          </ElSelect>
          <span class="lbl">目标类</span>
          <ElSelect v-model="srcClassId" filterable style="width: 200px" size="small">
            <ElOption v-for="c in classOptions" :key="c.id" :label="c.label" :value="c.id" />
          </ElSelect>
          <ElButton size="small" type="primary" :loading="suggesting" @click="runSuggest">生成建议</ElButton>
          <ElButton v-if="suggestRows.length" size="small" @click="acceptHighConfidence(0.5)">
            保留 ≥0.5
          </ElButton>
          <ElButton
            v-if="suggestRows.length"
            size="small"
            type="success"
            @click="applySuggest"
          >
            应用到映射（{{ mappedCount }}）
          </ElButton>
        </div>

        <div class="toolbar-row">
          <span class="lbl">业务键</span>
          <ElCheckboxGroup v-model="srcKeys">
            <ElCheckbox v-for="c in srcCols" :key="c.name" :value="c.name">
              {{ c.name }}
              <ElTag v-if="c.isPk" size="small" type="warning" style="margin-left: 2px">PK</ElTag>
            </ElCheckbox>
          </ElCheckboxGroup>
        </div>

        <ElAlert
          v-if="suggestRows.length"
          type="info"
          :closable="false"
          :title="`建议 ${mappedCount} / ${suggestRows.length} 列；可改「映射到」下拉，无匹配选「忽略」`"
          style="margin-bottom: 10px"
        />

        <ElEmpty v-if="!suggestRows.length" description="选择源和表后点「生成建议」" />

        <ElTable v-else :data="suggestRows" size="small" border>
          <ElTableColumn prop="column" label="物理列" min-width="120">
            <template #default="{ row }">
              <code>{{ row.column }}</code>
              <div class="col-type">{{ row.type }}</div>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="comment" label="注释" min-width="100" show-overflow-tooltip />
          <ElTableColumn label="映射到本体属性" min-width="200">
            <template #default="{ row }">
              <ElSelect
                v-model="row.propertyId"
                clearable
                filterable
                size="small"
                placeholder="忽略"
                style="width: 100%"
              >
                <ElOption
                  v-for="p in currentClass?.id === srcClassId
                    ? currentClass.properties
                    : onto.classes.find((c) => c.id === srcClassId)?.properties ?? []"
                  :key="p.id"
                  :label="`${p.label}（${p.id}）`"
                  :value="p.id"
                />
              </ElSelect>
            </template>
          </ElTableColumn>
          <ElTableColumn label="置信度" width="120">
            <template #default="{ row }">
              <ElProgress
                :percentage="Math.round(row.confidence * 100)"
                :stroke-width="8"
                :status="row.confidence >= 0.7 ? 'success' : row.confidence >= 0.4 ? 'warning' : 'exception'"
              />
            </template>
          </ElTableColumn>
          <ElTableColumn prop="reason" label="依据" min-width="140" show-overflow-tooltip />
        </ElTable>
      </ElTabPane>
    </ElTabs>

    <!-- 新建映射简易对话框 -->
    <ElDialog v-model="newMapOpen" title="新建物理映射" width="420px">
      <div class="form-line">
        <span class="lbl">数据源</span>
        <ElSelect v-model="newMapForm.sourceId" style="width: 160px">
          <ElOption v-for="s in maps.schemas" :key="s.sourceId" :label="sourceLabel(s.sourceId)" :value="s.sourceId" />
        </ElSelect>
      </div>
      <div class="form-line">
        <span class="lbl">物理表</span>
        <ElSelect v-model="newMapForm.table" filterable style="width: 220px">
          <ElOption
            v-for="t in newMapTableOptions"
            :key="t.name"
            :label="t.schema ? `${t.schema}.${t.name}` : t.name"
            :value="t.name"
          />
        </ElSelect>
      </div>
      <template #footer>
        <ElButton @click="newMapOpen = false">取消</ElButton>
        <ElButton type="primary" :loading="busy" @click="createMap">创建</ElButton>
      </template>
    </ElDialog>

    <ValueMapDialog
      v-model:open="vmOpen"
      v-model:model-value="vmValueMap"
      :property-label="vmPropLabel"
      @save="onSaveValueMap"
    />
  </div>
</template>

<style scoped>
.mapping-page {
  height: 100%;
  overflow: auto;
  padding: 16px 20px 32px;
  background: #f8fafc;
}

.page-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
}

.page-head h2 {
  margin: 0 0 4px;
  font-size: 18px;
}

.sub {
  margin: 0;
  font-size: 12px;
  color: #64748b;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.dirty {
  color: #b45309;
}

.head-actions {
  display: flex;
  gap: 8px;
}

.toolbar-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: nowrap;
}

.toolbar-row.wrap {
  flex-wrap: wrap;
}

.lbl {
  font-size: 12px;
  color: #64748b;
  flex-shrink: 0;
}

.map-card {
  margin-bottom: 12px;
}

.map-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.meta-tag {
  font-size: 11px;
  color: #2563eb;
  background: #eff6ff;
  padding: 2px 6px;
  border-radius: 4px;
}

.key-row,
.filter-row {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-top: 10px;
}

.sec {
  margin: 16px 0 8px;
  font-size: 13px;
  color: #374151;
  border-left: 3px solid #4c8bf5;
  padding-left: 8px;
}

.muted {
  color: #9ca3af;
  font-size: 12px;
}

.col-type {
  font-size: 11px;
  color: #94a3b8;
}

.form-line {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

/* 跨数据源物理来源汇总 */
.src-summary {
  margin: 4px 0 14px;
  padding: 12px 14px 14px;
  background: #fff;
  border: 1px solid #bee3f8;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(49, 130, 206, 0.08);
}

.src-summary-head {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  margin-bottom: 8px;
}

.sec.inline {
  margin: 0;
}

.src-summary-meta {
  font-size: 12px;
  color: #4a6fa5;
}

.src-summary-head .el-button {
  margin-left: auto;
}

.src-groups {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.src-group {
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: linear-gradient(135deg, #f7fbff 0%, #f0f7ff 100%);
  overflow: hidden;
}

.src-group-head {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  padding: 8px 12px;
  background: linear-gradient(135deg, #ebf8ff 0%, #e6f2fd 100%);
  border-bottom: 1px solid #bee3f8;
}

.src-group-title {
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
  font-weight: 700;
  color: #1a365d;
}

.src-group-count {
  font-size: 12px;
  color: #4a6fa5;
}

.src-group-conf {
  margin-left: auto;
  font-size: 12px;
  font-weight: 700;
}

.src-prop-list {
  list-style: none;
  margin: 0;
  padding: 8px 12px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.src-prop-list li {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  font-size: 12px;
  color: #1a365d;
}

.sp-prop {
  font-weight: 600;
  min-width: 120px;
}

.sp-prop code {
  margin-left: 4px;
  font-size: 11px;
  color: #7b9cc4;
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
}

.sp-arrow {
  color: #3182ce;
  font-weight: 700;
}

.sp-col {
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  background: #fff;
  border: 1px solid #bee3f8;
  border-radius: 4px;
  padding: 1px 6px;
  color: #2b6cb0;
}

.sp-conf {
  margin-left: auto;
  font-weight: 700;
}

.src-group-empty {
  margin: 0;
  padding: 8px 12px;
  font-size: 12px;
  color: #94a3b8;
}

.src-chip {
  margin: 2px 4px 2px 0;
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
}
</style>
