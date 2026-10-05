import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as dsApi from '@/api/datasource'
import type {
  DataSourceVO,
  DataSourceSchemaVO,
  PageResult,
  SchemaRawDiffVO,
  SchemaRefreshResultVO,
} from '@/types/api'
import type { SourceInfo, SourceType } from '@/types/ontology'
import { normalizeSchemaDiff } from '@/utils/schemaDiff'
import { useMappingStore } from './mapping'

// Re-export legacy types for backward compatibility with pages
export interface DataSourceInfo {
  id: string
  code?: string
  name: string
  type: DataSourceType
  host?: string
  port?: number
  database?: string
  user?: string
  defaultSchema?: string
  readOnly: boolean
  status: 'connected' | 'disconnected' | 'testing' | 'error'
  tableCount: number
  mappedClassCount: number
  lastTested?: string
  lastIntrospected?: string
  latencyMs?: number
  tags?: string[]
  description?: string
}

export type DataSourceType =
  | 'postgresql'
  | 'mysql'
  | 'dameng'
  | 'kingbase'
  | 'mongodb'
  | 'redis'
  | 'influxdb'
  | 'csv'
  | 'parquet'
  | 's3'
  | 'clickhouse'
  | 'sqlite'
  | 'oracle'
  | 'sqlserver'

export interface SchemaInfo {
  sourceId: string
  tables: TableInfo[]
  introspectedAt: number
}

export interface TableInfo {
  name: string
  schema?: string
  comment?: string
  columns: ColumnInfo[]
}

export interface ColumnInfo {
  name: string
  type: string
  nullable?: boolean
  comment?: string
  isPk?: boolean
  isFk?: boolean
}

export interface DataSourceForm {
  id: string
  name: string
  type: DataSourceType
  host: string
  port: number
  database: string
  user: string
  password: string
  defaultSchema: string
  readOnly: boolean
  tags: string[]
  description: string
}

// 数据源类型默认端口
export const DEFAULT_PORTS: Record<DataSourceType, number> = {
  postgresql: 5432,
  mysql: 3306,
  dameng: 5236,
  kingbase: 54321,
  mongodb: 27017,
  redis: 6379,
  influxdb: 8086,
  csv: 0,
  parquet: 0,
  s3: 0,
  clickhouse: 9000,
  sqlite: 0,
  oracle: 1521,
  sqlserver: 1433,
}

// 数据源类型标签
export const TYPE_LABELS: Record<DataSourceType, string> = {
  postgresql: 'PostgreSQL',
  mysql: 'MySQL',
  dameng: '达梦 DM',
  kingbase: '人大金仓',
  mongodb: 'MongoDB',
  redis: 'Redis',
  influxdb: 'InfluxDB',
  csv: 'CSV 文件',
  parquet: 'Parquet 文件',
  s3: 'S3/MinIO',
  clickhouse: 'ClickHouse',
  sqlite: 'SQLite',
  oracle: 'Oracle',
  sqlserver: 'SQL Server',
}

/** 后端类型 → 页面使用的 SourceType（types/ontology） */
const API_TYPE_TO_LEGACY: Record<string, SourceType> = {
  postgresql: 'postgres',
  postgres: 'postgres',
  mysql: 'mysql',
  dameng: 'dameng',
  kingbase: 'kingbase',
  opengauss: 'opengauss',
  gaussdb: 'gaussdb',
  oracle: 'oracle',
  sqlite: 'sqlite',
  sqlserver: 'sqlserver',
  clickhouse: 'clickhouse',
}

export function toLegacySourceType(t?: string): SourceType {
  return API_TYPE_TO_LEGACY[(t || '').toLowerCase()] ?? 'postgres'
}

/** 页面 SourceType → 后端类型 */
export function toApiSourceType(t?: string): string {
  const v = (t || '').toLowerCase()
  return v === 'postgres' ? 'postgresql' : v || 'postgresql'
}

