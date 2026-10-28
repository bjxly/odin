import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import {
  executeLogicalSql,
  dryRunQuery,
  getQueryTemplatePage,
  createQueryTemplate,
  getQueryHistoryPage,
} from '@/api/odin/query'
import type {
  QueryResultVO,
  QueryDryRunVO,
  QueryTemplateVO,
  QueryTemplateDTO,
  QueryHistoryVO,
} from '@/types/odin'

export const useOdinQueryStore = defineStore('odinQuery', () => {
  /** 当前查询规格（逻辑SQL + 参数） */
  const currentSpec = ref<{ logicalSql: string; ontologyId?: string; params?: Record<string, unknown> }>({
    logicalSql: '',
  })
  const dryRunResult = ref<QueryDryRunVO | null>(null)
  const executeResult = ref<QueryResultVO | null>(null)
  const templates = ref<QueryTemplateVO[]>([])
  const history = ref<QueryHistoryVO[]>([])
  const nlParse = ref<{ text: string; logicalSql?: string; parsePath?: string } | null>(null)
  const loading = ref(false)

  const hasResult = computed(() => !!executeResult.value?.success)

  /** 自然语言解析（通过助手 SSE 完成，此处仅提供状态存储） */
  function parseNl(text: string) {
    nlParse.value = { text }
  }

  /** Dry-run 校验逻辑 SQL */
  async function dryRun(logicalSql: string, ontologyId?: string): Promise<QueryDryRunVO | null> {
    loading.value = true
    try {
      const result = await dryRunQuery({ logicalSql, ontologyId })
      dryRunResult.value = result
      return result
    } catch (e) {
      console.error('[odinQuery] dryRun failed', e)
      return null
    } finally {
      loading.value = false
    }
  }

  /** 执行查询 */
  async function execute(spec: { logicalSql: string; ontologyId?: string; params?: Record<string, unknown> }): Promise<QueryResultVO | null> {
    loading.value = true
    try {
      currentSpec.value = spec
      const result = await executeLogicalSql(spec)
      executeResult.value = result
      return result
    } catch (e) {
      console.error('[odinQuery] execute failed', e)
      return null
    } finally {
      loading.value = false
    }
  }

  /** 保存为模板 */
  async function saveTemplate(data: QueryTemplateDTO): Promise<string | null> {
    try {
      const id = await createQueryTemplate(data)
      await loadTemplates()
      return id
    } catch (e) {
      console.error('[odinQuery] saveTemplate failed', e)
      return null
    }
  }

  /** 加载模板列表 */
  async function loadTemplates(): Promise<void> {
    try {
      const res = await getQueryTemplatePage({ current: 1, size: 100 })
      templates.value = res?.records ?? []
    } catch (e) {
      console.error('[odinQuery] loadTemplates failed', e)
    }
  }

  /** 加载查询历史 */
  async function loadHistory(): Promise<void> {
    try {
      const res = await getQueryHistoryPage({ current: 1, size: 50 })
      history.value = res?.records ?? []
    } catch (e) {
      console.error('[odinQuery] loadHistory failed', e)
    }
  }

  function reset() {
    currentSpec.value = { logicalSql: '' }
    dryRunResult.value = null
    executeResult.value = null
    nlParse.value = null
  }

  return {
    currentSpec,
    dryRunResult,
    executeResult,
    templates,
    history,
    nlParse,
    loading,
    hasResult,
    parseNl,
    dryRun,
    execute,
    saveTemplate,
    loadTemplates,
    loadHistory,
    reset,
  }
})
