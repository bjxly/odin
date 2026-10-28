import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import * as ontApi from '@/api/ontology'
import type { OntologyVO, OntClassVO, OntPropertyVO, OntRelationVO, OntRuleVO, PageResult } from '@/types/api'
import type {
  LayoutPos,
  OntoClass,
  OntoProperty,
  OntoRelation,
  OntoRule,
  SourceInfo,
  ChangeLogItem,
} from '@/types/ontology'
import { useDataSourceStore } from './datasource'
import { getClassMapRefs, registerClassIdResolver } from './classRegistry'

export const useOntologyStore = defineStore('ontology', () => {
  const dsStore = useDataSourceStore()

  // --- API-backed state ---
  const ontologies = ref<OntologyVO[]>([])
  const currentOntologyId = ref<number | null>(null)
  const currentOntology = ref<OntologyVO | null>(null)
  const classRows = ref<OntoClass[]>([])
  const relations = ref<OntoRelation[]>([])
  const rules = ref<OntoRule[]>([])
  const loading = ref(false)

  // --- UI state (not persisted) ---
  const focusId = ref<string | null>(null)
  const search = ref('')
  const showUnmappedOnly = ref(false)
  const activeDomain = ref<string | null>(null)
  const dirty = ref(false)
  const changeLog = ref<ChangeLogItem[]>([])

  // --- 布局状态（持久化到 ont_definition.layout_json）---
  /** 类名 → 画布坐标 */
  const layoutPositions = ref<Record<string, LayoutPos>>({})
  /** 坐标是否有未保存变更（拖动 / 自动布局） */
  const layoutDirty = ref(false)
  /** 布局变更在 changeLog 中固定占一条，避免顶部未保存计数被拖动次数放大 */
  const LAYOUT_LOG_ID = '__layout__'

  // 向 mapping store 提供「数字 class_id → 类名」解析（避免循环依赖）
  registerClassIdResolver((numericId) => classRows.value.find(c => c.numericId === numericId)?.id)

  /** 类名 → 已映射的数据源 id 列表（来自 mapping store 快照） */
  const mappedSourcesByClass = computed(() => {
    const acc = new Map<string, string[]>()
    for (const m of getClassMapRefs()) {
      if (!m.classId) continue
      const arr = acc.get(m.classId)
      if (arr) {
        if (!arr.includes(m.sourceId)) arr.push(m.sourceId)
      } else {
        acc.set(m.classId, [m.sourceId])
      }
    }
    return acc
  })

  /** 对外暴露的类列表：合并映射信息（只读视图） */
  const classes = computed<OntoClass[]>(() =>
    classRows.value.map(c => {
      const ms = mappedSourcesByClass.value.get(c.id)
      if (!ms || !ms.length) return c
      return c.mappedSources.length === ms.length
        && ms.every((s, i) => s === c.mappedSources[i])
        ? c
        : { ...c, mappedSources: ms }
    }),
  )

  // --- Legacy compat computed ---
  const version = computed(() => currentOntology.value?.version ?? '1.0.0')
  /** 数据源列表：委托给 datasource store */
  const sources = computed<SourceInfo[]>(() => dsStore.sourceInfos)
  /** 类名 → 坐标（图谱渲染与导出共用） */
  const layout = computed<Record<string, LayoutPos>>(() => layoutPositions.value)
  const activeOntologyId = computed(() => currentOntologyId.value != null ? String(currentOntologyId.value) : '')
  const activeOntologyName = computed(() => currentOntology.value?.name ?? '未选择')
  const ontologyRegistry = computed(() =>
    ontologies.value.map(o => ({
      id: String(o.id),
      name: o.name,
      description: o.description,
      version: o.version ?? '1.0.0',
      classCount: classRows.value.filter(c => !c.abstract).length,
      totalClasses: classRows.value.length,
      relationCount: relations.value.length,
      mapCount: mappedSourcesByClass.value.size,
    }))
  )

  const domains = computed(() => {
    const set = new Set<string>()
    for (const c of classRows.value) {
      if (c.domain) set.add(c.domain)
    }
    return [...set]
  })

  const focusedClass = computed(() =>
    classes.value.find(c => c.id === focusId.value) ?? null
  )

  const filteredClasses = computed(() => {
    const q = search.value.trim().toLowerCase()
    return classes.value.filter(c => {
      if (showUnmappedOnly.value && c.mappedSources.length > 0) return false
      if (activeDomain.value && c.domain !== activeDomain.value) return false
      if (!q) return true
      const hay = [
        c.id, c.label, c.description ?? '',
        ...(c.synonyms ?? []),
        ...c.properties.map(p => p.label),
      ].join(' ').toLowerCase()
      return hay.includes(q)
    })
  })

  const filteredClassIds = computed(() => new Set(filteredClasses.value.map(c => c.id)))

  const visibleRelations = computed(() => {
    const ids = filteredClassIds.value
    return relations.value.filter(r => ids.has(r.domain) && ids.has(r.range))
  })

  // --- Helpers ---
  /** 后端 OntClassProperty → 前端 OntoProperty（data_type→range、required→isKey） */
  function mapPropertyFromVO(p: any): OntoProperty {
    return {
      id: p.name || String(p.id),
      label: p.label || p.name || String(p.id),
      range: p.data_type || p.range || 'string',
      isKey: p.required ?? p.is_key ?? false,
      sensitive: p.sensitive,
      synonyms: p.synonyms,
    }
  }

  function mapClassFromVO(vo: OntClassVO): OntoClass {
    // 继承链：优先用后端 parent_class_name，其次由 parent_class_id 经 classRows 解析
    const subclassOf = vo.parent_class_name
      || (vo.parent_class_id != null
        ? classRows.value.find(c => c.numericId === vo.parent_class_id)?.id
        : undefined)
    // 设计期默认坐标：仅当后端写了非零 position_x/position_y 时才采用
    const px = Number(vo.position_x)
    const py = Number(vo.position_y)
    const position = Number.isFinite(px) && Number.isFinite(py) && (px !== 0 || py !== 0)
      ? { x: Math.round(px), y: Math.round(py) }
      : undefined
    return {
      id: vo.name || String(vo.id),
      numericId: vo.id,
      label: vo.label || vo.name,
      description: vo.description,
      // 后端 OntClass 无 domain 字段，改按 class_type 分组（保证领域过滤/图渲染可用）
      domain: vo.class_type === 'virtual' ? '虚拟类' : '实体类',
      synonyms: [],
      subclassOf,
      // 虚拟类：class_type === 'virtual'
      abstract: vo.class_type === 'virtual',
      properties: [],
      mappedSources: [],
      position,
    }
  }

  function mapRelationFromVO(vo: OntRelationVO): OntoRelation {
    // 两端类名：优先用后端附带的 from/to_class_name，其次由 from/to_class_id 经 classRows 解析
    const domain = vo.from_class_name
      || classRows.value.find(c => c.numericId === vo.from_class_id)?.id
      || (vo.from_class_id != null ? String(vo.from_class_id) : '')
    const range = vo.to_class_name
      || classRows.value.find(c => c.numericId === vo.to_class_id)?.id
      || (vo.to_class_id != null ? String(vo.to_class_id) : '')
    return {
      id: vo.name || String(vo.id),
      numericId: vo.id,
      label: vo.label || vo.name,
      domain,
      range,
      cardinality: vo.cardinality || 'N:1',
      inverse: undefined,
      crossSource: false,
    }
  }

  function mapRuleFromVO(vo: OntRuleVO): OntoRule {
    // 后端规则为 ontology 级，无 class_id；classId 置空，rulesOf 返回本体全部规则
    return {
      id: vo.name || String(vo.id),
      numericId: vo.id,
      label: vo.name,
      classId: '',
      description: vo.description || '',
      condition: vo.condition_json || '',
      action: vo.action_json || '',
      enabled: vo.enabled ?? true,
    }
  }

  // --- API Actions ---
  async function fetchOntologies() {
    loading.value = true
    try {
      const res = await ontApi.getOntologies() as PageResult<OntologyVO> | OntologyVO[]
      if (Array.isArray(res)) {
        ontologies.value = res
      } else {
        ontologies.value = res.records ?? []
      }
    } catch (e) {
      console.error('[ontology] fetchOntologies failed', e)
    } finally {
      loading.value = false
    }
  }

  async function fetchOntology(id: number) {
    loading.value = true
    try {
      currentOntology.value = await ontApi.getOntology(id) as OntologyVO
      currentOntologyId.value = id
      // 读回后端保存的坐标（仅在加载时同步，避免覆盖未保存的拖动）
      syncLayoutFromOntology()
      // Fetch classes and relations for this ontology
      await Promise.all([fetchClasses(id), fetchRelations(id), fetchRules(id)])
    } catch (e) {
      console.error('[ontology] fetchOntology failed', e)
    } finally {
      loading.value = false
    }
  }

  async function createOntologyAction(data: any) {
    const res = await ontApi.createOntology(data) as OntologyVO
    ontologies.value.push(res)
    return res
  }

  async function updateOntologyAction(id: number, data: any) {
    const res = await ontApi.updateOntology(id, data) as OntologyVO
    const idx = ontologies.value.findIndex(o => o.id === id)
    if (idx >= 0) ontologies.value[idx] = res
    if (currentOntologyId.value === id) currentOntology.value = res
    return res
  }

  async function deleteOntologyAction(id: number) {
    await ontApi.deleteOntology(id)
    ontologies.value = ontologies.value.filter(o => o.id !== id)
    if (currentOntologyId.value === id) {
      currentOntologyId.value = null
      currentOntology.value = null
      classRows.value = []
      relations.value = []
      rules.value = []
    }
  }

  // --- Class CRUD ---
  async function fetchClasses(ontologyId: number) {
    try {
      const res = await ontApi.getClasses(ontologyId) as OntClassVO[]
      const rows = Array.isArray(res) ? res.map(mapClassFromVO) : []
      // 先写入类列表（供 subclassOf / 关系解析使用）
      classRows.value = rows
      // 后端 classView 仅给 property_count，需逐类拉取属性明细
      const withProps = await Promise.all(rows.map(async (cls) => {
        if (cls.numericId == null) return cls
        try {
          const props = await ontApi.getProperties(ontologyId, cls.numericId) as any[]
          return { ...cls, properties: Array.isArray(props) ? props.map(mapPropertyFromVO) : [] }
        } catch { return cls }
      }))
      classRows.value = withProps
      // 用类自带的 position_x/position_y 补齐 layout_json 中缺失的坐标
      mergeClassDefaultPositions()
    } catch (e) {
      console.error('[ontology] fetchClasses failed', e)
    }
  }

  async function createClass(ontologyId: number, data: any) {
    const res = await ontApi.createClass(ontologyId, data) as OntClassVO
    classRows.value.push(mapClassFromVO(res))
    return res
  }

  async function updateClassAction(ontologyId: number, classId: number, data: any) {
    const res = await ontApi.updateClass(ontologyId, classId, data) as OntClassVO
    const mapped = mapClassFromVO(res)
    const idx = classRows.value.findIndex(c => c.numericId === classId || c.id === mapped.id)
    if (idx >= 0) classRows.value[idx] = mapped
    else classRows.value.push(mapped)
    return res
  }

  async function deleteClassAction(ontologyId: number, classId: number) {
    await ontApi.deleteClass(ontologyId, classId)
    classRows.value = classRows.value.filter(c => c.numericId !== classId && c.id !== String(classId))
  }

  // --- Property CRUD ---
  async function fetchProperties(ontologyId: number, classId: number) {
    try {
      return await ontApi.getProperties(ontologyId, classId) as OntPropertyVO[]
    } catch (e) {
      console.error('[ontology] fetchProperties failed', e)
      return []
    }
  }

  async function createProperty(ontologyId: number, classId: number, data: any) {
    return await ontApi.createProperty(ontologyId, classId, data) as OntPropertyVO
  }

  async function updateProperty(ontologyId: number, classId: number, propId: number, data: any) {
    return await ontApi.updateProperty(ontologyId, classId, propId, data) as OntPropertyVO
  }

  async function deleteProperty(ontologyId: number, classId: number, propId: number) {
    return await ontApi.deleteProperty(ontologyId, classId, propId)
  }

  // --- Relation CRUD ---
  async function fetchRelations(ontologyId: number) {
    try {
      const res = await ontApi.getRelations(ontologyId) as any
      // 后端返回 { relations: OntRelation[], items: [{ relation, from_class_name, to_class_name }] }
      // 优先用 items（已解析两端类名），其次 relations，最后兼容裸数组
      const items: any[] = Array.isArray(res?.items) ? res.items
        : Array.isArray(res?.relations) ? res.relations.map((r: any) => ({ relation: r }))
        : Array.isArray(res) ? res.map((r: any) => ({ relation: r }))
        : []
      relations.value = items.map((it: any) => {
        const rel = it?.relation ?? it
        return mapRelationFromVO({
          ...rel,
          from_class_name: it?.from_class_name,
          to_class_name: it?.to_class_name,
        } as OntRelationVO)
      })
    } catch (e) {
      console.error('[ontology] fetchRelations failed', e)
    }
  }

  async function createRelation(ontologyId: number, data: any) {
    const res = await ontApi.createRelation(ontologyId, data) as OntRelationVO
    relations.value.push(mapRelationFromVO(res))
    return res
  }

  async function updateRelation(ontologyId: number, relId: number, data: any) {
    const res = await ontApi.updateRelation(ontologyId, relId, data) as OntRelationVO
    const mapped = mapRelationFromVO(res)
    const idx = relations.value.findIndex(r => r.id === mapped.id)
    if (idx >= 0) relations.value[idx] = mapped
    return res
  }

  async function deleteRelation(ontologyId: number, relId: number) {
    await ontApi.deleteRelation(ontologyId, relId)
    relations.value = relations.value.filter(r => r.numericId !== relId && r.id !== String(relId))
  }

  // --- Rule CRUD ---
  async function fetchRules(ontologyId: number) {
    try {
      const res = await ontApi.getRules(ontologyId) as OntRuleVO[]
      rules.value = Array.isArray(res) ? res.map(mapRuleFromVO) : []
    } catch (e) {
      console.error('[ontology] fetchRules failed', e)
    }
  }

  async function createRule(ontologyId: number, data: any) {
    const res = await ontApi.createRule(ontologyId, data) as OntRuleVO
    rules.value.push(mapRuleFromVO(res))
    return res
  }

  async function updateRule(ontologyId: number, ruleId: number, data: any) {
    const res = await ontApi.updateRule(ontologyId, ruleId, data) as OntRuleVO
    const mapped = mapRuleFromVO(res)
    const idx = rules.value.findIndex(r => r.id === mapped.id)
    if (idx >= 0) rules.value[idx] = mapped
    return res
  }

  async function deleteRule(ontologyId: number, ruleId: number) {
    await ontApi.deleteRule(ontologyId, ruleId)
    rules.value = rules.value.filter(r => r.numericId !== ruleId && r.id !== String(ruleId))
  }

  // --- Legacy compatibility methods ---
  function setCurrentOntology(id: number) {
    currentOntologyId.value = id
  }

  function switchOntology(id: string) {
    const numId = Number(id)
    currentOntologyId.value = numId
    focusId.value = null
    search.value = ''
    showUnmappedOnly.value = false
    activeDomain.value = null
    dirty.value = false
    layoutDirty.value = false
    changeLog.value = []
    // Fetch full ontology data
    fetchOntology(numId)
  }

  function focus(id: string | null) { focusId.value = id }
  function setSearch(v: string) { search.value = v }
  function toggleUnmapped() { showUnmappedOnly.value = !showUnmappedOnly.value }
  function setDomain(d: string | null) { activeDomain.value = d }
  function classExists(id: string) { return classRows.value.some(c => c.id === id) }

  // --- 布局：解析 / 应用 / 持久化 ---
  /**
   * 解析后端 layout_json，兼容三种历史形态：
   *   1. { version, positions: { 类名: {x,y} } }（当前写入格式）
   *   2. { 类名: {x,y} }（扁平映射）
   *   3. [{ id, x, y }] / [{ id, position: [x,y] }]（数组）
   */
  function parseLayoutJson(raw?: string | null): Record<string, LayoutPos> {
    if (!raw) return {}
    let parsed: any
    try {
      parsed = JSON.parse(raw)
    } catch {
      return {}
    }
    const src = parsed && typeof parsed === 'object' && parsed.positions && typeof parsed.positions === 'object'
      ? parsed.positions
      : parsed
    const out: Record<string, LayoutPos> = {}
    const put = (id: unknown, x: unknown, y: unknown) => {
      const key = typeof id === 'string' ? id : String(id ?? '')
      const nx = Number(x)
      const ny = Number(y)
      if (key && Number.isFinite(nx) && Number.isFinite(ny)) out[key] = { x: Math.round(nx), y: Math.round(ny) }
    }
    if (Array.isArray(src)) {
      for (const it of src) put(it?.id ?? it?.class ?? it?.name, it?.x ?? it?.position?.[0], it?.y ?? it?.position?.[1])
    } else if (src && typeof src === 'object') {
      for (const [k, v] of Object.entries(src as Record<string, any>)) put(k, v?.x ?? v?.position?.[0], v?.y ?? v?.position?.[1])
    }
    return out
  }

  /** 序列化为后端存储结构 */
  function serializeLayout(): string {
    return JSON.stringify({
      version: 1,
      positions: layoutPositions.value,
      updated_at: new Date().toISOString(),
    })
  }

  /** 从当前本体读取坐标（仅加载 / 切换本体时调用） */
  function syncLayoutFromOntology() {
    layoutPositions.value = parseLayoutJson(currentOntology.value?.layout_json)
    layoutDirty.value = false
  }

  /**
   * 用后端类自带的 position_x/position_y 补齐缺失坐标：
   * 优先级 layout_json（用户保存）> ont_class.position_*（设计期）> 前端自动布局。
   * 不标记未保存，因为这些坐标本身已持久化在类记录上。
   */
  function mergeClassDefaultPositions() {
    const patch: Record<string, LayoutPos> = {}
    for (const c of classRows.value) {
      if (c.position && !layoutPositions.value[c.id]) patch[c.id] = { ...c.position }
    }
    if (Object.keys(patch).length) layoutPositions.value = { ...layoutPositions.value, ...patch }
  }

  /** 记录一条布局未保存变更 */
  function markLayoutDirty(summary: string, target = 'graph') {
    layoutDirty.value = true
    const item: ChangeLogItem = { id: LAYOUT_LOG_ID, kind: 'update_layout', target, summary, at: Date.now() }
    const idx = changeLog.value.findIndex(i => i.id === LAYOUT_LOG_ID)
    if (idx >= 0) {
      const arr = [...changeLog.value]
      arr[idx] = item
      changeLog.value = arr
    } else {
      changeLog.value = [...changeLog.value, item]
    }
  }

  /** 批量应用坐标；mark=false 用于加载时静默补齐（不置未保存） */
  function applyLayout(positions: Record<string, LayoutPos>, summary = '自动布局', mark = true) {
    if (!positions || !Object.keys(positions).length) return
    layoutPositions.value = { ...layoutPositions.value, ...positions }
    if (mark) markLayoutDirty(summary)
  }

  /** 把当前坐标写回后端 layout_json 持久化 */
  async function saveLayout(): Promise<boolean> {
    const id = currentOntologyId.value
    if (id == null || !currentOntology.value) return false
    const json = serializeLayout()
    try {
      await ontApi.updateOntology(id, { layout_json: json })
      currentOntology.value.layout_json = json
      layoutDirty.value = false
      changeLog.value = changeLog.value.filter(i => i.id !== LAYOUT_LOG_ID)
      return true
    } catch (e) {
      console.error('[ontology] saveLayout failed', e)
      return false
    }
  }

  /** 拖动节点：更新单个类坐标并标记未保存 */
  function setLayoutPos(classId: string, x: number, y: number) {
    const nx = Math.round(x)
    const ny = Math.round(y)
    const prev = layoutPositions.value[classId]
    if (prev && prev.x === nx && prev.y === ny) return
    layoutPositions.value = { ...layoutPositions.value, [classId]: { x: nx, y: ny } }
    const label = classRows.value.find(c => c.id === classId)?.label || classId
    markLayoutDirty(`移动节点：${label}`, classId)
  }

  function mappedSourcesOf(classId: string): string[] {
    return mappedSourcesByClass.value.get(classId) ?? []
  }

  function rulesOf(_classId: string): OntoRule[] {
    // 后端规则归属 ontology 级，无 class 维度；返回当前本体全部规则
    return rules.value
  }

  function relationsOf(classId: string): OntoRelation[] {
    return relations.value.filter(r => r.domain === classId || r.range === classId)
  }

  function sourceInfo(id: string): SourceInfo | undefined {
    return dsStore.sourceInfo(id)
  }

  // Legacy CRUD stubs that work locally
  function addClass(input: {
    id: string; label: string; description?: string; domain?: string
    synonyms?: string[]; subclassOf?: string; abstract?: boolean; properties?: OntoProperty[]
  }) {
    const id = input.id.trim()
    if (!id) throw new Error('类 ID 不能为空')
    if (classExists(id)) throw new Error(`类 ID 已存在：${id}`)
    const cls: OntoClass = {
      id,
      label: input.label.trim() || id,
      description: input.description,
      domain: input.domain,
      synonyms: input.synonyms,
      subclassOf: input.subclassOf,
      abstract: input.abstract,
      properties: input.properties ?? [],
      mappedSources: [],
    }
    classRows.value = [...classRows.value, cls]
    dirty.value = true
    focus(id)
    return cls
  }

  function updateClass(id: string, patch: Partial<Omit<OntoClass, 'id' | 'properties'>>) {
    const idx = classRows.value.findIndex(c => c.id === id)
    if (idx < 0) return
    const arr = [...classRows.value]
    arr[idx] = { ...arr[idx], ...patch }
    classRows.value = arr
    dirty.value = true
  }

  function deleteClass(id: string) {
    const cls = classRows.value.find(c => c.id === id)
    if (!cls) return
    classRows.value = classRows.value.filter(c => c.id !== id)
    relations.value = relations.value.filter(r => r.domain !== id && r.range !== id)
    if (focusId.value === id) focusId.value = null
    dirty.value = true
  }

  function addProperty(classId: string, prop: OntoProperty) {
    const cls = classRows.value.find(c => c.id === classId)
    if (!cls) throw new Error(`类不存在：${classId}`)
    const pid = prop.id.trim() || `${classId}.${prop.label}`
    if (cls.properties.some(p => p.id === pid)) throw new Error(`属性 ID 已存在：${pid}`)
    const next: OntoProperty = {
      id: pid, label: prop.label.trim() || pid, range: prop.range || 'string',
      isKey: prop.isKey, sensitive: prop.sensitive, synonyms: prop.synonyms,
    }
    cls.properties = [...cls.properties, next]
    dirty.value = true
    return next
  }

  function updatePropertyLocal(classId: string, propId: string, patch: Partial<OntoProperty>) {
    const cls = classRows.value.find(c => c.id === classId)
    if (!cls) return
    const idx = cls.properties.findIndex(p => p.id === propId)
    if (idx < 0) return
    const arr = [...cls.properties]
    arr[idx] = { ...arr[idx], ...patch, id: propId }
    cls.properties = arr
    dirty.value = true
  }

  function deletePropertyLocal(classId: string, propId: string) {
    const cls = classRows.value.find(c => c.id === classId)
    if (!cls) return
    cls.properties = cls.properties.filter(p => p.id !== propId)
    dirty.value = true
  }

  function addRelation(input: {
    domain: string; range: string; label: string
    cardinality: string; inverse?: string; crossSource?: boolean
  }) {
    if (!classExists(input.domain)) throw new Error(`起点类不存在：${input.domain}`)
    if (!classExists(input.range)) throw new Error(`终点类不存在：${input.range}`)
    const id = `${input.domain}.${input.label || 'rel'}.${input.range}`.replace(/\s+/g, '')
    if (relations.value.some(r => r.id === id)) throw new Error(`关系已存在：${id}`)
    const rel: OntoRelation = {
      id, label: input.label.trim() || `${input.domain}→${input.range}`,
      domain: input.domain, range: input.range,
      cardinality: input.cardinality || 'N:1',
      inverse: input.inverse, crossSource: input.crossSource,
    }
    relations.value = [...relations.value, rel]
    dirty.value = true
    return rel
  }

  function updateRelationLocal(id: string, patch: Partial<Omit<OntoRelation, 'id' | 'domain' | 'range'>>) {
    const idx = relations.value.findIndex(r => r.id === id)
    if (idx < 0) return
    const arr = [...relations.value]
    arr[idx] = { ...arr[idx], ...patch }
    relations.value = arr
    dirty.value = true
  }

  function deleteRelationLocal(id: string) {
    const rel = relations.value.find(r => r.id === id)
    if (!rel) return
    relations.value = relations.value.filter(r => r.id !== id)
    dirty.value = true
  }

  /** 保存：先持久化布局坐标（若有未保存变更），再清空未保存标记 */
  async function save(): Promise<string> {
    if (layoutDirty.value) {
      const ok = await saveLayout()
      if (!ok) throw new Error('布局坐标保存失败，请重试')
    }
    dirty.value = false
    changeLog.value = []
    // 版本号由后端管理，前端不再递增
    return version.value
  }

  function resetToMock() {
    dirty.value = false
    layoutDirty.value = false
    changeLog.value = []
    focusId.value = null
  }

  function persistSilently() {
    // 布局改为显式 saveLayout()；保留空实现以兼容旧调用
  }

  // Legacy source stubs (sources are in datasource store now)
  function sourceExists(id: string) { return dsStore.sourceInfos.some(s => s.id === id) }
  function addSource(input: any): any { return dsStore.addSource(input) }
  function updateSource(id: string, patch: any) { dsStore.updateSource(id, patch) }
  function deleteSource(id: string) { dsStore.deleteSource(id) }
  function testSource(id: string) {
    void dsStore.testConnection(id)
    return dsStore.sourceInfo(id)
      ?? { id, type: 'postgres' as const, status: 'unknown' as const, testMessage: '' }
  }
  function markIntrospected(id: string, _tableCount: number) {
    void dsStore.introspectSchema(id)
  }

  return {
    // API state
    ontologies, currentOntologyId, currentOntology,
    classes, relations, rules, loading,
    // Legacy compat state
    version, sources, layout, layoutPositions, layoutDirty,
    activeOntologyId, activeOntologyName, ontologyRegistry,
    domains, focusId, focusedClass, search, showUnmappedOnly, activeDomain,
    filteredClasses, visibleRelations, dirty, changeLog,
    // API actions
    fetchOntologies, fetchOntology,
    createOntology: createOntologyAction,
    updateOntology: updateOntologyAction,
    deleteOntology: deleteOntologyAction,
    fetchClasses, createClass, updateClassApi: updateClassAction, deleteClassApi: deleteClassAction,
    fetchProperties, createProperty, updatePropertyApi: updateProperty, deletePropertyApi: deleteProperty,
    fetchRelations, createRelation, updateRelationApi: updateRelation, deleteRelationApi: deleteRelation,
    fetchRules, createRule, updateRuleApi: updateRule, deleteRuleApi: deleteRule,
    // Legacy actions
    setCurrentOntology, switchOntology,
    focus, setSearch, toggleUnmapped, setDomain,
    classExists, setLayoutPos, applyLayout, markLayoutDirty, saveLayout,
    syncLayoutFromOntology, mergeClassDefaultPositions,
    mappedSourcesOf, rulesOf, relationsOf, sourceInfo,
    addClass, updateClass, deleteClass,
    addProperty, updateProperty: updatePropertyLocal, deleteProperty: deletePropertyLocal,
    addRelation, updateRelation: updateRelationLocal, deleteRelation: deleteRelationLocal,
    sourceExists, addSource, updateSource, deleteSource, testSource, markIntrospected,
    save, resetToMock, persistSilently,
  }
})

export type OntologyClass = OntoClass