export function parseTags(raw?: string): string[] {
  if (!raw) return []
  let arr: unknown = raw
  if (typeof raw === 'string') {
    const s = raw.trim()
    if (!s) return []
    if (s.startsWith('[')) {
      try { arr = JSON.parse(s) } catch { arr = s.split(/[,，]/) }
    } else {
      arr = s.split(/[,，]/)
    }
  }
  if (!Array.isArray(arr)) return []
  return arr.map(x => String(x).trim()).filter(Boolean)
}

/** 测试连接归一化结果 */
export interface TestResult {
  status: 'ok' | 'warn' | 'down'
  latencyMs: number
  testMessage: string
}

/** 本地记录的测试 / 内省元数据（后端 VO 未返回这些字段） */
interface SourceMeta {
  lastTestAt?: number
  lastIntrospectAt?: number
  latencyMs?: number
  testMessage?: string
}

export const useDataSourceStore = defineStore('datasource', () => {
  // --- API-backed state ---
  const datasources = ref<DataSourceVO[]>([])
  const currentDatasource = ref<DataSourceVO | null>(null)
  const schemas = ref<Record<number, DataSourceSchemaVO>>({})
  const loading = ref(false)
  const total = ref(0)
  /** 本地测试 / 内省元数据，key = String(datasource.id) */
  const testMeta = ref<Record<string, SourceMeta>>({})

  function recordMeta(id: string, patch: SourceMeta) {
    testMeta.value = { ...testMeta.value, [id]: { ...(testMeta.value[id] ?? {}), ...patch } }
  }

  /** 每个源已映射的本体类数（来自 mapping store） */
  const mappedClassCountBySource = computed<Record<string, number>>(() => {
    const acc: Record<string, number> = {}
    try {
      const ms = useMappingStore()
      const seen = new Map<string, Set<string>>()
      for (const m of ms.classMaps) {
        if (!seen.has(m.sourceId)) seen.set(m.sourceId, new Set())
        seen.get(m.sourceId)!.add(m.classId)
      }
      for (const [sid, set] of seen) acc[sid] = set.size
    } catch { /* store 尚未就绪 */ }
    return acc
  })

  // --- Legacy compatibility (mapped from API data) ---
  const sources = computed<DataSourceInfo[]>(() =>
    datasources.value.map(ds => {
      const meta = testMeta.value[String(ds.id)] ?? {}
      return {
        id: String(ds.id),
        code: ds.code,
        name: ds.name,
        type: (ds.type as DataSourceType) || 'postgresql',
        host: ds.host,
        port: ds.port,
        database: ds.database_name ?? ds.database,
        user: ds.username,
        defaultSchema: ds.schema_name,
        readOnly: true,
        status: ds.status === 'active' ? 'connected' : ds.status === 'testing' ? 'testing' : ds.status === 'error' ? 'error' : 'disconnected',
        tableCount: ds.table_count ?? schemas.value[ds.id]?.tables?.length ?? 0,
        mappedClassCount: mappedClassCountBySource.value[String(ds.id)] ?? 0,
        lastTested: meta.lastTestAt ? new Date(meta.lastTestAt).toISOString() : undefined,
        lastIntrospected: meta.lastIntrospectAt ? new Date(meta.lastIntrospectAt).toISOString() : undefined,
        latencyMs: meta.latencyMs,
        tags: parseTags(ds.tags),
        description: ds.description,
      }
    })
  )

  /** SourceManager / OntologyGraph 侧栏使用的 SourceInfo 形状 */
  const sourceInfos = computed<SourceInfo[]>(() =>
    datasources.value.map(ds => {
      const meta = testMeta.value[String(ds.id)] ?? {}
      const status: SourceInfo['status'] =
        ds.status === 'active' ? 'ok'
          : ds.status === 'testing' ? 'warn'
            : ds.status === 'error' ? 'down'
              : meta.lastTestAt ? 'ok' : 'unknown'
      return {
        id: String(ds.id),
        code: ds.code,
        label: ds.name,
        type: toLegacySourceType(ds.type),
        status,
        host: ds.host,
        port: ds.port,
        database: ds.database_name ?? ds.database,
        user: ds.username,
        defaultSchema: ds.schema_name,
        readOnly: true,
        tags: parseTags(ds.tags),
        description: ds.description,
        tables: ds.table_count ?? schemas.value[ds.id]?.tables?.length ?? 0,
        lastTestAt: meta.lastTestAt,
        lastIntrospectAt: meta.lastIntrospectAt,
        latencyMs: meta.latencyMs,
        testMessage: meta.testMessage,
      }
    })
  )

  const connectedSources = computed(() =>
    sources.value.filter(s => s.status === 'connected')
  )

  const sourceById = computed(() => {
    const map = new Map<string, DataSourceInfo>()
    sources.value.forEach(s => map.set(s.id, s))
    return (id: string) => map.get(id)
  })

  function sourceInfo(id: string): SourceInfo | undefined {
    return sourceInfos.value.find(s => s.id === id)
  }

  // --- Actions ---
  async function fetchDatasources(params?: { current?: number; size?: number; type?: string; status?: string }) {
    loading.value = true
    try {
      const res = await dsApi.getDatasources(params) as PageResult<DataSourceVO> | DataSourceVO[]
      if (Array.isArray(res)) {
        datasources.value = res
        total.value = res.length
      } else {
        datasources.value = res.records ?? []
        total.value = res.total ?? 0
      }
    } catch (e) {
      console.error('[datasource] fetchDatasources failed', e)
    } finally {
      loading.value = false
    }
  }

  async function fetchDatasource(id: number) {
    loading.value = true
    try {
      currentDatasource.value = await dsApi.getDatasource(id) as DataSourceVO
    } catch (e) {
      console.error('[datasource] fetchDatasource failed', e)
    } finally {
      loading.value = false
    }
  }

  async function createDatasource(data: any) {
    const res = await dsApi.createDatasource(data) as DataSourceVO
    datasources.value.push(res)
    return res
  }

  async function updateDatasource(id: number, data: any) {
    const res = await dsApi.updateDatasource(id, data) as DataSourceVO
    const idx = datasources.value.findIndex(d => d.id === id)
    if (idx >= 0) datasources.value[idx] = res
    return res
  }

  async function deleteDatasource(id: number) {
    await dsApi.deleteDatasource(id)
    datasources.value = datasources.value.filter(d => d.id !== id)
    delete schemas.value[id]
    const next = { ...testMeta.value }
    delete next[String(id)]
    testMeta.value = next
  }

  function normalizeStatus(raw: unknown): 'ok' | 'warn' | 'down' {
    const v = String(raw ?? '').toLowerCase()
    if (['ok', 'success', 'active', 'connected', 'true'].includes(v)) return 'ok'
    if (['warn', 'warning', 'testing'].includes(v)) return 'warn'
    if (['down', 'error', 'fail', 'failed', 'false', 'inactive'].includes(v)) return 'down'
    return 'ok'
  }

  async function testDatasource(id: number): Promise<TestResult> {
    const key = String(id)
    const started = Date.now()
    let res: any = null
    try {
      res = await dsApi.testDatasource(id)
    } catch (e: any) {
      const latencyMs = Date.now() - started
      const testMessage = e?.message || '连接失败'
      recordMeta(key, { lastTestAt: Date.now(), latencyMs, testMessage })
      const errIdx = datasources.value.findIndex(d => d.id === id)
      if (errIdx >= 0) datasources.value[errIdx] = { ...datasources.value[errIdx], status: 'error' }
      return { status: 'down', latencyMs, testMessage }
    }
    const latencyMs = Number(res?.latency_ms ?? res?.latencyMs) || (Date.now() - started)
    const status = normalizeStatus(res?.status ?? res?.result ?? (res?.success === false ? 'down' : 'ok'))
    const testMessage = String(
      res?.message ?? res?.testMessage ?? res?.error_message
      ?? (status === 'ok' ? `连接成功（${latencyMs} ms）` : status === 'warn' ? '连接成功，但存在告警' : '连接失败'),
    )
    recordMeta(key, { lastTestAt: Date.now(), latencyMs, testMessage })
    const idx = datasources.value.findIndex(d => d.id === id)
    if (idx >= 0) {
      datasources.value[idx] = {
        ...datasources.value[idx],
        status: status === 'down' ? 'error' : 'active',
      }
    }
    return { status, latencyMs, testMessage }
  }

  async function fetchSchema(id: number) {
    loading.value = true
    try {
      const res = await dsApi.getDatasourceSchema(id) as DataSourceSchemaVO
      schemas.value[id] = res
      recordMeta(String(id), { lastIntrospectAt: Date.now() })
      return res
    } catch (e) {
      console.error('[datasource] fetchSchema failed', e)
      return null
    } finally {
      loading.value = false
    }
  }

  /**
   * 差量刷新 Schema。
   *
   * 后端 `POST /datasources/:id/schema/refresh` 返回的是**嵌套结构**
   * `data = { schema, diff }`（axios 拦截器已剥掉 code/message 外壳）：
   *   - `schema`：刷新后的完整内省结果（含 tables / table_count / column_count / stats）；
   *   - `diff`：`{ added_tables, removed_tables, added_columns, removed_columns, changed_columns }`，
   *     其中列级差异为 `{table, column}` / `{table, column, old_type, new_type}` **对象数组**。
   *
   * 这里做两件事：把 `schema` 写入本地缓存（使 SchemaBrowser / 映射表下拉立即生效），
   * 并把 `diff` 归一化成前端友好的 `SchemaDiffVO`（含计数）后随返回值交给页面。
   */
  async function refreshSchema(id: number): Promise<DataSourceSchemaVO | null> {
    loading.value = true
    try {
      const raw = (await dsApi.refreshSchema(id)) as
        | Partial<SchemaRefreshResultVO>
        | Partial<DataSourceSchemaVO>
        | null
      // 主路径：嵌套 { schema, diff }；兼容早期把 schema 字段铺平在顶层的返回。
      // 两路都要求 tables 是数组，避免畸形响应用空 schema 覆盖掉已有缓存。
      const nestedRaw = (raw as SchemaRefreshResultVO | null)?.schema ?? null
      const nested = nestedRaw && Array.isArray(nestedRaw.tables) ? nestedRaw : null
      const flatSchema = Array.isArray((raw as DataSourceSchemaVO | null)?.tables)
        ? (raw as DataSourceSchemaVO)
        : null
      const schema = nested ?? flatSchema
      if (!schema) {
        console.warn('[datasource] refreshSchema: 响应中未找到 schema.tables', raw)
        return null
      }

      const tables = schema.tables
      const columnTotal = tables.reduce((n, t) => n + (t.columns?.length ?? 0), 0)
      const refreshed: DataSourceSchemaVO = {
        ...schema,
        tables,
        table_count: schema.table_count ?? schema.stats?.tables ?? tables.length,
        column_count: schema.column_count ?? schema.stats?.columns ?? columnTotal,
      }
      // 更新本地 Schema 缓存（刷新拿到的新结构必须入库，否则界面仍是旧快照）
      schemas.value[id] = refreshed
      recordMeta(String(id), { lastIntrospectAt: Date.now() })
      // 同步刷新数据源列表，以更新卡片上的 table_count
      await fetchDatasources()

      // 归一化 diff 字段名与形态（对象数组 → 展示串 + 计数）供页面使用
      const rawDiff = ((raw as SchemaRefreshResultVO | null)?.diff ??
        (schema as { diff?: SchemaRawDiffVO }).diff) as SchemaRawDiffVO | null | undefined
      const diff = normalizeSchemaDiff(rawDiff)
      return { ...refreshed, diff }
    } catch (e) {
      console.error('[datasource] refreshSchema failed', e)
      return null
    } finally {
      loading.value = false
    }
  }

  // --- Legacy compatibility methods ---
  function addSource(form: DataSourceForm): DataSourceInfo {
    // Legacy sync method - creates a local entry
    const source: DataSourceInfo = {
      id: form.id || `source_${Date.now()}`,
      name: form.name,
      type: form.type,
      host: form.host || undefined,
      port: form.port || undefined,
      database: form.database || undefined,
      user: form.user || undefined,
      defaultSchema: form.defaultSchema || undefined,
      readOnly: form.readOnly,
      status: 'disconnected',
      tableCount: 0,
      mappedClassCount: 0,
      tags: form.tags,
      description: form.description,
    }
    return source
  }

  function updateSource(_id: string, _updates: Partial<DataSourceForm>) {
    // Legacy stub - use updateDatasource instead
  }

  function deleteSource(_id: string) {
    // Legacy stub - use deleteDatasource instead
  }

  async function testConnection(id: string): Promise<{ success: boolean; message: string }> {
    try {
      const r = await testDatasource(Number(id))
      return { success: r.status !== 'down', message: r.testMessage }
    } catch (e: any) {
      return { success: false, message: e?.message || '连接失败' }
    }
  }

  async function introspectSchema(id: string): Promise<SchemaInfo | null> {
    const res = await fetchSchema(Number(id))
    if (!res) return null
    return {
      sourceId: id,
      tables: res.tables?.map(t => {
        // 后端 schemaColumnView 无 is_foreign_key，改从表的 foreign_keys 匹配列名推导
        const fkCols = new Set(
          ((t.foreign_keys as any[]) ?? []).map(fk => fk?.column).filter(Boolean) as string[],
        )
        return {
          name: t.name,
          schema: t.schema,
          comment: t.comment,
          columns: t.columns?.map(c => ({
            name: c.name,
            type: c.data_type ?? c.type ?? '',
            nullable: c.is_nullable ?? c.nullable,
            comment: c.comment,
            isPk: c.is_primary_key,
            isFk: fkCols.has(c.name),
          })) ?? [],
        }
      }) ?? [],
      introspectedAt: Date.now(),
    }
  }

  function getSchema(sourceId: string): SchemaInfo | undefined {
    const numId = Number(sourceId)
    const apiSchema = schemas.value[numId]
    if (!apiSchema) return undefined
    return {
      sourceId,
      tables: apiSchema.tables?.map(t => {
        // 后端 schemaColumnView 无 is_foreign_key，改从表的 foreign_keys 匹配列名推导
        const fkCols = new Set(
          ((t.foreign_keys as any[]) ?? []).map(fk => fk?.column).filter(Boolean) as string[],
        )
        return {
          name: t.name,
          schema: t.schema,
          comment: t.comment,
          columns: t.columns?.map(c => ({
            name: c.name,
            type: c.data_type ?? c.type ?? '',
            nullable: c.is_nullable ?? c.nullable,
            comment: c.comment,
            isPk: c.is_primary_key,
            isFk: fkCols.has(c.name),
          })) ?? [],
        }
      }) ?? [],
      introspectedAt: Date.now(),
    }
  }

  function resetToMock() {
    // No-op in API mode
  }

  return {
    // API state
    datasources,
    currentDatasource,
    schemas,
    loading,
    total,
    testMeta,
    // Legacy compat
    sources,
    sourceInfos,
    connectedSources,
    sourceById,
    sourceInfo,
    mappedClassCountBySource,
    // API actions
    fetchDatasources,
    fetchDatasource,
    createDatasource,
    updateDatasource,
    deleteDatasource,
    testDatasource,
    fetchSchema,
    refreshSchema,
    // Legacy actions
    addSource,
    updateSource,
    deleteSource,
    testConnection,
    introspectSchema,
    getSchema,
    resetToMock,
  }
})
