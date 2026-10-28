import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import {
  getDatasourceList,
  getDatasourceById,
  testDatasource,
  syncDatasource,
  getDatasourceTables,
  getTableColumns,
} from '@/api/odin/datasource'
import { getConnectorList } from '@/api/odin/connector'
import type { DatasourceVO, SchemaTableVO, SchemaColumnVO, ConnectorVO, DatasourceTestResult } from '@/types/odin'

export const useOdinDatasourceStore = defineStore('odinDatasource', () => {
  const sources = ref<DatasourceVO[]>([])
  const currentSource = ref<DatasourceVO | null>(null)
  /** 按数据源ID缓存的表列表 */
  const schemas = ref<Map<string, SchemaTableVO[]>>(new Map())
  /** 按表ID缓存的列列表 */
  const columns = ref<Map<string, SchemaColumnVO[]>>(new Map())
  const connectors = ref<ConnectorVO[]>([])
  const loading = ref(false)

  const sourceCount = computed(() => sources.value.length)

  /** 加载全部数据源 */
  async function loadSources(): Promise<void> {
    loading.value = true
    try {
      sources.value = await getDatasourceList()
    } catch (e) {
      console.error('[odinDatasource] loadSources failed', e)
    } finally {
      loading.value = false
    }
  }

  /** 选择并加载数据源详情 */
  async function selectSource(id: string): Promise<void> {
    if (!id) {
      currentSource.value = null
      return
    }
    loading.value = true
    try {
      currentSource.value = await getDatasourceById(id)
    } catch (e) {
      console.error('[odinDatasource] selectSource failed', e)
    } finally {
      loading.value = false
    }
  }

  /** 加载数据源 Schema（表列表） */
  async function loadSchema(dsId: string): Promise<SchemaTableVO[]> {
    try {
      const tables = await getDatasourceTables(dsId)
      schemas.value.set(dsId, tables)
      return tables
    } catch (e) {
      console.error('[odinDatasource] loadSchema failed', e)
      return []
    }
  }

  /** 加载表列 */
  async function loadColumns(tableId: string): Promise<SchemaColumnVO[]> {
    try {
      const cols = await getTableColumns(tableId)
      columns.value.set(tableId, cols)
      return cols
    } catch (e) {
      console.error('[odinDatasource] loadColumns failed', e)
      return []
    }
  }

  /** 测试连接 */
  async function testConnection(id: string): Promise<DatasourceTestResult | null> {
    try {
      return await testDatasource(id)
    } catch (e) {
      console.error('[odinDatasource] testConnection failed', e)
      return null
    }
  }

  /** 同步元数据 */
  async function syncSchema(id: string): Promise<number | null> {
    loading.value = true
    try {
      const count = await syncDatasource(id)
      // 同步后刷新 schema 缓存
      await loadSchema(id)
      return count
    } catch (e) {
      console.error('[odinDatasource] syncSchema failed', e)
      return null
    } finally {
      loading.value = false
    }
  }

  /** 加载连接器列表 */
  async function loadConnectors(): Promise<void> {
    try {
      connectors.value = await getConnectorList()
    } catch (e) {
      console.error('[odinDatasource] loadConnectors failed', e)
    }
  }

  function reset() {
    sources.value = []
    currentSource.value = null
    schemas.value = new Map()
    columns.value = new Map()
    connectors.value = []
  }

  return {
    sources,
    currentSource,
    schemas,
    columns,
    connectors,
    loading,
    sourceCount,
    loadSources,
    selectSource,
    loadSchema,
    loadColumns,
    testConnection,
    syncSchema,
    loadConnectors,
    reset,
  }
})
