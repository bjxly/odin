import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import * as traceApi from '@/api/trace'
import type { TraceVO } from '@/types/query'

/**
 * 链路溯源 store。
 *
 * 驱动全局 TraceDrawer：任何携带 trace_id 的位置（ProvenanceCard、QueryLab、
 * 助手消息）都只需调用 `open(traceId)` 即可唤起抽屉，抽屉自身负责渲染。
 */
export const useTraceStore = defineStore('trace', () => {
  const visible = ref(false)
  const traceId = ref('')
  const trace = ref<TraceVO | null>(null)
  const loading = ref(false)
  const error = ref('')
  /** 已加载过的链路缓存，避免重复请求 */
  const cache = new Map<string, TraceVO>()

  const hasTrace = computed(() => !!trace.value)
  const nodes = computed(() => trace.value?.nodes ?? [])
  const reasonTrace = computed(() => trace.value?.reason_trace ?? null)

  /** 拉取一次链路详情（带缓存） */
  async function loadTrace(id: string, force = false): Promise<TraceVO | null> {
    const key = (id ?? '').trim()
    if (!key) {
      error.value = '缺少 trace_id'
      return null
    }
    if (!force && cache.has(key)) {
      traceId.value = key
      trace.value = cache.get(key) ?? null
      error.value = ''
      return trace.value
    }
    loading.value = true
    error.value = ''
    traceId.value = key
    try {
      const res = await traceApi.getTrace(key)
      trace.value = res ?? null
      if (res) cache.set(key, res)
      return trace.value
    } catch (e: any) {
      trace.value = null
      error.value = e?.message || '加载链路失败'
      return null
    } finally {
      loading.value = false
    }
  }

  /** 打开抽屉并加载指定链路 */
  async function open(id: string): Promise<void> {
    const key = (id ?? '').trim()
    if (!key) return
    visible.value = true
    if (traceId.value !== key || !trace.value) {
      trace.value = cache.get(key) ?? null
    }
    await loadTrace(key)
  }

  function close(): void {
    visible.value = false
  }

  function reset(): void {
    visible.value = false
    traceId.value = ''
    trace.value = null
    error.value = ''
    loading.value = false
  }

  /** 清除缓存（例如后端重跑同一问题后希望强制刷新） */
  function clearCache(): void {
    cache.clear()
  }

  return {
    visible, traceId, trace, loading, error,
    hasTrace, nodes, reasonTrace,
    loadTrace, open, close, reset, clearCache,
  }
})
