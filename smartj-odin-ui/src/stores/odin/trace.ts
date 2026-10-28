import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { get } from '@/api/request'
import type { TraceData } from '@/types/odin'

export const useOdinTraceStore = defineStore('odinTrace', () => {
  const currentTraceId = ref('')
  const visible = ref(false)
  const traceData = ref<TraceData | null>(null)
  const loading = ref(false)

  /** 已加载过的链路缓存 */
  const cache = new Map<string, TraceData>()

  const hasTrace = computed(() => !!traceData.value)

  /** 打开溯源抽屉并加载数据 */
  async function openTrace(traceId: string): Promise<void> {
    const key = (traceId ?? '').trim()
    if (!key) return
    visible.value = true
    currentTraceId.value = key
    // 使用缓存
    if (cache.has(key)) {
      traceData.value = cache.get(key) ?? null
      return
    }
    await loadTraceData(key)
  }

  /** 关闭溯源抽屉 */
  function closeTrace(): void {
    visible.value = false
  }

  /** 加载溯源数据 */
  async function loadTraceData(id: string): Promise<TraceData | null> {
    loading.value = true
    try {
      const data = await get<TraceData>(`/odin/query/trace/${encodeURIComponent(id)}`)
      traceData.value = data ?? null
      if (data) cache.set(id, data)
      return traceData.value
    } catch (e) {
      console.error('[odinTrace] loadTraceData failed', e)
      traceData.value = null
      return null
    } finally {
      loading.value = false
    }
  }

  function reset(): void {
    currentTraceId.value = ''
    visible.value = false
    traceData.value = null
    loading.value = false
  }

  /** 清除缓存 */
  function clearCache(): void {
    cache.clear()
  }

  return {
    currentTraceId,
    visible,
    traceData,
    loading,
    hasTrace,
    openTrace,
    closeTrace,
    loadTraceData,
    reset,
    clearCache,
  }
})
