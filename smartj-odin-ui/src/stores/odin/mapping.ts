import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import {
  getMappingPage,
  createMapping,
  updateMapping,
  deleteMapping,
  suggestColumnMappings,
  validateMapping,
} from '@/api/odin/mapping'
import type { MappingConfigVO, MappingConfigDTO, MappingConfigQuery, MappingSuggestVO } from '@/types/odin'

export const useOdinMappingStore = defineStore('odinMapping', () => {
  const mappings = ref<MappingConfigVO[]>([])
  const dirty = ref(false)
  const suggestions = ref<MappingSuggestVO[]>([])
  const coverageRate = ref(0)
  const loading = ref(false)
  const total = ref(0)

  const mappingCount = computed(() => mappings.value.length)

  /** 加载映射配置列表 */
  async function loadMappings(ontologyId?: string): Promise<void> {
    loading.value = true
    try {
      const params: Record<string, unknown> = { current: 1, size: 200 }
      if (ontologyId) params.ontologyId = ontologyId
      const res = await getMappingPage(params as any)
      mappings.value = res?.records ?? []
      total.value = res?.total ?? 0
      // 计算覆盖率：有映射的类占全部类的比例
      if (mappings.value.length > 0) {
        coverageRate.value = Math.round(
          (mappings.value.filter((m) => m.columnMappings).length / mappings.value.length) * 100,
        )
      }
    } catch (e) {
      console.error('[odinMapping] loadMappings failed', e)
    } finally {
      loading.value = false
    }
  }

  /** 加载 AI 建议映射 */
  async function loadSuggestions(classId: string, tableId: string): Promise<MappingSuggestVO[]> {
    loading.value = true
    try {
      suggestions.value = await suggestColumnMappings(classId, tableId)
      return suggestions.value
    } catch (e) {
      console.error('[odinMapping] loadSuggestions failed', e)
      return []
    } finally {
      loading.value = false
    }
  }

  /** 按数据源加载其全部映射（表中心视图用，不影响本体作用域的 mappings） */
  async function loadMappingsByDatasource(datasourceId: string): Promise<MappingConfigVO[]> {
    if (!datasourceId) return []
    try {
      const res = await getMappingPage({ current: 1, size: 200, datasourceId } as MappingConfigQuery)
      return res?.records ?? []
    } catch (e) {
      console.error('[odinMapping] loadMappingsByDatasource failed', e)
      return []
    }
  }

  /** 应用建议（更新映射配置的 columnMappings） */
  async function applySuggestions(mappingId: string, columnMappings: string): Promise<void> {
    const existing = mappings.value.find((m) => m.id === mappingId)
    if (!existing) return
    const dto: MappingConfigDTO = {
      id: existing.id,
      ontologyId: existing.ontologyId,
      classId: existing.classId,
      datasourceId: existing.datasourceId,
      tableName: existing.tableName,
      schemaName: existing.schemaName,
      columnMappings,
      filterCondition: existing.filterCondition,
      priority: existing.priority,
      status: existing.status,
      remark: existing.remark,
    }
    await updateMapping(dto)
    dirty.value = true
    await loadMappings(existing.ontologyId)
  }

  /** 校验映射配置，返回后端校验结果描述 */
  async function validateMappings(id: string): Promise<string> {
    loading.value = true
    try {
      return await validateMapping(id)
    } catch (e) {
      console.error('[odinMapping] validateMappings failed', e)
      return ''
    } finally {
      loading.value = false
    }
  }

  /** 新增映射 */
  async function addMapping(data: MappingConfigDTO): Promise<string> {
    const id = await createMapping(data)
    dirty.value = true
    return id
  }

  /** 删除映射 */
  async function removeMapping(id: string): Promise<void> {
    await deleteMapping(id)
    mappings.value = mappings.value.filter((m) => m.id !== id)
    dirty.value = true
  }

  function reset() {
    mappings.value = []
    dirty.value = false
    suggestions.value = []
    coverageRate.value = 0
    total.value = 0
  }

  return {
    mappings,
    dirty,
    suggestions,
    coverageRate,
    loading,
    total,
    mappingCount,
    loadMappings,
    loadMappingsByDatasource,
    loadSuggestions,
    applySuggestions,
    validateMappings,
    addMapping,
    removeMapping,
    reset,
  }
})
