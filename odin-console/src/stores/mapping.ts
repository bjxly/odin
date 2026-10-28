import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as mapApi from '@/api/mapping'
import type { MappingVO, MappingSuggestionVO, MappingValidationVO, PageResult } from '@/types/api'
import type {
  ClassMap,
  PhysTable,
  PropertyMap,
  SourceSchema,
  SuggestRow,
  ValidateResult,
} from '@/types/mapping'
import { useDataSourceStore } from './datasource'
import { registerClassMapsProvider, resolveClassName } from './classRegistry'

export const useMappingStore = defineStore('mapping', () => {
  // --- API-backed state ---
  const mappings = ref<MappingVO[]>([])
  const currentMapping = ref<MappingVO | null>(null)
  const suggestions = ref<MappingSuggestionVO[]>([])
  const validationResult = ref<MappingValidationVO | null>(null)
  const loading = ref(false)

  // --- Legacy compat state ---
  /** 页面本地写入的 schema（内省结果以 datasource store 为主） */
  const localSchemas = ref<SourceSchema[]>([])
  const dirty = ref(false)
  const classMaps = computed<ClassMap[]>(() =>
    mappings.value.map(m => {
      let properties: PropertyMap[] = []
      // 优先使用后端返回的 property_mappings 数组，其次 property_mappings_json / 旧 mappings_json
      const rawList: any[] = Array.isArray(m.property_mappings) && m.property_mappings.length
        ? m.property_mappings
        : []
      if (rawList.length) {
        properties = rawList.map((p: any) => ({
          propertyId: p.property_name ?? (p.property_id != null ? String(p.property_id) : ''),
          column: p.column_name ?? p.column ?? '',
          valueMap: p.value_map ?? p.valueMap,
          // 保留后端存储的置信度 / 校验状态，供编辑器与汇总视图展示
          confidence: typeof p.confidence === 'number' ? p.confidence : null,
          validationStatus: p.validation_status,
        })).filter(p => p.propertyId || p.column)
      } else {
        const raw = m.property_mappings_json || m.mappings_json
        try {
          if (raw) {
            const parsed = JSON.parse(raw)
            if (Array.isArray(parsed)) {
              properties = parsed.map((p: any) => ({
                propertyId: p.property_name ?? p.propertyId ?? p.property_id ?? '',
                column: p.column_name ?? p.column ?? '',
                valueMap: p.value_map ?? p.valueMap,
                confidence: typeof p.confidence === 'number' ? p.confidence : null,
                validationStatus: p.validation_status,
              })).filter((p: PropertyMap) => p.propertyId || p.column)
            }
          }
        } catch { /* ignore */ }
      }
      let keyColumns: string[] = []
      try {
        if (m.key_columns) keyColumns = JSON.parse(m.key_columns)
      } catch { /* ignore */ }
      let filters: string[] = []
      try {
        if (m.filters) filters = JSON.parse(m.filters)
      } catch { /* ignore */ }
      return {
        id: String(m.id),
        classId: resolveClassName(m.class_id),
        sourceId: String(m.datasource_id),
        table: m.source_table,
        schema: m.source_schema,
        keyColumns,
        properties,
        filters,
        primary: m.is_primary,
        // 后端附带的可读信息，供本体侧汇总视图直接使用
        datasourceName: m.datasource_name,
        datasourceType: m.datasource_type,
        className: m.class_name,
      }
    })
  )

  const coverage = computed(() => {
    const mapped = new Set(classMaps.value.map(m => m.classId)).size
    return { total: mapped, mapped, percent: mapped ? 100 : 0 }
  })

  // 向 ontology store 提供映射快照（避免直接 import 形成循环依赖）
  registerClassMapsProvider(() =>
    classMaps.value.map(m => ({ classId: m.classId, sourceId: m.sourceId, primary: m.primary })),
  )

  /** 物理 Schema：以 datasource store 的内省结果为主，叠加本地覆盖 */
  const schemas = computed<SourceSchema[]>(() => {
    const out: SourceSchema[] = []
    try {
      const ds = useDataSourceStore()
      for (const d of ds.datasources) {
        const s = ds.schemas[d.id]
        out.push({
          sourceId: String(d.id),
          tables: (s?.tables ?? []).map(t => ({
            name: t.name,
            schema: t.schema,
            comment: t.comment,
            columns: (t.columns ?? []).map(c => ({
              name: c.name,
              type: c.data_type ?? c.type ?? '',
              nullable: c.is_nullable ?? c.nullable,
              comment: c.comment,
              isPk: c.is_primary_key,
            })),
          })),
        })
      }
    } catch { /* store 尚未就绪 */ }
    for (const l of localSchemas.value) {
      const i = out.findIndex(x => x.sourceId === l.sourceId)
      if (i >= 0) out[i] = l
      else out.push(l)
    }
    return out
  })

  // --- API Actions ---
  async function fetchMappings(params?: { ontology_id?: number; class_id?: number; datasource_id?: number }) {
    loading.value = true
    try {
      const res = await mapApi.getMappings(params) as PageResult<MappingVO> | MappingVO[]
      if (Array.isArray(res)) {
        mappings.value = res
      } else {
        mappings.value = res.records ?? []
      }
    } catch (e) {
      console.error('[mapping] fetchMappings failed', e)
    } finally {
      loading.value = false
    }
  }

  /** 按条件查询映射（不覆盖全局 state），供“查看映射”抽屉使用 */
  async function queryMappings(params?: { ontology_id?: number; class_id?: number; datasource_id?: number }): Promise<MappingVO[]> {
    try {
      const res = await mapApi.getMappings(params) as PageResult<MappingVO> | MappingVO[]
      return Array.isArray(res) ? res : (res.records ?? [])
    } catch (e) {
      console.error('[mapping] queryMappings failed', e)
      return []
    }
  }

  async function fetchMapping(id: number) {
    loading.value = true
    try {
      currentMapping.value = await mapApi.getMapping(id) as MappingVO
    } catch (e) {
      console.error('[mapping] fetchMapping failed', e)
    } finally {
      loading.value = false
    }
  }

  async function createMapping(data: any) {
    const res = await mapApi.createMapping(data) as MappingVO
    mappings.value.push(res)
    return res
  }

  async function updateMapping(id: number, data: any) {
    const res = await mapApi.updateMapping(id, data) as MappingVO
    const idx = mappings.value.findIndex(m => m.id === id)
    if (idx >= 0) mappings.value[idx] = res
    return res
  }

  async function deleteMapping(id: number) {
    await mapApi.deleteMapping(id)
    mappings.value = mappings.value.filter(m => m.id !== id)
  }

  async function validateMapping(id: number) {
    try {
      validationResult.value = await mapApi.validateMapping(id) as MappingValidationVO
      return validationResult.value
    } catch (e) {
      console.error('[mapping] validateMapping failed', e)
      return null
    }
  }

  async function suggestMappingsAction(params: { ontology_id: number; class_id: number; datasource_id: number; table_name: string }) {
    try {
      const res = await mapApi.suggestMappings(params) as any
      // 后端返回 { suggestions: [...] } 对象（也可能直接为数组），字段名为 column_name/property_name 等，需归一化
      const rawList: any[] = Array.isArray(res) ? res : (Array.isArray(res?.suggestions) ? res.suggestions : [])
      suggestions.value = rawList.map((s: any): MappingSuggestionVO => ({
        column: s.column ?? s.column_name ?? '',
        column_type: s.column_type ?? s.column_data_type ?? s.data_type ?? '',
        column_comment: s.column_comment,
        suggested_property_id: s.suggested_property_id ?? s.property_id,
        suggested_property_name: s.suggested_property_name ?? s.property_name,
        suggested_property_label: s.suggested_property_label ?? s.property_label,
        confidence: s.confidence ?? 0,
        reason: s.reason ?? '',
      }))
      return suggestions.value
    } catch (e) {
      console.error('[mapping] suggestMappings failed', e)
      return []
    }
  }

  // --- Legacy compatibility methods ---
  function reloadFromWorkspace() {
    // No-op in API mode - schemas come from datasource store
  }

  function persistSchemas() {
    // No-op
  }

  function markDirty() {
    dirty.value = true
  }

  function resetToMock() {
    dirty.value = false
  }

  function saveMappings() {
    dirty.value = false
  }

  function ensureSchema(sourceId: string) {
    if (!localSchemas.value.some(s => s.sourceId === sourceId)) {
      localSchemas.value = [...localSchemas.value, { sourceId, tables: [] }]
    }
  }

  function removeSchema(sourceId: string) {
    localSchemas.value = localSchemas.value.filter(s => s.sourceId !== sourceId)
  }

  function setSchemaTables(sourceId: string, tables: PhysTable[]) {
    ensureSchema(sourceId)
    localSchemas.value = localSchemas.value.map(s =>
      s.sourceId === sourceId ? { ...s, tables } : s
    )
  }

  function introspectSource(_sourceId: string, _opts?: { defaultSchema?: string }): PhysTable[] {
    return []
  }

  function removeMapsOfSource(sourceId: string) {
    mappings.value = mappings.value.filter(m => String(m.datasource_id) !== sourceId)
  }

  function schemaOf(sourceId: string): SourceSchema | undefined {
    return schemas.value.find(s => s.sourceId === sourceId)
  }

  function tableOf(sourceId: string, table: string, schema?: string) {
    const s = schemaOf(sourceId)
    return s?.tables.find(t => t.name === table && (!schema || t.schema === schema))
  }

  function mapsOfClass(classId: string): ClassMap[] {
    return classMaps.value.filter(m => m.classId === classId)
  }

  function mapsOfSource(sourceId: string): ClassMap[] {
    return classMaps.value.filter(m => m.sourceId === sourceId)
  }

  function findMap(id: string): ClassMap | undefined {
    return classMaps.value.find(m => m.id === id)
  }

  function syncOntologyMappedSources(_classId: string) {
    // No-op in API mode
  }

  function upsertClassMap(input: {
    id?: string; classId: string; sourceId: string; table: string; schema?: string
    keyColumns: string[]; properties: PropertyMap[]; filters?: string[]; primary?: boolean
  }) {
    // Legacy local-only upsert
    const id = input.id ?? `${input.classId}_${input.sourceId}_${input.table}`.toLowerCase()
    const map: ClassMap = {
      id, classId: input.classId, sourceId: input.sourceId,
      table: input.table, schema: input.schema,
      keyColumns: input.keyColumns, properties: input.properties,
      filters: input.filters, primary: input.primary,
    }
    dirty.value = true
    return map
  }

  function removeClassMap(_id: string) {
    dirty.value = true
  }

  function setPropertyMaps(_mapId: string, _properties: PropertyMap[]) {
    dirty.value = true
  }

  function setKeyColumns(_mapId: string, _keyColumns: string[]) {
    dirty.value = true
  }

  function setFilters(_mapId: string, _filters: string[]) {
    dirty.value = true
  }

  function setValueMap(_mapId: string, _propertyId: string, _valueMap: Record<string, string> | undefined) {
    dirty.value = true
  }

  function suggestMappingsLegacy(_classId: string, _sourceId: string, _table: string, _schema?: string): SuggestRow[] {
    return []
  }

  function applySuggestions(
    classId: string, sourceId: string, table: string, schema: string | undefined,
    rows: SuggestRow[], keyColumns: string[]
  ) {
    const properties: PropertyMap[] = rows
      .filter(r => r.propertyId)
      .map(r => ({ propertyId: r.propertyId!, column: r.column }))
    return upsertClassMap({ classId, sourceId, table, schema, keyColumns, properties })
  }

  function validateClassMap(mapId: string): ValidateResult {
    const m = findMap(mapId)
    if (!m) {
      return { ok: false, issues: [{ level: 'error', code: 'map_not_found', message: '映射不存在' }] }
    }
    return { ok: true, issues: [{ level: 'ok', code: 'pass', message: '校验通过' }] }
  }

  return {
    // API state
    mappings, currentMapping, suggestions, validationResult, loading,
    // Legacy compat
    schemas, localSchemas, dirty, classMaps, coverage,
    // API actions
    fetchMappings, fetchMapping, createMapping, updateMapping, deleteMapping,
    validateMapping, suggestMappings: suggestMappingsAction, queryMappings,
    // Legacy actions
    reloadFromWorkspace, persistSchemas, markDirty, resetToMock, saveMappings,
    ensureSchema, removeSchema, setSchemaTables, introspectSource, removeMapsOfSource,
    schemaOf, tableOf, mapsOfClass, mapsOfSource, findMap,
    upsertClassMap, removeClassMap, setPropertyMaps, setKeyColumns, setFilters, setValueMap,
    suggestMappingsLocal: suggestMappingsLegacy,
    applySuggestions, validateClassMap, syncOntologyMappedSources,
  }
})
